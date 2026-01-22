package services

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hhftechnology/middleware-manager/models"
	"golang.org/x/sync/singleflight"
)

// TraefikFetcher fetches resources from Traefik API
// Implements best practices from Mantrae:
// - Concurrent multi-endpoint fetching
// - Singleflight pattern to prevent duplicate requests
// - Rate limiting between updates
// - Configurable TLS verification
// - Proper error categorization (critical vs non-critical)
type TraefikFetcher struct {
	config       models.DataSourceConfig
	httpClient   *http.Client
	singleflight singleflight.Group
	lastFetch    time.Time
	lastFetchMu  sync.RWMutex
	minInterval  time.Duration

	// Cached data from last fetch
	cachedData   *models.FullTraefikData
	cachedDataMu sync.RWMutex
}

// fetchResult holds result from concurrent fetch operation
type fetchResult struct {
	name string
	data []byte
	err  error
}

// NewTraefikFetcher creates a new Traefik API fetcher with connection pooling
func NewTraefikFetcher(config models.DataSourceConfig) *TraefikFetcher {
	// Use the shared HTTP client pool but allow TLS configuration
	httpClient := createTraefikHTTPClient(config)

	return &TraefikFetcher{
		config:      config,
		httpClient:  httpClient,
		minInterval: 5 * time.Second, // Rate limit: minimum 5 seconds between fetches
	}
}

// createTraefikHTTPClient creates an HTTP client with proper TLS settings
// Following Mantrae's pattern: 5-second timeout, connection pooling (100 max idle, 10 per-host)
func createTraefikHTTPClient(config models.DataSourceConfig) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.SkipTLSVerify,
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second, // Mantrae uses 5 seconds
	}
}

// FetchResources fetches resources from Traefik API with fallback options
// Uses singleflight to prevent duplicate concurrent requests
func (f *TraefikFetcher) FetchResources(ctx context.Context) (*models.ResourceCollection, error) {
	// Use singleflight to deduplicate concurrent requests
	result, err, _ := f.singleflight.Do("fetch-resources", func() (interface{}, error) {
		return f.fetchResourcesInternal(ctx)
	})

	if err != nil {
		return nil, err
	}

	return result.(*models.ResourceCollection), nil
}

// FetchFullData fetches all data from Traefik API including TCP/UDP
// Uses singleflight to prevent duplicate concurrent requests
func (f *TraefikFetcher) FetchFullData(ctx context.Context) (*models.FullTraefikData, error) {
	// Use singleflight to deduplicate concurrent requests
	result, err, _ := f.singleflight.Do("fetch-full-data", func() (interface{}, error) {
		return f.fetchFullDataInternal(ctx)
	})

	if err != nil {
		return nil, err
	}

	return result.(*models.FullTraefikData), nil
}

// fetchResourcesInternal performs the actual fetch with rate limiting
func (f *TraefikFetcher) fetchResourcesInternal(ctx context.Context) (*models.ResourceCollection, error) {
	// Check rate limiting
	f.lastFetchMu.RLock()
	timeSinceLastFetch := time.Since(f.lastFetch)
	f.lastFetchMu.RUnlock()

	if timeSinceLastFetch < f.minInterval {
		log.Printf("Rate limiting: skipping fetch, last fetch was %v ago", timeSinceLastFetch)
		return nil, fmt.Errorf("rate limited: please wait %v before next fetch", f.minInterval-timeSinceLastFetch)
	}

	log.Println("Fetching resources from Traefik API...")

	// Try the configured URL first
	resources, err := f.fetchResourcesFromURL(ctx, f.config.URL)
	if err == nil {
		f.updateLastFetch()
		log.Printf("Successfully fetched resources from %s", f.config.URL)
		return resources, nil
	}

	// Log the initial error
	log.Printf("Failed to connect to primary Traefik API URL %s: %v", f.config.URL, err)

	// Try common fallback URLs
	fallbackURLs := []string{
		"http://host.docker.internal:8080",
		"http://localhost:8080",
		"http://127.0.0.1:8080",
		"http://traefik:8080",
	}

	// Don't try the same URL twice
	if f.config.URL != "" {
		for i := len(fallbackURLs) - 1; i >= 0; i-- {
			if fallbackURLs[i] == f.config.URL {
				fallbackURLs = append(fallbackURLs[:i], fallbackURLs[i+1:]...)
			}
		}
	}

	// Try each fallback URL
	var lastErr error
	for _, url := range fallbackURLs {
		log.Printf("Trying fallback Traefik API URL: %s", url)
		resources, err := f.fetchResourcesFromURL(ctx, url)
		if err == nil {
			f.updateLastFetch()
			f.suggestURLUpdate(url)
			return resources, nil
		}
		lastErr = err
		log.Printf("Fallback URL %s failed: %v", url, err)
	}

	return nil, fmt.Errorf("all Traefik API connection attempts failed, last error: %w", lastErr)
}

