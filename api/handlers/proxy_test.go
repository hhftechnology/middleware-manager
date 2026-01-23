package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hhftechnology/middleware-manager/internal/testutil"
	"github.com/hhftechnology/middleware-manager/services"
)

// mockConfigProxy creates a minimal config proxy for testing
func newTestConfigProxy(t *testing.T) *services.ConfigProxy {
	t.Helper()
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	return services.NewConfigProxy(db.DB, cm)
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
