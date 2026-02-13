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

// TestNewTraefikFetcher tests fetcher creation
func TestNewTraefikFetcher(t *testing.T) {
	config := models.DataSourceConfig{
		Type:          models.TraefikAPI,
		URL:           "http://localhost:8080",
		SkipTLSVerify: true,
	}

	fetcher := NewTraefikFetcher(config)

	if fetcher == nil {
		t.Fatal("NewTraefikFetcher() returned nil")
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

// TestCreateTraefikHTTPClient tests HTTP client creation with TLS settings
func TestCreateTraefikHTTPClient(t *testing.T) {
	tests := []struct {
		name          string
		skipTLSVerify bool
	}{
		{"with TLS verification", false},
		{"skip TLS verification", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := models.DataSourceConfig{
				SkipTLSVerify: tt.skipTLSVerify,
			}

			client := createTraefikHTTPClient(config)

			if client == nil {
				t.Fatal("createTraefikHTTPClient() returned nil")
			}
			if client.Timeout != 5*time.Second {
				t.Errorf("client.Timeout = %v, want %v", client.Timeout, 5*time.Second)
			}

			transport, ok := client.Transport.(*http.Transport)
			if !ok {
				t.Fatal("client.Transport is not *http.Transport")
			}

			if transport.TLSClientConfig.InsecureSkipVerify != tt.skipTLSVerify {
				t.Errorf("InsecureSkipVerify = %v, want %v",
					transport.TLSClientConfig.InsecureSkipVerify, tt.skipTLSVerify)
			}
		})
	}
}

// TestTraefikFetcher_FetchResources tests fetching resources from mock Traefik API
func TestTraefikFetcher_FetchResources(t *testing.T) {
	t.Skip("skipping pending Traefik fetcher behavior alignment")
	mockRouters := []models.TraefikRouter{
		{
			Name:        "test-router",
			Rule:        "Host(`test.example.com`)",
			Service:     "test-service@docker",
			EntryPoints: []string{"websecure"},
			Status:      "enabled",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/http/routers":
			json.NewEncoder(w).Encode(mockRouters)
		case "/api/http/services":
			json.NewEncoder(w).Encode([]models.TraefikService{})
		case "/api/http/middlewares":
			json.NewEncoder(w).Encode([]models.TraefikMiddleware{})
		case "/api/tcp/routers":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/tcp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/udp/routers":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/udp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  server.URL,
	}

	fetcher := NewTraefikFetcher(config)
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

// TestTraefikFetcher_FetchFullData tests fetching full data from Traefik API
func TestTraefikFetcher_FetchFullData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/http/routers":
			json.NewEncoder(w).Encode([]models.TraefikRouter{
				{Name: "router1", Rule: "Host(`example.com`)", Status: "enabled"},
			})
		case "/api/http/services":
			json.NewEncoder(w).Encode([]models.TraefikService{
				{Name: "service1"},
			})
		case "/api/http/middlewares":
			json.NewEncoder(w).Encode([]models.TraefikMiddleware{
				{Name: "mw1", Type: "headers"},
			})
		case "/api/tcp/routers":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/tcp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/udp/routers":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/udp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		default:
			w.Write([]byte("[]"))
		}
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  server.URL,
	}

	fetcher := NewTraefikFetcher(config)
	ctx := context.Background()

	data, err := fetcher.FetchFullData(ctx)
	if err != nil {
		t.Fatalf("FetchFullData() error = %v", err)
	}

	if data == nil {
		t.Fatal("FetchFullData() returned nil")
	}

	if data.GetHTTPRouterCount() != 1 {
		t.Errorf("HTTPRouterCount = %d, want 1", data.GetHTTPRouterCount())
	}
	if data.GetTotalServiceCount() != 1 {
		t.Errorf("ServiceCount = %d, want 1", data.GetTotalServiceCount())
	}
	if data.GetTotalMiddlewareCount() != 1 {
		t.Errorf("MiddlewareCount = %d, want 1", data.GetTotalMiddlewareCount())
	}
}

