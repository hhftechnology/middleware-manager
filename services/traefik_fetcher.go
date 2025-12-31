package services

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "strings"
    "time"

    "github.com/hhftechnology/middleware-manager/models"
)

// TraefikFetcher fetches resources from Traefik API
type TraefikFetcher struct {
    config     models.DataSourceConfig
    httpClient *http.Client
}

// NewTraefikFetcher creates a new Traefik API fetcher
func NewTraefikFetcher(config models.DataSourceConfig) *TraefikFetcher {
    return &TraefikFetcher{
        config: config,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

// FetchResources fetches resources from Traefik API with fallback options
func (f *TraefikFetcher) FetchResources(ctx context.Context) (*models.ResourceCollection, error) {
    log.Println("Fetching resources from Traefik API...")
    
    // Try the configured URL first
    resources, err := f.fetchResourcesFromURL(ctx, f.config.URL)
    if err == nil {
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
            // Success with fallback - remember this URL for next time
            f.suggestURLUpdate(url)
            return resources, nil
        }
        lastErr = err
        log.Printf("Fallback URL %s failed: %v", url, err)
    }
    
    // All fallbacks failed
    return nil, fmt.Errorf("all Traefik API connection attempts failed, last error: %w", lastErr)
}

// fetchResourcesFromURL fetches resources from a specific URL
func (f *TraefikFetcher) fetchResourcesFromURL(ctx context.Context, baseURL string) (*models.ResourceCollection, error) {
    // Create HTTP request to fetch HTTP routers
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/http/routers", nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Add basic auth if configured
    if f.config.BasicAuth.Username != "" {
        req.SetBasicAuth(f.config.BasicAuth.Username, f.config.BasicAuth.Password)
    }
    
    // Execute request
    resp, err := f.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("HTTP request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Check status code
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    
    // Read and parse response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    // Parse the Traefik routers response
    var traefikRouters []models.TraefikRouter
    if err := json.Unmarshal(body, &traefikRouters); err != nil {
        // Try parsing as a map if array unmarshaling fails (Traefik API might return different formats)
        var routersMap map[string]models.TraefikRouter
        if jsonErr := json.Unmarshal(body, &routersMap); jsonErr != nil {
            return nil, fmt.Errorf("failed to parse routers JSON: %w", err)
        }
        
        // Convert map to array
        for name, router := range routersMap {
            router.Name = name // Set the name from the map key
            traefikRouters = append(traefikRouters, router)
        }
    }
    
    // Convert Traefik routers to our internal model
    resources := &models.ResourceCollection{
        Resources: make([]models.Resource, 0),
    }
    
    // Get TLS domains for routers by making a separate request to the Traefik API
    tlsDomainsMap, err := f.fetchTLSDomains(ctx, baseURL)
    if err != nil {
        log.Printf("Warning: Failed to fetch TLS domains: %v", err)
        // Continue without TLS domains, as this is not critical
    }
    
    for _, router := range traefikRouters {
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
            // Use domains from the router if available
            resource.TLSDomains = models.JoinTLSDomains(router.TLS.Domains)
        }
        
        resources.Resources = append(resources.Resources, resource)
    }
    
    log.Printf("Fetched %d resources from Traefik API", len(resources.Resources))
    return resources, nil
}

// suggestURLUpdate logs a message suggesting the URL should be updated
func (f *TraefikFetcher) suggestURLUpdate(workingURL string) {
    log.Printf("IMPORTANT: Consider updating the Traefik API URL to %s in the settings", workingURL)
}

// fetchTLSDomains fetches TLS configuration for routers from Traefik API
func (f *TraefikFetcher) fetchTLSDomains(ctx context.Context, baseURL string) (map[string]string, error) {
    // Create HTTP request to fetch TLS configuration
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/http/routers", nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create TLS domains request: %w", err)
    }
    
    // Add basic auth if configured
    if f.config.BasicAuth.Username != "" {
        req.SetBasicAuth(f.config.BasicAuth.Username, f.config.BasicAuth.Password)
    }
    
    // Execute request
    resp, err := f.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("TLS domains HTTP request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Check status code
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("TLS domains unexpected status code: %d", resp.StatusCode)
    }
    
    // Read and parse response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read TLS domains response: %w", err)
    }
    
    // First try to parse as a map of routers
    var routersMap map[string]models.TraefikRouter
    if err := json.Unmarshal(body, &routersMap); err != nil {
        // If map parsing fails, try to parse as an array
        var routersArray []models.TraefikRouter
        if jsonErr := json.Unmarshal(body, &routersArray); jsonErr != nil {
            return nil, fmt.Errorf("failed to parse TLS domains JSON: %w", err)
        }
        
        // Extract TLS domains for each router from the array
        domainsMap := make(map[string]string)
        for _, router := range routersArray {
            if len(router.TLS.Domains) > 0 && router.Name != "" {
                domainsMap[router.Name] = models.JoinTLSDomains(router.TLS.Domains)
            }
        }
        return domainsMap, nil
    }
    
    // Extract TLS domains for each router from the map
    domainsMap := make(map[string]string)
    for name, router := range routersMap {
        if len(router.TLS.Domains) > 0 {
            domainsMap[name] = models.JoinTLSDomains(router.TLS.Domains)
        }
    }
    
    return domainsMap, nil
}

// fetchTCPRouters fetches TCP routers from Traefik API
func (f *TraefikFetcher) fetchTCPRouters(ctx context.Context, baseURL string) ([]models.TraefikRouter, error) {
    // Create HTTP request to fetch TCP routers
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/tcp/routers", nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create TCP routers request: %w", err)
    }
    
    // Add basic auth if configured
    if f.config.BasicAuth.Username != "" {
        req.SetBasicAuth(f.config.BasicAuth.Username, f.config.BasicAuth.Password)
    }
    
    // Execute request
    resp, err := f.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("TCP routers HTTP request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Check status code
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("TCP routers unexpected status code: %d", resp.StatusCode)
    }
    
    // Read and parse response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read TCP routers response: %w", err)
    }
    
    // Try to parse as an array of routers
    var tcpRouters []models.TraefikRouter
    if err := json.Unmarshal(body, &tcpRouters); err != nil {
        // Try parsing as a map if array unmarshaling fails
        var routersMap map[string]models.TraefikRouter
        if jsonErr := json.Unmarshal(body, &routersMap); jsonErr != nil {
            return nil, fmt.Errorf("failed to parse TCP routers JSON: %w", err)
        }
        
        // Convert map to array
        for name, router := range routersMap {
            router.Name = name // Set the name from the map key
            tcpRouters = append(tcpRouters, router)
        }
    }
    
    return tcpRouters, nil
}

// shouldIncludeNonTLSRouters returns whether non-TLS routers should be included
// This could be made configurable through system settings
func shouldIncludeNonTLSRouters() bool {
    return true // Changed to true to include non-TLS routers
}

// isTraefikSystemRouter checks if a router is a Traefik system router (to be skipped)
func isTraefikSystemRouter(routerID string) bool {
    // Keep original system prefixes but improve matching
    systemPrefixes := []string{
        "api@internal",
        "dashboard@internal",
        "acme-http@internal",
    }
    
    // But allow these specific patterns that are user routers and not internal system routers
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