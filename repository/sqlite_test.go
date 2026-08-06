package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ginMeterBox/models"
)

func TestMigrateJSONToSQLitePreservesBillingAndTotalMeterData(t *testing.T) {
	dir := t.TempDir()
	billingFile := filepath.Join(dir, "billing.json")
	totalMeterFile := filepath.Join(dir, "total_meter.json")
	createdAt := time.Date(2026, 7, 1, 9, 0, 0, 123456789, time.FixedZone("CST", 8*3600))
	billingRecords := []models.BillingRecord{{
		ID: 7, RoomNumber: "101", BillingMonth: "2026-07",
		CurrentWater: 20, PreviousWater: 10, WaterAdjustment: 1,
		CurrentElectric: 30, PreviousElectric: 10, ElectricAdjustment: 2,
		ManagementFee: 9, WaterPrice: 4.3, ElectricPrice: 0.72,
		ExtraFees: []models.ExtraFee{{Name: "维修", Amount: 5}},
		CreatedAt: createdAt, UpdatedAt: createdAt.Add(time.Hour),
	}, {
		ID: 8, RoomNumber: "总表", BillingMonth: "2026-07",
		CurrentWater: 100, PreviousWater: 80, CurrentElectric: 200, PreviousElectric: 150,
		WaterPrice: 4.3, ElectricPrice: 0.72, CreatedAt: createdAt, UpdatedAt: createdAt,
	}}
	for i := range billingRecords {
		billingRecords[i].CalculateCosts()
	}
	totalMeterRecords := []models.TotalMeterRecord{{ID: 3, Month: "2026-07", WaterReading: 100, ElectricReading: 200, CreatedAt: createdAt, UpdatedAt: createdAt}}

	writeJSON := func(filename string, data any) {
		t.Helper()
		encoded, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, encoded, 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeJSON(billingFile, billingRecords)
	writeJSON(totalMeterFile, totalMeterRecords)

	db, err := OpenSQLite(filepath.Join(dir, "billing.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := MigrateJSONToSQLiteIfNeeded(db, billingFile, totalMeterFile); err != nil {
		t.Fatal(err)
	}
	// 第二次执行必须是幂等的，不能重复导入。
	if err := MigrateJSONToSQLiteIfNeeded(db, billingFile, totalMeterFile); err != nil {
		t.Fatal(err)
	}

	billingRepo := NewBillingSQLiteRepo(db)
	migrated := billingRepo.GetAll()
	if len(migrated) != 2 {
		t.Fatalf("migrated billing count = %d, want 2", len(migrated))
	}
	first, err := billingRepo.GetByID(7)
	if err != nil {
		t.Fatal(err)
	}
	if first.RoomNumber != "101" || first.CreatedAt.Format(time.RFC3339Nano) != billingRecords[0].CreatedAt.Format(time.RFC3339Nano) || len(first.ExtraFees) != 1 || first.ExtraFees[0].Name != "维修" {
		t.Fatalf("migrated billing record lost fields: %#v", first)
	}
	master, err := billingRepo.GetByID(8)
	if err != nil || master.RoomNumber != "总表" {
		t.Fatalf("historical master record = %#v, %v", master, err)
	}

	totalRepo := NewTotalMeterSQLiteRepo(db)
	total, err := totalRepo.GetByMonth("2026-07")
	if err != nil || total.ID != 3 || total.WaterReading != 100 {
		t.Fatalf("migrated total meter = %#v, %v", total, err)
	}
}

func TestBackfillLegacyMasterBillsToTotalMetersPreservesManualRecords(t *testing.T) {
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "billing.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	billingRepo := NewBillingSQLiteRepo(db)
	createdAt := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	for _, record := range []*models.BillingRecord{
		{RoomNumber: "总表", BillingMonth: "2026-06", CurrentWater: 8200, PreviousWater: 7828, CurrentElectric: 295857, PreviousElectric: 285094, WaterPrice: 4.3, ElectricPrice: 0.72},
		{RoomNumber: "总表", BillingMonth: "2026-07", CurrentWater: 8300, PreviousWater: 8200, CurrentElectric: 296000, PreviousElectric: 295857, WaterPrice: 4.3, ElectricPrice: 0.72},
	} {
		if err := billingRepo.Create(record); err != nil {
			t.Fatal(err)
		}
	}
	totalRepo := NewTotalMeterSQLiteRepo(db)
	manual := &models.TotalMeterRecord{Month: "2026-06", WaterReading: 9999, ElectricReading: 8888, CreatedAt: createdAt, UpdatedAt: createdAt}
	if err := totalRepo.Create(manual); err != nil {
		t.Fatal(err)
	}

	count, err := BackfillLegacyMasterBillsToTotalMeters(db)
	if err != nil || count != 1 {
		t.Fatalf("BackfillLegacyMasterBillsToTotalMeters() = (%d, %v), want (1, nil)", count, err)
	}
	if count, err = BackfillLegacyMasterBillsToTotalMeters(db); err != nil || count != 0 {
		t.Fatalf("second backfill = (%d, %v), want (0, nil)", count, err)
	}
	preserved, err := totalRepo.GetByMonth("2026-06")
	if err != nil || preserved.WaterReading != 9999 {
		t.Fatalf("manual total meter record was overwritten: %#v, %v", preserved, err)
	}
	backfilled, err := totalRepo.GetByMonth("2026-07")
	if err != nil || backfilled.WaterReading != 8300 || backfilled.ElectricReading != 296000 {
		t.Fatalf("backfilled total meter record = %#v, %v", backfilled, err)
	}
}

func TestBillingSQLiteRepoPersistsExtraFeesAndBatchAdjustment(t *testing.T) {
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "billing.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewBillingSQLiteRepo(db)
	record := recordForTest("101")
	record.ExtraFees = []models.ExtraFee{{Name: "公共照明", Amount: 3}}
	if err := repo.Create(record); err != nil {
		t.Fatal(err)
	}
	if record.ID == 0 {
		t.Fatal("Create did not assign id")
	}

	adjustment := 2.5
	count, err := repo.BatchUpdateAdjustments([]int{record.ID}, &adjustment, nil)
	if err != nil || count != 1 {
		t.Fatalf("BatchUpdateAdjustments() = (%d, %v)", count, err)
	}
	stored, err := repo.GetByID(record.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.WaterAdjustment != adjustment || len(stored.ExtraFees) != 1 || stored.TotalCost != stored.ManagementFee+stored.TotalWaterCost+stored.TotalElectricCost+3 {
		t.Fatalf("stored record incorrect: %#v", stored)
	}
}
