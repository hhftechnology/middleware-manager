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

// TestNewPangolinFetcher tests fetcher creation
func TestNewPangolinFetcher(t *testing.T) {
	config := models.DataSourceConfig{
		Type: models.PangolinAPI,
		URL:  "http://localhost:8080",
	}

	fetcher := NewPangolinFetcher(config)

	if fetcher == nil {
		t.Fatal("NewPangolinFetcher() returned nil")
	}
	if fetcher.config.URL != config.URL {
		t.Errorf("config.URL = %q, want %q", fetcher.config.URL, config.URL)
	}
	if fetcher.httpClient == nil {
		t.Error("httpClient is nil")
	}
	if fetcher.minInterval != 5*time.Second {
		t.Errorf("minInterval = %v, want %v", fetcher.minInterval, 5*time.Second)
	}
}

// TestPangolinFetcher_FetchResources tests fetching resources from mock Pangolin API
func TestPangolinFetcher_FetchResources(t *testing.T) {
	// Create a mock Pangolin API server
	var mockConfig models.PangolinTraefikConfig
	mockConfig.HTTP.Routers = map[string]models.PangolinRouter{
		"test-router": {
			Rule:        "Host(`test.example.com`)",
			Service:     "test-service",
			EntryPoints: []string{"websecure"},
			Priority:    100,
		},
	}
	mockConfig.HTTP.Services = map[string]models.PangolinService{}
	mockConfig.HTTP.Middlewares = map[string]map[string]interface{}{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/traefik-config" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockConfig)
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.PangolinAPI,
		URL:  server.URL,
	}

	fetcher := NewPangolinFetcher(config)
	ctx := context.Background()

	resources, err := fetcher.FetchResources(ctx)
	if err != nil {
		t.Fatalf("FetchResources() error = %v", err)
	}

	if resources == nil {
		t.Fatal("FetchResources() returned nil")
	}

	if len(resources.Resources) != 1 {
		t.Errorf("len(Resources) = %d, want 1", len(resources.Resources))
	}

	if len(resources.Resources) > 0 {
		r := resources.Resources[0]
		if r.Host != "test.example.com" {
			t.Errorf("Resource.Host = %q, want %q", r.Host, "test.example.com")
		}
	}
}

