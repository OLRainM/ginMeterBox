package storage

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"go-ele/models"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	dataFile          = "data/billing_records.json"
)

// Storage 数据存储
type Storage struct {
	mu      sync.RWMutex
	records []models.BillingRecord
	nextID  int
}

// NewStorage 创建新的存储实例
func NewStorage() *Storage {
	s := &Storage{
		records: make([]models.BillingRecord, 0),
		nextID:  1,
	}
	s.loadFromFile()
	return s
}

// loadFromFile 从文件加载数据
func (s *Storage) loadFromFile() error {
	// 确保数据目录存在
	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}

	data, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在是正常的
		}
		return err
	}

	var records []models.BillingRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	s.records = records
	
	// 更新nextID
	maxID := 0
	for _, r := range records {
		if r.ID > maxID {
			maxID = r.ID
		}
	}
	s.nextID = maxID + 1

	return nil
}

// saveToFile 保存数据到文件
func (s *Storage) saveToFile() error {
	data, err := json.MarshalIndent(s.records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(dataFile, data, 0644)
}

// GetAll 获取所有记录
func (s *Storage) GetAll() []models.BillingRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.BillingRecord, len(s.records))
	copy(result, s.records)
	return result
}

// GetByID 根据ID获取记录
func (s *Storage) GetByID(id int) (*models.BillingRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.records {
		if s.records[i].ID == id {
			record := s.records[i]
			return &record, nil
		}
	}
	return nil, ErrRecordNotFound
}

// Create 创建新记录
func (s *Storage) Create(record *models.BillingRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record.ID = s.nextID
	s.nextID++
	record.CreatedAt = time.Now()
	record.UpdatedAt = time.Now()
	
	// 计算费用
	record.CalculateCosts()

	s.records = append(s.records, *record)
	return s.saveToFile()
}

// Update 更新记录
func (s *Storage) Update(id int, record *models.BillingRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.records {
		if s.records[i].ID == id {
			record.ID = id
			record.CreatedAt = s.records[i].CreatedAt
			record.UpdatedAt = time.Now()
			
			// 计算费用
			record.CalculateCosts()
			
			s.records[i] = *record
			return s.saveToFile()
		}
	}
	return ErrRecordNotFound
}

// Delete 删除记录
func (s *Storage) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.records {
		if s.records[i].ID == id {
			s.records = append(s.records[:i], s.records[i+1:]...)
			return s.saveToFile()
		}
	}
	return ErrRecordNotFound
}

// GetByMonth 根据月份获取记录
func (s *Storage) GetByMonth(month string) []models.BillingRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.BillingRecord, 0)
	for _, record := range s.records {
		if record.BillingMonth == month {
			result = append(result, record)
		}
	}
	return result
}

// GetByIDs 根据多个ID获取记录
func (s *Storage) GetByIDs(ids []int) []models.BillingRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.BillingRecord, 0)
	idMap := make(map[int]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	for _, record := range s.records {
		if idMap[record.ID] {
			result = append(result, record)
		}
	}
	return result
}

// GetLatestByRoomNumber 获取某住户的最新记录
func (s *Storage) GetLatestByRoomNumber(roomNumber string) (*models.BillingRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var latest *models.BillingRecord
	for i := range s.records {
		if s.records[i].RoomNumber == roomNumber {
			if latest == nil || s.records[i].CreatedAt.After(latest.CreatedAt) {
				record := s.records[i]
				latest = &record
			}
		}
	}

	if latest == nil {
		return nil, ErrRecordNotFound
	}
	return latest, nil
}

// CreateFromPrevious 从上月记录创建新记录（自动延续）
func (s *Storage) CreateFromPrevious(roomNumber, newMonth string) (*models.BillingRecord, error) {
	// 获取该住户的最新记录
	previous, err := s.GetLatestByRoomNumber(roomNumber)
	if err != nil {
		return nil, err
	}

	// 创建新记录，本月初始值 = 上月当前值
	newRecord := &models.BillingRecord{
		RoomNumber:         roomNumber,
		BillingMonth:       newMonth,
		CurrentWater:       previous.CurrentWater,      // 初始化为上月的当前值
		PreviousWater:      previous.CurrentWater,      // 上月读数
		WaterAdjustment:    0,
		CurrentElectric:    previous.CurrentElectric,   // 初始化为上月的当前值
		PreviousElectric:   previous.CurrentElectric,   // 上月读数
		ElectricAdjustment: 0,
		ManagementFee:      previous.ManagementFee,     // 继承管理费
		WaterPrice:         previous.WaterPrice,        // 继承水价
		ElectricPrice:      previous.ElectricPrice,     // 继承电价
	}

	// 创建记录
	if err := s.Create(newRecord); err != nil {
		return nil, err
	}

	return newRecord, nil
}

// BatchImport 批量导入记录
func (s *Storage) BatchImport(records []models.BillingRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range records {
		records[i].ID = s.nextID
		s.nextID++
		records[i].CreatedAt = time.Now()
		records[i].UpdatedAt = time.Now()
		
		// 计算费用
		records[i].CalculateCosts()
		
		s.records = append(s.records, records[i])
	}
	
	return s.saveToFile()
}

// ExportToJSON 导出所有记录到JSON
func (s *Storage) ExportToJSON(filepath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}
