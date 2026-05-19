package services

import (
	"ginMeterBox/models"
	"ginMeterBox/repository"
)

// BillingService 账单业务逻辑
type BillingService struct {
	repo repository.BillingRepo
}

func NewBillingService(repo repository.BillingRepo) *BillingService {
	return &BillingService{repo: repo}
}

func (s *BillingService) GetAll() []models.BillingRecord {
	return s.repo.GetAll()
}

func (s *BillingService) GetByID(id int) (*models.BillingRecord, error) {
	return s.repo.GetByID(id)
}

func (s *BillingService) GetByMonth(month string) []models.BillingRecord {
	return s.repo.GetByMonth(month)
}

func (s *BillingService) GetByIDs(ids []int) []models.BillingRecord {
	return s.repo.GetByIDs(ids)
}

func (s *BillingService) GetLatestByRoom(roomNumber string) (*models.BillingRecord, error) {
	return s.repo.GetLatestByRoomNumber(roomNumber)
}

func (s *BillingService) Create(record *models.BillingRecord) error {
	return s.repo.Create(record)
}

func (s *BillingService) Update(id int, record *models.BillingRecord) error {
	return s.repo.Update(id, record)
}

func (s *BillingService) Delete(id int) error {
	return s.repo.Delete(id)
}

// ContinueFromPrevious 从上月记录创建新记录（自动延续）
func (s *BillingService) ContinueFromPrevious(roomNumber, newMonth string) (*models.BillingRecord, error) {
	previous, err := s.repo.GetLatestByRoomNumber(roomNumber)
	if err != nil {
		return nil, err
	}

	newRecord := &models.BillingRecord{
		RoomNumber:       roomNumber,
		BillingMonth:     newMonth,
		CurrentWater:     previous.CurrentWater,
		PreviousWater:    previous.CurrentWater,
		CurrentElectric:  previous.CurrentElectric,
		PreviousElectric: previous.CurrentElectric,
		ManagementFee:    previous.ManagementFee,
		WaterPrice:       previous.WaterPrice,
		ElectricPrice:    previous.ElectricPrice,
	}

	if err := s.repo.Create(newRecord); err != nil {
		return nil, err
	}
	return newRecord, nil
}

// BatchDelete 批量删除
func (s *BillingService) BatchDelete(ids []int) int {
	count := 0
	for _, id := range ids {
		if err := s.repo.Delete(id); err == nil {
			count++
		}
	}
	return count
}

// BatchImport 批量导入
func (s *BillingService) BatchImport(records []models.BillingRecord) error {
	return s.repo.BatchImport(records)
}

// ExportToJSON 导出JSON
func (s *BillingService) ExportToJSON(filepath string) error {
	return s.repo.ExportToJSON(filepath)
}
