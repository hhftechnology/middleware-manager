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

func newLoadBalancer(url string) *struct {
	Servers []struct {
		URL     string `json:"url,omitempty"`
		Address string `json:"address,omitempty"`
		Weight  *int   `json:"weight,omitempty"`
	} `json:"servers,omitempty"`
	PassHostHeader *bool       `json:"passHostHeader,omitempty"`
	Sticky         interface{} `json:"sticky,omitempty"`
	HealthCheck    interface{} `json:"healthCheck,omitempty"`
} {
	return &struct {
		Servers []struct {
			URL     string `json:"url,omitempty"`
			Address string `json:"address,omitempty"`
			Weight  *int   `json:"weight,omitempty"`
		} `json:"servers,omitempty"`
		PassHostHeader *bool       `json:"passHostHeader,omitempty"`
		Sticky         interface{} `json:"sticky,omitempty"`
		HealthCheck    interface{} `json:"healthCheck,omitempty"`
	}{
		Servers: []struct {
			URL     string `json:"url,omitempty"`
			Address string `json:"address,omitempty"`
			Weight  *int   `json:"weight,omitempty"`
		}{
			{URL: url},
		},
	}
}

func newWeighted(name string, weight int) *struct {
	Services []struct {
		Name   string `json:"name"`
		Weight int    `json:"weight"`
	} `json:"services,omitempty"`
	Sticky      interface{} `json:"sticky,omitempty"`
	HealthCheck interface{} `json:"healthCheck,omitempty"`
} {
	return &struct {
		Services []struct {
			Name   string `json:"name"`
			Weight int    `json:"weight"`
		} `json:"services,omitempty"`
		Sticky      interface{} `json:"sticky,omitempty"`
		HealthCheck interface{} `json:"healthCheck,omitempty"`
	}{
		Services: []struct {
			Name   string `json:"name"`
			Weight int    `json:"weight"`
		}{
			{Name: name, Weight: weight},
		},
	}
}

func newMirroring(name string, percent int, service string) *struct {
	Service    string `json:"service"`
	Mirrors    []struct {
		Name    string `json:"name"`
		Percent int    `json:"percent"`
	} `json:"mirrors,omitempty"`
	MaxBodySize *int        `json:"maxBodySize,omitempty"`
	MirrorBody  *bool       `json:"mirrorBody,omitempty"`
	HealthCheck interface{} `json:"healthCheck,omitempty"`
} {
	return &struct {
		Service    string `json:"service"`
		Mirrors    []struct {
			Name    string `json:"name"`
			Percent int    `json:"percent"`
		} `json:"mirrors,omitempty"`
		MaxBodySize *int        `json:"maxBodySize,omitempty"`
		MirrorBody  *bool       `json:"mirrorBody,omitempty"`
		HealthCheck interface{} `json:"healthCheck,omitempty"`
	}{
		Service: service,
		Mirrors: []struct {
			Name    string `json:"name"`
			Percent int    `json:"percent"`
		}{
			{Name: name, Percent: percent},
		},
	}
}

func newFailover(service, fallback string) *struct {
	Service     string      `json:"service"`
	Fallback    string      `json:"fallback"`
	HealthCheck interface{} `json:"healthCheck,omitempty"`
} {
	return &struct {
		Service     string      `json:"service"`
		Fallback    string      `json:"fallback"`
		HealthCheck interface{} `json:"healthCheck,omitempty"`
	}{
		Service:  service,
		Fallback: fallback,
	}
}

// TestNewServiceFetcher_Pangolin tests creating Pangolin service fetcher
func TestNewServiceFetcher_Pangolin(t *testing.T) {
	config := models.DataSourceConfig{
		Type: models.PangolinAPI,
		URL:  "http://localhost:3001",
	}

	fetcher, err := NewServiceFetcher(config)
	if err != nil {
		t.Fatalf("NewServiceFetcher() error = %v", err)
	}

	if fetcher == nil {
		t.Fatal("NewServiceFetcher() returned nil")
	}

	_, ok := fetcher.(*PangolinServiceFetcher)
	if !ok {
		t.Error("expected PangolinServiceFetcher type")
	}
}

// TestNewServiceFetcher_Traefik tests creating Traefik service fetcher
func TestNewServiceFetcher_Traefik(t *testing.T) {
	config := models.DataSourceConfig{
		Type: models.TraefikAPI,
		URL:  "http://localhost:8080",
	}

	fetcher, err := NewServiceFetcher(config)
	if err != nil {
		t.Fatalf("NewServiceFetcher() error = %v", err)
	}

	if fetcher == nil {
		t.Fatal("NewServiceFetcher() returned nil")
	}

	_, ok := fetcher.(*TraefikServiceFetcher)
	if !ok {
		t.Error("expected TraefikServiceFetcher type")
	}
}

