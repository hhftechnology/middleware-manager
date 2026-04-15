package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hhftechnology/middleware-manager/models"
)

var traefikEndpointPaths = []string{
	"/api/http/routers",
	"/api/http/services",
	"/api/http/middlewares",
	"/api/tcp/routers",
	"/api/tcp/services",
	"/api/tcp/middlewares",
	"/api/udp/routers",
	"/api/udp/services",
	"/api/overview",
	"/api/version",
	"/api/entrypoints",
}

type traefikRequestCounter struct {
	mu     sync.Mutex
	counts map[string]int
}

func newTraefikRequestCounter() *traefikRequestCounter {
	return &traefikRequestCounter{
		counts: make(map[string]int),
	}
}

func (c *traefikRequestCounter) increment(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.counts[path]++
}

func (c *traefikRequestCounter) count(path string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.counts[path]
}

func defaultTraefikResponse(path string) (interface{}, bool) {
	switch path {
	case "/api/http/routers":
		return []models.TraefikRouter{
			{
				Name:        "test-router",
				Rule:        "Host(`test.example.com`)",
				Service:     "test-service@docker",
				EntryPoints: []string{"websecure"},
				Status:      "enabled",
			},
		}, true
	case "/api/http/services":
		return []models.TraefikService{
			{Name: "test-service@docker"},
		}, true
	case "/api/http/middlewares":
		return []models.TraefikMiddleware{
			{Name: "test-middleware@docker", Type: "headers"},
		}, true
	case "/api/tcp/routers":
		return []models.TCPRouter{}, true
	case "/api/tcp/services":
		return []models.TCPService{}, true
	case "/api/tcp/middlewares":
		return []models.TCPMiddleware{}, true
	case "/api/udp/routers":
		return []models.UDPRouter{}, true
	case "/api/udp/services":
		return []models.UDPService{}, true
	case "/api/overview":
		return models.TraefikOverview{}, true
	case "/api/version":
		return models.TraefikVersion{Version: "3.0.0", Codename: "test"}, true
	case "/api/entrypoints":
		return []models.TraefikEntrypoint{}, true
	default:
		return nil, false
	}
}

