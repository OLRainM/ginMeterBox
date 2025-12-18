package models

import "time"

// ExtraFee 额外费用项
type ExtraFee struct {
	Name   string  `json:"name"`   // 费用名称（如"水管维修费"）
	Amount float64 `json:"amount"` // 费用金额
}

// BillingRecord 水电费账单记录
type BillingRecord struct {
	ID                int        `json:"id"`
	RoomNumber        string     `json:"roomNumber"`        // 住户编号
	CurrentWater      float64    `json:"currentWater"`      // 本月水表读数
	PreviousWater     float64    `json:"previousWater"`     // 上月水表读数
	WaterAdjustment   float64    `json:"waterAdjustment"`   // 水表补差
	WaterUsage        float64    `json:"waterUsage"`        // 用水量
	CurrentElectric   float64    `json:"currentElectric"`   // 本月电表读数
	PreviousElectric  float64    `json:"previousElectric"`  // 上月电表读数
	ElectricAdjustment float64   `json:"electricAdjustment"` // 电表补差
	ElectricUsage     float64    `json:"electricUsage"`     // 用电量
	ManagementFee     float64    `json:"managementFee"`     // 管理费
	WaterPrice        float64    `json:"waterPrice"`        // 水单价
	ElectricPrice     float64    `json:"electricPrice"`     // 电单价
	TotalWaterCost    float64    `json:"totalWaterCost"`    // 水费
	TotalElectricCost float64    `json:"totalElectricCost"` // 电费
	ExtraFees         []ExtraFee `json:"extraFees,omitempty"` // 额外费用列表（为空时不输出）
	TotalCost         float64    `json:"totalCost"`         // 总费用
	BillingMonth      string     `json:"billingMonth"`      // 缴费月份
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// CalculateCosts 计算各项费用
func (b *BillingRecord) CalculateCosts() {
	// 用水量 = 本月读数 - 上月读数 + 补差
	b.WaterUsage = b.CurrentWater - b.PreviousWater + b.WaterAdjustment
	
	// 用电量 = 本月读数 - 上月读数 + 补差
	b.ElectricUsage = b.CurrentElectric - b.PreviousElectric + b.ElectricAdjustment
	
	// 水费 = 用水量 × 水单价
	b.TotalWaterCost = b.WaterUsage * b.WaterPrice
	
	// 电费 = 用电量 × 电单价
	b.TotalElectricCost = b.ElectricUsage * b.ElectricPrice
	
	// 计算额外费用总和
	extraTotal := 0.0
	for _, fee := range b.ExtraFees {
		extraTotal += fee.Amount
	}
	
	// 总费用 = 管理费 + 水费 + 电费 + 额外费用
	b.TotalCost = b.ManagementFee + b.TotalWaterCost + b.TotalElectricCost + extraTotal
}

// PriceConfig 价格配置
type PriceConfig struct {
	WaterPrice    float64 `json:"waterPrice"`    // 水单价
	ElectricPrice float64 `json:"electricPrice"` // 电单价
}
