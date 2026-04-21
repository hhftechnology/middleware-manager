package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTraefikHandlerPingOK(t *testing.T) {
	env := newTestEnv(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/version" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer upstream.Close()

	settings := env.settings.Defaults()
	settings.TraefikAPIURL = upstream.URL
	if err := env.settings.Save(settings); err != nil {
		t.Fatalf("save settings: %v", err)
	}

	handler := NewTraefikHandler(env.settings, env.files, upstream.Client())
	router := gin.New()
	router.GET("/ping", handler.Ping)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("ping status %d", rec.Code)
	}
	body := map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if ok, _ := body["ok"].(bool); !ok {
		t.Fatalf("expected ok=true, got %v", body)
	}
}

func TestTraefikHandlerPingFallbacksToPingEndpoint(t *testing.T) {
	env := newTestEnv(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			w.WriteHeader(http.StatusNotFound)
		case "/ping":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer upstream.Close()

	settings := env.settings.Defaults()
	settings.TraefikAPIURL = upstream.URL
	if err := env.settings.Save(settings); err != nil {
		t.Fatalf("save settings: %v", err)
	}

	handler := NewTraefikHandler(env.settings, env.files, upstream.Client())
	router := gin.New()
	router.GET("/ping", handler.Ping)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("ping status %d", rec.Code)
	}
	body := map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if ok, _ := body["ok"].(bool); !ok {
		t.Fatalf("expected ok=true via /ping fallback, got %v", body)
	}
}

func TestTraefikHandlerPingDown(t *testing.T) {
	env := newTestEnv(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer upstream.Close()

	settings := env.settings.Defaults()
	settings.TraefikAPIURL = upstream.URL
	if err := env.settings.Save(settings); err != nil {
		t.Fatalf("save settings: %v", err)
	}

	handler := NewTraefikHandler(env.settings, env.files, upstream.Client())
	router := gin.New()
	router.GET("/ping", handler.Ping)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("ping status %d", rec.Code)
	}
	body := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if ok, _ := body["ok"].(bool); ok {
		t.Fatalf("expected ok=false, got %v", body)
	}
}

func TestTraefikHandlerOverviewProxiesUpstream(t *testing.T) {
	env := newTestEnv(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/overview" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"http":{"routers":{"total":3}}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer upstream.Close()

	settings := env.settings.Defaults()
	settings.TraefikAPIURL = upstream.URL
	_ = env.settings.Save(settings)

	handler := NewTraefikHandler(env.settings, env.files, upstream.Client())
	router := gin.New()
	router.GET("/overview", handler.Overview)

	req := httptest.NewRequest(http.MethodGet, "/overview", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("overview status %d", rec.Code)
	}
	payload := map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := payload["http"]; !ok {
		t.Fatalf("expected upstream body to be proxied, got %v", payload)
	}
}
