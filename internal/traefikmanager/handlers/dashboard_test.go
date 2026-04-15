package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	tmconfig "github.com/hhftechnology/middleware-manager/internal/traefikmanager/config"
)

func TestDashboardHandlerGetReturnsEmptyDefault(t *testing.T) {
	env := newTestEnv(t)
	store := tmconfig.NewDashboardStore(env.cfg)
	handler := NewDashboardHandler(store, &http.Client{})

	router := gin.New()
	router.GET("/dashboard", handler.GetConfig)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	payload := map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["custom_groups"] == nil {
		t.Fatalf("expected custom_groups key, got %v", payload)
	}
}

func TestDashboardHandlerSaveThenGet(t *testing.T) {
	env := newTestEnv(t)
	store := tmconfig.NewDashboardStore(env.cfg)
	handler := NewDashboardHandler(store, &http.Client{})

	router := gin.New()
	router.POST("/dashboard", handler.SaveConfig)
	router.GET("/dashboard", handler.GetConfig)

	body := map[string]any{
		"custom_groups":   []map[string]any{{"name": "Prod", "routers": []string{"api@file"}}},
		"route_overrides": map[string]any{"api@file": map[string]any{"icon": "nginx"}},
	}
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/dashboard", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status %d: %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get status %d", rec.Code)
	}
	payload := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &payload)
	groups, _ := payload["custom_groups"].([]any)
	if len(groups) != 1 {
		t.Fatalf("expected 1 custom group, got %v", groups)
	}
}

func TestDashboardHandlerIconServesCachedFile(t *testing.T) {
	env := newTestEnv(t)
	store := tmconfig.NewDashboardStore(env.cfg)
	cacheDir, err := store.EnsureCacheDir()
	if err != nil {
		t.Fatalf("ensure cache: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "nginx.png"), []byte("png-bytes"), 0o644); err != nil {
		t.Fatalf("write cache: %v", err)
	}
	handler := NewDashboardHandler(store, &http.Client{})

	router := gin.New()
	router.GET("/icon/:slug", handler.Icon)

	req := httptest.NewRequest(http.MethodGet, "/icon/nginx", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("icon status %d", rec.Code)
	}
	if rec.Body.String() != "png-bytes" {
		t.Fatalf("expected cached bytes, got %q", rec.Body.String())
	}
}

func TestDashboardHandlerIconRespectsMissMarker(t *testing.T) {
	env := newTestEnv(t)
	store := tmconfig.NewDashboardStore(env.cfg)
	cacheDir, _ := store.EnsureCacheDir()
	_ = os.WriteFile(filepath.Join(cacheDir, "unknown.404"), []byte(""), 0o644)
	handler := NewDashboardHandler(store, &http.Client{})

	router := gin.New()
	router.GET("/icon/:slug", handler.Icon)

	req := httptest.NewRequest(http.MethodGet, "/icon/unknown", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for miss marker, got %d", rec.Code)
	}
}
