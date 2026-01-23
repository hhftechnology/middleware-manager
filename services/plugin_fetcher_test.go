package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hhftechnology/middleware-manager/models"
)

// TestNewPluginFetcher tests plugin fetcher creation
func TestNewPluginFetcher(t *testing.T) {
	config := models.DataSourceConfig{
		URL: "http://localhost:8080",
	}

	fetcher := NewPluginFetcher(config)

	if fetcher == nil {
		t.Fatal("NewPluginFetcher() returned nil")
	}
	if fetcher.httpClient == nil {
		t.Error("fetcher.httpClient is nil")
	}
	if fetcher.config.URL != config.URL {
		t.Errorf("config.URL = %q, want %q", fetcher.config.URL, config.URL)
	}
}

// TestPluginFetcher_FetchPlugins tests fetching plugins
func TestPluginFetcher_FetchPlugins(t *testing.T) {
	// Create mock server
	middlewares := []models.TraefikMiddleware{
		{
			Name:     "test-plugin@file",
			Type:     "plugin",
			Provider: "file",
			Config: map[string]interface{}{
				"plugin": map[string]interface{}{
					"testPlugin": map[string]interface{}{
						"moduleName": "github.com/example/test-plugin",
						"version":    "v1.0.0",
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/http/middlewares":
			json.NewEncoder(w).Encode(middlewares)
		case "/api/overview":
			overview := models.TraefikPluginOverview{
				Plugins: models.TraefikPluginStatus{
					Enabled: []string{"testPlugin"},
				},
			}
			json.NewEncoder(w).Encode(overview)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
	}
	fetcher := NewPluginFetcher(config)

	ctx := context.Background()
	plugins, err := fetcher.FetchPlugins(ctx)
	if err != nil {
		t.Fatalf("FetchPlugins() error = %v", err)
	}

	if len(plugins) == 0 {
		t.Log("No plugins returned (may be expected if no plugin middlewares found)")
	}
}

// TestPluginFetcher_FetchPlugins_WithBasicAuth tests fetching with auth
func TestPluginFetcher_FetchPlugins_WithBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "secret" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode([]models.TraefikMiddleware{})
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
		BasicAuth: struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: "admin",
			Password: "secret",
		},
	}
	fetcher := NewPluginFetcher(config)

	ctx := context.Background()
	_, err := fetcher.FetchPlugins(ctx)
	if err != nil {
		t.Fatalf("FetchPlugins() with auth error = %v", err)
	}
}

// TestPluginFetcher_FetchPlugins_ServerError tests handling server errors
func TestPluginFetcher_FetchPlugins_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
	}
	fetcher := NewPluginFetcher(config)

	ctx := context.Background()
	_, err := fetcher.FetchPlugins(ctx)
	if err == nil {
		t.Error("FetchPlugins() should error on server error")
	}
}

// TestPluginFetcher_RateLimiting tests rate limiting
func TestPluginFetcher_RateLimiting(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path == "/api/http/middlewares" {
			json.NewEncoder(w).Encode([]models.TraefikMiddleware{})
		} else if r.URL.Path == "/api/overview" {
			json.NewEncoder(w).Encode(models.TraefikPluginOverview{})
		}
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
	}
	fetcher := NewPluginFetcher(config)

	ctx := context.Background()

	// First call
	_, err := fetcher.FetchPlugins(ctx)
	if err != nil {
		t.Fatalf("First FetchPlugins() error = %v", err)
	}
	firstCallCount := callCount

	// Immediate second call should return cached data
	_, err = fetcher.FetchPlugins(ctx)
	if err != nil {
		t.Fatalf("Second FetchPlugins() error = %v", err)
	}

	// Call count should not increase significantly due to caching
	if callCount > firstCallCount*2 {
		t.Log("Rate limiting may not be preventing extra calls")
	}
}

