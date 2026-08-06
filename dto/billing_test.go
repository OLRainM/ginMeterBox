package dto

import (
	"math"
	"testing"

	"ginMeterBox/models"
)

func validBillingRecordRequest() BillingRecordRequest {
	return BillingRecordRequest{
		RoomNumber:       "A101",
		BillingMonth:     "2026-07",
		CurrentWater:     20,
		PreviousWater:    10,
		CurrentElectric:  100,
		PreviousElectric: 50,
		ManagementFee:    10,
		WaterPrice:       2,
		ElectricPrice:    0.5,
		ExtraFees:        []models.ExtraFee{{Name: "维修费", Amount: 5}},
	}
}

func TestBillingRecordRequestValidate(t *testing.T) {
	if err := validBillingRecordRequest().Validate(); err != nil {
		t.Fatalf("有效请求校验失败: %v", err)
	}

	invalid := validBillingRecordRequest()
	invalid.RoomNumber = " "
	if err := invalid.Validate(); err == nil {
		t.Fatal("空房号应校验失败")
	}

	invalid = validBillingRecordRequest()
	invalid.BillingMonth = "2026-13"
	if err := invalid.Validate(); err == nil {
		t.Fatal("无效月份应校验失败")
	}

	invalid = validBillingRecordRequest()
	invalid.CurrentWater = math.Inf(1)
	if err := invalid.Validate(); err == nil {
		t.Fatal("无穷读数应校验失败")
	}

	invalid = validBillingRecordRequest()
	invalid.CurrentElectric = 1
	if err := invalid.Validate(); err == nil {
		t.Fatal("倒退电表读数应校验失败")
	}
}

func TestValidateIDs(t *testing.T) {
	if err := ValidateIDs([]int{1, 2, 3}); err != nil {
		t.Fatalf("有效 ID 校验失败: %v", err)
	}
	if err := ValidateIDs([]int{1, 1}); err == nil {
		t.Fatal("重复 ID 应校验失败")
	}
	if err := ValidateIDs(make([]int, MaxBatchSize+1)); err == nil {
		t.Fatal("超过批量上限应校验失败")
	}
}

func TestContinueRequestsRejectMasterMeter(t *testing.T) {
	if err := (ContinueRequest{RoomNumber: "总表", NewMonth: "2026-07"}).Validate(); err == nil {
		t.Fatal("master meter must not be continued as a household bill")
	}
	if err := (BatchContinueRequest{RoomNumbers: []string{"101", "总表"}, NewMonth: "2026-07"}).Validate(); err == nil {
		t.Fatal("batch continuation must reject master meter")
	}
	if err := (BatchContinueRequest{RoomNumbers: []string{"101", "103"}, NewMonth: "2026-07"}).Validate(); err != nil {
		t.Fatalf("valid household continuation rejected: %v", err)
	}
}

func TestBillingRecordRequestToRecordDropsDerivedFields(t *testing.T) {
	record := validBillingRecordRequest().ToRecord()
	if record.ID != 0 || !record.CreatedAt.IsZero() || !record.UpdatedAt.IsZero() {
		t.Fatal("DTO 转换不应写入服务端字段")
	}
	if record.WaterUsage != 0 || record.ElectricUsage != 0 || record.TotalCost != 0 {
		t.Fatal("DTO 转换不应写入派生费用字段")
	}
}
