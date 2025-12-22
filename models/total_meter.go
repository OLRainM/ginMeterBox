package models

import "time"

// TotalMeterRecord 总表记录
type TotalMeterRecord struct {
	ID            int       `json:"id"`
	Month         string    `json:"month"`         // 月份 (YYYY-MM)
	WaterReading  float64   `json:"waterReading"`  // 水表总表读数
	ElectricReading float64 `json:"electricReading"` // 电表总表读数
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