// fetchFullDataInternal fetches all Traefik data with caching
func (f *TraefikFetcher) fetchFullDataInternal(ctx context.Context) (*models.FullTraefikData, error) {
	// Check rate limiting and return cached data if available
	f.lastFetchMu.RLock()
	timeSinceLastFetch := time.Since(f.lastFetch)
	f.lastFetchMu.RUnlock()

	f.cachedDataMu.RLock()
	cachedData := f.cachedData
	f.cachedDataMu.RUnlock()

	if timeSinceLastFetch < f.minInterval && cachedData != nil {
		log.Printf("Rate limiting: using cached data, last fetch was %v ago", timeSinceLastFetch)
		return cachedData, nil
	}

	log.Println("Fetching full data from Traefik API...")

	// Fetch all endpoints concurrently
	data, err := f.fetchAllEndpointsConcurrently(ctx, f.config.URL)
	if err != nil {
		return nil, err
	}

	// Update cache
	f.cachedDataMu.Lock()
	f.cachedData = data
	f.cachedDataMu.Unlock()

	// Update last fetch time
	f.updateLastFetch()

	log.Printf("Fetched full data: %d HTTP routers, %d TCP routers, %d UDP routers, %d services, %d middlewares",
		data.GetHTTPRouterCount(),
		data.GetTCPRouterCount(),
		data.GetUDPRouterCount(),
		data.GetTotalServiceCount(),
		data.GetTotalMiddlewareCount())

	return data, nil
}

// updateLastFetch updates the last fetch timestamp
func (f *TraefikFetcher) updateLastFetch() {
	f.lastFetchMu.Lock()
	f.lastFetch = time.Now()
	f.lastFetchMu.Unlock()
}

// fetchResourcesFromURL fetches resources from a specific URL using concurrent fetching
func (f *TraefikFetcher) fetchResourcesFromURL(ctx context.Context, baseURL string) (*models.ResourceCollection, error) {
	// Fetch all endpoints concurrently (like Mantrae pattern)
	fullData, err := f.fetchAllEndpointsConcurrently(ctx, baseURL)
	if err != nil {
		return nil, err
	}

	// Convert Traefik routers to our internal model
	resources := &models.ResourceCollection{
		Resources: make([]models.Resource, 0, len(fullData.HTTPRouters)),
	}

	// Build TLS domains map from routers
	tlsDomainsMap := make(map[string]string)
	for _, router := range fullData.HTTPRouters {
		if len(router.TLS.Domains) > 0 && router.Name != "" {
			tlsDomainsMap[router.Name] = models.JoinTLSDomains(router.TLS.Domains)
		}
	}

	for _, router := range fullData.HTTPRouters {
		// Skip internal routers
		if router.Provider == "internal" {
			continue
		}

		// Skip routers without TLS only if configured to do so
		if router.TLS.CertResolver == "" && !shouldIncludeNonTLSRouters() {
			continue
		}

		// Skip system routers (dashboard, api, etc.)
		if isTraefikSystemRouter(router.Name) {
			continue
		}

		// Extract host from rule
		host := extractHostFromRule(router.Rule)
		if host == "" {
			log.Printf("Could not extract host from rule: %s", router.Rule)
			continue
		}

		// Create resource
		resource := models.Resource{
			ID:             router.Name,
			Host:           host,
			ServiceID:      router.Service,
			Status:         "active",
			SourceType:     string(models.TraefikAPI),
			Entrypoints:    joinEntrypoints(router.EntryPoints),
			RouterPriority: router.Priority,
		}

		// Add TLS domains if available
		if tlsDomains, exists := tlsDomainsMap[router.Name]; exists {
			resource.TLSDomains = tlsDomains
		} else if len(router.TLS.Domains) > 0 {
			resource.TLSDomains = models.JoinTLSDomains(router.TLS.Domains)
		}

		resources.Resources = append(resources.Resources, resource)
	}

	log.Printf("Fetched %d HTTP routers, %d TCP routers, %d UDP routers, %d services, %d middlewares from Traefik API",
		fullData.GetHTTPRouterCount(),
		fullData.GetTCPRouterCount(),
		fullData.GetUDPRouterCount(),
		fullData.GetTotalServiceCount(),
		fullData.GetTotalMiddlewareCount())

	return resources, nil
}

