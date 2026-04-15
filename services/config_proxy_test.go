package services

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hhftechnology/middleware-manager/models"
)

func TestConfigProxyCachesAndInvalidates(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, map[string]interface{}{
			"http": map[string]interface{}{
				"middlewares": map[string]interface{}{},
				"routers":     map[string]interface{}{},
				"services":    map[string]interface{}{},
			},
		})
	}))
	defer server.Close()

	// Set active data source so GetMergedConfig can resolve it.
	// NOTE: UpdateDataSource makes a connection-test HTTP call, so reset
	// the counter afterward to measure only GetMergedConfig traffic.
	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")
	hits = 0 // reset: connection-test hit during UpdateDataSource is not under test here

	cp := NewConfigProxy(db, cm, server.URL)
	cp.httpClient = server.Client()

	if _, err := cp.GetMergedConfig(); err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}
	if hits != 1 {
		t.Fatalf("expected first call to hit server once, got %d", hits)
	}

	// Cached path should not hit server again.
	if _, err := cp.GetMergedConfig(); err != nil {
		t.Fatalf("cached fetch failed: %v", err)
	}
	if hits != 1 {
		t.Fatalf("expected cached call to avoid server, hits=%d", hits)
	}

	cp.InvalidateCache()
	if _, err := cp.GetMergedConfig(); err != nil {
		t.Fatalf("post-invalidate fetch failed: %v", err)
	}
	if hits != 2 {
		t.Fatalf("expected cache invalidation to refetch, hits=%d", hits)
	}
}

func TestConfigProxyPreservesServersTransports(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, map[string]interface{}{
			"http": map[string]interface{}{
				"middlewares": map[string]interface{}{},
				"routers":     map[string]interface{}{},
				"services": map[string]interface{}{
					"14-example-service": map[string]interface{}{
						"loadBalancer": map[string]interface{}{
							"servers": []map[string]interface{}{
								{"url": "https://10.0.0.1:12345"},
							},
							"serversTransport": "14-transport",
						},
					},
				},
				"serversTransports": map[string]interface{}{
					"14-transport": map[string]interface{}{
						"serverName":         "example.com",
						"insecureSkipVerify": true,
					},
				},
			},
		})
	}))
	defer server.Close()

	// Set active data source so GetMergedConfig can resolve it.
	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	cp := NewConfigProxy(db, cm, server.URL)
	cp.httpClient = server.Client()

	config, err := cp.GetMergedConfig()
	if err != nil {
		t.Fatalf("GetMergedConfig() error = %v", err)
	}

	if config.HTTP == nil {
		t.Fatal("config.HTTP is nil")
	}

	transportRaw, exists := config.HTTP.ServersTransports["14-transport"]
	if !exists {
		t.Fatalf("serversTransport %q was not preserved", "14-transport")
	}

	transport, ok := transportRaw.(map[string]interface{})
	if !ok {
		t.Fatalf("serversTransport has type %T, want map[string]interface{}", transportRaw)
	}

	if got, ok := transport["serverName"].(string); !ok || got != "example.com" {
		t.Fatalf("serverName = %#v, want %q", transport["serverName"], "example.com")
	}
	if got, ok := transport["insecureSkipVerify"].(bool); !ok || !got {
		t.Fatalf("insecureSkipVerify = %#v, want true", transport["insecureSkipVerify"])
	}

	serviceRaw, exists := config.HTTP.Services["14-example-service"]
	if !exists {
		t.Fatalf("service %q not found", "14-example-service")
	}
	service, ok := serviceRaw.(map[string]interface{})
	if !ok {
		t.Fatalf("service has type %T, want map[string]interface{}", serviceRaw)
	}

	loadBalancerRaw, exists := service["loadBalancer"]
	if !exists {
		t.Fatalf("service %q missing loadBalancer", "14-example-service")
	}
	loadBalancer, ok := loadBalancerRaw.(map[string]interface{})
	if !ok {
		t.Fatalf("loadBalancer has type %T, want map[string]interface{}", loadBalancerRaw)
	}

	if got, ok := loadBalancer["serversTransport"].(string); !ok || got != "14-transport" {
		t.Fatalf("loadBalancer.serversTransport = %#v, want %q", loadBalancer["serversTransport"], "14-transport")
	}
}

