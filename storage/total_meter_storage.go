package storage

import (
	"encoding/json"
	"errors"
	"go-ele/models"
	"os"
	"sort"
	"sync"
	"time"
)

type TotalMeterStorage struct {
	records  []models.TotalMeterRecord
	nextID   int
	mu       sync.RWMutex
	filename string
}

func NewTotalMeterStorage() *TotalMeterStorage {
	storage := &TotalMeterStorage{
		records:  []models.TotalMeterRecord{},
		nextID:   1,
		filename: "data/total_meter_records.json",
	}
	storage.load()
	return storage
}

// load 从文件加载数据
func (s *TotalMeterStorage) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, &s.records); err != nil {
		return err
	}

	// 更新nextID
	for _, record := range s.records {
		if record.ID >= s.nextID {
			s.nextID = record.ID + 1
		}
	}

	return nil
}

// save 保存数据到文件
func (s *TotalMeterStorage) save() error {
	data, err := json.MarshalIndent(s.records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filename, data, 0644)
}

// GetAll 获取所有记录
func (s *TotalMeterStorage) GetAll() []models.TotalMeterRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.TotalMeterRecord, len(s.records))
	copy(result, s.records)

	// 按月份降序排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Month > result[j].Month
	})

	return result
}

// GetByMonth 根据月份获取记录
func (s *TotalMeterStorage) GetByMonth(month string) (*models.TotalMeterRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, record := range s.records {
		if record.Month == month {
			return &record, nil
		}
	}

	return nil, errors.New("record not found")
}

// Create 创建新记录
func (s *TotalMeterStorage) Create(record *models.TotalMeterRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查月份是否已存在
	for _, r := range s.records {
		if r.Month == record.Month {
			return errors.New("该月份的总表记录已存在")
		}
	}

	record.ID = s.nextID
	s.nextID++
	record.CreatedAt = time.Now()
	record.UpdatedAt = time.Now()

	s.records = append(s.records, *record)

	return s.save()
}

// Update 更新记录
func (s *TotalMeterStorage) Update(month string, record *models.TotalMeterRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, r := range s.records {
		if r.Month == month {
			record.ID = r.ID
			record.CreatedAt = r.CreatedAt
			record.UpdatedAt = time.Now()
			s.records[i] = *record
			return s.save()
		}
	}

	return errors.New("record not found")
}

// Delete 删除记录
func (s *TotalMeterStorage) Delete(month string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, record := range s.records {
		if record.Month == month {
			s.records = append(s.records[:i], s.records[i+1:]...)
			return s.save()
		}
	}

	return errors.New("record not found")
}
