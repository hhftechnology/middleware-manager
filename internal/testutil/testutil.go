package testutil

import (
	"io"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/database"
	"github.com/hhftechnology/middleware-manager/services"
)

// NewTempDB returns a SQLite database initialized with the project's migrations.
func NewTempDB(t *testing.T) *database.DB {
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

// MustExec is a helper that fails the test if the statement errors.
func MustExec(t *testing.T, db *database.DB, query string, args ...interface{}) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("exec failed: %v", err)
	}
}

// NewTestConfigManager creates a ConfigManager backed by a temp file.
func NewTestConfigManager(t *testing.T) *services.ConfigManager {
	t.Helper()

	cfgPath := filepath.Join(t.TempDir(), "config.json")
	cm, err := services.NewConfigManager(cfgPath)
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}
	return cm
}

// NewContext returns a Gin test context and recorder.
func NewContext(t *testing.T, method, path string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, path, body)
	return c, rec
}
