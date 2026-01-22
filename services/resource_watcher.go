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

	"github.com/google/uuid"
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
    
    // Keep track of resources we find (by internal ID)
    foundInternalIDs := make(map[string]bool)

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

    // Process resources
    for _, resource := range resources.Resources {
        // Skip invalid resources
        if resource.Host == "" || resource.ServiceID == "" {
            continue
        }

        // Process resource and get its internal ID
        internalID, err := rw.updateOrCreateResource(resource)
        if err != nil {
            log.Printf("Error processing resource %s: %v", resource.ID, err)
            // Continue processing other resources even if one fails
            continue
        }
        
        // Mark this internal resource ID as found
        foundInternalIDs[internalID] = true
    }
    
    // Mark resources as disabled if they no longer exist in the data source
    // Now we compare internal UUIDs, which is correct
    for _, resourceID := range existingResources {
        if !foundInternalIDs[resourceID] {
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
// Uses internal UUID for stable tracking, pangolin_router_id for Pangolin reference
// Returns the internal UUID of the resource
func (rw *ResourceWatcher) updateOrCreateResource(resource models.Resource) (string, error) {
    pangolinRouterID := util.NormalizeID(resource.ID)

    // Step 1: Try to find existing resource by pangolin_router_id
    var internalID, status string
    err := rw.db.QueryRow(`
        SELECT id, status FROM resources
        WHERE pangolin_router_id = ? AND status = 'active'
    `, pangolinRouterID).Scan(&internalID, &status)

    if err == nil {
        // Found by pangolin_router_id - update it
        log.Printf("Found resource by pangolin_router_id %s (internal: %s)", pangolinRouterID, internalID)
        if err := rw.updateExistingResourceByInternalID(internalID, pangolinRouterID, resource); err != nil {
            return "", err
        }
        return internalID, nil
    }

    // Step 2: Try to find by host (handles Pangolin router ID changes)
    err = rw.db.QueryRow(`
        SELECT id, status FROM resources
        WHERE host = ? AND status = 'active'
    `, resource.Host).Scan(&internalID, &status)

    if err == nil {
        // Found by host - Pangolin changed the router ID, just update pangolin_router_id
        log.Printf("Found resource by host %s (internal: %s), updating pangolin_router_id from old to %s",
            resource.Host, internalID, pangolinRouterID)
        if err := rw.updateExistingResourceByInternalID(internalID, pangolinRouterID, resource); err != nil {
            return "", err
        }
        return internalID, nil
    }

    // Step 3: Check for legacy resources (where id = pangolin_router_id, no internal UUID yet)
    err = rw.db.QueryRow(`
        SELECT id, status FROM resources
        WHERE id = ? OR pangolin_router_id IS NULL AND host = ?
    `, pangolinRouterID, resource.Host).Scan(&internalID, &status)

    if err == nil {
        // Found legacy resource - update it
        log.Printf("Found legacy resource %s, updating", internalID)
        if err := rw.updateExistingResourceByInternalID(internalID, pangolinRouterID, resource); err != nil {
            return "", err
        }
        return internalID, nil
    }

    // Step 4: No existing resource found, create a new one with UUID
    return rw.createNewResourceWithUUID(resource, pangolinRouterID)
}

// updateExistingResourceByInternalID updates an existing resource using its internal UUID
func (rw *ResourceWatcher) updateExistingResourceByInternalID(internalID, pangolinRouterID string, resource models.Resource) error {
    return rw.db.WithTransaction(func(tx *sql.Tx) error {
        log.Printf("Updating resource (internal: %s, pangolin: %s, host: %s)",
            internalID, pangolinRouterID, resource.Host)

        // Update essential fields and pangolin_router_id, preserve custom configuration
        _, err := tx.Exec(`
            UPDATE resources
            SET pangolin_router_id = ?, host = ?, service_id = ?,
                status = 'active', source_type = ?, updated_at = ?
            WHERE id = ?
        `, pangolinRouterID, resource.Host, resource.ServiceID, resource.SourceType, time.Now(), internalID)

        if err != nil {
            return fmt.Errorf("failed to update resource %s: %w", internalID, err)
        }

        // Update router_priority from Pangolin only if not manually overridden
        if resource.RouterPriority > 0 {
            _, err = tx.Exec(`
                UPDATE resources
                SET router_priority = ?
                WHERE id = ? AND COALESCE(router_priority_manual, 0) = 0
            `, resource.RouterPriority, internalID)

            if err != nil {
                log.Printf("Warning: failed to update router_priority for resource %s: %v", internalID, err)
            }
        }

        return nil
    })
}

// createNewResourceWithUUID creates a new resource with a stable internal UUID
// The UUID remains constant even if Pangolin changes the router ID
// Returns the new internal UUID
func (rw *ResourceWatcher) createNewResourceWithUUID(resource models.Resource, pangolinRouterID string) (string, error) {
    // Generate a new UUID for internal tracking
    internalID := uuid.New().String()

    // Set default values for new resources
    entrypoints := resource.Entrypoints
    if entrypoints == "" {
        entrypoints = "websecure"
    }

    orgID := resource.OrgID
    if orgID == "" {
        orgID = "unknown"
    }

    siteID := resource.SiteID
    if siteID == "" {
        siteID = "unknown"
    }

    tcpEnabledValue := 0
    if resource.TCPEnabled {
        tcpEnabledValue = 1
    }

    // Use default router priority if not set
    routerPriority := resource.RouterPriority
    if routerPriority == 0 {
        routerPriority = 100 // Default priority
    }

    err := rw.db.WithTransaction(func(tx *sql.Tx) error {
        log.Printf("Creating new resource: internal=%s, pangolin=%s, host=%s",
            internalID, pangolinRouterID, resource.Host)

        _, err := tx.Exec(`
            INSERT INTO resources (
                id, pangolin_router_id, host, service_id, org_id, site_id, status, source_type,
                entrypoints, tls_domains, tcp_enabled, tcp_entrypoints, tcp_sni_rule,
                custom_headers, router_priority, router_priority_manual, created_at, updated_at
            ) VALUES (?, ?, ?, ?, ?, ?, 'active', ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)
        `, internalID, pangolinRouterID, resource.Host, resource.ServiceID, orgID, siteID,
            resource.SourceType, entrypoints, resource.TLSDomains, tcpEnabledValue,
            resource.TCPEntrypoints, resource.TCPSNIRule, resource.CustomHeaders,
            routerPriority, time.Now(), time.Now())

        if err != nil {
            return fmt.Errorf("failed to create resource (internal=%s, pangolin=%s): %w",
                internalID, pangolinRouterID, err)
        }

        log.Printf("Added new resource: %s (internal: %s, pangolin: %s)",
            resource.Host, internalID, pangolinRouterID)
        return nil
    })

    if err != nil {
        return "", err
    }

    return internalID, nil
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