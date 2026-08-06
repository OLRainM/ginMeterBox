package repository

import "ginMeterBox/models"

// BillingRepo 账单数据访问接口
type BillingRepo interface {
	GetAll() []models.BillingRecord
	GetByID(id int) (*models.BillingRecord, error)
	GetByMonth(month string) []models.BillingRecord
	GetByIDs(ids []int) []models.BillingRecord
	GetLatestByRoomNumber(roomNumber string) (*models.BillingRecord, error)
	Create(record *models.BillingRecord) error
	Update(id int, record *models.BillingRecord) error
	Delete(id int) error
	BatchDelete(ids []int) (int, error)
	BatchUpdateAdjustments(ids []int, waterAdjustment, electricAdjustment *float64) (int, error)
	BatchSetExtraFees(ids []int, extraFees []models.ExtraFee, mode string) (int, error)
	BatchImport(records []models.BillingRecord) error
	ExportToJSON(filepath string) error
}

// TotalMeterRepo 总表数据访问接口
type TotalMeterRepo interface {
	GetAll() []models.TotalMeterRecord
	GetByMonth(month string) (*models.TotalMeterRecord, error)
	Create(record *models.TotalMeterRecord) error
	Update(month string, record *models.TotalMeterRecord) error
	Delete(month string) error
}
