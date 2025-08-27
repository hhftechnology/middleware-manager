package database

import (
    "database/sql"
    "fmt"
    "log"
    "strings"
    "time"
    "context"
    "sync"
    
    "github.com/hhftechnology/middleware-manager/util"
)

// CleanupOptions contains options for controlling cleanup operations
type CleanupOptions struct {
    DryRun           bool // If true, logs what would be done without making changes
    LogLevel         int  // 0=errors only, 1=basic info, 2=verbose
    MaxDeleteBatch   int  // Maximum number of items to delete in one batch
    ReapDisabled     bool // If true, physically delete disabled resources
    RecoverCorrupted bool // If true, attempt to recover corrupted resources
}

// DefaultCleanupOptions returns the default cleanup options
func DefaultCleanupOptions() CleanupOptions {
    return CleanupOptions{
        DryRun:           false,
        LogLevel:         1,
        MaxDeleteBatch:   100,
        ReapDisabled:     false,
        RecoverCorrupted: true,
    }
}

// Add this function locally if util package doesn't exist
func normalizeID(id string) string {
    // Extract the base name (everything before the first @)
    baseName := id
    if idx := strings.Index(id, "@"); idx > 0 {
        baseName = id[:idx]
    }
    return baseName
}

// CleanupDuplicateServices - CORRECTED VERSION
func (db *DB) CleanupDuplicateServices(opts CleanupOptions) error {
    if opts.LogLevel >= 1 {
        log.Println("Starting cleanup of duplicate services...")
    }
    
    // Get all services
    rows, err := db.Query("SELECT id, name, type, config FROM services")
    if err != nil {
        return fmt.Errorf("failed to query services: %w", err)
    }
    defer rows.Close()
    
    // Map to track unique base names
    type serviceInfo struct {
        ID     string
        Config string
    }
    uniqueServices := make(map[string]serviceInfo)
    var servicesToDelete []string
    
    // Process each service - COMPLETE the duplicate detection logic
    for rows.Next() {
        var id, name, typ, configStr string
        if err := rows.Scan(&id, &name, &typ, &configStr); err != nil {
            return fmt.Errorf("failed to scan service: %w", err)
        }
        
        // Get normalized ID
        normalizedID := normalizeID(id) // Use local function instead of util.NormalizeID
        
        // If we've already seen this normalized ID, check which one to keep
        if existing, found := uniqueServices[normalizedID]; found {
            // Determine which one to keep:
            // 1. Prefer versions without provider suffixes or with @file suffix
            // 2. If both have same suffix type, keep the one with simpler/shorter ID 
            keepNew := false
            
            existingHasSuffix := strings.Contains(existing.ID, "@")
            newHasSuffix := strings.Contains(id, "@")
            
            if existingHasSuffix && !newHasSuffix {
                // Keep the one without suffix
                keepNew = true
            } else if !existingHasSuffix && newHasSuffix {
                // Keep existing without suffix
                keepNew = false
            } else if strings.HasSuffix(id, "@file") && !strings.HasSuffix(existing.ID, "@file") {
                // Prefer @file suffix
                keepNew = true
            } else if !strings.HasSuffix(id, "@file") && strings.HasSuffix(existing.ID, "@file") {
                // Keep existing with @file
                keepNew = false
            } else {
                // Both have same suffix type, keep the one with simpler ID
                if len(existing.ID) > len(id) {
                    keepNew = true
                }
            }
            
            if keepNew {
                // The new one is better, mark the old one for deletion
                if opts.LogLevel >= 2 {
                    log.Printf("Duplicate found: keeping %s, will delete %s", id, existing.ID)
                }
                servicesToDelete = append(servicesToDelete, existing.ID)
                uniqueServices[normalizedID] = serviceInfo{id, configStr}
            } else {
                // The existing one is better, mark this one for deletion
                if opts.LogLevel >= 2 {
                    log.Printf("Duplicate found: keeping %s, will delete %s", existing.ID, id)
                }
                servicesToDelete = append(servicesToDelete, id)
            }
        } else {
            // First time seeing this normalized ID
            uniqueServices[normalizedID] = serviceInfo{id, configStr}
        }
    }
    
    if err := rows.Err(); err != nil {
        return fmt.Errorf("error iterating services: %w", err)
    }

    if len(servicesToDelete) == 0 {
        if opts.LogLevel >= 1 {
            log.Println("No duplicate services found.")
        }
        return nil
    }

    if opts.DryRun {
        log.Printf("DRY RUN: Would delete %d duplicate services", len(servicesToDelete))
        for _, id := range servicesToDelete {
            log.Printf("  - %s", id)
        }
        return nil
    }

    // Use timeout transaction to prevent indefinite locks
    ctx := context.Background()
    timeout := 30 * time.Second
    
    return db.WithTimeoutTransaction(ctx, timeout, func(tx *sql.Tx) error {
        // Process in smaller batches to reduce lock time
        batchSize := opts.MaxDeleteBatch
        if batchSize <= 0 {
            batchSize = 50 // Default batch size
        }
        
        for i := 0; i < len(servicesToDelete); i += batchSize {
            end := i + batchSize
            if end > len(servicesToDelete) {
                end = len(servicesToDelete)
            }
            
            batch := servicesToDelete[i:end]
            
            // Use batch DELETE with IN clause for better performance
            if len(batch) > 1 {
                placeholders := strings.Repeat("?,", len(batch)-1) + "?"
                args := make([]interface{}, len(batch))
                for i, id := range batch {
                    args[i] = id
                }
                
                // First remove relationships in batch
                _, err := tx.Exec(
                    fmt.Sprintf("DELETE FROM resource_services WHERE service_id IN (%s)", placeholders),
                    args...,
                )
                if err != nil {
                    return fmt.Errorf("failed to delete service relationships: %w", err)
                }
                
                // Then delete services in batch
                _, err = tx.Exec(
                    fmt.Sprintf("DELETE FROM services WHERE id IN (%s)", placeholders),
                    args...,
                )
                if err != nil {
                    return fmt.Errorf("failed to delete services: %w", err)
                }
                
                if opts.LogLevel >= 1 {
                    log.Printf("Deleted batch of %d services", len(batch))
                }
            } else {
                // Single item - original logic
                id := batch[0]
                if _, err := tx.Exec("DELETE FROM resource_services WHERE service_id = ?", id); err != nil {
                    return fmt.Errorf("failed to delete resource_service references for %s: %w", id, err)
                }
                if _, err := tx.Exec("DELETE FROM services WHERE id = ?", id); err != nil {
                    return fmt.Errorf("failed to delete service %s: %w", id, err)
                }
            }
        }
        
        return nil
    })
}

