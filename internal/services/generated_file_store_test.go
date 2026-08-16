package services

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratedFileStoreAllocatesAndResolvesGeneratedFiles(t *testing.T) {
	root := t.TempDir()
	store := NewGeneratedFileStore(filepath.Join(root, "exports"), filepath.Join(root, "reports"))

	reportName, reportPath, err := store.NewReportFile()
	if err != nil {
		t.Fatalf("NewReportFile() error = %v", err)
	}
	if filepath.Base(reportPath) != reportName {
		t.Fatalf("report path basename = %q, want %q", filepath.Base(reportPath), reportName)
	}
	if err := os.WriteFile(reportPath, []byte("png"), 0644); err != nil {
		t.Fatal(err)
	}
	resolvedReport, err := store.ResolveReportDownload(reportName)
	if err != nil {
		t.Fatalf("ResolveReportDownload() error = %v", err)
	}
	if resolvedReport != reportPath {
		t.Fatalf("resolved report = %q, want %q", resolvedReport, reportPath)
	}

	exportName, exportPath, err := store.NewExportFile(".json")
	if err != nil {
		t.Fatalf("NewExportFile() error = %v", err)
	}
	if err := os.WriteFile(exportPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	resolvedExport, err := store.ResolveExportDownload(exportName)
	if err != nil {
		t.Fatalf("ResolveExportDownload() error = %v", err)
	}
	if resolvedExport != exportPath {
		t.Fatalf("resolved export = %q, want %q", resolvedExport, exportPath)
	}
}

func TestGeneratedFileStoreRejectsUnsafeDownloadIdentifiers(t *testing.T) {
	store := NewGeneratedFileStore(filepath.Join(t.TempDir(), "exports"), filepath.Join(t.TempDir(), "reports"))

	unsafeNames := []string{
		"../generated_20260731120000_0123456789abcdef.png",
		"generated_20260731120000_0123456789abcdef.txt",
		"generated_20260731120000_0123456789abcdef.json",
		"C:\\generated_20260731120000_0123456789abcdef.png",
		"reports/generated_20260731120000_0123456789abcdef.png",
		"not-generated.png",
	}
	for _, name := range unsafeNames {
		_, err := store.ResolveReportDownload(name)
		if !errors.Is(err, ErrInvalidGeneratedFile) {
			t.Errorf("ResolveReportDownload(%q) error = %v, want ErrInvalidGeneratedFile", name, err)
		}
	}

	if _, _, err := store.NewExportFile(".csv"); !errors.Is(err, ErrInvalidGeneratedFile) {
		t.Fatalf("NewExportFile(.csv) error = %v, want ErrInvalidGeneratedFile", err)
	}
}

func TestGeneratedFileStoreRejectsMissingAndSymlinkedFiles(t *testing.T) {
	root := t.TempDir()
	reportDir := filepath.Join(root, "reports")
	store := NewGeneratedFileStore(filepath.Join(root, "exports"), reportDir)
	name := "generated_20260731120000_0123456789abcdef.png"

	if _, err := store.ResolveReportDownload(name); !errors.Is(err, ErrGeneratedFileNotFound) {
		t.Fatalf("missing file error = %v, want ErrGeneratedFileNotFound", err)
	}

	outside := filepath.Join(root, "outside.txt")
	if err := os.WriteFile(outside, []byte("private"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		t.Fatal(err)
	}
	linkPath := filepath.Join(reportDir, name)
	if err := os.Symlink(outside, linkPath); err != nil {
		t.Skipf("symbolic links unavailable in this environment: %v", err)
	}

	if _, err := store.ResolveReportDownload(name); !errors.Is(err, ErrInvalidGeneratedFile) {
		t.Fatalf("symlinked file error = %v, want ErrInvalidGeneratedFile", err)
	}
}
