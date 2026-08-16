package services

import (
	"errors"
	"path/filepath"
	"testing"

	"ginMeterBox/internal/models"
	"ginMeterBox/internal/repository"
)

func TestBatchContinueIsAtomicWhenTargetMonthAlreadyExists(t *testing.T) {
	repo, err := repository.NewBillingJSONRepo(filepath.Join(t.TempDir(), "billing.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range []*models.BillingRecord{
		{RoomNumber: "101", BillingMonth: "2026-07", CurrentWater: 10, PreviousWater: 5, CurrentElectric: 10, PreviousElectric: 5, WaterPrice: 1, ElectricPrice: 1},
		{RoomNumber: "101", BillingMonth: "2026-08", CurrentWater: 15, PreviousWater: 10, CurrentElectric: 15, PreviousElectric: 10, WaterPrice: 1, ElectricPrice: 1},
		{RoomNumber: "102", BillingMonth: "2026-07", CurrentWater: 20, PreviousWater: 10, CurrentElectric: 20, PreviousElectric: 10, WaterPrice: 1, ElectricPrice: 1},
	} {
		if err := repo.Create(record); err != nil {
			t.Fatal(err)
		}
	}

	service := NewBillingService(repo)
	if err := service.BatchContinueFromPrevious([]string{"101", "102"}, "2026-08"); !errors.Is(err, repository.ErrBillingPeriodExists) {
		t.Fatalf("BatchContinueFromPrevious() error = %v, want ErrBillingPeriodExists", err)
	}
	if _, err := repo.GetByRoomAndMonth("102", "2026-08"); !errors.Is(err, repository.ErrRecordNotFound) {
		t.Fatalf("room 102 continuation should not exist after failed batch, got %v", err)
	}
}

func TestBatchContinueWorksWithSQLite(t *testing.T) {
	db, err := repository.OpenSQLite(filepath.Join(t.TempDir(), "billing.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewBillingSQLiteRepo(db)
	for _, record := range []*models.BillingRecord{
		{RoomNumber: "101", BillingMonth: "2026-06", CurrentWater: 12, PreviousWater: 8, CurrentElectric: 100, PreviousElectric: 80, ManagementFee: 20, WaterPrice: 4.3, ElectricPrice: 0.72},
		{RoomNumber: "103", BillingMonth: "2026-06", CurrentWater: 30, PreviousWater: 25, CurrentElectric: 220, PreviousElectric: 200, ManagementFee: 20, WaterPrice: 4.3, ElectricPrice: 0.72},
	} {
		if err := repo.Create(record); err != nil {
			t.Fatal(err)
		}
	}

	service := NewBillingService(repo)
	if err := service.BatchContinueFromPrevious([]string{"101", "103"}, "2026-07"); err != nil {
		t.Fatalf("BatchContinueFromPrevious() with SQLite error = %v", err)
	}
	for _, roomNumber := range []string{"101", "103"} {
		record, err := repo.GetByRoomAndMonth(roomNumber, "2026-07")
		if err != nil {
			t.Fatalf("continuation record for %s missing: %v", roomNumber, err)
		}
		if record.CurrentWater != record.PreviousWater || record.CurrentElectric != record.PreviousElectric {
			t.Fatalf("continuation readings for %s were not carried forward", roomNumber)
		}
	}
}
