package repository

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"sync"
	"time"

	"ginMeterBox/models"
)

// TotalMeterJSONRepo JSON文件实现的总表仓储
type TotalMeterJSONRepo struct {
	records  []models.TotalMeterRecord
	nextID   int
	mu       sync.RWMutex
	filename string
}

func NewTotalMeterJSONRepo(filename string) *TotalMeterJSONRepo {
	r := &TotalMeterJSONRepo{
		records:  []models.TotalMeterRecord{},
		nextID:   1,
		filename: filename,
	}
	r.load()
	return r
}

func (r *TotalMeterJSONRepo) load() error {
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
	for _, rec := range r.records {
		if rec.ID >= r.nextID {
			r.nextID = rec.ID + 1
		}
	}
	return nil
}

func (r *TotalMeterJSONRepo) save() error {
	data, err := json.MarshalIndent(r.records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.filename, data, 0644)
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
	for _, rec := range r.records {
		if rec.Month == month {
			return &rec, nil
		}
	}
	return nil, errors.New("record not found")
}

func (r *TotalMeterJSONRepo) Create(record *models.TotalMeterRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, rec := range r.records {
		if rec.Month == record.Month {
			return errors.New("该月份的总表记录已存在")
		}
	}
	record.ID = r.nextID
	r.nextID++
	record.CreatedAt = time.Now()
	record.UpdatedAt = time.Now()
	r.records = append(r.records, *record)
	return r.save()
}

func (r *TotalMeterJSONRepo) Update(month string, record *models.TotalMeterRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, rec := range r.records {
		if rec.Month == month {
			record.ID = rec.ID
			record.CreatedAt = rec.CreatedAt
			record.UpdatedAt = time.Now()
			r.records[i] = *record
			return r.save()
		}
	}
	return errors.New("record not found")
}

func (r *TotalMeterJSONRepo) Delete(month string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, rec := range r.records {
		if rec.Month == month {
			r.records = append(r.records[:i], r.records[i+1:]...)
			return r.save()
		}
	}
	return errors.New("record not found")
}
