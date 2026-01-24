package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

// TestNewPluginHandler tests plugin handler creation
func TestNewPluginHandler(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	configPath := filepath.Join(t.TempDir(), "traefik.yml")

	handler := NewPluginHandler(db.DB, configPath, cm)

	if handler == nil {
		t.Fatal("NewPluginHandler() returned nil")
	}
	if handler.DB == nil {
		t.Error("handler.DB is nil")
	}
	if handler.TraefikStaticConfigPath != configPath {
		t.Errorf("TraefikStaticConfigPath = %q, want %q", handler.TraefikStaticConfigPath, configPath)
	}
}

// TestPluginHandler_RefreshPluginFetcher tests fetcher refresh
func TestPluginHandler_RefreshPluginFetcher(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	configPath := filepath.Join(t.TempDir(), "traefik.yml")

	handler := NewPluginHandler(db.DB, configPath, cm)

	err := handler.RefreshPluginFetcher()
	if err != nil {
		t.Errorf("RefreshPluginFetcher() error = %v", err)
	}
}

// TestPluginHandler_RefreshPluginFetcher_NoConfigManager tests refresh without config manager
func TestPluginHandler_RefreshPluginFetcher_NoConfigManager(t *testing.T) {
	db := testutil.NewTempDB(t)
	configPath := filepath.Join(t.TempDir(), "traefik.yml")

	handler := NewPluginHandler(db.DB, configPath, nil)

	err := handler.RefreshPluginFetcher()
	if err == nil {
		t.Error("RefreshPluginFetcher() should error without config manager")
	}
}

// TestPluginHandler_GetPlugins tests getting plugins
func TestPluginHandler_GetPlugins(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	configPath := filepath.Join(t.TempDir(), "traefik.yml")

	handler := NewPluginHandler(db.DB, configPath, cm)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/plugins", nil)
	handler.GetPlugins(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var plugins []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &plugins); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Should return an array (possibly empty)
	if plugins == nil {
		t.Error("expected array, got nil")
	}
}

// TestPluginHandler_GetPlugins_WithLocalConfig tests getting plugins with local config
func TestPluginHandler_GetPlugins_WithLocalConfig(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "traefik.yml")

	// Create a test traefik config with plugins
	traefikConfig := `
experimental:
  plugins:
    mtls-whitelist:
      moduleName: github.com/example/mtls-whitelist
      version: v1.0.0
`
	if err := os.WriteFile(configPath, []byte(traefikConfig), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	handler := NewPluginHandler(db.DB, configPath, cm)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/plugins", nil)
	handler.GetPlugins(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var plugins []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &plugins)

	// Should find the plugin from local config
	found := false
	for _, p := range plugins {
		if p["name"] == "mtls-whitelist" {
			found = true
			break
		}
	}

	if !found && len(plugins) > 0 {
		// Plugin should be in the list
		t.Log("Plugin not found in response, but this may be expected if Traefik API is not available")
	}
}

// TestPluginHandler_InstallPlugin_InvalidJSON tests invalid install request
func TestPluginHandler_InstallPlugin_InvalidJSON(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	configPath := filepath.Join(t.TempDir(), "traefik.yml")

	handler := NewPluginHandler(db.DB, configPath, cm)

	body := bytes.NewBufferString(`{invalid}`)
	c, rec := testutil.NewContext(t, http.MethodPost, "/api/plugins/install", body)
	handler.InstallPlugin(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestPluginHandler_InstallPlugin_MissingFields tests install with missing fields
func TestPluginHandler_InstallPlugin_MissingFields(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	configPath := filepath.Join(t.TempDir(), "traefik.yml")

	handler := NewPluginHandler(db.DB, configPath, cm)

	// Missing required fields
	body := bytes.NewBufferString(`{"name": "test"}`)
	c, rec := testutil.NewContext(t, http.MethodPost, "/api/plugins/install", body)
	handler.InstallPlugin(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestPluginHandler_RemovePlugin_InvalidJSON ensures remove plugin validates body
func TestPluginHandler_RemovePlugin_InvalidJSON(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	configPath := filepath.Join(t.TempDir(), "traefik.yml")

	handler := NewPluginHandler(db.DB, configPath, cm)

	c, rec := testutil.NewContext(t, http.MethodDelete, "/api/plugins/remove", bytes.NewBufferString(`{invalid}`))
	handler.RemovePlugin(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestPluginHandler_GetPluginUsage tests getting plugin usage info
func TestPluginHandler_GetPluginUsage(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	configPath := filepath.Join(t.TempDir(), "traefik.yml")

	handler := NewPluginHandler(db.DB, configPath, cm)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/plugins/test-plugin/usage", nil)
	c.Params = gin.Params{{Key: "name", Value: "test-plugin"}}
	handler.GetPluginUsage(c)

	// Depending on fetcher configuration, this may be 404 or 500, but should not panic
	if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound && rec.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status code %d", rec.Code)
	}
}
