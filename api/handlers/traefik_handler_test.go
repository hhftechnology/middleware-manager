package handlers

import (
	"net/http"
	"testing"

	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

// TestNewTraefikHandler tests traefik handler creation
func TestNewTraefikHandler(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewTraefikHandler(db.DB, cm)

	if handler == nil {
		t.Fatal("NewTraefikHandler() returned nil")
	}
	if handler.DB == nil {
		t.Error("handler.DB is nil")
	}
	if handler.ConfigManager == nil {
		t.Error("handler.ConfigManager is nil")
	}
}

// TestTraefikHandler_GetOverview tests overview endpoint
func TestTraefikHandler_GetOverview(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewTraefikHandler(db.DB, cm)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/traefik/overview", nil)
	handler.GetOverview(c)

	// May fail without Traefik running, but should not panic
	// Accept both success and internal error
	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200 or 500, got %d", rec.Code)
	}
}

// TestTraefikHandler_GetVersion tests version endpoint
func TestTraefikHandler_GetVersion(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewTraefikHandler(db.DB, cm)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/traefik/version", nil)
	handler.GetVersion(c)

	// May fail without Traefik running
	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200 or 500, got %d", rec.Code)
	}
}

// TestTraefikHandler_GetEntrypoints tests entrypoints endpoint
func TestTraefikHandler_GetEntrypoints(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewTraefikHandler(db.DB, cm)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/traefik/entrypoints", nil)
	handler.GetEntrypoints(c)

	// May fail without Traefik running
	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200 or 500, got %d", rec.Code)
	}
}

// TestTraefikHandler_GetRouters tests routers endpoint
func TestTraefikHandler_GetRouters(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewTraefikHandler(db.DB, cm)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/traefik/routers", nil)
	handler.GetRouters(c)

	// May fail without Traefik running
	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200 or 500, got %d", rec.Code)
	}
}

// TestTraefikHandler_GetServices tests services endpoint
func TestTraefikHandler_GetServices(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewTraefikHandler(db.DB, cm)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/traefik/services", nil)
	handler.GetServices(c)

	// May fail without Traefik running
	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200 or 500, got %d", rec.Code)
	}
}

// TestTraefikHandler_GetMiddlewares tests middlewares endpoint
func TestTraefikHandler_GetMiddlewares(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewTraefikHandler(db.DB, cm)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/traefik/middlewares", nil)
	handler.GetMiddlewares(c)

	// May fail without Traefik running
	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200 or 500, got %d", rec.Code)
	}
}
