package services

import (
	"context"
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

// PangolinFetcher fetches resources from Pangolin API
// Implements the same patterns as TraefikFetcher:
// - Singleflight pattern to prevent duplicate requests
// - Rate limiting between updates
// - Connection pooling via shared HTTP client
type PangolinFetcher struct {
	config       models.DataSourceConfig
	httpClient   *http.Client
	singleflight singleflight.Group
	lastFetch    time.Time
	lastFetchMu  sync.RWMutex
	minInterval  time.Duration

	// Cached data from last fetch
	cachedConfig   *models.PangolinTraefikConfig
	cachedConfigMu sync.RWMutex
}

// NewPangolinFetcher creates a new Pangolin API fetcher with connection pooling
func NewPangolinFetcher(config models.DataSourceConfig) *PangolinFetcher {
	// Use the shared HTTP client pool for better connection reuse
	httpClient := GetHTTPClient()

	return &PangolinFetcher{
		config:      config,
		httpClient:  httpClient,
		minInterval: 5 * time.Second, // Rate limit: minimum 5 seconds between fetches
	}
}

// FetchResources fetches resources from Pangolin API with singleflight deduplication
func (f *PangolinFetcher) FetchResources(ctx context.Context) (*models.ResourceCollection, error) {
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
func (f *PangolinFetcher) fetchResourcesInternal(ctx context.Context) (*models.ResourceCollection, error) {
	// Check rate limiting
	f.lastFetchMu.RLock()
	timeSinceLastFetch := time.Since(f.lastFetch)
	f.lastFetchMu.RUnlock()

	if timeSinceLastFetch < f.minInterval && f.cachedConfig != nil {
		log.Printf("Rate limiting: using cached config, last fetch was %v ago", timeSinceLastFetch)
		return f.convertConfigToResources(f.cachedConfig), nil
	}

	log.Println("Fetching resources from Pangolin API...")

	// Fetch the traefik-config endpoint
	config, err := f.fetchTraefikConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Update cache
	f.cachedConfigMu.Lock()
	f.cachedConfig = config
	f.cachedConfigMu.Unlock()

	// Update last fetch time
	f.lastFetchMu.Lock()
	f.lastFetch = time.Now()
	f.lastFetchMu.Unlock()

	// Convert to resources
	resources := f.convertConfigToResources(config)

	log.Printf("Fetched %d resources, %d services, %d middlewares from Pangolin API",
		len(resources.Resources),
		len(config.HTTP.Services),
		len(config.HTTP.Middlewares))

	return resources, nil
}

// fetchTraefikConfig fetches the complete traefik config from Pangolin
func (f *PangolinFetcher) fetchTraefikConfig(ctx context.Context) (*models.PangolinTraefikConfig, error) {
	url := f.config.URL + "/traefik-config"

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

	// Read response with size limit
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 10MB limit
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var config models.PangolinTraefikConfig
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Initialize empty maps if nil
	if config.HTTP.Routers == nil {
		config.HTTP.Routers = make(map[string]models.PangolinRouter)
	}
	if config.HTTP.Services == nil {
		config.HTTP.Services = make(map[string]models.PangolinService)
	}
	if config.HTTP.Middlewares == nil {
		config.HTTP.Middlewares = make(map[string]map[string]interface{})
	}

	return &config, nil
}

// convertConfigToResources converts Pangolin config to ResourceCollection
func (f *PangolinFetcher) convertConfigToResources(config *models.PangolinTraefikConfig) *models.ResourceCollection {
	resources := &models.ResourceCollection{
		Resources: make([]models.Resource, 0, len(config.HTTP.Routers)),
	}

	for id, router := range config.HTTP.Routers {
		// Extract host from rule
		host := extractHostFromRule(router.Rule)
		if host == "" {
			continue
		}

		// Skip system routers
		if isPangolinSystemRouter(id) {
			continue
		}

		// Use Pangolin's priority if provided, otherwise default to 100
		priority := router.Priority
		if priority == 0 {
			priority = 100
		}

		resource := models.Resource{
			ID:             id,
			Host:           host,
			ServiceID:      router.Service,
			Status:         "active",
			SourceType:     string(models.PangolinAPI),
			Entrypoints:    strings.Join(router.EntryPoints, ","),
			TLSDomains:     models.JoinTLSDomains(router.TLS.Domains),
			RouterPriority: priority,
		}

		resources.Resources = append(resources.Resources, resource)
	}

	return resources
}

// GetTraefikMiddlewares returns middlewares from the cached Pangolin config
// This allows the UI to display middlewares fetched from Pangolin API
func (f *PangolinFetcher) GetTraefikMiddlewares(ctx context.Context) ([]models.TraefikMiddleware, error) {
	// Ensure we have fresh data
	config, err := f.getOrFetchConfig(ctx)
	if err != nil {
		return nil, err
	}

	middlewares := make([]models.TraefikMiddleware, 0, len(config.HTTP.Middlewares))

	for name, middlewareConfig := range config.HTTP.Middlewares {
		// Determine middleware type from the config structure
		middlewareType := detectMiddlewareType(middlewareConfig)

		middleware := models.TraefikMiddleware{
			Name:     name,
			Type:     middlewareType,
			Provider: "pangolin",
			Status:   "enabled",
			Config:   middlewareConfig,
		}

		middlewares = append(middlewares, middleware)
	}

	return middlewares, nil
}

// GetTraefikServices returns services from the cached Pangolin config
// This allows the UI to display services fetched from Pangolin API
func (f *PangolinFetcher) GetTraefikServices(ctx context.Context) ([]models.TraefikService, error) {
	// Ensure we have fresh data
	config, err := f.getOrFetchConfig(ctx)
	if err != nil {
		return nil, err
	}

	services := make([]models.TraefikService, 0, len(config.HTTP.Services))

	for name, service := range config.HTTP.Services {
		traefikService := models.TraefikService{
			Name:     name,
			Provider: "pangolin",
		}

		// Convert PangolinService to TraefikService format
		// The PangolinService uses interface{} for flexibility
		if service.LoadBalancer != nil {
			traefikService.LoadBalancer = convertToLoadBalancer(service.LoadBalancer)
		}
		if service.Weighted != nil {
			traefikService.Weighted = convertToWeighted(service.Weighted)
		}
		if service.Mirroring != nil {
			traefikService.Mirroring = convertToMirroring(service.Mirroring)
		}
		if service.Failover != nil {
			traefikService.Failover = convertToFailover(service.Failover)
		}

		services = append(services, traefikService)
	}

	return services, nil
}

// GetTraefikRouters returns routers from the cached Pangolin config
func (f *PangolinFetcher) GetTraefikRouters(ctx context.Context) ([]models.TraefikRouter, error) {
	// Ensure we have fresh data
	config, err := f.getOrFetchConfig(ctx)
	if err != nil {
		return nil, err
	}

	routers := make([]models.TraefikRouter, 0, len(config.HTTP.Routers))

	for name, router := range config.HTTP.Routers {
		traefikRouter := models.TraefikRouter{
			Name:        name,
			Rule:        router.Rule,
			Service:     router.Service,
			EntryPoints: router.EntryPoints,
			Middlewares: router.Middlewares,
			Priority:    100, // Default priority
			Provider:    "pangolin",
			Status:      "enabled",
			TLS: models.TraefikTLSConfig{
				CertResolver: router.TLS.CertResolver,
			},
		}

		routers = append(routers, traefikRouter)
	}

	return routers, nil
}

// getOrFetchConfig returns cached config or fetches fresh data
func (f *PangolinFetcher) getOrFetchConfig(ctx context.Context) (*models.PangolinTraefikConfig, error) {
	f.cachedConfigMu.RLock()
	config := f.cachedConfig
	f.cachedConfigMu.RUnlock()

	if config != nil {
		return config, nil
	}

	// Need to fetch fresh data
	_, err := f.FetchResources(ctx)
	if err != nil {
		return nil, err
	}

	f.cachedConfigMu.RLock()
	config = f.cachedConfig
	f.cachedConfigMu.RUnlock()

	if config == nil {
		return nil, fmt.Errorf("failed to fetch config")
	}

	return config, nil
}

// detectMiddlewareType determines the middleware type from its config
func detectMiddlewareType(config map[string]interface{}) string {
	// Check for known middleware type keys
	typeKeys := []string{
		"basicAuth", "digestAuth", "forwardAuth",
		"ipAllowList",
		"rateLimit", "headers",
		"stripPrefix", "stripPrefixRegex",
		"addPrefix", "redirectRegex", "redirectScheme",
		"replacePath", "replacePathRegex",
		"chain", "plugin",
		"buffering", "circuitBreaker", "compress",
		"contentType", "errors", "grpcWeb",
		"inFlightReq", "passTLSClientCert", "retry",
	}

	for _, key := range typeKeys {
		if _, exists := config[key]; exists {
			return key
		}
	}

	return "unknown"
}

// convertToLoadBalancer converts interface{} to LoadBalancer struct
func convertToLoadBalancer(data interface{}) *struct {
	Servers []struct {
		URL     string `json:"url,omitempty"`
		Address string `json:"address,omitempty"`
		Weight  *int   `json:"weight,omitempty"`
	} `json:"servers,omitempty"`
	PassHostHeader *bool       `json:"passHostHeader,omitempty"`
	Sticky         interface{} `json:"sticky,omitempty"`
	HealthCheck    interface{} `json:"healthCheck,omitempty"`
} {
	if data == nil {
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	var lb struct {
		Servers []struct {
			URL     string `json:"url,omitempty"`
			Address string `json:"address,omitempty"`
			Weight  *int   `json:"weight,omitempty"`
		} `json:"servers,omitempty"`
		PassHostHeader *bool       `json:"passHostHeader,omitempty"`
		Sticky         interface{} `json:"sticky,omitempty"`
		HealthCheck    interface{} `json:"healthCheck,omitempty"`
	}

	if err := json.Unmarshal(jsonData, &lb); err != nil {
		return nil
	}

	return &lb
}

// convertToWeighted converts interface{} to Weighted struct
func convertToWeighted(data interface{}) *struct {
	Services []struct {
		Name   string `json:"name"`
		Weight int    `json:"weight"`
	} `json:"services,omitempty"`
	Sticky      interface{} `json:"sticky,omitempty"`
	HealthCheck interface{} `json:"healthCheck,omitempty"`
} {
	if data == nil {
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	var w struct {
		Services []struct {
			Name   string `json:"name"`
			Weight int    `json:"weight"`
		} `json:"services,omitempty"`
		Sticky      interface{} `json:"sticky,omitempty"`
		HealthCheck interface{} `json:"healthCheck,omitempty"`
	}

	if err := json.Unmarshal(jsonData, &w); err != nil {
		return nil
	}

	return &w
}

// convertToMirroring converts interface{} to Mirroring struct
func convertToMirroring(data interface{}) *struct {
	Service string `json:"service"`
	Mirrors []struct {
		Name    string `json:"name"`
		Percent int    `json:"percent"`
	} `json:"mirrors,omitempty"`
	MaxBodySize *int        `json:"maxBodySize,omitempty"`
	MirrorBody  *bool       `json:"mirrorBody,omitempty"`
	HealthCheck interface{} `json:"healthCheck,omitempty"`
} {
	if data == nil {
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	var m struct {
		Service string `json:"service"`
		Mirrors []struct {
			Name    string `json:"name"`
			Percent int    `json:"percent"`
		} `json:"mirrors,omitempty"`
		MaxBodySize *int        `json:"maxBodySize,omitempty"`
		MirrorBody  *bool       `json:"mirrorBody,omitempty"`
		HealthCheck interface{} `json:"healthCheck,omitempty"`
	}

	if err := json.Unmarshal(jsonData, &m); err != nil {
		return nil
	}

	return &m
}

// convertToFailover converts interface{} to Failover struct
func convertToFailover(data interface{}) *struct {
	Service     string      `json:"service"`
	Fallback    string      `json:"fallback"`
	HealthCheck interface{} `json:"healthCheck,omitempty"`
} {
	if data == nil {
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	var fo struct {
		Service     string      `json:"service"`
		Fallback    string      `json:"fallback"`
		HealthCheck interface{} `json:"healthCheck,omitempty"`
	}

	if err := json.Unmarshal(jsonData, &fo); err != nil {
		return nil
	}

	return &fo
}

// isPangolinSystemRouter checks if a router is a Pangolin system router (to be skipped)
func isPangolinSystemRouter(routerID string) bool {
	systemPrefixes := []string{
		"api-router",
		"next-router",
		"ws-router",
	}

	for _, prefix := range systemPrefixes {
		if strings.Contains(routerID, prefix) {
			return true
		}
	}

	return false
}
