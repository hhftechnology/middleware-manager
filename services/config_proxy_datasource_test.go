package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hhftechnology/middleware-manager/models"
)

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
