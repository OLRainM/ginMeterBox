package dto

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"ginMeterBox/internal/models"
)

const (
	maxRoomNumberLength = 64
	maxExtraFees        = 20
	maxFeeNameLength    = 64
	maxBatchSize        = 500
)

// MaxBatchSize 是单次批量操作允许的最大记录数。
const MaxBatchSize = maxBatchSize

var (
	monthPattern = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)
	roomPattern  = regexp.MustCompile(`^[\p{Han}A-Za-z0-9][\p{Han}A-Za-z0-9 _#-]*$`)
)

type BillingRecordRequest struct {
	RoomNumber         string            `json:"roomNumber"`
	CurrentWater       float64           `json:"currentWater"`
	PreviousWater      float64           `json:"previousWater"`
	WaterAdjustment    float64           `json:"waterAdjustment"`
	CurrentElectric    float64           `json:"currentElectric"`
	PreviousElectric   float64           `json:"previousElectric"`
	ElectricAdjustment float64           `json:"electricAdjustment"`
	ManagementFee      float64           `json:"managementFee"`
	WaterPrice         float64           `json:"waterPrice"`
	ElectricPrice      float64           `json:"electricPrice"`
	ExtraFees          []models.ExtraFee `json:"extraFees"`
	BillingMonth       string            `json:"billingMonth"`
}

func (r BillingRecordRequest) Validate() error {
	if err := ValidateRoomNumber(r.RoomNumber); err != nil {
		return err
	}
	if err := ValidateMonth(r.BillingMonth); err != nil {
		return err
	}
	values := []float64{
		r.CurrentWater, r.PreviousWater, r.WaterAdjustment,
		r.CurrentElectric, r.PreviousElectric, r.ElectricAdjustment,
		r.ManagementFee, r.WaterPrice, r.ElectricPrice,
	}
	for _, value := range values {
		if !isNonNegativeFinite(value) {
			return fmt.Errorf("表读数、补差、单价和费用必须为非负有限数值")
		}
	}
	if r.CurrentWater+r.WaterAdjustment < r.PreviousWater || r.CurrentElectric+r.ElectricAdjustment < r.PreviousElectric {
		return fmt.Errorf("补差后的本月读数不能小于上月读数")
	}
	return ValidateExtraFees(r.ExtraFees)
}

func (r BillingRecordRequest) ToRecord() models.BillingRecord {
	return models.BillingRecord{
		RoomNumber:         strings.TrimSpace(r.RoomNumber),
		CurrentWater:       r.CurrentWater,
		PreviousWater:      r.PreviousWater,
		WaterAdjustment:    r.WaterAdjustment,
		CurrentElectric:    r.CurrentElectric,
		PreviousElectric:   r.PreviousElectric,
		ElectricAdjustment: r.ElectricAdjustment,
		ManagementFee:      r.ManagementFee,
		WaterPrice:         r.WaterPrice,
		ElectricPrice:      r.ElectricPrice,
		ExtraFees:          append([]models.ExtraFee(nil), r.ExtraFees...),
		BillingMonth:       r.BillingMonth,
	}
}

type TotalMeterRecordRequest struct {
	Month           string  `json:"month"`
	WaterReading    float64 `json:"waterReading"`
	ElectricReading float64 `json:"electricReading"`
}

func (r TotalMeterRecordRequest) Validate() error {
	if err := ValidateMonth(r.Month); err != nil {
		return err
	}
	if !isNonNegativeFinite(r.WaterReading) || !isNonNegativeFinite(r.ElectricReading) {
		return fmt.Errorf("总表读数必须为非负有限数值")
	}
	return nil
}

func (r TotalMeterRecordRequest) ToRecord() models.TotalMeterRecord {
	return models.TotalMeterRecord{Month: r.Month, WaterReading: r.WaterReading, ElectricReading: r.ElectricReading}
}

type BatchDeleteRequest struct {
	IDs []int `json:"ids"`
}

func (r BatchDeleteRequest) Validate() error {
	return ValidateIDs(r.IDs)
}

type BatchAdjustmentRequest struct {
	IDs                []int    `json:"ids"`
	WaterAdjustment    *float64 `json:"waterAdjustment"`
	ElectricAdjustment *float64 `json:"electricAdjustment"`
}

func (r BatchAdjustmentRequest) Validate() error {
	if err := ValidateIDs(r.IDs); err != nil {
		return err
	}
	if r.WaterAdjustment == nil && r.ElectricAdjustment == nil {
		return fmt.Errorf("请至少设置一个补差值")
	}
	for _, value := range []*float64{r.WaterAdjustment, r.ElectricAdjustment} {
		if value != nil && !isNonNegativeFinite(*value) {
			return fmt.Errorf("补差必须为非负有限数值")
		}
	}
	return nil
}

type BatchExtraFeeRequest struct {
	IDs       []int             `json:"ids"`
	ExtraFees []models.ExtraFee `json:"extraFees"`
	Mode      string            `json:"mode"`
}

