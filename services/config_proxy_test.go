package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigProxyCachesAndInvalidates(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"http": map[string]interface{}{
				"middlewares": map[string]interface{}{},
				"routers":     map[string]interface{}{},
				"services":    map[string]interface{}{},
			},
		})
	}))
	defer server.Close()

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
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
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
