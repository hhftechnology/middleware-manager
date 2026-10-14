package services

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/hhftechnology/middleware-manager/database"
)

func waitUntil(t *testing.T, cond func() bool, message string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal(message)
}

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

// newTestSQLDB returns the underlying *sql.DB for tests that require it
func newTestSQLDB(t *testing.T) *sql.DB {
	t.Helper()
	return newTestDB(t).DB
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
