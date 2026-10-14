package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hhftechnology/middleware-manager/models"
)

func TestGetMergedConfigStandaloneTraefikSkipsPangolin(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)
	if err := cm.SetActiveDataSource("traefik"); err != nil {
		t.Fatalf("set active source: %v", err)
	}

	_, err := db.Exec(
		`INSERT INTO middlewares (id, name, type, config) VALUES (?, ?, ?, ?)`,
		"rate-limit", "Rate Limit", "rateLimit", `{"average":10,"burst":5}`,
	)
	if err != nil {
		t.Fatalf("insert middleware: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO resources (id, host, service_id, org_id, site_id, status, entrypoints, source_type)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"res-1", "app.example.com", "app", "org", "site", "active", "websecure", "traefik",
	)
	if err != nil {
		t.Fatalf("insert resource: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO resource_middlewares (resource_id, middleware_id, priority) VALUES (?, ?, ?)`,
		"res-1", "rate-limit", 100,
	)
	if err != nil {
		t.Fatalf("assign middleware: %v", err)
	}

	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		http.Error(w, "pangolin should not be called", http.StatusInternalServerError)
	}))
	defer server.Close()

	cp := NewConfigProxy(db, cm, server.URL)
	cp.httpClient = server.Client()

	cfg, err := cp.GetMergedConfig()
	if err != nil {
		t.Fatalf("GetMergedConfig: %v", err)
	}
	if hits != 0 {
		t.Fatalf("standalone mode contacted Pangolin %d times", hits)
	}
	if cfg == nil || cfg.HTTP == nil {
		t.Fatal("expected an HTTP config")
	}
	if _, ok := cfg.HTTP.Middlewares["Rate Limit"]; !ok {
		t.Fatalf("expected middleware keyed by name, got %#v", cfg.HTTP.Middlewares)
	}
	if len(cfg.HTTP.Routers) == 0 {
		t.Fatal("expected a router built from the local resource")
	}
}

func TestGetMergedConfigPangolinModeStillFetches(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)
	active, err := cm.GetActiveDataSourceConfig()
	if err != nil {
		t.Fatal(err)
	}
	if active.Type != models.PangolinAPI {
		t.Fatalf("expected default source pangolin, got %s", active.Type)
	}

	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"http": map[string]interface{}{
				"routers":     map[string]interface{}{},
				"services":    map[string]interface{}{},
				"middlewares": map[string]interface{}{},
			},
		})
	}))
	defer server.Close()

	cp := NewConfigProxy(db, cm, server.URL)
	cp.httpClient = server.Client()
	if _, err := cp.GetMergedConfig(); err != nil {
		t.Fatalf("GetMergedConfig: %v", err)
	}
	if hits != 1 {
		t.Fatalf("expected one Pangolin fetch, got %d", hits)
	}
}

func TestGetMergedConfigPangolinFailureWithoutCache(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	defer server.Close()

	cp := NewConfigProxy(db, cm, server.URL)
	cp.httpClient = server.Client()
	if _, err := cp.GetMergedConfig(); err == nil {
		t.Fatal("expected error when Pangolin fails and there is no cache")
	}
}
