package repository

import (
	"os"
	"path/filepath"
	"testing"

	"ginMeterBox/internal/models"
)

func newTestBillingRepo(t *testing.T) *BillingJSONRepo {
	t.Helper()
	repo, err := NewBillingJSONRepo(filepath.Join(t.TempDir(), "nested", "billing.json"))
	if err != nil {
		t.Fatalf("NewBillingJSONRepo() error = %v", err)
	}
	return repo
}

func recordForTest(room string) *models.BillingRecord {
	return &models.BillingRecord{
		RoomNumber: room, BillingMonth: "2026-07",
		CurrentWater: 10, PreviousWater: 5, CurrentElectric: 20, PreviousElectric: 10,
		WaterPrice: 1, ElectricPrice: 1, ManagementFee: 5,
	}
}

func TestBillingJSONRepoBatchUpdateIsPersistedOnceAndReloadable(t *testing.T) {
	repo := newTestBillingRepo(t)
	first, second := recordForTest("101"), recordForTest("102")
	if err := repo.Create(first); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(second); err != nil {
		t.Fatal(err)
	}
	adjustment := 2.0
	count, err := repo.BatchUpdateAdjustments([]int{first.ID, second.ID}, &adjustment, nil)
	if err != nil || count != 2 {
		t.Fatalf("BatchUpdateAdjustments() = (%d, %v), want (2, nil)", count, err)
	}
	reloaded, err := NewBillingJSONRepo(repo.dataFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range reloaded.GetAll() {
		if record.WaterAdjustment != adjustment {
			t.Fatalf("record %d water adjustment = %v, want %v", record.ID, record.WaterAdjustment, adjustment)
		}
	}
}

func TestWriteJSONAtomicallyReplacesExistingFile(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "data.json")
	if err := os.WriteFile(filename, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := writeJSONAtomically(filename, []byte("new")); err != nil {
		t.Fatalf("writeJSONAtomically() error = %v", err)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Fatalf("file content = %q, want new", data)
	}
}
