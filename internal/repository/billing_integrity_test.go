package repository

import (
	"errors"
	"path/filepath"
	"testing"

	"ginMeterBox/internal/models"
)

func TestBillingSQLiteRepoRejectsDuplicateRoomAndMonth(t *testing.T) {
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "billing.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewBillingSQLiteRepo(db)
	if err := repo.Create(recordForTest("101")); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(recordForTest("101")); !errors.Is(err, ErrBillingPeriodExists) {
		t.Fatalf("Create() error = %v, want ErrBillingPeriodExists", err)
	}
}

func TestBillingJSONRepoRejectsDuplicateRoomAndMonth(t *testing.T) {
	repo := newTestBillingRepo(t)
	if err := repo.Create(recordForTest("101")); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(recordForTest("101")); !errors.Is(err, ErrBillingPeriodExists) {
		t.Fatalf("Create() error = %v, want ErrBillingPeriodExists", err)
	}
}

func TestBatchAdjustmentRejectsNegativeUsageWithoutPersisting(t *testing.T) {
	tests := []struct {
		name string
		new  func(*testing.T) BillingRepo
	}{
		{
			name: "sqlite",
			new: func(t *testing.T) BillingRepo {
				db, err := OpenSQLite(filepath.Join(t.TempDir(), "billing.db"))
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { db.Close() })
				return NewBillingSQLiteRepo(db)
			},
		},
		{
			name: "json",
			new:  func(t *testing.T) BillingRepo { return newTestBillingRepo(t) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.new(t)
			record := &models.BillingRecord{
				RoomNumber: "101", BillingMonth: "2026-07",
				CurrentWater: 4, PreviousWater: 5, WaterAdjustment: 1,
				CurrentElectric: 10, PreviousElectric: 5,
				WaterPrice: 1, ElectricPrice: 1,
			}
			if err := repo.Create(record); err != nil {
				t.Fatal(err)
			}
			zero := 0.0
			if _, err := repo.BatchUpdateAdjustments([]int{record.ID}, &zero, nil); !errors.Is(err, ErrInvalidUsage) {
				t.Fatalf("BatchUpdateAdjustments() error = %v, want ErrInvalidUsage", err)
			}
			stored, err := repo.GetByID(record.ID)
			if err != nil {
				t.Fatal(err)
			}
			if stored.WaterAdjustment != 1 || stored.WaterUsage != 0 {
				t.Fatalf("record was changed despite rejected batch update: %#v", stored)
			}
		})
	}
}

func TestBatchWaterReadingUpdateIsAtomic(t *testing.T) {
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "billing.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewBillingSQLiteRepo(db)
	first, second := recordForTest("101"), recordForTest("102")
	if err := repo.Create(first); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(second); err != nil {
		t.Fatal(err)
	}
	if err := repo.BatchUpdateWaterReadings([]WaterReadingUpdate{
		{ID: first.ID, CurrentWater: 20},
		{ID: second.ID, CurrentWater: 0},
	}); !errors.Is(err, ErrInvalidUsage) {
		t.Fatalf("BatchUpdateWaterReadings() error = %v, want ErrInvalidUsage", err)
	}
	stored, err := repo.GetByID(first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.CurrentWater != first.CurrentWater {
		t.Fatalf("first record was partially updated: got %v, want %v", stored.CurrentWater, first.CurrentWater)
	}
}
