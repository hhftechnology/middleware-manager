package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hhftechnology/middleware-manager/database"
	"github.com/hhftechnology/middleware-manager/models"
	"github.com/hhftechnology/middleware-manager/util"
)

// ServiceWatcher watches for services using configured data source
type ServiceWatcher struct {
	db            *database.DB
	fetcher       ServiceFetcher
	configManager *ConfigManager
	stopChan      chan struct{}
	isRunning     bool
}

// NewServiceWatcher creates a new service watcher
func NewServiceWatcher(db *database.DB, configManager *ConfigManager) (*ServiceWatcher, error) {
	// Get the active data source config
	dsConfig, err := configManager.GetActiveDataSourceConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get active data source config: %w", err)
	}

	// Create the fetcher
	fetcher, err := NewServiceFetcher(dsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create service fetcher: %w", err)
	}

	return &ServiceWatcher{
		db:            db,
		fetcher:       fetcher,
		configManager: configManager,
		stopChan:      make(chan struct{}),
		isRunning:     false,
	}, nil
}

// Start begins watching for services
func (sw *ServiceWatcher) Start(interval time.Duration) {
	if sw.isRunning {
		return
	}

	sw.isRunning = true
	log.Printf("Service watcher started, checking every %v", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Do an initial check
	if err := sw.checkServices(); err != nil {
		log.Printf("Initial service check failed: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			// Check if data source config has changed
			if err := sw.refreshFetcher(); err != nil {
				log.Printf("Failed to refresh service fetcher: %v", err)
			}

			if err := sw.checkServices(); err != nil {
				log.Printf("Service check failed: %v", err)
			}
		case <-sw.stopChan:
			log.Println("Service watcher stopped")
			return
		}
	}
}

// refreshFetcher updates the fetcher if the data source config has changed
func (sw *ServiceWatcher) refreshFetcher() error {
	dsConfig, err := sw.configManager.GetActiveDataSourceConfig()
	if err != nil {
		return fmt.Errorf("failed to get data source config: %w", err)
	}

	// Create a new fetcher with the updated config
	fetcher, err := NewServiceFetcher(dsConfig)
	if err != nil {
		return fmt.Errorf("failed to create service fetcher: %w", err)
	}

	// Update the fetcher
	sw.fetcher = fetcher
	return nil
}

// Stop stops the service watcher
func (sw *ServiceWatcher) Stop() {
	if !sw.isRunning {
		return
	}

	close(sw.stopChan)
	sw.isRunning = false
}

// checkServices fetches services from the configured data source and updates the database
func (sw *ServiceWatcher) checkServices() error {

	// Create a context with timeout for the operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch services using the configured fetcher
	services, err := sw.fetcher.FetchServices(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch services: %w", err)
	}

	// Get all existing active Pangolin-synced services from the database
	// We only track Pangolin-synced services for cleanup (source_type = 'pangolin')
	var existingServices []string
	rows, err := sw.db.Query("SELECT id FROM services WHERE status = 'active' AND source_type = 'pangolin'")
	if err != nil {
		return fmt.Errorf("failed to query existing services: %w", err)
	}

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Printf("Error scanning service ID: %v", err)
			continue
		}
		existingServices = append(existingServices, id)
	}
	rows.Close()

	// Keep track of services we find
	foundServices := make(map[string]bool)

	// Check if there are any services
	if len(services.Services) == 0 {
		log.Println("No services found in data source")
		return nil
	}

	// Process services
	for _, service := range services.Services {
		// Skip invalid services
		if service.ID == "" || service.Type == "" {
			continue
		}

		// Process service
		if err := sw.updateOrCreateService(service); err != nil {
			log.Printf("Error processing service %s: %v", service.ID, err)
			// Continue processing other services even if one fails
			continue
		}

		// Mark normalized version of this service as found
		normalizedID := util.NormalizeID(service.ID)
		foundServices[normalizedID] = true
	}

	// Mark Pangolin-synced services as disabled if they no longer exist in the data source
	// Only affects services with source_type = 'pangolin' (already filtered in the query above)
	for _, serviceID := range existingServices {
		normalizedID := util.NormalizeID(serviceID)
		if !foundServices[normalizedID] {
			log.Printf("Service %s no longer exists in Pangolin, marking as disabled", serviceID)
			_, err := sw.db.Exec(
				"UPDATE services SET status = 'disabled', updated_at = ? WHERE id = ?",
				time.Now(), serviceID,
			)
			if err != nil {
				log.Printf("Error marking service as disabled: %v", err)
			}
		}
	}

	return nil
}