// fetchAllEndpointsConcurrently fetches multiple Traefik API endpoints in parallel
// This pattern is inspired by Mantrae's concurrent fetching approach
func (f *TraefikFetcher) fetchAllEndpointsConcurrently(ctx context.Context, baseURL string) (*models.FullTraefikData, error) {
	// Full list of endpoints following Mantrae pattern
	endpoints := map[string]string{
		// HTTP Protocol
		"http_routers":     "/api/http/routers",
		"http_services":    "/api/http/services",
		"http_middlewares": "/api/http/middlewares",
		// TCP Protocol
		"tcp_routers":     "/api/tcp/routers",
		"tcp_services":    "/api/tcp/services",
		"tcp_middlewares": "/api/tcp/middlewares",
		// UDP Protocol
		"udp_routers":  "/api/udp/routers",
		"udp_services": "/api/udp/services",
		// Metadata
		"overview":    "/api/overview",
		"version":     "/api/version",
		"entrypoints": "/api/entrypoints",
	}

	// Buffered channel to collect results
	results := make(chan fetchResult, len(endpoints))

	// Launch concurrent fetches
	var wg sync.WaitGroup
	for name, path := range endpoints {
		wg.Add(1)
		go func(name, path string) {
			defer wg.Done()
			data, err := f.fetch(ctx, baseURL+path)
			results <- fetchResult{name: name, data: data, err: err}
		}(name, path)
	}

	// Wait for all fetches and close channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	response := &models.FullTraefikData{}
	var criticalErrors []string
	var nonCriticalErrors []string

	for result := range results {
		if result.err != nil {
			// Categorize errors: HTTP routers and overview are critical
			if result.name == "http_routers" || result.name == "version" {
				criticalErrors = append(criticalErrors, fmt.Sprintf("%s: %v", result.name, result.err))
			} else {
				nonCriticalErrors = append(nonCriticalErrors, fmt.Sprintf("%s: %v", result.name, result.err))
				log.Printf("Warning: non-critical endpoint failed: %s: %v", result.name, result.err)
			}
			continue
		}

		// Decode based on endpoint type
		switch result.name {
		// HTTP Protocol
		case "http_routers":
			routers, err := f.decodeHTTPRouters(result.data)
			if err != nil {
				criticalErrors = append(criticalErrors, fmt.Sprintf("http_routers decode: %v", err))
			} else {
				response.HTTPRouters = routers
			}

		case "http_services":
			services, err := f.decodeHTTPServices(result.data)
			if err != nil {
				log.Printf("Warning: failed to decode http_services: %v", err)
			} else {
				response.HTTPServices = services
			}

		case "http_middlewares":
			middlewares, err := f.decodeHTTPMiddlewares(result.data)
			if err != nil {
				log.Printf("Warning: failed to decode http_middlewares: %v", err)
			} else {
				response.HTTPMiddlewares = middlewares
			}

		// TCP Protocol
		case "tcp_routers":
			routers, err := f.decodeTCPRouters(result.data)
			if err != nil {
				log.Printf("Warning: failed to decode tcp_routers: %v", err)
			} else {
				response.TCPRouters = routers
			}

		case "tcp_services":
			services, err := f.decodeTCPServices(result.data)
			if err != nil {
				log.Printf("Warning: failed to decode tcp_services: %v", err)
			} else {
				response.TCPServices = services
			}

		case "tcp_middlewares":
			middlewares, err := f.decodeTCPMiddlewares(result.data)
			if err != nil {
				log.Printf("Warning: failed to decode tcp_middlewares: %v", err)
			} else {
				response.TCPMiddlewares = middlewares
			}

		// UDP Protocol
		case "udp_routers":
			routers, err := f.decodeUDPRouters(result.data)
			if err != nil {
				log.Printf("Warning: failed to decode udp_routers: %v", err)
			} else {
				response.UDPRouters = routers
			}

		case "udp_services":
			services, err := f.decodeUDPServices(result.data)
			if err != nil {
				log.Printf("Warning: failed to decode udp_services: %v", err)
			} else {
				response.UDPServices = services
			}

		// Metadata
		case "overview":
			var overview models.TraefikOverview
			if err := json.Unmarshal(result.data, &overview); err != nil {
				log.Printf("Warning: failed to decode overview: %v", err)
			} else {
				response.Overview = &overview
			}

		case "version":
			var version models.TraefikVersion
			if err := json.Unmarshal(result.data, &version); err != nil {
				log.Printf("Warning: failed to decode version: %v", err)
			} else {
				response.Version = &version
				log.Printf("Connected to Traefik %s (%s)", version.Version, version.Codename)
			}

		case "entrypoints":
			entrypoints, err := f.decodeEntrypoints(result.data)
			if err != nil {
				log.Printf("Warning: failed to decode entrypoints: %v", err)
			} else {
				response.Entrypoints = entrypoints
			}
		}
	}

	// If critical endpoints failed, return error
	if len(criticalErrors) > 0 {
		return nil, fmt.Errorf("critical endpoints failed: %s", strings.Join(criticalErrors, "; "))
	}

	// Log non-critical errors summary
	if len(nonCriticalErrors) > 0 {
		log.Printf("Note: %d non-critical endpoints failed but continuing with available data", len(nonCriticalErrors))
	}

	return response, nil
}

