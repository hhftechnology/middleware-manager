package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hhftechnology/middleware-manager/models"
)

func setActiveDataSourceForConfigProxy(t *testing.T, cm *ConfigManager, name string, cfg models.DataSourceConfig) {
	t.Helper()

	if err := cm.UpdateDataSource(name, cfg); err != nil {
		t.Fatalf("UpdateDataSource(%s) failed: %v", name, err)
	}
	if err := cm.SetActiveDataSource(name); err != nil {
		t.Fatalf("SetActiveDataSource(%s) failed: %v", name, err)
	}
}

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

func TestConfigProxyUsesStandaloneConfigWhenTraefikIsActive(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	traefikServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/version" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"Version": "3.6.12",
		})
	}))
	defer traefikServer.Close()

	setActiveDataSourceForConfigProxy(t, cm, "traefik", models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  traefikServer.URL,
	})

	cp := NewConfigProxy(db, cm, "http://pangolin.invalid:3001/api/v1")
	config, err := cp.GetMergedConfig()
	if err != nil {
		t.Fatalf("GetMergedConfig() should not require Pangolin when Traefik is active: %v", err)
	}

	if config == nil {
		t.Fatal("GetMergedConfig() returned nil config")
	}
	if config.HTTP == nil {
		t.Fatal("config.HTTP is nil")
	}
	if config.HTTP.Middlewares == nil {
		t.Fatal("config.HTTP.Middlewares is nil")
	}
	if config.HTTP.Routers == nil {
		t.Fatal("config.HTTP.Routers is nil")
	}
	if config.HTTP.Services == nil {
		t.Fatal("config.HTTP.Services is nil")
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

// TestConfigProxyFindMatchingRouter covers findMatchingRouter's host-fallback
// matching, in particular the case that used to be non-deterministic: a host
// served by more than one non-redirect router, which happens with Pangolin's
// multi-target path routing (one resource, several Targets/paths, each
// producing its own router that shares an identical Host() but differs in
// PathPrefix()). See GH issue #113.
func TestConfigProxyFindMatchingRouter(t *testing.T) {
	cp := &ConfigProxy{}

	t.Run("single non-redirect router with redirect companion", func(t *testing.T) {
		routers := map[string]interface{}{
			"app": map[string]interface{}{
				"rule":        "Host(`single.example.com`)",
				"entryPoints": []interface{}{"websecure"},
			},
			"app-redirect": map[string]interface{}{
				"rule":        "Host(`single.example.com`)",
				"entryPoints": []interface{}{"web"},
			},
		}

		name, router := cp.findMatchingRouter(routers, "single.example.com")
		if name != "app" {
			t.Fatalf("expected the non-redirect router to be selected, got %q", name)
		}
		if router == nil {
			t.Fatal("expected a non-nil router")
		}
	})

	t.Run("no router matches the host", func(t *testing.T) {
		routers := map[string]interface{}{
			"other": map[string]interface{}{
				"rule":        "Host(`other.example.com`)",
				"entryPoints": []interface{}{"websecure"},
			},
		}

		name, router := cp.findMatchingRouter(routers, "missing.example.com")
		if name != "" || router != nil {
			t.Fatalf("expected no match, got name=%q router=%v", name, router)
		}
	})

	t.Run("ambiguous multi-target host returns no match, not a guess", func(t *testing.T) {
		routers := map[string]interface{}{
			"catch-all": map[string]interface{}{
				"rule":        "Host(`multi.example.com`) && PathPrefix(`/`)",
				"entryPoints": []interface{}{"websecure"},
				"priority":    10,
			},
			"specific": map[string]interface{}{
				"rule":        "Host(`multi.example.com`) && PathPrefix(`/app`)",
				"entryPoints": []interface{}{"websecure"},
				"priority":    200,
			},
		}

		// Run many times: with the pre-fix implementation this returned
		// whichever router Go's randomized map iteration visited first,
		// which flips from call to call. The fix must consistently report
		// "no match" instead of guessing, on every single call.
		for i := 0; i < 200; i++ {
			name, router := cp.findMatchingRouter(routers, "multi.example.com")
			if name != "" || router != nil {
				t.Fatalf("iteration %d: expected no match for an ambiguous multi-target host, got name=%q router=%v", i, name, router)
			}
		}
	})

	t.Run("only redirect routers match the host: deterministic fallback to first match", func(t *testing.T) {
		// Every router matching this host has the "-redirect" suffix, so
		// nonRedirect is empty (case 0) and the function falls back to
		// matches[0] in the deterministic (name-sorted) order established
		// before the redirect filter runs. Regression guard: this fallback
		// must stay deterministic across repeated calls despite `routers`
		// being a Go map with randomized iteration order.
		routers := map[string]interface{}{
			"zzz-redirect": map[string]interface{}{
				"rule":        "Host(`redirect-only.example.com`)",
				"entryPoints": []interface{}{"web"},
			},
			"aaa-redirect": map[string]interface{}{
				"rule":        "Host(`redirect-only.example.com`)",
				"entryPoints": []interface{}{"web"},
			},
		}

		for i := 0; i < 20; i++ {
			name, router := cp.findMatchingRouter(routers, "redirect-only.example.com")
			if name != "aaa-redirect" {
				t.Fatalf("iteration %d: expected deterministic fallback to %q, got %q", i, "aaa-redirect", name)
			}
			if router == nil {
				t.Fatalf("iteration %d: expected a non-nil router", i)
			}
		}
	})

	t.Run("ambiguous host including redirect companions still returns no match", func(t *testing.T) {
		routers := map[string]interface{}{
			"catch-all": map[string]interface{}{
				"rule":        "Host(`multi2.example.com`) && PathPrefix(`/`)",
				"entryPoints": []interface{}{"websecure"},
			},
			"catch-all-redirect": map[string]interface{}{
				"rule":        "Host(`multi2.example.com`) && PathPrefix(`/`)",
				"entryPoints": []interface{}{"web"},
			},
			"specific": map[string]interface{}{
				"rule":        "Host(`multi2.example.com`) && PathPrefix(`/app`)",
				"entryPoints": []interface{}{"websecure"},
			},
			"specific-redirect": map[string]interface{}{
				"rule":        "Host(`multi2.example.com`) && PathPrefix(`/app`)",
				"entryPoints": []interface{}{"web"},
			},
		}

		for i := 0; i < 50; i++ {
			name, router := cp.findMatchingRouter(routers, "multi2.example.com")
			if name != "" || router != nil {
				t.Fatalf("iteration %d: expected no match, got name=%q router=%v", i, name, router)
			}
		}
	})
}

// TestApplyResourceOverridesSkipsAmbiguousMultiTargetHost is the end-to-end
// regression test for the bug reported in GH issue #113: when the direct
// pangolin_router_id lookup misses (e.g. because Pangolin's router ID
// changed) and the host-based fallback is ambiguous (two non-redirect
// routers share the host), applyResourceOverrides must leave BOTH routers'
// priority untouched instead of corrupting one of them with the resource's
// RouterPriority.
func TestApplyResourceOverridesSkipsAmbiguousMultiTargetHost(t *testing.T) {
	cp := &ConfigProxy{}

	routers := map[string]interface{}{
		"catch-all": map[string]interface{}{
			"rule":        "Host(`multi.example.com`) && PathPrefix(`/`)",
			"entryPoints": []interface{}{"websecure"},
			"priority":    10,
		},
		"specific": map[string]interface{}{
			"rule":        "Host(`multi.example.com`) && PathPrefix(`/app`)",
			"entryPoints": []interface{}{"websecure"},
			"priority":    200,
		},
	}

	config := &ProxiedTraefikConfig{
		HTTP: &HTTPConfig{
			Routers:     routers,
			Middlewares: map[string]interface{}{},
		},
	}

	resources := []*resourceData{
		{
			ID: "resource-1",
			// Simulates "Pangolin changed the router ID": the stored ID no
			// longer exists as a router key, so findRouterByPangolinID
			// misses and the host-based fallback (findMatchingRouter) is
			// used - which is ambiguous for this host.
			PangolinRouterID: "stale-router-id-that-no-longer-exists",
			Host:             "multi.example.com",
			// Sentinel value: must never end up on either router below.
			RouterPriority: 999,
		},
	}

	for i := 0; i < 20; i++ {
		if err := cp.applyResourceOverrides(config, resources, nil, nil); err != nil {
			t.Fatalf("iteration %d: applyResourceOverrides() error = %v", i, err)
		}

		catchAll, ok := routers["catch-all"].(map[string]interface{})
		if !ok {
			t.Fatalf("iteration %d: catch-all router missing or wrong type", i)
		}
		specific, ok := routers["specific"].(map[string]interface{})
		if !ok {
			t.Fatalf("iteration %d: specific router missing or wrong type", i)
		}

		if catchAll["priority"] != 10 {
			t.Fatalf("iteration %d: catch-all router priority was mutated: got %v, want 10 (ambiguous host must not be overridden)", i, catchAll["priority"])
		}
		if specific["priority"] != 200 {
			t.Fatalf("iteration %d: specific router priority was mutated: got %v, want 200 (ambiguous host must not be overridden)", i, specific["priority"])
		}
		if _, ok := catchAll["middlewares"]; ok {
			t.Fatalf("iteration %d: catch-all router should not have received middleware overrides for an ambiguous host match", i)
		}
		if _, ok := specific["middlewares"]; ok {
			t.Fatalf("iteration %d: specific router should not have received middleware overrides for an ambiguous host match", i)
		}
	}
}
