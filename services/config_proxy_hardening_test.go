package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPruneEmptySectionsRemovesBlankProtocols(t *testing.T) {
	cp := &ConfigProxy{}
	config := &ProxiedTraefikConfig{}
	cp.initializeConfigMaps(config)
	config.HTTP.Routers["app"] = map[string]interface{}{"rule": "Host(`app.example.com`)"}

	cp.pruneEmptySections(config)

	if config.HTTP == nil || len(config.HTTP.Routers) != 1 {
		t.Fatal("HTTP routers should be kept")
	}
	if config.TCP != nil {
		t.Fatal("empty TCP section should be removed")
	}
	if config.UDP != nil {
		t.Fatal("empty UDP section should be removed")
	}
	if config.TLS != nil {
		t.Fatal("empty TLS section should be removed")
	}
}

func TestFindMatchingRouterSingleCandidate(t *testing.T) {
	cp := &ConfigProxy{}
	routers := map[string]interface{}{
		"app-redirect": map[string]interface{}{
			"rule": "Host(`app.example.com`)",
		},
		"app": map[string]interface{}{
			"rule":     "Host(`app.example.com`)",
			"priority": 100,
		},
		"other": map[string]interface{}{
			"rule": "Host(`other.example.com`)",
		},
	}

	for i := 0; i < 20; i++ {
		name, router := cp.findMatchingRouter(routers, "app.example.com")
		if name != "app" || router == nil {
			t.Fatalf("iteration %d: expected app, got %q", i, name)
		}
	}

	if name, router := cp.findMatchingRouter(routers, "missing.example.com"); name != "" || router != nil {
		t.Fatalf("expected no match, got %q", name)
	}
}

func TestMapToOrderedTLSKeepsStringDomains(t *testing.T) {
	cp := &ConfigProxy{}
	ordered := cp.mapToOrderedTLS(map[string]interface{}{
		"certResolver": "letsencrypt",
		"domains":      []interface{}{"example.com"},
		"options":      "tls-hardened",
	})
	if ordered.CertResolver != "letsencrypt" || ordered.Options != "tls-hardened" {
		t.Fatalf("unexpected tls metadata: %+v", ordered)
	}
	if len(ordered.Domains) != 1 || ordered.Domains[0] != "example.com" {
		t.Fatalf("string domain was not preserved: %#v", ordered.Domains)
	}
}

func TestGetMergedConfigPangolinDownWithoutCache(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusBadGateway)
	}))
	defer server.Close()

	cp := NewConfigProxy(db, cm, server.URL)
	cp.httpClient = server.Client()
	if _, err := cp.GetMergedConfig(); err == nil {
		t.Fatal("expected an error when Pangolin is down and the cache is empty")
	}
}
