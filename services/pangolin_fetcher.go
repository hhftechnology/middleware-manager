package services

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
    "time"
    
    "github.com/hhftechnology/middleware-manager/models"
)

// PangolinFetcher fetches resources from Pangolin API
type PangolinFetcher struct {
    config     models.DataSourceConfig
    httpClient *http.Client
}

// NewPangolinFetcher creates a new Pangolin API fetcher
func NewPangolinFetcher(config models.DataSourceConfig) *PangolinFetcher {
    return &PangolinFetcher{
        config: config,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

// FetchResources fetches resources from Pangolin API
func (f *PangolinFetcher) FetchResources(ctx context.Context) (*models.ResourceCollection, error) {
    // Create HTTP request
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.config.URL+"/traefik-config", nil)
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
    
    // Process response
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    // Parse the Pangolin config
    var config models.PangolinTraefikConfig
    if err := json.Unmarshal(body, &config); err != nil {
        return nil, fmt.Errorf("failed to parse JSON: %w", err)
    }
    
    // Convert Pangolin config to our internal model
    resources := &models.ResourceCollection{
        Resources: make([]models.Resource, 0, len(config.HTTP.Routers)),
    }
    
    for id, router := range config.HTTP.Routers {
        // Skip non-SSL routers (usually HTTP redirects)
        if router.TLS.CertResolver == "" {
            continue
        }
        
        // Extract host from rule
        host := extractHostFromRule(router.Rule)
        if host == "" {
            continue
        }
        
        // Skip system routers
        if isPangolinSystemRouter(id) {
            continue
        }
        
        resource := models.Resource{
            ID:             id,
            Host:           host,
            ServiceID:      router.Service,
            Status:         "active",
            SourceType:     string(models.PangolinAPI),
            Entrypoints:    strings.Join(router.EntryPoints, ","),
            RouterPriority: 200, // Default
        }
        
        resources.Resources = append(resources.Resources, resource)
    }
    
    log.Printf("Fetched %d resources from Pangolin API", len(resources.Resources))
    return resources, nil
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

// Helper function to extract TLS domains into a comma-separated string
// Note: This function is updated to use the model package's function
func extractTLSDomains(domains []models.TraefikTLSDomain) string {
    return models.JoinTLSDomains(domains)
}