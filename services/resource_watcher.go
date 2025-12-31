package services

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "strings"
    "time"

    "github.com/hhftechnology/middleware-manager/database"
    "github.com/hhftechnology/middleware-manager/models"
    "github.com/hhftechnology/middleware-manager/util"
)

// ResourceWatcher watches for resources using configured data source
type ResourceWatcher struct {
    db              *database.DB
    fetcher         ResourceFetcher
    configManager   *ConfigManager
    stopChan        chan struct{}
    isRunning       bool
    httpClient      *http.Client
}

// NewResourceWatcher creates a new resource watcher
func NewResourceWatcher(db *database.DB, configManager *ConfigManager) (*ResourceWatcher, error) {
    // Get the active data source config
    dsConfig, err := configManager.GetActiveDataSourceConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to get active data source config: %w", err)
    }

    // Create the fetcher
    fetcher, err := NewResourceFetcher(dsConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create resource fetcher: %w", err)
    }

    // Use the shared HTTP client pool for better connection reuse
    httpClient := GetHTTPClient()

    return &ResourceWatcher{
        db:             db,
        fetcher:        fetcher,
        configManager:  configManager,
        stopChan:       make(chan struct{}),
        isRunning:      false,
        httpClient:     httpClient,
    }, nil
}

// Start begins watching for resources
func (rw *ResourceWatcher) Start(interval time.Duration) {
    if rw.isRunning {
        return
    }
    
    rw.isRunning = true
    log.Printf("Resource watcher started, checking every %v", interval)

    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    // Do an initial check
    if err := rw.checkResources(); err != nil {
        log.Printf("Initial resource check failed: %v", err)
    }

    for {
        select {
        case <-ticker.C:
            // Check if data source config has changed
            if err := rw.refreshFetcher(); err != nil {
                log.Printf("Failed to refresh resource fetcher: %v", err)
            }
            
            if err := rw.checkResources(); err != nil {
                log.Printf("Resource check failed: %v", err)
            }
        case <-rw.stopChan:
            log.Println("Resource watcher stopped")
            return
        }
    }
}

// refreshFetcher updates the fetcher if the data source config has changed
func (rw *ResourceWatcher) refreshFetcher() error {
    dsConfig, err := rw.configManager.GetActiveDataSourceConfig()
    if err != nil {
        return fmt.Errorf("failed to get data source config: %w", err)
    }
    
    // Create a new fetcher with the updated config
    fetcher, err := NewResourceFetcher(dsConfig)
    if err != nil {
        return fmt.Errorf("failed to create resource fetcher: %w", err)
    }
    
    // Update the fetcher
    rw.fetcher = fetcher
    return nil
}

// Stop stops the resource watcher
func (rw *ResourceWatcher) Stop() {
    if !rw.isRunning {
        return
    }
    
    close(rw.stopChan)
    rw.isRunning = false
}

