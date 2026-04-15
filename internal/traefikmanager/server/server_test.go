package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	tmconfig "github.com/hhftechnology/middleware-manager/internal/traefikmanager/config"
	tmtypes "github.com/hhftechnology/middleware-manager/internal/traefikmanager/types"
)

func TestServerHealthAndExcludedAuthRoutes(t *testing.T) {
	uiDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(uiDir, "index.html"), []byte("<html>ok</html>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	cfg := tmtypes.RuntimeConfig{
		Port:             "0",
		UIPath:           uiDir,
		SettingsPath:     filepath.Join(t.TempDir(), "manager.yml"),
		BackupDir:        t.TempDir(),
		ConfigPath:       filepath.Join(t.TempDir(), "dynamic.yml"),
		TraefikAPIURL:    "http://traefik:8080",
		SettingsDir:      t.TempDir(),
		GroupsConfigFile: filepath.Join(t.TempDir(), "dashboard.yml"),
		GroupsCacheDir:   filepath.Join(t.TempDir(), "cache"),
	}
	files, err := tmconfig.NewFileStore(cfg)
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}
	settings := tmconfig.NewSettingsStore(cfg)
	dashboard := tmconfig.NewDashboardStore(cfg)
	srv := New(cfg, files, settings, dashboard, http.DefaultClient)

	healthRec := httptest.NewRecorder()
	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	srv.router.ServeHTTP(healthRec, healthReq)
	if healthRec.Code != http.StatusOK {
		t.Fatalf("expected /health 200, got %d", healthRec.Code)
	}

	authRec := httptest.NewRecorder()
	authReq := httptest.NewRequest(http.MethodGet, "/api/auth/toggle", nil)
	srv.router.ServeHTTP(authRec, authReq)
	if authRec.Code != http.StatusNotFound {
		t.Fatalf("expected /api/auth/toggle 404, got %d", authRec.Code)
	}
}
