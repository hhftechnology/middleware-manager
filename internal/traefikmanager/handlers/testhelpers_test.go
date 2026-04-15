package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	tmconfig "github.com/hhftechnology/middleware-manager/internal/traefikmanager/config"
	tmtypes "github.com/hhftechnology/middleware-manager/internal/traefikmanager/types"
)

type testEnv struct {
	cfg      tmtypes.RuntimeConfig
	files    *tmconfig.FileStore
	settings *tmconfig.SettingsStore
	dir      string
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "dynamic.yml")
	if err := os.WriteFile(configPath, []byte("http: {}\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	backupDir := filepath.Join(dir, "backups")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatalf("mkdir backup: %v", err)
	}
	cfg := tmtypes.RuntimeConfig{
		ConfigPath:       configPath,
		BackupDir:        backupDir,
		SettingsPath:     filepath.Join(dir, "manager.yml"),
		SettingsDir:      dir,
		GroupsConfigFile: filepath.Join(dir, "dashboard.yml"),
		GroupsCacheDir:   filepath.Join(dir, "cache"),
		TraefikAPIURL:    "http://traefik.local",
		AcmeJSONPath:     filepath.Join(dir, "acme.json"),
		AccessLogPath:    filepath.Join(dir, "access.log"),
		StaticConfigPath: filepath.Join(dir, "traefik.yml"),
	}
	files, err := tmconfig.NewFileStore(cfg)
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}
	return &testEnv{cfg: cfg, files: files, settings: tmconfig.NewSettingsStore(cfg), dir: dir}
}