// checkResources fetches resources from the configured data source and updates the database
func (rw *ResourceWatcher) checkResources() error {
    log.Println("Checking for resources using configured data source...")
    
    // Create a context with timeout for the operation
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Fetch resources using the configured fetcher
    resources, err := rw.fetcher.FetchResources(ctx)
    if err != nil {
        return fmt.Errorf("failed to fetch resources: %w", err)
    }

    // Get all existing resources from the database
    var existingResources []string
    rows, err := rw.db.Query("SELECT id FROM resources WHERE status = 'active'")
    if err != nil {
        return fmt.Errorf("failed to query existing resources: %w", err)
    }
    
    for rows.Next() {
        var id string
        if err := rows.Scan(&id); err != nil {
            log.Printf("Error scanning resource ID: %v", err)
            continue
        }
        existingResources = append(existingResources, id)
    }
    rows.Close()
    
    // Keep track of resources we find
    foundResources := make(map[string]bool)

    // Check if there are any resources
    if len(resources.Resources) == 0 {
        log.Println("No resources found in data source")
        // Mark all existing resources as disabled since there are no active resources
        for _, resourceID := range existingResources {
            log.Printf("No active resources, marking resource %s as disabled", resourceID)
            _, err := rw.db.Exec(
                "UPDATE resources SET status = 'disabled', updated_at = ? WHERE id = ?",
                time.Now(), resourceID,
            )
            if err != nil {
                log.Printf("Error marking resource as disabled: %v", err)
            }
        }
        return nil
    }

    // Build a map of normalized IDs to original resources
    normalizedMap := make(map[string]models.Resource)
    // Process resources
    for _, resource := range resources.Resources {
        // Skip invalid resources
        if resource.Host == "" || resource.ServiceID == "" {
            continue
        }

        normalizedID := util.NormalizeID(resource.ID)
        normalizedMap[normalizedID] = resource
        
        // Process resource
        if err := rw.updateOrCreateResource(resource); err != nil {
            log.Printf("Error processing resource %s: %v", resource.ID, err)
            // Continue processing other resources even if one fails
            continue
        }
        
        // Mark this resource as found (using normalized ID)
        foundResources[normalizedID] = true
    }
    
    // Mark resources as disabled if they no longer exist in the data source
    for _, resourceID := range existingResources {
        normalizedID := util.NormalizeID(resourceID)
        if !foundResources[normalizedID] {
            log.Printf("Resource %s no longer exists, marking as disabled", resourceID)
            _, err := rw.db.Exec(
                "UPDATE resources SET status = 'disabled', updated_at = ? WHERE id = ?",
                time.Now(), resourceID,
            )
            if err != nil {
                log.Printf("Error marking resource as disabled: %v", err)
            }
        }
    }
    
    return nil
}

// updateOrCreateResource updates an existing resource or creates a new one
func (rw *ResourceWatcher) updateOrCreateResource(resource models.Resource) error {
    // Use our centralized normalization function
    normalizedID := util.NormalizeID(resource.ID)
    
    // For logging purposes, keep track if we normalized the ID
    originalID := resource.ID
    wasNormalized := normalizedID != originalID
    
    if wasNormalized {
        log.Printf("Normalized resource ID from %s to %s", originalID, normalizedID)
    }
    
    // First try exact match with the normalized ID
    var exists int
    var status string
    var entrypoints, tlsDomains, tcpEntrypoints, tcpSNIRule, customHeaders string
    var tcpEnabled int
    var routerPriority sql.NullInt64
    
    err := rw.db.QueryRow(`
        SELECT 1, status, entrypoints, tls_domains, tcp_enabled, tcp_entrypoints, tcp_sni_rule, 
               custom_headers, router_priority
        FROM resources WHERE id = ?
    `, normalizedID).Scan(&exists, &status, &entrypoints, &tlsDomains, &tcpEnabled, 
                       &tcpEntrypoints, &tcpSNIRule, &customHeaders, &routerPriority)
    
    if err == nil {
        // Resource exists with normalized ID, update it
        return rw.updateExistingResource(normalizedID, resource, status)
    }
    
    // If not found with normalized ID, try with original ID
    if normalizedID != originalID {
        err = rw.db.QueryRow(`
            SELECT 1, status, entrypoints, tls_domains, tcp_enabled, tcp_entrypoints, tcp_sni_rule, 
                   custom_headers, router_priority
            FROM resources WHERE id = ?
        `, originalID).Scan(&exists, &status, &entrypoints, &tlsDomains, &tcpEnabled, 
                         &tcpEntrypoints, &tcpSNIRule, &customHeaders, &routerPriority)
        
        if err == nil {
            // Resource exists with original ID, update it
            return rw.updateExistingResource(originalID, resource, status)
        }
    }
    
    // If still not found, try to find a resource with a similar normalized pattern
    var existingID string
    err = rw.db.QueryRow(`
        SELECT id FROM resources 
        WHERE id LIKE ? OR id LIKE ? 
        LIMIT 1
    `, normalizedID+"%", originalID+"%").Scan(&existingID)
    
    if err == nil {
        // Found a similar resource
        log.Printf("Found resource via pattern matching: %s matches pattern %s", 
                 existingID, normalizedID+"%")
        
        // Get its status
        err = rw.db.QueryRow("SELECT status FROM resources WHERE id = ?", 
                           existingID).Scan(&status)
        
        if err == nil {
            // Update the resource using the existing ID
            return rw.updateExistingResource(existingID, resource, status)
        }
    }
    
    // No existing resource found, create a new one
    return rw.createNewResource(resource, normalizedID, wasNormalized)
}