// updateOrCreateService updates an existing service or creates a new one
func (sw *ServiceWatcher) updateOrCreateService(service models.Service) error {
	// Use our centralized normalization function
	normalizedID := util.NormalizeID(service.ID)
	originalID := service.ID

	// Check if service already exists using normalized ID
	var exists int
	var existingType, existingConfig string

	err := sw.db.QueryRow(
		"SELECT 1, type, config FROM services WHERE id = ?",
		normalizedID,
	).Scan(&exists, &existingType, &existingConfig)

	if err == nil {
		// Service exists, only update if it changed
		if shouldUpdateService(sw.db, service, normalizedID) {
			log.Printf("Updating existing service: %s (normalized from %s)", normalizedID, originalID)
			return sw.updateService(service, normalizedID)
		}
		// Service exists and hasn't changed, skip update
		return nil
	} else if err != sql.ErrNoRows {
		// Unexpected error
		return fmt.Errorf("error checking if service exists: %w", err)
	}

	// Try checking if service exists with different provider suffixes
	var found bool
	err = sw.db.QueryRow(
		"SELECT 1 FROM services WHERE id LIKE ?",
		normalizedID+"%",
	).Scan(&exists)

	if err == nil {
		// Found a service with this base name but different suffix
		found = true
		var altID string
		err = sw.db.QueryRow(
			"SELECT id FROM services WHERE id LIKE ? LIMIT 1",
			normalizedID+"%",
		).Scan(&altID)

		if err == nil {
			// Check if update is actually needed before updating
			if shouldUpdateService(sw.db, service, altID) {
				log.Printf("Updating service with different suffix: %s", altID)
				return sw.updateService(service, altID)
			}
			// Service exists and hasn't changed, skip update
			return nil
		}
	}

	if !found {
		// Service doesn't exist with any suffix, create it
		service.ID = normalizedID
		return sw.createService(service)
	}

	// This shouldn't be reached, but just in case
	return nil
}

// shouldUpdateService determines if an existing service needs to be updated
func shouldUpdateService(db *database.DB, newService models.Service, normalizedID string) bool {
	var existingType, existingConfig string

	err := db.QueryRow(
		"SELECT type, config FROM services WHERE id = ?",
		normalizedID,
	).Scan(&existingType, &existingConfig)

	if err != nil {
		// If there's an error, assume we should update
		log.Printf("Error checking existing service %s: %v", normalizedID, err)
		return true
	}

	// Check if the type has changed
	if existingType != newService.Type {
		return true
	}

	// Check if the configuration has changed
	// Parse both configs to compare them semantically
	var existingConfigMap map[string]interface{}
	var newConfigMap map[string]interface{}

	if err := json.Unmarshal([]byte(existingConfig), &existingConfigMap); err != nil {
		log.Printf("Error parsing existing config for %s: %v", normalizedID, err)
		return true
	}

	if err := json.Unmarshal([]byte(newService.Config), &newConfigMap); err != nil {
		log.Printf("Error parsing new config for %s: %v", normalizedID, err)
		return true
	}

	// Compare the configurations
	return configsAreDifferent(existingConfigMap, newConfigMap)
}

// configsAreDifferent compares two service configurations
func configsAreDifferent(config1, config2 map[string]interface{}) bool {
	// Check for key differences
	for key := range config1 {
		if _, exists := config2[key]; !exists {
			return true
		}
	}

	for key := range config2 {
		if _, exists := config1[key]; !exists {
			return true
		}
	}

	// Check server configurations
	servers1, hasServers1 := config1["servers"].([]interface{})
	servers2, hasServers2 := config2["servers"].([]interface{})

	if hasServers1 != hasServers2 {
		return true
	}

	if hasServers1 && hasServers2 {
		if len(servers1) != len(servers2) {
			return true
		}

		// Compare each server
		for i, server1 := range servers1 {
			if i >= len(servers2) {
				return true
			}

			server1Map, ok1 := server1.(map[string]interface{})
			server2Map, ok2 := servers2[i].(map[string]interface{})

			if !ok1 || !ok2 {
				return true
			}

			// Check URL/address fields
			url1, hasURL1 := server1Map["url"].(string)
			url2, hasURL2 := server2Map["url"].(string)

			if hasURL1 != hasURL2 || (hasURL1 && url1 != url2) {
				return true
			}

			addr1, hasAddr1 := server1Map["address"].(string)
			addr2, hasAddr2 := server2Map["address"].(string)

			if hasAddr1 != hasAddr2 || (hasAddr1 && addr1 != addr2) {
				return true
			}
		}
	}

	// For other service types, we would need to check specific fields
	// For simplicity, we'll consider them different if any common key has a different value
	for key, val1 := range config1 {
		if val2, exists := config2[key]; exists {
			// Skip servers as we've handled them above
			if key == "servers" {
				continue
			}

			// Handle primitive types
			switch v1 := val1.(type) {
			case string:
				v2, ok := val2.(string)
				if !ok || v1 != v2 {
					return true
				}
			case float64:
				v2, ok := val2.(float64)
				if !ok || v1 != v2 {
					return true
				}
			case bool:
				v2, ok := val2.(bool)
				if !ok || v1 != v2 {
					return true
				}
			}
		}
	}

	return false
}

