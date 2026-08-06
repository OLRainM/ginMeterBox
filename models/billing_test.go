package models

import "testing"

func TestBillingRecordCalculateCosts(t *testing.T) {
	record := BillingRecord{
		RoomNumber:         "A101",
		CurrentWater:       20,
		PreviousWater:      10,
		WaterAdjustment:    2,
		CurrentElectric:    150,
		PreviousElectric:   100,
		ElectricAdjustment: 5,
		ManagementFee:      30,
		WaterPrice:         3,
		ElectricPrice:      0.5,
		ExtraFees: []ExtraFee{
			{Name: "维修费", Amount: 12.5},
			{Name: "清洁费", Amount: 7.5},
		},
	}

	record.CalculateCosts()

	if record.WaterUsage != 12 {
		t.Fatalf("WaterUsage = %v, want 12", record.WaterUsage)
	}
	if record.ElectricUsage != 55 {
		t.Fatalf("ElectricUsage = %v, want 55", record.ElectricUsage)
	}
	if record.TotalWaterCost != 36 {
		t.Fatalf("TotalWaterCost = %v, want 36", record.TotalWaterCost)
	}
	if record.TotalElectricCost != 27.5 {
		t.Fatalf("TotalElectricCost = %v, want 27.5", record.TotalElectricCost)
	}
	if record.TotalCost != 113.5 {
		t.Fatalf("TotalCost = %v, want 113.5", record.TotalCost)
	}
}

func TestBillingRecordCalculateCostsForMainMeter(t *testing.T) {
	record := BillingRecord{
		RoomNumber:    "总表",
		ManagementFee: 100,
		CurrentWater:  10,
		WaterPrice:    2,
	}

	record.CalculateCosts()

	if record.ManagementFee != 0 {
		t.Fatalf("ManagementFee = %v, want 0", record.ManagementFee)
	}
	if record.TotalCost != 20 {
		t.Fatalf("TotalCost = %v, want 20", record.TotalCost)
	}
}
