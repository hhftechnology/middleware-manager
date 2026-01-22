package services

import (
	"path/filepath"
	"testing"

	"github.com/hhftechnology/middleware-manager/database"
)

func newTestDB(t *testing.T) *database.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init temp db: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	return db
}

func newTestConfigManager(t *testing.T) *ConfigManager {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	cm, err := NewConfigManager(cfgPath)
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}
	return cm
}
