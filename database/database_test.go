package database

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init temp db: %v", err)
	}
	return db
}

func mustExec(t *testing.T, db *DB, query string, args ...interface{}) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("exec failed: %v", err)
	}
}

func TestWithTransactionCommitAndRollback(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	mustExec(t, db, `CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)`)

	if err := db.WithTransaction(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO items (name) VALUES (?)", "ok")
		return err
	}); err != nil {
		t.Fatalf("commit transaction failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row after commit, got %d", count)
	}

	// Force a rollback by returning an error from the TxFn.
	_ = db.WithTransaction(func(tx *sql.Tx) error {
		if _, err := tx.Exec("INSERT INTO items (name) VALUES (?)", "rollback"); err != nil {
			return err
		}
		return errors.New("boom")
	})

	if err := db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected rollback to keep 1 row, got %d", count)
	}
}

func TestWithTimeoutTransaction(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	err := db.WithTimeoutTransaction(context.Background(), 10*time.Millisecond, func(tx *sql.Tx) error {
		time.Sleep(50 * time.Millisecond)
		_, execErr := tx.Exec("SELECT 1")
		return execErr
	})

	if err == nil || !strings.Contains(err.Error(), "transaction timed out") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestCleanupDuplicateServices(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	// Seed duplicate services and a relationship.
	mustExec(t, db, `INSERT INTO services (id, name, type, config) VALUES (?, ?, ?, ?)`,
		"svc@file", "svc", "loadBalancer", "{}")
	mustExec(t, db, `INSERT INTO services (id, name, type, config) VALUES (?, ?, ?, ?)`,
		"svc", "svc", "loadBalancer", "{}")
	mustExec(t, db, `INSERT INTO services (id, name, type, config) VALUES (?, ?, ?, ?)`,
		"other", "other", "loadBalancer", "{}")

	mustExec(t, db, `INSERT INTO resources (id, host, service_id, org_id, site_id, status) VALUES (?, ?, ?, ?, ?, ?)`,
		"res1", "example.com", "svc@file", "org", "site", "active")
	mustExec(t, db, `INSERT INTO resource_services (resource_id, service_id) VALUES (?, ?)`,
		"res1", "svc@file")

	if err := db.CleanupDuplicateServices(DefaultCleanupOptions()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	// Only the unsuffixed service should remain.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM services WHERE id LIKE 'svc%'`).Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 svc* row after cleanup, got %d", count)
	}

	var remainingID string
	if err := db.QueryRow(`SELECT id FROM services WHERE id LIKE 'svc%'`).Scan(&remainingID); err != nil {
		t.Fatalf("select failed: %v", err)
	}
	if remainingID != "svc" {
		t.Fatalf("expected svc to remain, got %s", remainingID)
	}

	// Relationship to deleted service should be gone.
	if err := db.QueryRow(`SELECT COUNT(*) FROM resource_services`).Scan(&count); err != nil {
		t.Fatalf("resource_services count failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected relationships cleaned up, got %d", count)
	}
}