func (r BatchExtraFeeRequest) Validate() error {
	if err := ValidateIDs(r.IDs); err != nil {
		return err
	}
	if len(r.ExtraFees) == 0 {
		return fmt.Errorf("额外费用项数量无效")
	}
	if err := ValidateExtraFees(r.ExtraFees); err != nil {
		return err
	}
	if r.Mode != "" && r.Mode != "append" && r.Mode != "replace" {
		return fmt.Errorf("额外费用模式无效")
	}
	return nil
}

type SmartWaterMatchRequest struct {
	IDs           []int     `json:"ids"`
	WaterReadings []float64 `json:"waterReadings"`
}

type GenerateReportRequest struct {
	IDs    []int  `json:"ids"`
	Month  string `json:"month"`
	SortBy string `json:"sortBy"`
	Order  string `json:"order"`
}

func (r GenerateReportRequest) Validate() error {
	if len(r.IDs) == 0 && r.Month == "" {
		return fmt.Errorf("请提供 ids 或 month 参数")
	}
	if len(r.IDs) > 0 {
		if err := ValidateIDs(r.IDs); err != nil {
			return err
		}
	}
	if r.Month != "" {
		if err := ValidateMonth(r.Month); err != nil {
			return err
		}
	}
	if r.SortBy != "" && r.SortBy != "room" {
		return fmt.Errorf("排序字段无效")
	}
	if r.Order != "" && r.Order != "asc" && r.Order != "desc" {
		return fmt.Errorf("排序方向无效")
	}
	return nil
}

func (r SmartWaterMatchRequest) Validate() error {
	if err := ValidateIDs(r.IDs); err != nil {
		return err
	}
	if len(r.IDs) > 10 {
		return fmt.Errorf("为保证性能，单次匹配用户数量不能超过 10 个")
	}
	if len(r.WaterReadings) != len(r.IDs) {
		return fmt.Errorf("用户数量与水表读数数量不匹配")
	}
	for _, reading := range r.WaterReadings {
		if !isNonNegativeFinite(reading) {
			return fmt.Errorf("水表读数必须为非负有限数值")
		}
	}
	return nil
}

type ContinueRequest struct {
	RoomNumber string `json:"roomNumber"`
	NewMonth   string `json:"newMonth"`
}

func (r ContinueRequest) Validate() error {
	if err := ValidateContinuationRoomNumber(r.RoomNumber); err != nil {
		return err
	}
	return ValidateMonth(r.NewMonth)
}

type BatchContinueRequest struct {
	RoomNumbers []string `json:"roomNumbers"`
	NewMonth    string   `json:"newMonth"`
}

func (r BatchContinueRequest) Validate() error {
	if len(r.RoomNumbers) == 0 || len(r.RoomNumbers) > maxBatchSize {
		return fmt.Errorf("房号数量必须在 1 到 %d 之间", maxBatchSize)
	}
	for _, roomNumber := range r.RoomNumbers {
		if err := ValidateContinuationRoomNumber(roomNumber); err != nil {
			return err
		}
	}
	return ValidateMonth(r.NewMonth)
}

func ValidateMonth(month string) error {
	if !monthPattern.MatchString(month) {
		return fmt.Errorf("月份必须为 YYYY-MM 格式")
	}
	return nil
}

func ValidateRoomNumber(roomNumber string) error {
	roomNumber = strings.TrimSpace(roomNumber)
	if !roomPattern.MatchString(roomNumber) || len(roomNumber) > maxRoomNumberLength {
		return fmt.Errorf("房号格式无效")
	}
	return nil
}

func ValidateContinuationRoomNumber(roomNumber string) error {
	if err := ValidateRoomNumber(roomNumber); err != nil {
		return err
	}
	if strings.TrimSpace(roomNumber) == "总表" {
		return fmt.Errorf("总表为独立计量项，不能参与账单延续")
	}
	return nil
}

func ValidateExtraFees(extraFees []models.ExtraFee) error {
	if len(extraFees) > maxExtraFees {
		return fmt.Errorf("额外费用项不能超过 %d 项", maxExtraFees)
	}
	for _, fee := range extraFees {
		if strings.TrimSpace(fee.Name) == "" || len(fee.Name) > maxFeeNameLength || !isNonNegativeFinite(fee.Amount) {
			return fmt.Errorf("额外费用格式无效")
		}
	}
	return nil
}

func ValidateIDs(ids []int) error {
	if len(ids) == 0 || len(ids) > maxBatchSize {
		return fmt.Errorf("记录数量必须在 1 到 %d 之间", maxBatchSize)
	}
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return fmt.Errorf("记录 ID 必须为正整数")
		}
		if _, ok := seen[id]; ok {
			return fmt.Errorf("记录 ID 不能重复")
		}
		seen[id] = struct{}{}
	}
	return nil
}

func isNonNegativeFinite(value float64) bool {
	return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}
