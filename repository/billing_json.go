package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ginMeterBox/models"
)

var ErrRecordNotFound = errors.New("record not found")

// BillingJSONRepo JSON文件实现的账单仓储。
type BillingJSONRepo struct {
	mu       sync.RWMutex
	records  []models.BillingRecord
	nextID   int
	dataFile string
}

func NewBillingJSONRepo(dataFile string) (*BillingJSONRepo, error) {
	r := &BillingJSONRepo{
		records:  make([]models.BillingRecord, 0),
		nextID:   1,
		dataFile: dataFile,
	}
	if err := r.loadFromFile(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *BillingJSONRepo) loadFromFile() error {
	if err := os.MkdirAll(filepath.Dir(r.dataFile), 0755); err != nil {
		return err
	}
	data, err := os.ReadFile(r.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var records []models.BillingRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}
	maxID := 0
	for _, rec := range records {
		if rec.ID > maxID {
			maxID = rec.ID
		}
	}
	r.records = records
	r.nextID = maxID + 1
	return nil
}

func (r *BillingJSONRepo) save(records []models.BillingRecord) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return writeJSONAtomically(r.dataFile, data)
}

func cloneRecords(records []models.BillingRecord) []models.BillingRecord {
	cloned := make([]models.BillingRecord, len(records))
	copy(cloned, records)
	for i := range cloned {
		cloned[i].ExtraFees = append([]models.ExtraFee(nil), records[i].ExtraFees...)
	}
	return cloned
}

func (r *BillingJSONRepo) GetAll() []models.BillingRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return cloneRecords(r.records)
}

func (r *BillingJSONRepo) GetByID(id int) (*models.BillingRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for i := range r.records {
		if r.records[i].ID == id {
			record := cloneRecords(r.records[i : i+1])[0]
			return &record, nil
		}
	}
	return nil, ErrRecordNotFound
}

func (r *BillingJSONRepo) GetByMonth(month string) []models.BillingRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]models.BillingRecord, 0)
	for _, record := range r.records {
		if record.BillingMonth == month {
			result = append(result, record)
		}
	}
	return cloneRecords(result)
}

func (r *BillingJSONRepo) GetByIDs(ids []int) []models.BillingRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	byID := make(map[int]models.BillingRecord, len(r.records))
	for _, record := range r.records {
		byID[record.ID] = record
	}
	result := make([]models.BillingRecord, 0, len(ids))
	for _, id := range ids {
		if record, ok := byID[id]; ok {
			result = append(result, record)
		}
	}
	return cloneRecords(result)
}

func (r *BillingJSONRepo) GetLatestByRoomNumber(roomNumber string) (*models.BillingRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var latest *models.BillingRecord
	for i := range r.records {
		if r.records[i].RoomNumber == roomNumber && (latest == nil || r.records[i].CreatedAt.After(latest.CreatedAt)) {
			record := cloneRecords(r.records[i : i+1])[0]
			latest = &record
		}
	}
	if latest == nil {
		return nil, ErrRecordNotFound
	}
	return latest, nil
}

func (r *BillingJSONRepo) Create(record *models.BillingRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	candidate := cloneRecords(r.records)
	created := *record
	created.ID = r.nextID
	created.CreatedAt = now
	created.UpdatedAt = now
	created.CalculateCosts()
	candidate = append(candidate, created)
	if err := r.save(candidate); err != nil {
		return err
	}
	r.records = candidate
	r.nextID++
	*record = created
	return nil
}

func (r *BillingJSONRepo) Update(id int, record *models.BillingRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	candidate := cloneRecords(r.records)
	for i := range candidate {
		if candidate[i].ID == id {
			updated := *record
			updated.ID = id
			updated.CreatedAt = candidate[i].CreatedAt
			updated.UpdatedAt = time.Now()
			updated.CalculateCosts()
			candidate[i] = updated
			if err := r.save(candidate); err != nil {
				return err
			}
			r.records = candidate
			*record = updated
			return nil
		}
	}
	return ErrRecordNotFound
}

func (r *BillingJSONRepo) Delete(id int) error {
	count, err := r.BatchDelete([]int{id})
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (r *BillingJSONRepo) BatchDelete(ids []int) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	idSet := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	if len(idSet) == 0 {
		return 0, nil
	}
	for _, record := range r.records {
		delete(idSet, record.ID)
	}
	if len(idSet) != 0 {
		return 0, ErrRecordNotFound
	}

	candidate := make([]models.BillingRecord, 0, len(r.records)-len(ids))
	for _, record := range r.records {
		found := false
		for _, id := range ids {
			if record.ID == id {
				found = true
				break
			}
		}
		if !found {
			candidate = append(candidate, record)
		}
	}
	if err := r.save(candidate); err != nil {
		return 0, err
	}
	r.records = candidate
	return len(ids), nil
}

func (r *BillingJSONRepo) BatchUpdateAdjustments(ids []int, waterAdjustment, electricAdjustment *float64) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	idSet := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	if len(idSet) == 0 {
		return 0, nil
	}
	candidate := cloneRecords(r.records)
	for i := range candidate {
		if _, ok := idSet[candidate[i].ID]; ok {
			delete(idSet, candidate[i].ID)
			if waterAdjustment != nil {
				candidate[i].WaterAdjustment = *waterAdjustment
			}
			if electricAdjustment != nil {
				candidate[i].ElectricAdjustment = *electricAdjustment
			}
			candidate[i].UpdatedAt = time.Now()
			candidate[i].CalculateCosts()
		}
	}
	if len(idSet) != 0 {
		return 0, ErrRecordNotFound
	}
	if err := r.save(candidate); err != nil {
		return 0, err
	}
	r.records = candidate
	return len(ids), nil
}

func (r *BillingJSONRepo) BatchSetExtraFees(ids []int, extraFees []models.ExtraFee, mode string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	idSet := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	if len(idSet) == 0 {
		return 0, nil
	}
	candidate := cloneRecords(r.records)
	for i := range candidate {
		if _, ok := idSet[candidate[i].ID]; ok {
			delete(idSet, candidate[i].ID)
			if mode == "replace" {
				candidate[i].ExtraFees = append([]models.ExtraFee(nil), extraFees...)
			} else {
				candidate[i].ExtraFees = append(candidate[i].ExtraFees, extraFees...)
			}
			candidate[i].UpdatedAt = time.Now()
			candidate[i].CalculateCosts()
		}
	}
	if len(idSet) != 0 {
		return 0, ErrRecordNotFound
	}
	if err := r.save(candidate); err != nil {
		return 0, err
	}
	r.records = candidate
	return len(ids), nil
}

func (r *BillingJSONRepo) BatchImport(records []models.BillingRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	candidate := cloneRecords(r.records)
	nextID := r.nextID
	now := time.Now()
	for i := range records {
		records[i].ID = nextID
		nextID++
		records[i].CreatedAt = now
		records[i].UpdatedAt = now
		records[i].CalculateCosts()
		candidate = append(candidate, records[i])
	}
	if err := r.save(candidate); err != nil {
		return err
	}
	r.records = candidate
	r.nextID = nextID
	return nil
}

func (r *BillingJSONRepo) ExportToJSON(filename string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	data, err := json.MarshalIndent(r.records, "", "  ")
	if err != nil {
		return err
	}
	return writeJSONAtomically(filename, data)
}
