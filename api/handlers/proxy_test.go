package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hhftechnology/middleware-manager/internal/testutil"
	"github.com/hhftechnology/middleware-manager/models"
	"github.com/hhftechnology/middleware-manager/services"
)

// mockConfigProxy creates a minimal config proxy for testing
func newTestConfigProxy(t *testing.T) *services.ConfigProxy {
	t.Helper()
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	return services.NewConfigProxy(db, cm, "")
}

func newTraefikActiveTestConfigProxy(t *testing.T) (*services.ConfigProxy, func()) {
	t.Helper()

	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)

	traefikServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/version" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"Version": "3.6.12",
		})
	}))
	if err := cm.UpdateDataSource("traefik", models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  traefikServer.URL,
	}); err != nil {
		t.Fatalf("UpdateDataSource(traefik) failed: %v", err)
	}
	if err := cm.SetActiveDataSource("traefik"); err != nil {
		t.Fatalf("SetActiveDataSource(traefik) failed: %v", err)
	}

	return services.NewConfigProxy(db, cm, ""), traefikServer.Close
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
	json.Unmarshal(rec.Body.Bytes(), &response)

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
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["status"] == nil {
		t.Error("expected status in response")
	}
	if response["message"] == nil {
		t.Error("expected message in response")
	}
}

func TestProxyHandler_GetProxyStatus_TraefikActiveHealthy(t *testing.T) {
	configProxy, cleanup := newTraefikActiveTestConfigProxy(t)
	defer cleanup()
	handler := NewProxyHandler(configProxy)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/traefik-config/status", nil)
	handler.GetProxyStatus(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if got := response["status"]; got != "healthy" {
		t.Fatalf("expected healthy status, got %#v", got)
	}
}

func TestProxyHandler_GetTraefikConfig_TraefikActive(t *testing.T) {
	configProxy, cleanup := newTraefikActiveTestConfigProxy(t)
	defer cleanup()
	handler := NewProxyHandler(configProxy)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/traefik-config", nil)
	handler.GetTraefikConfig(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["http"] == nil {
		t.Fatal("expected http section in response")
	}
}

// TestProxyHandler_GetTraefikConfig tests getting merged config
func TestProxyHandler_GetTraefikConfig(t *testing.T) {
	configProxy := newTestConfigProxy(t)
	handler := NewProxyHandler(configProxy)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/traefik-config", nil)
	handler.GetTraefikConfig(c)

	// This may fail if no data source is configured, which is expected
	// We're mainly testing that the handler doesn't panic
	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200 or 500, got %d", rec.Code)
	}
}
