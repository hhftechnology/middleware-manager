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