// updateExistingResource updates an existing resource by ID
func (rw *ResourceWatcher) updateExistingResource(id string, resource models.Resource, status string) error {
    // Use a transaction for the update
    return rw.db.WithTransaction(func(tx *sql.Tx) error {
        log.Printf("Updating resource %s using existing ID %s in database", resource.ID, id)
        
        // Update essential fields but preserve custom configuration
        _, err := tx.Exec(`
            UPDATE resources 
            SET host = ?, service_id = ?, status = 'active', 
                source_type = ?, updated_at = ? 
            WHERE id = ?
        `, resource.Host, resource.ServiceID, resource.SourceType, time.Now(), id)
        
        if err != nil {
            return fmt.Errorf("failed to update resource %s: %w", id, err)
        }
        
        if status == "disabled" {
            log.Printf("Resource %s was disabled but is now active again", id)
        }
        
        return nil
    })
}

// createNewResource creates a new resource in the database
func (rw *ResourceWatcher) createNewResource(resource models.Resource, normalizedID string, wasNormalized bool) error {
    // Set default values for new resources
    if resource.Entrypoints == "" {
        resource.Entrypoints = "websecure"
    }
    
    if resource.OrgID == "" {
        resource.OrgID = "unknown"
    }
    
    if resource.SiteID == "" {
        resource.SiteID = "unknown"
    }
    
    tcpEnabledValue := 0
    if resource.TCPEnabled {
        tcpEnabledValue = 1
    }
    
    // Use default router priority if not set
    if resource.RouterPriority == 0 {
        resource.RouterPriority = 200 // Default priority
    }
    
    // Use a transaction for the insert
    return rw.db.WithTransaction(func(tx *sql.Tx) error {
        // For new resources, always use the normalized ID to prevent duplication
        resourceID := resource.ID
        if wasNormalized {
            log.Printf("Creating new resource with normalized ID: %s (was %s)", normalizedID, resource.ID)
            resourceID = normalizedID
        }
        
        // Try to create with the ideal ID first
        log.Printf("Adding new resource: %s (%s)", resource.Host, resourceID)
        
        result, err := tx.Exec(`
            INSERT INTO resources (
                id, host, service_id, org_id, site_id, status, source_type,
                entrypoints, tls_domains, tcp_enabled, tcp_entrypoints, tcp_sni_rule,
                custom_headers, router_priority, created_at, updated_at
            ) VALUES (?, ?, ?, ?, ?, 'active', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `, resourceID, resource.Host, resource.ServiceID, resource.OrgID, resource.SiteID,
           resource.SourceType, resource.Entrypoints, resource.TLSDomains, tcpEnabledValue,
           resource.TCPEntrypoints, resource.TCPSNIRule, resource.CustomHeaders, 
           resource.RouterPriority, time.Now(), time.Now())
        
        if err != nil {
            // Check if it's a duplicate key error
            if strings.Contains(err.Error(), "UNIQUE constraint") {
                // Try with a different ID format (append -auth if it's a router)
                if strings.Contains(resourceID, "-router") && !strings.Contains(resourceID, "-auth") {
                    alternativeID := resourceID + "-auth"
                    log.Printf("Encountered duplicate, trying alternative ID: %s", alternativeID)
                    
                    result, err = tx.Exec(`
                        INSERT INTO resources (
                            id, host, service_id, org_id, site_id, status, source_type,
                            entrypoints, tls_domains, tcp_enabled, tcp_entrypoints, tcp_sni_rule,
                            custom_headers, router_priority, created_at, updated_at
                        ) VALUES (?, ?, ?, ?, ?, 'active', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                    `, alternativeID, resource.Host, resource.ServiceID, resource.OrgID, resource.SiteID,
                       resource.SourceType, resource.Entrypoints, resource.TLSDomains, tcpEnabledValue,
                       resource.TCPEntrypoints, resource.TCPSNIRule, resource.CustomHeaders, 
                       resource.RouterPriority, time.Now(), time.Now())
                    
                    if err != nil {
                        return fmt.Errorf("failed to create resource with alternative ID %s: %w", alternativeID, err)
                    }
                    
                    log.Printf("Added new resource with alternative ID: %s (%s)", resource.Host, alternativeID)
                    return nil
                }
                
                return fmt.Errorf("failed to create resource due to ID conflict: %w", err)
            }
            
            return fmt.Errorf("failed to create resource %s: %w", resourceID, err)
        }
        rowsAffected, err := result.RowsAffected()
if err != nil {
    log.Printf("Error getting rows affected: %v", err)
} else if rowsAffected > 0 {
    log.Printf("Successfully updated/inserted %d rows", rowsAffected)
}
        log.Printf("Added new resource: %s (%s)", resource.Host, resourceID)
        return nil
    })
}