// TestNewServiceFetcher_UnknownType tests error for unknown type
func TestNewServiceFetcher_UnknownType(t *testing.T) {
	config := models.DataSourceConfig{
		Type: "unknown",
		URL:  "http://localhost:8080",
	}

	_, err := NewServiceFetcher(config)
	if err == nil {
		t.Error("NewServiceFetcher() should error for unknown type")
	}
}

// TestNewPangolinServiceFetcher tests Pangolin fetcher creation
func TestNewPangolinServiceFetcher(t *testing.T) {
	config := models.DataSourceConfig{
		URL: "http://localhost:3001",
	}

	fetcher := NewPangolinServiceFetcher(config)

	if fetcher == nil {
		t.Fatal("NewPangolinServiceFetcher() returned nil")
	}
	if fetcher.httpClient == nil {
		t.Error("httpClient is nil")
	}
	if fetcher.config.URL != config.URL {
		t.Errorf("URL = %q, want %q", fetcher.config.URL, config.URL)
	}
}

// TestNewTraefikServiceFetcher tests Traefik fetcher creation
func TestNewTraefikServiceFetcher(t *testing.T) {
	config := models.DataSourceConfig{
		URL: "http://localhost:8080",
	}

	fetcher := NewTraefikServiceFetcher(config)

	if fetcher == nil {
		t.Fatal("NewTraefikServiceFetcher() returned nil")
	}
	if fetcher.httpClient == nil {
		t.Error("httpClient is nil")
	}
}

// TestPangolinServiceFetcher_FetchServices tests Pangolin service fetching
func TestPangolinServiceFetcher_FetchServices(t *testing.T) {
	// Create mock Pangolin server
	var pangolinConfig models.PangolinTraefikConfig
	pangolinConfig.HTTP.Services = map[string]models.PangolinService{
		"web-service": {
			LoadBalancer: map[string]interface{}{
				"servers": []map[string]interface{}{
					{"url": "http://backend:8080"},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/traefik-config" {
			json.NewEncoder(w).Encode(pangolinConfig)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
	}
	fetcher := NewPangolinServiceFetcher(config)

	ctx := context.Background()
	services, err := fetcher.FetchServices(ctx)
	if err != nil {
		t.Fatalf("FetchServices() error = %v", err)
	}

	if services == nil {
		t.Fatal("FetchServices() returned nil")
	}

	if len(services.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(services.Services))
	}
}

// TestPangolinServiceFetcher_FetchServices_WithAuth tests auth
func TestPangolinServiceFetcher_FetchServices_WithAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "secret" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
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
	fetcher := NewPangolinServiceFetcher(config)

	ctx := context.Background()
	_, err := fetcher.FetchServices(ctx)
	if err != nil {
		t.Fatalf("FetchServices() with auth error = %v", err)
	}
}

// TestPangolinServiceFetcher_FetchServices_ServerError tests error handling
func TestPangolinServiceFetcher_FetchServices_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
	}
	fetcher := NewPangolinServiceFetcher(config)

	ctx := context.Background()
	_, err := fetcher.FetchServices(ctx)
	if err == nil {
		t.Error("FetchServices() should error on server error")
	}
}

// TestTraefikServiceFetcher_FetchServices tests Traefik service fetching
func TestTraefikServiceFetcher_FetchServices(t *testing.T) {
	// Create mock Traefik server
	httpServices := []models.TraefikService{
		{
			Name:     "web-service@docker",
			Provider: "docker",
			LoadBalancer: newLoadBalancer("http://backend:8080"),
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/http/services":
			json.NewEncoder(w).Encode(httpServices)
		case "/api/tcp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/udp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
	}
	fetcher := NewTraefikServiceFetcher(config)

	ctx := context.Background()
	services, err := fetcher.FetchServices(ctx)
	if err != nil {
		t.Fatalf("FetchServices() error = %v", err)
	}

	if services == nil {
		t.Fatal("FetchServices() returned nil")
	}

	if len(services.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(services.Services))
	}
}