func writeJSONResponse(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeResponseBody(w http.ResponseWriter, body string) {
	if _, err := w.Write([]byte(body)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func newCountingTraefikServer(delay time.Duration) (*httptest.Server, *traefikRequestCounter) {
	counter := newTraefikRequestCounter()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter.increment(r.URL.Path)
		if delay > 0 {
			time.Sleep(delay)
		}

		payload, ok := defaultTraefikResponse(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

		writeJSONResponse(w, payload)
	}))

	return server, counter
}

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
		switch r.URL.Path {
		case "/api/http/routers":
			writeJSONResponse(w, mockRouters)
		case "/api/http/services":
			writeJSONResponse(w, []models.TraefikService{})
		case "/api/http/middlewares":
			writeJSONResponse(w, []models.TraefikMiddleware{})
		case "/api/tcp/routers":
			writeJSONResponse(w, []interface{}{})
		case "/api/tcp/services":
			writeJSONResponse(w, []interface{}{})
		case "/api/tcp/middlewares":
			writeJSONResponse(w, []interface{}{})
		case "/api/udp/routers":
			writeJSONResponse(w, []interface{}{})
		case "/api/udp/services":
			writeJSONResponse(w, []interface{}{})
		case "/api/overview":
			writeJSONResponse(w, models.TraefikOverview{})
		case "/api/version":
			writeJSONResponse(w, models.TraefikVersion{Version: "3.0.0", Codename: "test"})
		case "/api/entrypoints":
			writeJSONResponse(w, []models.TraefikEntrypoint{})
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
		switch r.URL.Path {
		case "/api/http/routers":
			writeJSONResponse(w, []models.TraefikRouter{
				{Name: "router1", Rule: "Host(`example.com`)", Status: "enabled"},
			})
		case "/api/http/services":
			writeJSONResponse(w, []models.TraefikService{
				{Name: "service1"},
			})
		case "/api/http/middlewares":
			writeJSONResponse(w, []models.TraefikMiddleware{
				{Name: "mw1", Type: "headers"},
			})
		case "/api/tcp/routers":
			writeJSONResponse(w, []interface{}{})
		case "/api/tcp/services":
			writeJSONResponse(w, []interface{}{})
		case "/api/tcp/middlewares":
			writeJSONResponse(w, []interface{}{})
		case "/api/udp/routers":
			writeJSONResponse(w, []interface{}{})
		case "/api/udp/services":
			writeJSONResponse(w, []interface{}{})
		case "/api/overview":
			writeJSONResponse(w, models.TraefikOverview{})
		case "/api/version":
			writeJSONResponse(w, models.TraefikVersion{Version: "3.0.0", Codename: "test"})
		case "/api/entrypoints":
			writeJSONResponse(w, []models.TraefikEntrypoint{})
		default:
			writeResponseBody(w, "[]")
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
		switch r.URL.Path {
		case "/api/http/middlewares":
			writeJSONResponse(w, mockMiddlewares)
		case "/api/http/routers":
			writeJSONResponse(w, []interface{}{})
		case "/api/http/services":
			writeJSONResponse(w, []interface{}{})
		case "/api/tcp/routers":
			writeJSONResponse(w, []interface{}{})
		case "/api/tcp/services":
			writeJSONResponse(w, []interface{}{})
		case "/api/tcp/middlewares":
			writeJSONResponse(w, []interface{}{})
		case "/api/udp/routers":
			writeJSONResponse(w, []interface{}{})
		case "/api/udp/services":
			writeJSONResponse(w, []interface{}{})
		case "/api/overview":
			writeJSONResponse(w, models.TraefikOverview{})
		case "/api/version":
			writeJSONResponse(w, models.TraefikVersion{Version: "3.0.0", Codename: "test"})
		case "/api/entrypoints":
			writeJSONResponse(w, []models.TraefikEntrypoint{})
		default:
			writeResponseBody(w, "[]")
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
		switch r.URL.Path {
		case "/api/http/services":
			writeJSONResponse(w, mockServices)
		case "/api/http/routers":
			writeJSONResponse(w, []interface{}{})
		case "/api/http/middlewares":
			writeJSONResponse(w, []interface{}{})
		case "/api/tcp/routers":
			writeJSONResponse(w, []interface{}{})
		case "/api/tcp/services":
			writeJSONResponse(w, []interface{}{})
		case "/api/tcp/middlewares":
			writeJSONResponse(w, []interface{}{})
		case "/api/udp/routers":
			writeJSONResponse(w, []interface{}{})
		case "/api/udp/services":
			writeJSONResponse(w, []interface{}{})
		case "/api/overview":
			writeJSONResponse(w, models.TraefikOverview{})
		case "/api/version":
			writeJSONResponse(w, models.TraefikVersion{Version: "3.0.0", Codename: "test"})
		case "/api/entrypoints":
			writeJSONResponse(w, []models.TraefikEntrypoint{})
		default:
			writeResponseBody(w, "[]")
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
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		writeResponseBody(w, "[]")
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  server.URL,
	}

	fetcher := NewTraefikFetcher(config)
	ctx := context.Background()

	// First request
	if _, err := fetcher.FetchResources(ctx); err != nil {
		t.Fatalf("First FetchResources() error = %v", err)
	}
	firstCount := requestCount.Load()

	// Immediate second request should be rate limited
	_, err := fetcher.FetchResources(ctx)
	if err == nil {
		t.Error("Second immediate request should be rate limited")
	}

	// Request count should not have increased
	if requestCount.Load() != firstCount {
		t.Errorf("requestCount changed from %d to %d during rate limiting", firstCount, requestCount.Load())
	}
}

// TestTraefikFetcher_FetchResources_ServerError tests error handling on server errors
func TestTraefikFetcher_FetchResources_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		writeResponseBody(w, "internal server error")
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
		writeResponseBody(w, "not valid json")
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
	server, counter := newCountingTraefikServer(100 * time.Millisecond)
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  server.URL,
	}

	fetcher := NewTraefikFetcher(config)
	ctx := context.Background()

	errCh := make(chan error, 3)
	for i := 0; i < 3; i++ {
		go func() {
			_, err := fetcher.FetchResources(ctx)
			errCh <- err
		}()
	}

	for i := 0; i < 3; i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("FetchResources() error = %v", err)
		}
	}

	for _, path := range traefikEndpointPaths {
		if got := counter.count(path); got != 1 {
			t.Errorf("request count for %s = %d, want 1", path, got)
		}
	}
}

func TestTraefikFetcher_FetchResourcesAndFullDataShareSingleflight(t *testing.T) {
	server, counter := newCountingTraefikServer(100 * time.Millisecond)
	defer server.Close()

	config := models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  server.URL,
	}

	fetcher := NewTraefikFetcher(config)
	ctx := context.Background()

	resourceCh := make(chan error, 1)
	fullDataCh := make(chan error, 1)

	go func() {
		resources, err := fetcher.FetchResources(ctx)
		if err == nil && len(resources.Resources) != 1 {
			err = fmt.Errorf("len(Resources) = %d, want 1", len(resources.Resources))
		}
		resourceCh <- err
	}()

	go func() {
		data, err := fetcher.FetchFullData(ctx)
		if err == nil && data.GetHTTPRouterCount() != 1 {
			err = fmt.Errorf("HTTPRouterCount = %d, want 1", data.GetHTTPRouterCount())
		}
		fullDataCh <- err
	}()

	if err := <-resourceCh; err != nil {
		t.Fatal(err)
	}
	if err := <-fullDataCh; err != nil {
		t.Fatal(err)
	}

	for _, path := range traefikEndpointPaths {
		if got := counter.count(path); got != 1 {
			t.Errorf("request count for %s = %d, want 1", path, got)
		}
	}
}