// fetch performs an HTTP GET request and returns the response body
func (f *TraefikFetcher) fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add basic auth if configured
	if f.config.BasicAuth.Username != "" {
		req.SetBasicAuth(f.config.BasicAuth.Username, f.config.BasicAuth.Password)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Use limited reader to prevent memory issues (10MB limit)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

// Decode functions using generic DecodeArrayOrMap for reduced code duplication

// decodeHTTPRouters decodes HTTP router response
func (f *TraefikFetcher) decodeHTTPRouters(data []byte) ([]models.TraefikRouter, error) {
	return DecodeArrayOrMap[models.TraefikRouter](data, func(r *models.TraefikRouter, name string) {
		r.Name = name
	})
}

// decodeHTTPServices decodes HTTP service response
func (f *TraefikFetcher) decodeHTTPServices(data []byte) ([]models.TraefikService, error) {
	return DecodeArrayOrMap[models.TraefikService](data, func(s *models.TraefikService, name string) {
		s.Name = name
	})
}

// decodeHTTPMiddlewares decodes HTTP middleware response
func (f *TraefikFetcher) decodeHTTPMiddlewares(data []byte) ([]models.TraefikMiddleware, error) {
	return DecodeArrayOrMap[models.TraefikMiddleware](data, func(m *models.TraefikMiddleware, name string) {
		m.Name = name
	})
}

// decodeTCPRouters decodes TCP router response
func (f *TraefikFetcher) decodeTCPRouters(data []byte) ([]models.TCPRouter, error) {
	return DecodeArrayOrMap[models.TCPRouter](data, func(r *models.TCPRouter, name string) {
		r.Name = name
	})
}

// decodeTCPServices decodes TCP service response
func (f *TraefikFetcher) decodeTCPServices(data []byte) ([]models.TCPService, error) {
	return DecodeArrayOrMap[models.TCPService](data, func(s *models.TCPService, name string) {
		s.Name = name
	})
}

// decodeTCPMiddlewares decodes TCP middleware response
func (f *TraefikFetcher) decodeTCPMiddlewares(data []byte) ([]models.TCPMiddleware, error) {
	return DecodeArrayOrMap[models.TCPMiddleware](data, func(m *models.TCPMiddleware, name string) {
		m.Name = name
	})
}

// decodeUDPRouters decodes UDP router response
func (f *TraefikFetcher) decodeUDPRouters(data []byte) ([]models.UDPRouter, error) {
	return DecodeArrayOrMap[models.UDPRouter](data, func(r *models.UDPRouter, name string) {
		r.Name = name
	})
}

// decodeUDPServices decodes UDP service response
func (f *TraefikFetcher) decodeUDPServices(data []byte) ([]models.UDPService, error) {
	return DecodeArrayOrMap[models.UDPService](data, func(s *models.UDPService, name string) {
		s.Name = name
	})
}

// decodeEntrypoints decodes entrypoints response
func (f *TraefikFetcher) decodeEntrypoints(data []byte) ([]models.TraefikEntrypoint, error) {
	return DecodeArrayOrMap[models.TraefikEntrypoint](data, func(e *models.TraefikEntrypoint, name string) {
		e.Name = name
	})
}

// GetTraefikServices returns the last fetched Traefik services
// This allows the UI to display services fetched from Traefik API
func (f *TraefikFetcher) GetTraefikServices(ctx context.Context) ([]models.TraefikService, error) {
	data, err := f.FetchFullData(ctx)
	if err != nil {
		return nil, err
	}
	return data.HTTPServices, nil
}

// GetTraefikMiddlewares returns the last fetched Traefik middlewares
// This allows the UI to display middlewares fetched from Traefik API
func (f *TraefikFetcher) GetTraefikMiddlewares(ctx context.Context) ([]models.TraefikMiddleware, error) {
	data, err := f.FetchFullData(ctx)
	if err != nil {
		return nil, err
	}
	return data.HTTPMiddlewares, nil
}

// GetTraefikRouters returns all routers (HTTP, TCP, UDP can be filtered)
func (f *TraefikFetcher) GetTraefikRouters(ctx context.Context) ([]models.TraefikRouter, error) {
	data, err := f.FetchFullData(ctx)
	if err != nil {
		return nil, err
	}
	return data.HTTPRouters, nil
}

// GetTCPRouters returns TCP routers
func (f *TraefikFetcher) GetTCPRouters(ctx context.Context) ([]models.TCPRouter, error) {
	data, err := f.FetchFullData(ctx)
	if err != nil {
		return nil, err
	}
	return data.TCPRouters, nil
}

// GetUDPRouters returns UDP routers
func (f *TraefikFetcher) GetUDPRouters(ctx context.Context) ([]models.UDPRouter, error) {
	data, err := f.FetchFullData(ctx)
	if err != nil {
		return nil, err
	}
	return data.UDPRouters, nil
}

// GetTCPServices returns TCP services
func (f *TraefikFetcher) GetTCPServices(ctx context.Context) ([]models.TCPService, error) {
	data, err := f.FetchFullData(ctx)
	if err != nil {
		return nil, err
	}
	return data.TCPServices, nil
}

// GetUDPServices returns UDP services
func (f *TraefikFetcher) GetUDPServices(ctx context.Context) ([]models.UDPService, error) {
	data, err := f.FetchFullData(ctx)
	if err != nil {
		return nil, err
	}
	return data.UDPServices, nil
}

// GetTCPMiddlewares returns TCP middlewares
func (f *TraefikFetcher) GetTCPMiddlewares(ctx context.Context) ([]models.TCPMiddleware, error) {
	data, err := f.FetchFullData(ctx)
	if err != nil {
		return nil, err
	}
	return data.TCPMiddlewares, nil
}

// GetOverview returns the Traefik overview
func (f *TraefikFetcher) GetOverview(ctx context.Context) (*models.TraefikOverview, error) {
	data, err := f.FetchFullData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Overview, nil
}

// GetVersion returns the Traefik version
func (f *TraefikFetcher) GetVersion(ctx context.Context) (*models.TraefikVersion, error) {
	data, err := f.FetchFullData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Version, nil
}

// GetEntrypoints returns the Traefik entrypoints
func (f *TraefikFetcher) GetEntrypoints(ctx context.Context) ([]models.TraefikEntrypoint, error) {
	data, err := f.FetchFullData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Entrypoints, nil
}

// suggestURLUpdate logs a message suggesting the URL should be updated
func (f *TraefikFetcher) suggestURLUpdate(workingURL string) {
	log.Printf("IMPORTANT: Consider updating the Traefik API URL to %s in the settings", workingURL)
}

// shouldIncludeNonTLSRouters returns whether non-TLS routers should be included
func shouldIncludeNonTLSRouters() bool {
	return true
}

// isTraefikSystemRouter checks if a router is a Traefik system router (to be skipped)
func isTraefikSystemRouter(routerID string) bool {
	systemPrefixes := []string{
		"api@internal",
		"dashboard@internal",
		"acme-http@internal",
	}

	userPatterns := []string{
		"-router",
		"api-router@file",
		"next-router@file",
		"ws-router@file",
	}

	// First check if it matches any user patterns - if so, don't skip it
	for _, pattern := range userPatterns {
		if strings.Contains(routerID, pattern) {
			return false
		}
	}

	// Then check if it matches any system prefixes
	for _, prefix := range systemPrefixes {
		if strings.Contains(routerID, prefix) {
			return true
		}
	}

	return false
}
