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

func TestRouteHandlerCreateHTTPRoute(t *testing.T) {
	env := newTestEnv(t)
	handler := NewRouteHandler(env.files, env.settings)

	router := gin.New()
	router.POST("/routes", handler.Create)

	body := map[string]any{
		"protocol":    "http",
		"serviceName": "whoami",
		"target":      "http://whoami:80",
		"subdomain":   "whoami",
		"domains":     []string{"example.com"},
		"entryPoints": []string{"websecure"},
	}
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/routes", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status %d: %s", rec.Code, rec.Body.String())
	}

	raw, _ := os.ReadFile(env.cfg.ConfigPath)
	parsed := map[string]any{}
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse: %v", err)
	}
	routers, _ := parsed["http"].(map[string]any)["routers"].(map[string]any)
	if _, ok := routers["whoami"]; !ok {
		t.Fatalf("expected router whoami, got %v", routers)
	}
	services, _ := parsed["http"].(map[string]any)["services"].(map[string]any)
	if _, ok := services["whoami-service"]; !ok {
		t.Fatalf("expected service whoami-service, got %v", services)
	}
}

func TestRouteHandlerListAndToggle(t *testing.T) {
	env := newTestEnv(t)
	handler := NewRouteHandler(env.files, env.settings)

	// Seed an HTTP route via direct config write.
	seed := map[string]any{
		"http": map[string]any{
			"routers": map[string]any{
				"api": map[string]any{
					"rule":        "Host(`api.example.com`)",
					"service":     "api-service",
					"entryPoints": []string{"websecure"},
				},
			},
			"services": map[string]any{
				"api-service": map[string]any{
					"loadBalancer": map[string]any{
						"servers": []map[string]any{{"url": "http://api:8080"}},
					},
				},
			},
		},
	}
	encoded, _ := yaml.Marshal(seed)
	if err := os.WriteFile(env.cfg.ConfigPath, encoded, 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	router := gin.New()
	router.GET("/routes", handler.List)
	router.PATCH("/routes/:id/toggle", handler.Toggle)

	// List
	req := httptest.NewRequest(http.MethodGet, "/routes", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status %d", rec.Code)
	}
	payload := map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	apps, _ := payload["apps"].([]any)
	if len(apps) != 1 {
		t.Fatalf("expected 1 app, got %v", apps)
	}
	id, _ := apps[0].(map[string]any)["id"].(string)
	if id == "" {
		t.Fatalf("missing route id in %v", apps[0])
	}

	// Toggle disable
	buf, _ := json.Marshal(map[string]any{"enable": false})
	req = httptest.NewRequest(http.MethodPatch, "/routes/"+id+"/toggle", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("toggle status %d: %s", rec.Code, rec.Body.String())
	}

	// After disabling, router should be removed from config; disabled_routes persisted in settings.
	raw, _ := os.ReadFile(env.cfg.ConfigPath)
	current := map[string]any{}
	_ = yaml.Unmarshal(raw, &current)
	if httpSec, ok := current["http"].(map[string]any); ok {
		if routers, ok := httpSec["routers"].(map[string]any); ok {
			if _, exists := routers["api"]; exists {
				t.Fatalf("expected api router to be removed when disabled, got %v", routers)
			}
		}
	}
	settings, _, err := env.settings.Load()
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if _, ok := settings.DisabledRoutes[id]; !ok {
		t.Fatalf("expected %s in disabled_routes, got %v", id, settings.DisabledRoutes)
	}
}

func TestRouteHandlerDeleteMissingRouteReturns404(t *testing.T) {
	env := newTestEnv(t)
	handler := NewRouteHandler(env.files, env.settings)

	router := gin.New()
	router.DELETE("/routes/:id", handler.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/routes/ghost", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing route, got %d: %s", rec.Code, rec.Body.String())
	}
}