// TestPangolinFetcher_FetchResources_BasicAuth tests basic auth is sent correctly
func TestPangolinFetcher_FetchResources_BasicAuth(t *testing.T) {
	receivedAuth := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"http":{"routers":{},"services":{},"middlewares":{}}}`))
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.PangolinAPI,
		URL:  server.URL,
		BasicAuth: struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: "testuser",
			Password: "testpass",
		},
	}

	fetcher := NewPangolinFetcher(config)
	ctx := context.Background()

	_, err := fetcher.FetchResources(ctx)
	if err != nil {
		t.Fatalf("FetchResources() error = %v", err)
	}

	if receivedAuth == "" {
		t.Error("Basic auth header was not sent")
	}
}

// TestPangolinFetcher_FetchResources_Error tests error handling
func TestPangolinFetcher_FetchResources_Error(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
	}{
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			body:       "internal error",
			wantErr:    true,
		},
		{
			name:       "invalid JSON",
			statusCode: http.StatusOK,
			body:       "not valid json",
			wantErr:    true,
		},
		{
			name:       "empty response",
			statusCode: http.StatusOK,
			body:       `{"http":{"routers":{},"services":{},"middlewares":{}}}`,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			config := models.DataSourceConfig{
				Type: models.PangolinAPI,
				URL:  server.URL,
			}

			fetcher := NewPangolinFetcher(config)
			ctx := context.Background()

			_, err := fetcher.FetchResources(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPangolinFetcher_GetTraefikMiddlewares tests middleware fetching
func TestPangolinFetcher_GetTraefikMiddlewares(t *testing.T) {
	var mockConfig models.PangolinTraefikConfig
	mockConfig.HTTP.Routers = map[string]models.PangolinRouter{}
	mockConfig.HTTP.Services = map[string]models.PangolinService{}
	mockConfig.HTTP.Middlewares = map[string]map[string]interface{}{
		"test-headers": {
			"headers": map[string]interface{}{
				"customRequestHeaders": map[string]string{
					"X-Custom": "value",
				},
			},
		},
		"test-ratelimit": {
			"rateLimit": map[string]interface{}{
				"average": 100,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockConfig)
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.PangolinAPI,
		URL:  server.URL,
	}

	fetcher := NewPangolinFetcher(config)
	ctx := context.Background()

	middlewares, err := fetcher.GetTraefikMiddlewares(ctx)
	if err != nil {
		t.Fatalf("GetTraefikMiddlewares() error = %v", err)
	}

	if len(middlewares) != 2 {
		t.Errorf("len(middlewares) = %d, want 2", len(middlewares))
	}

	// Check middleware types are detected
	foundHeaders := false
	foundRateLimit := false
	for _, mw := range middlewares {
		if mw.Type == "headers" {
			foundHeaders = true
		}
		if mw.Type == "rateLimit" {
			foundRateLimit = true
		}
	}

	if !foundHeaders {
		t.Error("headers middleware type not detected")
	}
	if !foundRateLimit {
		t.Error("rateLimit middleware type not detected")
	}
}

// TestPangolinFetcher_GetTraefikServices tests service fetching
func TestPangolinFetcher_GetTraefikServices(t *testing.T) {
	var mockConfig models.PangolinTraefikConfig
	mockConfig.HTTP.Routers = map[string]models.PangolinRouter{}
	mockConfig.HTTP.Services = map[string]models.PangolinService{
		"test-service": {
			LoadBalancer: map[string]interface{}{
				"servers": []map[string]string{
					{"url": "http://backend:8080"},
				},
			},
		},
	}
	mockConfig.HTTP.Middlewares = map[string]map[string]interface{}{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockConfig)
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.PangolinAPI,
		URL:  server.URL,
	}

	fetcher := NewPangolinFetcher(config)
	ctx := context.Background()

	services, err := fetcher.GetTraefikServices(ctx)
	if err != nil {
		t.Fatalf("GetTraefikServices() error = %v", err)
	}

	if len(services) != 1 {
		t.Errorf("len(services) = %d, want 1", len(services))
	}

	if len(services) > 0 && services[0].Name != "test-service" {
		t.Errorf("service.Name = %q, want %q", services[0].Name, "test-service")
	}
}

// TestDetectMiddlewareType tests middleware type detection
func TestDetectMiddlewareType(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]interface{}
		want   string
	}{
		{
			name:   "headers middleware",
			config: map[string]interface{}{"headers": map[string]interface{}{}},
			want:   "headers",
		},
		{
			name:   "rateLimit middleware",
			config: map[string]interface{}{"rateLimit": map[string]interface{}{}},
			want:   "rateLimit",
		},
		{
			name:   "basicAuth middleware",
			config: map[string]interface{}{"basicAuth": map[string]interface{}{}},
			want:   "basicAuth",
		},
		{
			name:   "stripPrefix middleware",
			config: map[string]interface{}{"stripPrefix": map[string]interface{}{}},
			want:   "stripPrefix",
		},
		{
			name:   "plugin middleware",
			config: map[string]interface{}{"plugin": map[string]interface{}{}},
			want:   "plugin",
		},
		{
			name:   "chain middleware",
			config: map[string]interface{}{"chain": map[string]interface{}{}},
			want:   "chain",
		},
		{
			name:   "unknown middleware",
			config: map[string]interface{}{"customType": map[string]interface{}{}},
			want:   "unknown",
		},
		{
			name:   "empty config",
			config: map[string]interface{}{},
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectMiddlewareType(tt.config)
			if got != tt.want {
				t.Errorf("detectMiddlewareType() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestPangolinFetcher_RateLimiting tests rate limiting behavior
func TestPangolinFetcher_RateLimiting(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"http":{"routers":{},"services":{},"middlewares":{}}}`))
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.PangolinAPI,
		URL:  server.URL,
	}

	fetcher := NewPangolinFetcher(config)
	// Override min interval for testing
	fetcher.minInterval = 100 * time.Millisecond
	ctx := context.Background()

	// First request should succeed
	_, err := fetcher.FetchResources(ctx)
	if err != nil {
		t.Fatalf("First FetchResources() error = %v", err)
	}

	// Second immediate request should use cached data
	_, err = fetcher.FetchResources(ctx)
	if err != nil {
		t.Fatalf("Second FetchResources() error = %v", err)
	}

	// Should have only made 1 request to the server (second used cache)
	if requestCount != 1 {
		t.Errorf("requestCount = %d, want 1 (rate limiting should use cache)", requestCount)
	}
}