// CleanupManager - CORRECTED VERSION with proper DB reference
type CleanupManager struct {
    db               *DB       // This should match your actual DB type
    cleanupMutex     sync.Mutex
    isCleanupRunning bool
}

func NewCleanupManager(database *DB) *CleanupManager {
    return &CleanupManager{
        db:               database,
        cleanupMutex:     sync.Mutex{},
        isCleanupRunning: false,
    }
}

func (cm *CleanupManager) PerformFullCleanup(opts CleanupOptions) error {
    cm.cleanupMutex.Lock()
    defer cm.cleanupMutex.Unlock()
    
    if cm.isCleanupRunning {
        return fmt.Errorf("cleanup already in progress")
    }
    
    cm.isCleanupRunning = true
    defer func() {
        cm.isCleanupRunning = false
    }()
    
    // Add warning log
    if opts.LogLevel >= 1 {
        log.Println("⚠️  Database cleanup starting - this may cause brief service interruptions")
    }
    
    // First clean up services
    if err := cm.db.CleanupDuplicateServices(opts); err != nil {
        return fmt.Errorf("service cleanup failed: %w", err)
    }
    
    // Then clean up resources
    if err := cm.db.CleanupDuplicateResources(opts); err != nil {
        return fmt.Errorf("resource cleanup failed: %w", err)
    }
    
    // Finally clean up orphaned relationships
    if err := cm.db.CleanupOrphanedRelationships(opts); err != nil {
        return fmt.Errorf("relationship cleanup failed: %w", err)
    }
    
    return nil
}