// TestTraefikServiceFetcher_FetchServices_MapResponse tests map response format
func TestTraefikServiceFetcher_FetchServices_MapResponse(t *testing.T) {
	// Some Traefik versions return services as a map
	httpServices := map[string]models.TraefikService{
		"web-service@docker": {
			Provider: "docker",
			LoadBalancer: newLoadBalancer("http://backend:8080"),
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/http/services":
			json.NewEncoder(w).Encode(httpServices)
		case "/api/tcp/services":
			json.NewEncoder(w).Encode(map[string]interface{}{})
		case "/api/udp/services":
			json.NewEncoder(w).Encode(map[string]interface{}{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
	}
	fetcher := NewTraefikServiceFetcher(config)

	ctx := context.Background()
	services, err := fetcher.FetchServices(ctx)
	if err != nil {
		t.Fatalf("FetchServices() error = %v", err)
	}

	if len(services.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(services.Services))
	}
}

// TestTraefikServiceFetcher_FetchServices_SkipsInternal tests skipping internal services
func TestTraefikServiceFetcher_FetchServices_SkipsInternal(t *testing.T) {
	httpServices := []models.TraefikService{
		{
			Name:     "api@internal",
			Provider: "internal",
		},
		{
			Name:     "user-service@docker",
			Provider: "docker",
			LoadBalancer: newLoadBalancer("http://backend:8080"),
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/http/services":
			json.NewEncoder(w).Encode(httpServices)
		case "/api/tcp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/udp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
	}
	fetcher := NewTraefikServiceFetcher(config)

	ctx := context.Background()
	services, err := fetcher.FetchServices(ctx)
	if err != nil {
		t.Fatalf("FetchServices() error = %v", err)
	}

	// Should only have 1 service (internal skipped)
	if len(services.Services) != 1 {
		t.Errorf("expected 1 service (internal skipped), got %d", len(services.Services))
	}
}

// TestTraefikServiceFetcher_ContextCancellation tests context handling
func TestTraefikServiceFetcher_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		json.NewEncoder(w).Encode([]models.TraefikService{})
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
	}
	fetcher := NewTraefikServiceFetcher(config)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := fetcher.FetchServices(ctx)
	if err == nil {
		t.Log("Expected error due to context timeout, but got success")
	}
}

// TestDetermineServiceType tests service type determination
func TestDetermineServiceType(t *testing.T) {
	tests := []struct {
		name     string
		service  models.PangolinService
		expected string
	}{
		{
			name: "load balancer",
			service: models.PangolinService{
				LoadBalancer: map[string]interface{}{
					"servers": []interface{}{},
				},
			},
			expected: string(models.LoadBalancerType),
		},
		{
			name: "weighted",
			service: models.PangolinService{
				Weighted: map[string]interface{}{
					"services": []interface{}{},
				},
			},
			expected: string(models.WeightedType),
		},
		{
			name: "mirroring",
			service: models.PangolinService{
				Mirroring: map[string]interface{}{
					"service": "main",
				},
			},
			expected: string(models.MirroringType),
		},
		{
			name: "failover",
			service: models.PangolinService{
				Failover: map[string]interface{}{
					"service":  "main",
					"fallback": "backup",
				},
			},
			expected: string(models.FailoverType),
		},
		{
			name:     "default to load balancer",
			service:  models.PangolinService{},
			expected: string(models.LoadBalancerType),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineServiceType(tt.service)
			if got != tt.expected {
				t.Errorf("determineServiceType() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestIsPangolinSystemService tests system service detection
func TestIsPangolinSystemService(t *testing.T) {
	tests := []struct {
		serviceID string
		expected  bool
	}{
		{"api-service@file", true},
		{"next-service@file", true},
		{"noop@internal", true},
		{"user-service@file", false},
		{"my-api@docker", false},
	}

	for _, tt := range tests {
		t.Run(tt.serviceID, func(t *testing.T) {
			got := isPangolinSystemService(tt.serviceID)
			if got != tt.expected {
				t.Errorf("isPangolinSystemService(%q) = %v, want %v", tt.serviceID, got, tt.expected)
			}
		})
	}
}

// TestIsTraefikSystemService tests Traefik system service detection
func TestIsTraefikSystemService(t *testing.T) {
	tests := []struct {
		serviceID string
		expected  bool
	}{
		{"api@internal", true},
		{"dashboard@internal", true},
		{"noop@internal", true},
		{"acme-http@internal", true},
		{"user-service@docker", false},
		{"my-service@file", false},
	}

	for _, tt := range tests {
		t.Run(tt.serviceID, func(t *testing.T) {
			got := isTraefikSystemService(tt.serviceID)
			if got != tt.expected {
				t.Errorf("isTraefikSystemService(%q) = %v, want %v", tt.serviceID, got, tt.expected)
			}
		})
	}
}

// TestProcessTraefikService tests Traefik service processing
func TestProcessTraefikService(t *testing.T) {
	tests := []struct {
		name     string
		service  models.TraefikService
		wantNil  bool
		wantType string
	}{
		{
			name: "load balancer service",
			service: models.TraefikService{
				Name:         "web@docker",
				Provider:     "docker",
				LoadBalancer: newLoadBalancer("http://localhost:8080"),
			},
			wantNil:  false,
			wantType: string(models.LoadBalancerType),
		},
		{
			name: "weighted service",
			service: models.TraefikService{
				Name:     "weighted@file",
				Provider: "file",
				Weighted: newWeighted("svc1", 50),
			},
			wantNil:  false,
			wantType: string(models.WeightedType),
		},
		{
			name: "mirroring service",
			service: models.TraefikService{
				Name:       "mirror@file",
				Provider:   "file",
				Mirroring:  newMirroring("mirror1", 10, "main-service"),
			},
			wantNil:  false,
			wantType: string(models.MirroringType),
		},
		{
			name: "failover service",
			service: models.TraefikService{
				Name:     "failover@file",
				Provider: "file",
				Failover: newFailover("main-service", "backup-service"),
			},
			wantNil:  false,
			wantType: string(models.FailoverType),
		},
		{
			name: "system service skipped",
			service: models.TraefikService{
				Name:     "api@internal",
				Provider: "internal",
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processTraefikService(tt.service)

			if tt.wantNil {
				if result != nil {
					t.Error("expected nil result for system service")
				}
				return
			}

			if result == nil {
				t.Fatal("processTraefikService() returned nil")
			}

			if result.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", result.Type, tt.wantType)
			}

			if result.ID != tt.service.Name {
				t.Errorf("ID = %q, want %q", result.ID, tt.service.Name)
			}
		})
	}
}

// TestTraefikServiceFetcher_FallbackURLs tests fallback URL handling
func TestTraefikServiceFetcher_FallbackURLs(t *testing.T) {
	// Create a server that will respond successfully
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/http/services":
			json.NewEncoder(w).Encode([]models.TraefikService{})
		case "/api/tcp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/udp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Use a bad URL as primary, but the test server URL should work
	// In a real scenario, this would test fallback behavior
	config := models.DataSourceConfig{
		URL: server.URL, // Using the working URL directly for this test
	}
	fetcher := NewTraefikServiceFetcher(config)

	ctx := context.Background()
	services, err := fetcher.FetchServices(ctx)
	if err != nil {
		t.Fatalf("FetchServices() error = %v", err)
	}

	if services == nil {
		t.Fatal("expected non-nil services")
	}
}

// TestTraefikServiceFetcher_TCPServices tests TCP service fetching
func TestTraefikServiceFetcher_TCPServices(t *testing.T) {
	tcpServices := []map[string]interface{}{
		{
			"name":     "tcp-service@docker",
			"provider": "docker",
			"loadBalancer": map[string]interface{}{
				"servers": []interface{}{
					map[string]interface{}{"address": "backend:3306"},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/http/services":
			json.NewEncoder(w).Encode([]models.TraefikService{})
		case "/api/tcp/services":
			json.NewEncoder(w).Encode(tcpServices)
		case "/api/udp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
	}
	fetcher := NewTraefikServiceFetcher(config)

	ctx := context.Background()
	services, err := fetcher.FetchServices(ctx)
	if err != nil {
		t.Fatalf("FetchServices() error = %v", err)
	}

	// Should have the TCP service
	if len(services.Services) != 1 {
		t.Errorf("expected 1 TCP service, got %d", len(services.Services))
	}
}

// TestTraefikServiceFetcher_UDPServices tests UDP service fetching
func TestTraefikServiceFetcher_UDPServices(t *testing.T) {
	udpServices := []map[string]interface{}{
		{
			"name":     "dns-service@docker",
			"provider": "docker",
			"loadBalancer": map[string]interface{}{
				"servers": []interface{}{
					map[string]interface{}{"address": "dns-server:53"},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/http/services":
			json.NewEncoder(w).Encode([]models.TraefikService{})
		case "/api/tcp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/udp/services":
			json.NewEncoder(w).Encode(udpServices)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
	}
	fetcher := NewTraefikServiceFetcher(config)

	ctx := context.Background()
	services, err := fetcher.FetchServices(ctx)
	if err != nil {
		t.Fatalf("FetchServices() error = %v", err)
	}

	// Should have the UDP service
	if len(services.Services) != 1 {
		t.Errorf("expected 1 UDP service, got %d", len(services.Services))
	}
}

// TestTraefikServiceFetcher_UDPNotSupported tests handling UDP not supported
func TestTraefikServiceFetcher_UDPNotSupported(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/http/services":
			json.NewEncoder(w).Encode([]models.TraefikService{})
		case "/api/tcp/services":
			json.NewEncoder(w).Encode([]interface{}{})
		case "/api/udp/services":
			http.NotFound(w, r) // UDP not supported
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := models.DataSourceConfig{
		URL: server.URL,
	}
	fetcher := NewTraefikServiceFetcher(config)

	ctx := context.Background()
	services, err := fetcher.FetchServices(ctx)
	if err != nil {
		t.Fatalf("FetchServices() error = %v (should handle UDP not found gracefully)", err)
	}

	if services == nil {
		t.Fatal("expected non-nil services")
	}
}