// createService creates a new service in the database
func (sw *ServiceWatcher) createService(service models.Service) error {
	// Validate service type
	if !models.IsValidServiceType(service.Type) {
		// Try to determine proper type if it's invalid
		if strings.Contains(strings.ToLower(service.Type), "load") ||
			strings.Contains(service.Config, "servers") {
			service.Type = string(models.LoadBalancerType)
		} else if strings.Contains(strings.ToLower(service.Type), "weight") {
			service.Type = string(models.WeightedType)
		} else if strings.Contains(strings.ToLower(service.Type), "mirror") {
			service.Type = string(models.MirroringType)
		} else if strings.Contains(strings.ToLower(service.Type), "fail") {
			service.Type = string(models.FailoverType)
		} else {
			// Default to LoadBalancer if we can't determine
			service.Type = string(models.LoadBalancerType)
		}
	}

	// Process the service configuration
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(service.Config), &configMap); err != nil {
		log.Printf("Error parsing service config for %s: %v, using empty config", service.ID, err)
		configMap = make(map[string]interface{})
	}

	// Apply any service-specific processing
	configMap = models.ProcessServiceConfig(service.Type, configMap)

	// Convert processed config back to JSON
	configJSON, err := json.Marshal(configMap)
	if err != nil {
		log.Printf("Error marshaling processed config for %s: %v", service.ID, err)
		configJSON = []byte("{}")
	}

	// Create a reasonable name if none provided
	if service.Name == "" {
		service.Name = formatServiceName(service.ID)
	}

	// Get active data source to determine provider suffix
	dsConfig, err := sw.configManager.GetActiveDataSourceConfig()
	if err != nil {
		log.Printf("Warning: Could not get active data source: %v. Using default file provider.", err)
		dsConfig.Type = models.PangolinAPI
	}

	// Use a database transaction for insert
	return sw.db.WithTransaction(func(tx *sql.Tx) error {
		log.Printf("Creating new service: %s", service.ID)

		// Check for existing service one more time within transaction
		var exists int
		err := tx.QueryRow("SELECT 1 FROM services WHERE id = ?", service.ID).Scan(&exists)
		if err == nil {
			// Service exists, silently skip
			return nil
		} else if err != sql.ErrNoRows {
			// Unexpected error
			return fmt.Errorf("error checking service existence in transaction: %w", err)
		}

		// Insert the service with source_type for tracking origin
		_, err = tx.Exec(
			"INSERT INTO services (id, name, type, config, status, source_type, created_at, updated_at) VALUES (?, ?, ?, ?, 'active', 'pangolin', ?, ?)",
			service.ID, service.Name, service.Type, string(configJSON), time.Now(), time.Now(),
		)

		if err != nil {
			// Check if it's a duplicate key error
			if strings.Contains(err.Error(), "UNIQUE constraint") {
				// Log but don't return error to continue processing other services
				log.Printf("Service %s already exists, skipping", service.ID)
				return nil
			}
			return fmt.Errorf("failed to insert service %s: %w", service.ID, err)
		}

		log.Printf("Created new service: %s", service.ID)
		return nil
	})
}

// updateService updates an existing service in the database
func (sw *ServiceWatcher) updateService(service models.Service, existingID string) error {
	// Get the existing service to preserve the name
	var existingName string
	err := sw.db.QueryRow("SELECT name FROM services WHERE id = ?", existingID).Scan(&existingName)

	if err != nil {
		log.Printf("Error fetching existing service name for %s: %v, using provided name", existingID, err)
	} else if existingName != "" {
		// Preserve existing name unless the new name is meaningful
		if service.Name == service.ID || service.Name == "" {
			service.Name = existingName
		}
	}

	// Process the service configuration
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(service.Config), &configMap); err != nil {
		log.Printf("Error parsing service config for %s: %v, using empty config", service.ID, err)
		configMap = make(map[string]interface{})
	}

	// Apply any service-specific processing
	configMap = models.ProcessServiceConfig(service.Type, configMap)

	// Convert processed config back to JSON
	configJSON, err := json.Marshal(configMap)
	if err != nil {
		log.Printf("Error marshaling processed config for %s: %v", service.ID, err)
		configJSON = []byte("{}")
	}

	// Update the service using a transaction
	return sw.db.WithTransaction(func(tx *sql.Tx) error {
		// Update the service using the existing ID, ensure status is active and source_type is pangolin
		result, err := tx.Exec(
			"UPDATE services SET name = ?, type = ?, config = ?, status = 'active', source_type = 'pangolin', updated_at = ? WHERE id = ?",
			service.Name, service.Type, string(configJSON), time.Now(), existingID,
		)

		if err != nil {
			return fmt.Errorf("failed to update service %s: %w", service.ID, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("Error getting rows affected: %v", err)
		} else if rowsAffected == 0 {
			log.Printf("Warning: Update did not affect any rows for service %s", existingID)
		} else {
			log.Printf("Updated existing service: %s", existingID)
		}

		return nil
	})
}

// formatServiceName creates a readable name from a service ID
func formatServiceName(id string) string {
	// Remove provider suffix if present
	name := id
	if idx := strings.Index(name, "@"); idx > 0 {
		name = name[:idx]
	}

	// Replace dashes and underscores with spaces
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")

	// Capitalize words
	parts := strings.Fields(name)
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}

	return strings.Join(parts, " ")
}
