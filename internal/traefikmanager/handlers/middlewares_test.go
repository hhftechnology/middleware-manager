package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func TestMiddlewareHandlerCreateListDelete(t *testing.T) {
	env := newTestEnv(t)
	handler := NewMiddlewareHandler(env.files)

	router := gin.New()
	router.POST("/middlewares", handler.Create)
	router.GET("/middlewares", handler.List)
	router.DELETE("/middlewares/:name", handler.Delete)

	body := map[string]any{
		"name":       "security-headers",
		"configFile": "",
		"yaml":       "headers:\n  customResponseHeaders:\n    X-Managed-By: traefik-manager\n",
	}
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/middlewares", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status %d: %s", rec.Code, rec.Body.String())
	}

	raw, err := os.ReadFile(env.cfg.ConfigPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	parsed := map[string]any{}
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse config: %v", err)
	}
	mws, _ := parsed["http"].(map[string]any)["middlewares"].(map[string]any)
	if _, ok := mws["security-headers"]; !ok {
		t.Fatalf("expected middleware persisted, got %v", parsed)
	}

	// List returns the entry
	req = httptest.NewRequest(http.MethodGet, "/middlewares", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status %d", rec.Code)
	}
	var list []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 1 || list[0]["name"] != "security-headers" {
		t.Fatalf("unexpected list output: %v", list)
	}

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/middlewares/security-headers", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status %d: %s", rec.Code, rec.Body.String())
	}

	raw, _ = os.ReadFile(env.cfg.ConfigPath)
	parsed = map[string]any{}
	_ = yaml.Unmarshal(raw, &parsed)
	httpSec, _ := parsed["http"].(map[string]any)
	if mws, ok := httpSec["middlewares"].(map[string]any); ok {
		if _, exists := mws["security-headers"]; exists {
			t.Fatalf("expected middleware removed, got %v", mws)
		}
	}
}

func TestMiddlewareHandlerCreateRejectsMissingName(t *testing.T) {
	env := newTestEnv(t)
	handler := NewMiddlewareHandler(env.files)
	router := gin.New()
	router.POST("/middlewares", handler.Create)

	buf, _ := json.Marshal(map[string]any{"yaml": "headers: {}"})
	req := httptest.NewRequest(http.MethodPost, "/middlewares", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d", rec.Code)
	}
}