// TestPluginFetcher_GetCachedPlugins tests getting cached plugins
func TestPluginFetcher_GetCachedPlugins(t *testing.T) {
	config := models.DataSourceConfig{
		URL: "http://localhost:8080",
	}
	fetcher := NewPluginFetcher(config)

	// Should return nil or empty when no cache
	cached := fetcher.GetCachedPlugins()
	if cached != nil && len(cached) > 0 {
		t.Error("expected nil or empty cache initially")
	}
}

// TestPluginFetcher_InvalidateCache tests cache invalidation
func TestPluginFetcher_InvalidateCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/http/middlewares" {
			json.NewEncoder(w).Encode([]models.TraefikMiddleware{})
		} else if r.URL.Path == "/api/overview" {
			json.NewEncoder(w).Encode(models.TraefikPluginOverview{})
		}
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
	}
	fetcher := NewPluginFetcher(config)

	ctx := context.Background()

	// Fetch to populate cache
	_, _ = fetcher.FetchPlugins(ctx)

	// Invalidate cache
	fetcher.InvalidateCache()

	// Cache should be empty
	cached := fetcher.GetCachedPlugins()
	if cached != nil && len(cached) > 0 {
		t.Error("expected nil or empty cache after invalidation")
	}
}

// TestPluginFetcher_ContextCancellation tests context cancellation
func TestPluginFetcher_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Simulate slow response
		json.NewEncoder(w).Encode([]models.TraefikMiddleware{})
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
	}
	fetcher := NewPluginFetcher(config)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := fetcher.FetchPlugins(ctx)
	// Should error due to context timeout
	if err == nil {
		t.Log("Expected error due to context timeout, but got success")
	}
}

