package repository

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"ginMeterBox/models"
)

var ErrRecordNotFound = errors.New("record not found")

// BillingJSONRepo JSON文件实现的账单仓储
type BillingJSONRepo struct {
	mu       sync.RWMutex
	records  []models.BillingRecord
	nextID   int
	dataFile string
}

func NewBillingJSONRepo(dataFile string) *BillingJSONRepo {
	r := &BillingJSONRepo{
		records:  make([]models.BillingRecord, 0),
		nextID:   1,
		dataFile: dataFile,
	}
	r.loadFromFile()
	return r
}

func (r *BillingJSONRepo) loadFromFile() error {
	if err := os.MkdirAll("data", 0755); err != nil {
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
	r.records = records
	maxID := 0
	for _, rec := range records {
		if rec.ID > maxID {
			maxID = rec.ID
		}
	}
	r.nextID = maxID + 1
	return nil
}

func (r *BillingJSONRepo) saveToFile() error {
	data, err := json.MarshalIndent(r.records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.dataFile, data, 0644)
}

func (r *BillingJSONRepo) GetAll() []models.BillingRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]models.BillingRecord, len(r.records))
	copy(result, r.records)
	return result
}

func (r *BillingJSONRepo) GetByID(id int) (*models.BillingRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for i := range r.records {
		if r.records[i].ID == id {
			rec := r.records[i]
			return &rec, nil
		}
	}
	return nil, ErrRecordNotFound
}

func (r *BillingJSONRepo) GetByMonth(month string) []models.BillingRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]models.BillingRecord, 0)
	for _, rec := range r.records {
		if rec.BillingMonth == month {
			result = append(result, rec)
		}
	}
	return result
}

func (r *BillingJSONRepo) GetByIDs(ids []int) []models.BillingRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	idMap := make(map[int]bool, len(ids))
	for _, id := range ids {
		idMap[id] = true
	}
	result := make([]models.BillingRecord, 0)
	for _, rec := range r.records {
		if idMap[rec.ID] {
			result = append(result, rec)
		}
	}
	return result
}

func (r *BillingJSONRepo) GetLatestByRoomNumber(roomNumber string) (*models.BillingRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var latest *models.BillingRecord
	for i := range r.records {
		if r.records[i].RoomNumber == roomNumber {
			if latest == nil || r.records[i].CreatedAt.After(latest.CreatedAt) {
				rec := r.records[i]
				latest = &rec
			}
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
	record.ID = r.nextID
	r.nextID++
	record.CreatedAt = time.Now()
	record.UpdatedAt = time.Now()
	record.CalculateCosts()
	r.records = append(r.records, *record)
	return r.saveToFile()
}

func (r *BillingJSONRepo) Update(id int, record *models.BillingRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.records {
		if r.records[i].ID == id {
			record.ID = id
			record.CreatedAt = r.records[i].CreatedAt
			record.UpdatedAt = time.Now()
			record.CalculateCosts()
			r.records[i] = *record
			return r.saveToFile()
		}
	}
	return ErrRecordNotFound
}

func (r *BillingJSONRepo) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.records {
		if r.records[i].ID == id {
			r.records = append(r.records[:i], r.records[i+1:]...)
			return r.saveToFile()
		}
	}
	return ErrRecordNotFound
}

func (r *BillingJSONRepo) BatchImport(records []models.BillingRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range records {
		records[i].ID = r.nextID
		r.nextID++
		records[i].CreatedAt = time.Now()
		records[i].UpdatedAt = time.Now()
		records[i].CalculateCosts()
		r.records = append(r.records, records[i])
	}
	return r.saveToFile()
}

func (r *BillingJSONRepo) ExportToJSON(filepath string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	data, err := json.MarshalIndent(r.records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}
