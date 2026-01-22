package services

import (
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	defaultHTTPClient *http.Client
	httpClientOnce    sync.Once
)

// HTTPClientConfig contains configuration for the HTTP client pool
type HTTPClientConfig struct {
	Timeout             time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
	DisableKeepAlives   bool
}

// DefaultHTTPClientConfig returns the default HTTP client configuration
func DefaultHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		Timeout:             10 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}
}

// GetHTTPClient returns the shared HTTP client instance
// This client is configured for optimal connection pooling
func GetHTTPClient() *http.Client {
	httpClientOnce.Do(func() {
		config := DefaultHTTPClientConfig()
		defaultHTTPClient = NewHTTPClient(config)
	})
	return defaultHTTPClient
}

// NewHTTPClient creates a new HTTP client with the given configuration
func NewHTTPClient(config HTTPClientConfig) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     config.DisableKeepAlives,
		ForceAttemptHTTP2:     true,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}
}

// HTTPClientWithTimeout returns a client with custom timeout
// Note: This reuses the transport from the shared client for connection pooling
func HTTPClientWithTimeout(timeout time.Duration) *http.Client {
	client := GetHTTPClient()
	return &http.Client{
		Transport: client.Transport,
		Timeout:   timeout,
	}
}
