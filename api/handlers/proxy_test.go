package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hhftechnology/middleware-manager/internal/testutil"
	"github.com/hhftechnology/middleware-manager/models"
	"github.com/hhftechnology/middleware-manager/services"
)

// modelDataSource builds a DataSourceConfig for tests.
func modelDataSource(dsType, url string) models.DataSourceConfig {
	return models.DataSourceConfig{
		Type: models.DataSourceType(dsType),
		URL:  url,
	}
}

// mockConfigProxy creates a minimal config proxy for testing
func newTestConfigProxy(t *testing.T) *services.ConfigProxy {
	t.Helper()
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	return services.NewConfigProxy(db, cm, "")
}

// TestNewProxyHandler tests proxy handler creation
func TestNewProxyHandler(t *testing.T) {
	configProxy := newTestConfigProxy(t)
	handler := NewProxyHandler(configProxy)

	if handler == nil {
		t.Fatal("NewProxyHandler() returned nil")
	}
	if handler.ConfigProxy == nil {
		t.Error("handler.ConfigProxy is nil")
	}
}

// TestProxyHandler_InvalidateCache tests cache invalidation
func TestProxyHandler_InvalidateCache(t *testing.T) {
	configProxy := newTestConfigProxy(t)
	handler := NewProxyHandler(configProxy)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/traefik-config/invalidate", nil)
	handler.InvalidateCache(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	mustUnmarshalResponse(t, rec.Body.Bytes(), &response)

	if response["message"] == nil {
		t.Error("expected message in response")
	}
}

// TestProxyHandler_GetProxyStatus tests proxy status endpoint
func TestProxyHandler_GetProxyStatus(t *testing.T) {
	configProxy := newTestConfigProxy(t)
	handler := NewProxyHandler(configProxy)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/traefik-config/status", nil)
	handler.GetProxyStatus(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	mustUnmarshalResponse(t, rec.Body.Bytes(), &response)

	if response["status"] == nil {
		t.Error("expected status in response")
	}
	if response["message"] == nil {
		t.Error("expected message in response")
	}
}

// TestProxyHandler_GetTraefikConfig tests getting merged config.
// When no active data source is configured the handler returns 500; this is expected.
func TestProxyHandler_GetTraefikConfig(t *testing.T) {
	configProxy := newTestConfigProxy(t)
	handler := NewProxyHandler(configProxy)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/traefik-config", nil)
	handler.GetTraefikConfig(c)

	// 500 is expected when no active data source is configured.
	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200 or 500, got %d", rec.Code)
	}
}

// TestProxyHandler_GetTraefikConfig_TraefikSource_OK verifies that GetTraefikConfig
// returns 200 with a valid JSON body when the active source is traefik.
func TestProxyHandler_GetTraefikConfig_TraefikSource_OK(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)

	// Register and activate the traefik source (no network call expected).
	if err := cm.UpdateDataSource("traefik", modelDataSource("traefik", "")); err != nil {
		t.Fatalf("UpdateDataSource: %v", err)
	}
	if err := cm.SetActiveDataSource("traefik"); err != nil {
		t.Fatalf("SetActiveDataSource: %v", err)
	}

	configProxy := services.NewConfigProxy(db, cm, "")
	handler := NewProxyHandler(configProxy)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/traefik-config", nil)
	handler.GetTraefikConfig(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Body must be valid JSON.
	var body interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not valid JSON: %v\nbody: %s", err, rec.Body.String())
	}
}