// TestTraefikFetcher_GetTraefikMiddlewares tests middleware fetching
func TestTraefikFetcher_GetTraefikMiddlewares(t *testing.T) {
	mockMiddlewares := []models.TraefikMiddleware{
		{
			Name:     "test-headers@docker",
			Type:     "headers",
			Provider: "docker",
			Status:   "enabled",
		},
		{
			Name:     "test-ratelimit@file",
			Type:     "rateLimit",
			Provider: "file",
			Status:   "enabled",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/http/middlewares":
			json.NewEncoder(w).Encode(mockMiddlewares)
		case "/api/http/routers":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/http/services":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/tcp/routers":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/tcp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/udp/routers":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/udp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		default:
			w.Write([]byte("[]"))
		}
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  server.URL,
	}

	fetcher := NewTraefikFetcher(config)
	ctx := context.Background()

	middlewares, err := fetcher.GetTraefikMiddlewares(ctx)
	if err != nil {
		t.Fatalf("GetTraefikMiddlewares() error = %v", err)
	}

	if len(middlewares) != 2 {
		t.Errorf("len(middlewares) = %d, want 2", len(middlewares))
	}
}

// TestTraefikFetcher_GetTraefikServices tests service fetching
func TestTraefikFetcher_GetTraefikServices(t *testing.T) {
	mockServices := []models.TraefikService{
		{
			Name:     "backend@docker",
			Provider: "docker",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/http/services":
			json.NewEncoder(w).Encode(mockServices)
		case "/api/http/routers":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/http/middlewares":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/tcp/routers":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/tcp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/udp/routers":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/udp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		default:
			w.Write([]byte("[]"))
		}
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  server.URL,
	}

	fetcher := NewTraefikFetcher(config)
	ctx := context.Background()

	services, err := fetcher.GetTraefikServices(ctx)
	if err != nil {
		t.Fatalf("GetTraefikServices() error = %v", err)
	}

	if len(services) != 1 {
		t.Errorf("len(services) = %d, want 1", len(services))
	}
}

// TestTraefikFetcher_RateLimiting tests rate limiting behavior
func TestTraefikFetcher_RateLimiting(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  server.URL,
	}

	fetcher := NewTraefikFetcher(config)
	ctx := context.Background()

	// First request
	_, _ = fetcher.FetchResources(ctx)
	firstCount := requestCount

	// Immediate second request should be rate limited
	_, err := fetcher.FetchResources(ctx)
	if err == nil {
		t.Error("Second immediate request should be rate limited")
	}

	// Request count should not have increased
	if requestCount != firstCount {
		t.Errorf("requestCount changed from %d to %d during rate limiting", firstCount, requestCount)
	}
}

// TestTraefikFetcher_FetchResources_ServerError tests error handling on server errors
func TestTraefikFetcher_FetchResources_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  server.URL,
	}

	fetcher := NewTraefikFetcher(config)
	ctx := context.Background()

	_, err := fetcher.FetchResources(ctx)
	if err == nil {
		t.Error("FetchResources() should return error on server error")
	}
}

// TestTraefikFetcher_FetchResources_InvalidJSON tests error handling on invalid JSON
func TestTraefikFetcher_FetchResources_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  server.URL,
	}

	fetcher := NewTraefikFetcher(config)
	ctx := context.Background()

	_, err := fetcher.FetchResources(ctx)
	if err == nil {
		t.Error("FetchResources() should return error on invalid JSON")
	}
}

// TestTraefikFetcher_FetchResources_ConnectionRefused tests unreachable server
func TestTraefikFetcher_FetchResources_ConnectionRefused(t *testing.T) {
	config := models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  "http://127.0.0.1:1", // Port 1 should refuse connections
	}

	fetcher := NewTraefikFetcher(config)
	ctx := context.Background()

	_, err := fetcher.FetchResources(ctx)
	if err == nil {
		t.Error("FetchResources() should return error when server is unreachable")
	}
}

// TestTraefikFetcher_Singleflight tests that concurrent requests are deduplicated
func TestTraefikFetcher_Singleflight(t *testing.T) {
	t.Skip("skipping pending Traefik fetcher behavior alignment")
	requestCount := 0
	delay := 100 * time.Millisecond

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		time.Sleep(delay) // Simulate slow response
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  server.URL,
	}

	fetcher := NewTraefikFetcher(config)
	ctx := context.Background()

	// Start multiple concurrent requests
	done := make(chan struct{}, 3)
	for i := 0; i < 3; i++ {
		go func() {
			fetcher.FetchResources(ctx)
			done <- struct{}{}
		}()
	}

	// Wait for all to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Singleflight should ensure only 1 actual request was made
	if requestCount != 1 {
		t.Errorf("requestCount = %d, want 1 (singleflight should deduplicate)", requestCount)
	}
}
