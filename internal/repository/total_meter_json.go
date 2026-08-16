package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"ginMeterBox/internal/models"
)

var ErrTotalMeterRecordNotFound = errors.New("total meter record not found")
var ErrTotalMeterMonthExists = errors.New("total meter month already exists")

// TotalMeterJSONRepo JSON文件实现的总表仓储。
type TotalMeterJSONRepo struct {
	records  []models.TotalMeterRecord
	nextID   int
	mu       sync.RWMutex
	filename string
}

func NewTotalMeterJSONRepo(filename string) (*TotalMeterJSONRepo, error) {
	r := &TotalMeterJSONRepo{
		records:  []models.TotalMeterRecord{},
		nextID:   1,
		filename: filename,
	}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *TotalMeterJSONRepo) load() error {
	if err := os.MkdirAll(filepath.Dir(r.filename), 0755); err != nil {
		return err
	}
	data, err := os.ReadFile(r.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := json.Unmarshal(data, &r.records); err != nil {
		return err
	}
	for _, record := range r.records {
		if record.ID >= r.nextID {
			r.nextID = record.ID + 1
		}
	}
	return nil
}

func (r *TotalMeterJSONRepo) save(records []models.TotalMeterRecord) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return writeJSONAtomically(r.filename, data)
}

func (r *TotalMeterJSONRepo) GetAll() []models.TotalMeterRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]models.TotalMeterRecord, len(r.records))
	copy(result, r.records)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Month > result[j].Month
	})
	return result
}

func (r *TotalMeterJSONRepo) GetByMonth(month string) (*models.TotalMeterRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, record := range r.records {
		if record.Month == month {
			return &record, nil
		}
	}
	return nil, ErrTotalMeterRecordNotFound
}

func (r *TotalMeterJSONRepo) Create(record *models.TotalMeterRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.records {
		if existing.Month == record.Month {
			return ErrTotalMeterMonthExists
		}
	}

	created := *record
	created.ID = r.nextID
	created.CreatedAt = time.Now()
	created.UpdatedAt = created.CreatedAt
	candidate := append(append([]models.TotalMeterRecord(nil), r.records...), created)
	if err := r.save(candidate); err != nil {
		return err
	}
	r.records = candidate
	r.nextID++
	*record = created
	return nil
}

func (r *TotalMeterJSONRepo) Update(month string, record *models.TotalMeterRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	candidate := append([]models.TotalMeterRecord(nil), r.records...)
	for i, existing := range candidate {
		if existing.Month == month {
			updated := *record
			updated.ID = existing.ID
			updated.Month = month
			updated.CreatedAt = existing.CreatedAt
			updated.UpdatedAt = time.Now()
			candidate[i] = updated
			if err := r.save(candidate); err != nil {
				return err
			}
			r.records = candidate
			*record = updated
			return nil
		}
	}
	return ErrTotalMeterRecordNotFound
}

func (r *TotalMeterJSONRepo) Delete(month string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, record := range r.records {
		if record.Month == month {
			candidate := append([]models.TotalMeterRecord(nil), r.records[:i]...)
			candidate = append(candidate, r.records[i+1:]...)
			if err := r.save(candidate); err != nil {
				return err
			}
			r.records = candidate
			return nil
		}
	}
	return ErrTotalMeterRecordNotFound
}