// CleanupDuplicateResources removes resource duplication from the database
func (db *DB) CleanupDuplicateResources(opts CleanupOptions) error {
    if opts.LogLevel >= 1 {
        log.Println("Starting cleanup of duplicate resources...")
    }
    
    // Get all resources
    rows, err := db.Query("SELECT id, host, service_id, status FROM resources")
    if err != nil {
        return fmt.Errorf("failed to query resources: %w", err)
    }
    defer rows.Close()
    
    // Map to track resources by normalized ID
    type resourceInfo struct {
        ID        string
        Host      string
        ServiceID string
        Status    string
    }
    
    // Group by host to find multiple resources for the same host
    hostMap := make(map[string][]resourceInfo)
    
    // Process each resource
    for rows.Next() {
        var id, host, serviceID, status string
        if err := rows.Scan(&id, &host, &serviceID, &status); err != nil {
            return fmt.Errorf("failed to scan resource: %w", err)
        }
        
        // Add to host map
        hostMap[host] = append(hostMap[host], resourceInfo{
            ID:        id,
            Host:      host,
            ServiceID: serviceID,
            Status:    status,
        })
    }
    
    if err := rows.Err(); err != nil {
        return fmt.Errorf("error iterating resources: %w", err)
    }
    
    // Find hosts with multiple resources
    var resourcesToDelete []string
    var resourcesToActivate []string
    
    for host, resources := range hostMap {
        if len(resources) <= 1 {
            continue // No duplicates
        }
        
        if opts.LogLevel >= 2 {
            log.Printf("Found %d resources for host %s", len(resources), host)
        }
        
        // Sort resources by status (active first) and then by ID complexity
        // We'll keep the active one with the simplest ID
        activeResources := make([]resourceInfo, 0)
        disabledResources := make([]resourceInfo, 0)
        
        for _, res := range resources {
            if res.Status == "active" {
                activeResources = append(activeResources, res)
            } else {
                disabledResources = append(disabledResources, res)
            }
        }
        
        // If there are multiple active resources, disable extras
        if len(activeResources) > 1 {
            // Sort to find the one to keep (prioritize simpler IDs)
            bestID := ""
            bestIdx := 0
            
            for i, res := range activeResources {
                normalizedID := util.NormalizeID(res.ID)
                
                if bestID == "" {
                    bestID = normalizedID
                    bestIdx = i
                } else {
                    // Prefer router-auth pattern for consistency
                    if strings.Contains(normalizedID, "-router-auth") && 
                       !strings.Contains(bestID, "-router-auth") {
                        bestID = normalizedID
                        bestIdx = i
                    } else if !strings.Contains(normalizedID, "-router-auth") && 
                              strings.Contains(bestID, "-router-auth") {
                        // Keep current best
                    } else if len(normalizedID) < len(bestID) {
                        // Prefer shorter/simpler IDs
                        bestID = normalizedID
                        bestIdx = i
                    }
                }
            }
            
            // Keep the best one, mark others for deletion
            for i, res := range activeResources {
                if i != bestIdx {
                    if opts.LogLevel >= 2 {
                        log.Printf("  - Will disable duplicate active resource: %s", res.ID)
                    }
                    resourcesToDelete = append(resourcesToDelete, res.ID)
                } else if opts.LogLevel >= 2 {
                    log.Printf("  - Keeping active resource: %s", res.ID)
                }
            }
        } else if len(activeResources) == 0 && len(disabledResources) > 0 && opts.RecoverCorrupted {
            // No active resources, recover one
            bestIdx := 0
            bestID := ""
            
            for i, res := range disabledResources {
                normalizedID := util.NormalizeID(res.ID)
                
                if bestID == "" {
                    bestID = normalizedID
                    bestIdx = i
                } else if len(normalizedID) < len(bestID) {
                    // Prefer shorter/simpler IDs
                    bestID = normalizedID
                    bestIdx = i
                }
            }
            
            // Activate the best one
            if opts.LogLevel >= 2 {
                log.Printf("  - Will activate resource: %s", disabledResources[bestIdx].ID)
            }
            resourcesToActivate = append(resourcesToActivate, disabledResources[bestIdx].ID)
            
            // If reaping disabled resources, delete the rest
            if opts.ReapDisabled {
                for i, res := range disabledResources {
                    if i != bestIdx {
                        if opts.LogLevel >= 2 {
                            log.Printf("  - Will delete disabled resource: %s", res.ID)
                        }
                        resourcesToDelete = append(resourcesToDelete, res.ID)
                    }
                }
            }
        } else if opts.ReapDisabled {
            // Delete all disabled resources if ReapDisabled is true
            for _, res := range disabledResources {
                if opts.LogLevel >= 2 {
                    log.Printf("  - Will delete disabled resource: %s", res.ID)
                }
                resourcesToDelete = append(resourcesToDelete, res.ID)
            }
        }
    }
    
    if len(resourcesToDelete) == 0 && len(resourcesToActivate) == 0 {
        if opts.LogLevel >= 1 {
            log.Println("No resources need cleanup.")
        }
        return nil
    }
    
    if opts.DryRun {
        log.Printf("DRY RUN: Would delete %d resources and activate %d resources", 
                  len(resourcesToDelete), len(resourcesToActivate))
        return nil
    }
    
    // Process changes in a transaction
    return db.WithTransaction(func(tx *sql.Tx) error {
        // Activate resources that need activation
        for _, id := range resourcesToActivate {
            if opts.LogLevel >= 1 {
                log.Printf("Activating resource: %s", id)
            }
            
            _, err := tx.Exec(
                "UPDATE resources SET status = 'active', updated_at = ? WHERE id = ?",
                time.Now(), id,
            )
            
            if err != nil {
                return fmt.Errorf("failed to activate resource %s: %w", id, err)
            }
        }
        
        // Delete or disable resources
        for _, id := range resourcesToDelete {
            if opts.ReapDisabled {
                // Physically delete the resource
                if opts.LogLevel >= 1 {
                    log.Printf("Deleting resource: %s", id)
                }
                
                // First delete any middleware relationships
                if _, err := tx.Exec("DELETE FROM resource_middlewares WHERE resource_id = ?", id); err != nil {
                    return fmt.Errorf("failed to delete resource_middlewares for %s: %w", id, err)
                }
                
                // Then delete any service relationships
                if _, err := tx.Exec("DELETE FROM resource_services WHERE resource_id = ?", id); err != nil {
                    return fmt.Errorf("failed to delete resource_services for %s: %w", id, err)
                }
                
                // Finally delete the resource
                if _, err := tx.Exec("DELETE FROM resources WHERE id = ?", id); err != nil {
                    return fmt.Errorf("failed to delete resource %s: %w", id, err)
                }
            } else {
                // Just mark as disabled
                if opts.LogLevel >= 1 {
                    log.Printf("Disabling resource: %s", id)
                }
                
                _, err := tx.Exec(
                    "UPDATE resources SET status = 'disabled', updated_at = ? WHERE id = ?",
                    time.Now(), id,
                )
                
                if err != nil {
                    return fmt.Errorf("failed to disable resource %s: %w", id, err)
                }
            }
        }
        
        if opts.LogLevel >= 1 {
            log.Printf("Resource cleanup complete. Deleted/disabled %d resources, activated %d resources",
                      len(resourcesToDelete), len(resourcesToActivate))
        }
        return nil
    })
}

