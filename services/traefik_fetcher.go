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
type TraefikFetcher struct {
	config       models.DataSourceConfig
	httpClient   *http.Client
	singleflight singleflight.Group
	lastFetch    time.Time
	lastFetchMu  sync.RWMutex
	minInterval  time.Duration
}

// TraefikAPIResponse holds the concurrent fetch results
type TraefikAPIResponse struct {
	HTTPRouters     []models.TraefikRouter
	TCPRouters      []models.TraefikRouter
	HTTPServices    []models.TraefikService
	HTTPMiddlewares []models.TraefikMiddleware
	Version         *TraefikVersion
	Entrypoints     []TraefikEntrypoint
}

// TraefikVersion represents version info from Traefik API
type TraefikVersion struct {
	Version   string `json:"version"`
	Codename  string `json:"codename"`
	GoVersion string `json:"goVersion"`
}

// TraefikEntrypoint represents an entrypoint from Traefik API
type TraefikEntrypoint struct {
	Name    string `json:"name"`
	Address string `json:"address"`
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
		Timeout:   10 * time.Second,
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

// updateLastFetch updates the last fetch timestamp
func (f *TraefikFetcher) updateLastFetch() {
	f.lastFetchMu.Lock()
	f.lastFetch = time.Now()
	f.lastFetchMu.Unlock()
}

// fetchResourcesFromURL fetches resources from a specific URL using concurrent fetching
func (f *TraefikFetcher) fetchResourcesFromURL(ctx context.Context, baseURL string) (*models.ResourceCollection, error) {
	// Fetch all endpoints concurrently (like Mantrae pattern)
	apiResponse, err := f.fetchAllEndpointsConcurrently(ctx, baseURL)
	if err != nil {
		return nil, err
	}

	// Convert Traefik routers to our internal model
	resources := &models.ResourceCollection{
		Resources: make([]models.Resource, 0, len(apiResponse.HTTPRouters)),
	}

	// Build TLS domains map from routers
	tlsDomainsMap := make(map[string]string)
	for _, router := range apiResponse.HTTPRouters {
		if len(router.TLS.Domains) > 0 && router.Name != "" {
			tlsDomainsMap[router.Name] = models.JoinTLSDomains(router.TLS.Domains)
		}
	}

	for _, router := range apiResponse.HTTPRouters {
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

	log.Printf("Fetched %d HTTP routers, %d TCP routers, %d services, %d middlewares from Traefik API",
		len(apiResponse.HTTPRouters),
		len(apiResponse.TCPRouters),
		len(apiResponse.HTTPServices),
		len(apiResponse.HTTPMiddlewares))

	return resources, nil
}

// fetchAllEndpointsConcurrently fetches multiple Traefik API endpoints in parallel
// This pattern is inspired by Mantrae's concurrent fetching approach
func (f *TraefikFetcher) fetchAllEndpointsConcurrently(ctx context.Context, baseURL string) (*TraefikAPIResponse, error) {
	endpoints := map[string]string{
		"http_routers":     "/api/http/routers",
		"tcp_routers":      "/api/tcp/routers",
		"http_services":    "/api/http/services",
		"http_middlewares": "/api/http/middlewares",
		"version":          "/api/version",
		"entrypoints":      "/api/entrypoints",
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
	response := &TraefikAPIResponse{}
	var fetchErrors []string

	for result := range results {
		if result.err != nil {
			// Log warning but continue - partial success is acceptable for some endpoints
			log.Printf("Warning: failed to fetch %s: %v", result.name, result.err)
			fetchErrors = append(fetchErrors, fmt.Sprintf("%s: %v", result.name, result.err))
			continue
		}

		// Decode based on endpoint type
		switch result.name {
		case "http_routers":
			routers, err := f.decodeRouters(result.data)
			if err != nil {
				log.Printf("Warning: failed to decode http_routers: %v", err)
			} else {
				response.HTTPRouters = routers
			}

		case "tcp_routers":
			routers, err := f.decodeRouters(result.data)
			if err != nil {
				log.Printf("Warning: failed to decode tcp_routers: %v", err)
			} else {
				response.TCPRouters = routers
			}

		case "http_services":
			services, err := f.decodeServices(result.data)
			if err != nil {
				log.Printf("Warning: failed to decode http_services: %v", err)
			} else {
				response.HTTPServices = services
			}

		case "http_middlewares":
			middlewares, err := f.decodeMiddlewares(result.data)
			if err != nil {
				log.Printf("Warning: failed to decode http_middlewares: %v", err)
			} else {
				response.HTTPMiddlewares = middlewares
			}

		case "version":
			var version TraefikVersion
			if err := json.Unmarshal(result.data, &version); err != nil {
				log.Printf("Warning: failed to decode version: %v", err)
			} else {
				response.Version = &version
				log.Printf("Connected to Traefik %s (%s)", version.Version, version.Codename)
			}

		case "entrypoints":
			var entrypoints []TraefikEntrypoint
			if err := json.Unmarshal(result.data, &entrypoints); err != nil {
				log.Printf("Warning: failed to decode entrypoints: %v", err)
			} else {
				response.Entrypoints = entrypoints
			}
		}
	}

	// If HTTP routers failed, this is a critical error
	if response.HTTPRouters == nil && len(fetchErrors) > 0 {
		return nil, fmt.Errorf("critical endpoints failed: %s", strings.Join(fetchErrors, "; "))
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

// decodeRouters decodes router response (handles both array and map formats)
func (f *TraefikFetcher) decodeRouters(data []byte) ([]models.TraefikRouter, error) {
	// Try array first
	var routers []models.TraefikRouter
	if err := json.Unmarshal(data, &routers); err == nil {
		return routers, nil
	}

	// Try map format
	var routersMap map[string]models.TraefikRouter
	if err := json.Unmarshal(data, &routersMap); err != nil {
		return nil, fmt.Errorf("failed to parse routers: %w", err)
	}

	routers = make([]models.TraefikRouter, 0, len(routersMap))
	for name, router := range routersMap {
		router.Name = name
		routers = append(routers, router)
	}

	return routers, nil
}

// decodeServices decodes service response (handles both array and map formats)
func (f *TraefikFetcher) decodeServices(data []byte) ([]models.TraefikService, error) {
	// Try array first
	var services []models.TraefikService
	if err := json.Unmarshal(data, &services); err == nil {
		return services, nil
	}

	// Try map format
	var servicesMap map[string]models.TraefikService
	if err := json.Unmarshal(data, &servicesMap); err != nil {
		return nil, fmt.Errorf("failed to parse services: %w", err)
	}

	services = make([]models.TraefikService, 0, len(servicesMap))
	for name, service := range servicesMap {
		service.Name = name
		services = append(services, service)
	}

	return services, nil
}

// decodeMiddlewares decodes middleware response (handles both array and map formats)
func (f *TraefikFetcher) decodeMiddlewares(data []byte) ([]models.TraefikMiddleware, error) {
	// Try array first
	var middlewares []models.TraefikMiddleware
	if err := json.Unmarshal(data, &middlewares); err == nil {
		return middlewares, nil
	}

	// Try map format
	var middlewaresMap map[string]models.TraefikMiddleware
	if err := json.Unmarshal(data, &middlewaresMap); err != nil {
		return nil, fmt.Errorf("failed to parse middlewares: %w", err)
	}

	middlewares = make([]models.TraefikMiddleware, 0, len(middlewaresMap))
	for name, middleware := range middlewaresMap {
		middleware.Name = name
		middlewares = append(middlewares, middleware)
	}

	return middlewares, nil
}

// GetTraefikServices returns the last fetched Traefik services
// This allows the UI to display services fetched from Traefik API
func (f *TraefikFetcher) GetTraefikServices(ctx context.Context) ([]models.TraefikService, error) {
	apiResponse, err := f.fetchAllEndpointsConcurrently(ctx, f.config.URL)
	if err != nil {
		return nil, err
	}
	return apiResponse.HTTPServices, nil
}

// GetTraefikMiddlewares returns the last fetched Traefik middlewares
// This allows the UI to display middlewares fetched from Traefik API
func (f *TraefikFetcher) GetTraefikMiddlewares(ctx context.Context) ([]models.TraefikMiddleware, error) {
	apiResponse, err := f.fetchAllEndpointsConcurrently(ctx, f.config.URL)
	if err != nil {
		return nil, err
	}
	return apiResponse.HTTPMiddlewares, nil
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