func TestConfigGeneratorWritesConfigFile(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)
	confDir := t.TempDir()

	generator := NewConfigGenerator(db, confDir, cm)
	if err := generator.generateConfig(); err != nil {
		t.Fatalf("generateConfig failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(confDir, "resource-overrides.yml")); err != nil {
		t.Fatalf("expected config file to be written: %v", err)
	}

	// Subsequent generation with no changes should be a no-op but still succeed.
	time.Sleep(10 * time.Millisecond)
	if err := generator.generateConfig(); err != nil {
		t.Fatalf("second generateConfig failed: %v", err)
	}
}

// TestConfigProxy_GetMergedConfig_TraefikStandalone_NoPangolinCall verifies that
// when the active source is traefik, GetMergedConfig makes zero HTTP calls to Pangolin.
// The test server tracks requests; we pass a distinct URL as the pangolinURL so any
// accidental Pangolin fetch would hit the counter. The active-source URL points
// elsewhere so connection health probes cannot inflate the count.
func TestConfigProxy_GetMergedConfig_TraefikStandalone_NoPangolinCall(t *testing.T) {
	var callCount atomic.Int64
	// pangolinSpy counts any HTTP request routed through the ConfigProxy's pangolinURL.
	pangolinSpy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusOK)
		writeResponseBody(w, "{}")
	}))
	defer pangolinSpy.Close()

	db := newTestDB(t)
	cm := newTestConfigManager(t)

	// Active source is traefik — the URL here is the data source URL, not used for config fetch.
	setActiveDataSource(t, cm, "traefik", "http://traefik-api:8080", "", "")

	// Pass pangolinSpy.URL as the pangolinURL so any accidental fetchPangolinConfig hits are counted.
	cp := NewConfigProxy(db, cm, pangolinSpy.URL)
	config, err := cp.GetMergedConfig()
	if err != nil {
		t.Fatalf("GetMergedConfig() error = %v", err)
	}
	if config == nil {
		t.Fatal("GetMergedConfig() returned nil config")
	}
	if got := callCount.Load(); got != 0 {
		t.Errorf("expected 0 Pangolin calls, got %d", got)
	}
}

// TestConfigProxy_GetMergedConfig_PangolinPath_StillWorks verifies that the Pangolin
// fetch path continues to work after the branching change.
func TestConfigProxy_GetMergedConfig_PangolinPath_StillWorks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponse(w, map[string]interface{}{
			"http": map[string]interface{}{
				"routers":     map[string]interface{}{},
				"services":    map[string]interface{}{},
				"middlewares": map[string]interface{}{},
			},
		})
	}))
	defer server.Close()

	db := newTestDB(t)
	cm := newTestConfigManager(t)

	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	cp := NewConfigProxy(db, cm, server.URL)
	config, err := cp.GetMergedConfig()
	if err != nil {
		t.Fatalf("GetMergedConfig() error = %v", err)
	}
	if config == nil {
		t.Fatal("GetMergedConfig() returned nil config")
	}
}

// TestConfigProxy_GetMergedConfig_UnknownSource_Errors verifies that an unsupported
// data source type yields an explicit error mentioning "unsupported".
func TestConfigProxy_GetMergedConfig_UnknownSource_Errors(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	bogus := models.DataSourceConfig{
		Type: models.DataSourceType("bogus-source"),
		URL:  "http://nowhere",
	}
	if err := cm.UpdateDataSource("bogus-source", bogus); err != nil {
		t.Fatalf("UpdateDataSource: %v", err)
	}
	if err := cm.SetActiveDataSource("bogus-source"); err != nil {
		t.Fatalf("SetActiveDataSource: %v", err)
	}

	cp := NewConfigProxy(db, cm, "http://nowhere")
	_, err := cp.GetMergedConfig()
	if err == nil {
		t.Fatal("expected error for unknown data source type, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected error to mention 'unsupported', got: %v", err)
	}
}

// TestConfigProxy_BuildStandaloneTraefikConfig_ReturnsEmptyByDefault verifies that
// buildStandaloneTraefikConfig on an empty DB returns without error.
func TestConfigProxy_BuildStandaloneTraefikConfig_ReturnsEmptyByDefault(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)
	cp := NewConfigProxy(db, cm, "")

	config, err := cp.buildStandaloneTraefikConfig()
	if err != nil {
		t.Fatalf("buildStandaloneTraefikConfig() error = %v", err)
	}
	if config == nil {
		t.Fatal("buildStandaloneTraefikConfig() returned nil")
	}
}