// PerformFullCleanup runs a comprehensive cleanup of the database
func (db *DB) PerformFullCleanup(opts CleanupOptions) error {
    // First clean up services
    if err := db.CleanupDuplicateServices(opts); err != nil {
        return fmt.Errorf("service cleanup failed: %w", err)
    }
    
    // Then clean up resources
    if err := db.CleanupDuplicateResources(opts); err != nil {
        return fmt.Errorf("resource cleanup failed: %w", err)
    }
    
    return nil
}

// CleanupOrphanedRelationships removes relationship rows that reference missing resources, services or middlewares.
func (db *DB) CleanupOrphanedRelationships(opts CleanupOptions) error {
    if opts.LogLevel >= 1 {
        log.Println("Starting cleanup of orphaned relationships...")
    }

    queries := []struct {
        desc string
        qry  string
    }{
        {"orphaned resource_services by missing service", "SELECT COUNT(*) FROM resource_services rs LEFT JOIN services s ON rs.service_id = s.id WHERE s.id IS NULL"},
        {"orphaned resource_services by missing resource", "SELECT COUNT(*) FROM resource_services rs LEFT JOIN resources r ON rs.resource_id = r.id WHERE r.id IS NULL"},
        {"orphaned resource_middlewares by missing middleware", "SELECT COUNT(*) FROM resource_middlewares rm LEFT JOIN middlewares m ON rm.middleware_id = m.id WHERE m.id IS NULL"},
        {"orphaned resource_middlewares by missing resource", "SELECT COUNT(*) FROM resource_middlewares rm LEFT JOIN resources r ON rm.resource_id = r.id WHERE r.id IS NULL"},
    }

    // Dry run: just report counts
    if opts.DryRun {
        for _, q := range queries {
            var count int64
            if err := db.QueryRow(q.qry).Scan(&count); err != nil {
                // Non-fatal: log and continue
                if opts.LogLevel >= 0 {
                    log.Printf("DRY RUN: failed to count %s: %v", q.desc, err)
                }
                continue
            }
            log.Printf("DRY RUN: %s: %d", q.desc, count)
        }
        return nil
    }

    // Execute deletes in a transaction
    return db.WithTransaction(func(tx *sql.Tx) error {
        delQueries := []struct {
            desc string
            qry  string
        }{
            {"delete resource_services with missing service", "DELETE FROM resource_services WHERE service_id NOT IN (SELECT id FROM services)"},
            {"delete resource_services with missing resource", "DELETE FROM resource_services WHERE resource_id NOT IN (SELECT id FROM resources)"},
            {"delete resource_middlewares with missing middleware", "DELETE FROM resource_middlewares WHERE middleware_id NOT IN (SELECT id FROM middlewares)"},
            {"delete resource_middlewares with missing resource", "DELETE FROM resource_middlewares WHERE resource_id NOT IN (SELECT id FROM resources)"},
        }

        for _, dq := range delQueries {
            res, err := tx.Exec(dq.qry)
            if err != nil {
                return fmt.Errorf("failed to %s: %w", dq.desc, err)
            }
            if opts.LogLevel >= 1 {
                if n, err := res.RowsAffected(); err == nil {
                    log.Printf("Deleted %d rows: %s", n, dq.desc)
                } else {
                    log.Printf("Deleted rows (unknown count): %s", dq.desc)
                }
            }
        }

        if opts.LogLevel >= 1 {
            log.Println("Orphaned relationship cleanup complete.")
        }
        return nil
    })
}