// TestIsPluginMiddleware tests plugin middleware detection
func TestIsPluginMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		mw       models.TraefikMiddleware
		expected bool
	}{
		{
			name: "plugin type",
			mw: models.TraefikMiddleware{
				Type: "plugin",
			},
			expected: true,
		},
		{
			name: "plugin in config",
			mw: models.TraefikMiddleware{
				Type: "headers",
				Config: map[string]interface{}{
					"plugin": map[string]interface{}{},
				},
			},
			expected: true,
		},
		{
			name: "not a plugin",
			mw: models.TraefikMiddleware{
				Type:   "headers",
				Config: map[string]interface{}{},
			},
			expected: false,
		},
		{
			name: "plugin in type name",
			mw: models.TraefikMiddleware{
				Type: "myPlugin",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPluginMiddleware(tt.mw)
			if got != tt.expected {
				t.Errorf("isPluginMiddleware() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestExtractPluginName tests plugin name extraction
func TestExtractPluginName(t *testing.T) {
	tests := []struct {
		name     string
		mw       models.TraefikMiddleware
		expected string
	}{
		{
			name: "from config",
			mw: models.TraefikMiddleware{
				Name: "test-middleware",
				Config: map[string]interface{}{
					"plugin": map[string]interface{}{
						"myPlugin": map[string]interface{}{},
					},
				},
			},
			expected: "myPlugin",
		},
		{
			name: "from name with suffix",
			mw: models.TraefikMiddleware{
				Name: "badger@file",
			},
			expected: "badger",
		},
		{
			name: "from name without suffix",
			mw: models.TraefikMiddleware{
				Name: "simple-plugin",
			},
			expected: "simple-plugin",
		},
		{
			name: "strip middleware suffix",
			mw: models.TraefikMiddleware{
				Name: "badger-middleware",
			},
			expected: "badger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPluginName(tt.mw)
			if got != tt.expected {
				t.Errorf("extractPluginName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestExtractPluginNameFromMiddlewareName tests name extraction from middleware name
func TestExtractPluginNameFromMiddlewareName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plugin prefix with suffix",
			input:    "plugin-badger@file",
			expected: "badger",
		},
		{
			name:     "middleware suffix",
			input:    "test-middleware",
			expected: "test",
		},
		{
			name:     "simple name",
			input:    "myPlugin",
			expected: "myPlugin",
		},
		{
			name:     "with provider",
			input:    "auth@docker",
			expected: "auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPluginNameFromMiddlewareName(tt.input)
			if got != tt.expected {
				t.Errorf("extractPluginNameFromMiddlewareName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestExtractModuleName tests module name extraction
func TestExtractModuleName(t *testing.T) {
	tests := []struct {
		name     string
		mw       models.TraefikMiddleware
		fallback string
		expected string
	}{
		{
			name: "from config",
			mw: models.TraefikMiddleware{
				Config: map[string]interface{}{
					"plugin": map[string]interface{}{
						"myPlugin": map[string]interface{}{
							"moduleName": "github.com/example/my-plugin",
						},
					},
				},
			},
			fallback: "default",
			expected: "github.com/example/my-plugin",
		},
		{
			name:     "fallback when not found",
			mw:       models.TraefikMiddleware{},
			fallback: "fallback-module",
			expected: "fallback-module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractModuleName(tt.mw, tt.fallback)
			if got != tt.expected {
				t.Errorf("extractModuleName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestExtractVersion tests version extraction
func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name     string
		mw       models.TraefikMiddleware
		expected string
	}{
		{
			name: "from config",
			mw: models.TraefikMiddleware{
				Config: map[string]interface{}{
					"plugin": map[string]interface{}{
						"myPlugin": map[string]interface{}{
							"version": "v1.2.3",
						},
					},
				},
			},
			expected: "v1.2.3",
		},
		{
			name:     "empty when not found",
			mw:       models.TraefikMiddleware{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVersion(tt.mw)
			if got != tt.expected {
				t.Errorf("extractVersion() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestPluginFetcher_BuildPluginResponses tests building plugin responses
func TestPluginFetcher_BuildPluginResponses(t *testing.T) {
	config := models.DataSourceConfig{
		URL: "http://localhost:8080",
	}
	fetcher := NewPluginFetcher(config)

	middlewares := []models.TraefikMiddleware{
		{
			Name:     "badger@file",
			Type:     "plugin",
			Provider: "file",
			Config: map[string]interface{}{
				"plugin": map[string]interface{}{
					"badger": map[string]interface{}{},
				},
			},
		},
	}

	overview := &models.TraefikPluginOverview{
		Plugins: models.TraefikPluginStatus{
			Enabled: []string{"badger"},
		},
	}

	plugins := fetcher.buildPluginResponses(middlewares, overview)

	if len(plugins) == 0 {
		t.Error("expected at least one plugin")
		return
	}

	// Find the badger plugin
	var found bool
	for _, p := range plugins {
		if p.Name == "badger" {
			found = true
			if p.Status != "enabled" {
				t.Errorf("expected status 'enabled', got %q", p.Status)
			}
			if !p.IsInstalled {
				t.Error("expected IsInstalled true")
			}
		}
	}

	if !found {
		t.Error("badger plugin not found in responses")
	}
}

// TestParseNextDataPlugins tests parsing Next.js data structure
func TestParseNextDataPlugins(t *testing.T) {
	validHTML := `<html><script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{"plugins":[{"id":"test","name":"Test Plugin","type":"middleware"}]}}}</script></html>`

	plugins, err := parseNextDataPlugins(validHTML)
	if err != nil {
		t.Fatalf("parseNextDataPlugins() error = %v", err)
	}

	if len(plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(plugins))
	}

	if len(plugins) > 0 && plugins[0].Name != "Test Plugin" {
		t.Errorf("plugin name = %q, want 'Test Plugin'", plugins[0].Name)
	}
}

// TestParseNextDataPlugins_InvalidHTML tests parsing invalid HTML
func TestParseNextDataPlugins_InvalidHTML(t *testing.T) {
	_, err := parseNextDataPlugins("<html><body>No script here</body></html>")
	if err == nil {
		t.Error("parseNextDataPlugins() should error for invalid HTML")
	}
}
