package api

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

func TestServerHealthAndDatasourceRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	traefikPath := filepath.Join(t.TempDir(), "traefik.yml")

	srv := NewServer(db, ServerConfig{
		Port:   "0",
		UIPath: "",
		Debug:  false,
	}, cm, traefikPath)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	srv.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected /health 200, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/datasource/active", nil)
	srv.router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected /api/datasource/active 200, got %d", rec2.Code)
	}
}
