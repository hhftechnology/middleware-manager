package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

func TestSaveTemplateFileIdempotent(t *testing.T) {
	dir := t.TempDir()

	if err := SaveTemplateFile(dir); err != nil {
		t.Fatalf("first save failed: %v", err)
	}

	firstBytes, err := os.ReadFile(filepath.Join(dir, "templates.yaml"))
	if err != nil {
		t.Fatalf("templates file missing after save: %v", err)
	}

	if err := SaveTemplateFile(dir); err != nil {
		t.Fatalf("second save should be idempotent: %v", err)
	}

	secondBytes, err := os.ReadFile(filepath.Join(dir, "templates.yaml"))
	if err != nil {
		t.Fatalf("templates file missing after second save: %v", err)
	}

	if string(firstBytes) != string(secondBytes) {
		t.Fatalf("templates content should stay unchanged on repeated save")
	}
}

func TestSaveTemplateServicesFileIdempotent(t *testing.T) {
	dir := t.TempDir()

	if err := SaveTemplateServicesFile(dir); err != nil {
		t.Fatalf("first save failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "templates_services.yaml")); err != nil {
		t.Fatalf("services template missing: %v", err)
	}

	if err := SaveTemplateServicesFile(dir); err != nil {
		t.Fatalf("second save should be idempotent: %v", err)
	}
}

func TestLoadDefaultTemplatesInsertRecords(t *testing.T) {
	db := testutil.NewTempDB(t)

	if err := LoadDefaultTemplates(db); err != nil {
		t.Fatalf("load templates failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM middlewares WHERE id = 'basic-auth'").Scan(&count); err != nil {
		t.Fatalf("select failed: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected default middleware to be inserted")
	}
}

func TestLoadDefaultServiceTemplatesInsertRecords(t *testing.T) {
	db := testutil.NewTempDB(t)

	if err := LoadDefaultServiceTemplates(db); err != nil {
		t.Fatalf("load service templates failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM services WHERE id = 'simple-http'").Scan(&count); err != nil {
		t.Fatalf("select failed: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected default service to be inserted")
	}
}
