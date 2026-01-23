package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestDefaultHTTPClientConfig tests that default configuration values are sensible
func TestDefaultHTTPClientConfig(t *testing.T) {
	config := DefaultHTTPClientConfig()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Timeout", config.Timeout, 10 * time.Second},
		{"MaxIdleConns", config.MaxIdleConns, 100},
		{"MaxIdleConnsPerHost", config.MaxIdleConnsPerHost, 10},
		{"IdleConnTimeout", config.IdleConnTimeout, 90 * time.Second},
		{"DisableKeepAlives", config.DisableKeepAlives, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestNewHTTPClient tests HTTP client creation with custom config
func TestNewHTTPClient(t *testing.T) {
	config := HTTPClientConfig{
		Timeout:             5 * time.Second,
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   true,
	}

	client := NewHTTPClient(config)

	if client == nil {
		t.Fatal("NewHTTPClient() returned nil")
	}

	if client.Timeout != config.Timeout {
		t.Errorf("client.Timeout = %v, want %v", client.Timeout, config.Timeout)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("client.Transport is not *http.Transport")
	}

	if transport.MaxIdleConns != config.MaxIdleConns {
		t.Errorf("MaxIdleConns = %d, want %d", transport.MaxIdleConns, config.MaxIdleConns)
	}

	if transport.MaxIdleConnsPerHost != config.MaxIdleConnsPerHost {
		t.Errorf("MaxIdleConnsPerHost = %d, want %d", transport.MaxIdleConnsPerHost, config.MaxIdleConnsPerHost)
	}

	if transport.DisableKeepAlives != config.DisableKeepAlives {
		t.Errorf("DisableKeepAlives = %v, want %v", transport.DisableKeepAlives, config.DisableKeepAlives)
	}
}

// TestGetHTTPClient tests the singleton HTTP client getter
func TestGetHTTPClient(t *testing.T) {
	client1 := GetHTTPClient()
	if client1 == nil {
		t.Fatal("GetHTTPClient() returned nil")
	}

	// Calling again should return the same instance
	client2 := GetHTTPClient()
	if client1 != client2 {
		t.Error("GetHTTPClient() should return the same instance")
	}
}

// TestHTTPClientWithTimeout tests creating a client with custom timeout
func TestHTTPClientWithTimeout(t *testing.T) {
	customTimeout := 30 * time.Second
	client := HTTPClientWithTimeout(customTimeout)

	if client == nil {
		t.Fatal("HTTPClientWithTimeout() returned nil")
	}

	if client.Timeout != customTimeout {
		t.Errorf("client.Timeout = %v, want %v", client.Timeout, customTimeout)
	}

	// Should reuse the transport from the shared client
	sharedClient := GetHTTPClient()
	if client.Transport != sharedClient.Transport {
		t.Error("HTTPClientWithTimeout should reuse the shared transport")
	}
}

// TestHTTPClientMakesRequests tests that the HTTP client can make actual requests
func TestHTTPClientMakesRequests(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := GetHTTPClient()

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestHTTPClientTimeout tests that timeout is enforced
func TestHTTPClientTimeout(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with very short timeout
	client := HTTPClientWithTimeout(100 * time.Millisecond)

	_, err := client.Get(server.URL)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

// TestHTTPClientConfigZeroValues tests handling of zero-value config
func TestHTTPClientConfigZeroValues(t *testing.T) {
	config := HTTPClientConfig{} // All zero values

	client := NewHTTPClient(config)

	if client == nil {
		t.Fatal("NewHTTPClient() with zero config returned nil")
	}

	// Client should still be functional with zero timeout (no timeout)
	if client.Timeout != 0 {
		t.Errorf("client.Timeout = %v, want 0", client.Timeout)
	}
}
