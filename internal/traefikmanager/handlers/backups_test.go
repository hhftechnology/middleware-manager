package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBackupHandlerCreateListDelete(t *testing.T) {
	env := newTestEnv(t)
	handler := NewBackupHandler(env.files)

	router := gin.New()
	router.POST("/backups", handler.Create)
	router.GET("/backups", handler.List)
	router.DELETE("/backups/:filename", handler.Delete)

	// Create backup
	req := httptest.NewRequest(http.MethodPost, "/backups", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("create backup status %d: %s", rec.Code, rec.Body.String())
	}
	created := map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	name, _ := created["name"].(string)
	if name == "" {
		t.Fatalf("expected non-empty backup name in %v", created)
	}

	// Verify the backup file was written inside the backup dir.
	entries, err := os.ReadDir(env.cfg.BackupDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("expected backup in dir %s: entries=%v err=%v", env.cfg.BackupDir, entries, err)
	}

	// List
	req = httptest.NewRequest(http.MethodGet, "/backups", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status %d", rec.Code)
	}
	var backups []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &backups); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(backups) == 0 {
		t.Fatalf("expected at least one backup, got %v", backups)
	}

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/backups/"+name, nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status %d: %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(env.cfg.BackupDir, name)); !os.IsNotExist(err) {
		t.Fatalf("expected backup file to be removed, got err=%v", err)
	}
}

func TestBackupHandlerDeleteTraversalRejected(t *testing.T) {
	env := newTestEnv(t)
	handler := NewBackupHandler(env.files)
	router := gin.New()
	router.DELETE("/backups/:filename", handler.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/backups/..%2Fsecret.bak", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatalf("expected traversal delete to be rejected, got 200")
	}
}