// fetchTraefikConfig fetches the Traefik configuration from the data source
func (rw *ResourceWatcher) fetchTraefikConfig(ctx context.Context) (*models.PangolinTraefikConfig, error) {
    // Get the active data source config
    dsConfig, err := rw.configManager.GetActiveDataSourceConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to get data source config: %w", err)
    }
    
    // Build the URL based on data source type
    var url string
    if dsConfig.Type == models.PangolinAPI {
        url = fmt.Sprintf("%s/traefik-config", dsConfig.URL)
    } else {
        return nil, fmt.Errorf("unsupported data source type for this operation: %s", dsConfig.Type)
    }
    
    // Create a request with context
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Add basic auth if configured
    if dsConfig.BasicAuth.Username != "" {
        req.SetBasicAuth(dsConfig.BasicAuth.Username, dsConfig.BasicAuth.Password)
    }
    
    // Make the request
    resp, err := rw.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("HTTP request failed: %w", err)
    }
    defer resp.Body.Close()

    // Check status code
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("HTTP request returned status %d", resp.StatusCode)
    }

    // Read response body with a limit to prevent memory issues
    body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 10MB limit
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    // Parse JSON
    var config models.PangolinTraefikConfig
    if err := json.Unmarshal(body, &config); err != nil {
        return nil, fmt.Errorf("failed to parse JSON: %w", err)
    }

    // Initialize empty maps if they're nil to prevent nil pointer dereferences
    if config.HTTP.Routers == nil {
        config.HTTP.Routers = make(map[string]models.PangolinRouter)
    }
    if config.HTTP.Services == nil {
        config.HTTP.Services = make(map[string]models.PangolinService)
    }

    return &config, nil
}

// isSystemRouter checks if a router is a system router (to be skipped)
func isSystemRouter(routerID string) bool {
    systemPrefixes := []string{
        "api@internal",
        "dashboard@internal",
        "acme-http@internal",
        "noop@internal",
    }
    
    // Check exact internal system routers
    for _, prefix := range systemPrefixes {
        if routerID == prefix {
            return true
        }
    }
    
    // Allow user routers with these patterns 
    userPatterns := []string{
        "api-router@file",
        "next-router@file",
        "ws-router@file",
    }
    
    for _, pattern := range userPatterns {
        if strings.Contains(routerID, pattern) {
            return false
        }
    }
    
    // Check other system prefixes
    otherSystemPrefixes := []string{
        "api@",
        "dashboard@",
        "traefik@",
    }
    
    for _, prefix := range otherSystemPrefixes {
        if strings.HasPrefix(routerID, prefix) {
            return true
        }
    }
    
    return false
}