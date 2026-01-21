package services

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/hhftechnology/middleware-manager/models"
)

// DuplicateDetector checks for middleware name conflicts with Traefik
type DuplicateDetector struct {
	configManager *ConfigManager
}

// NewDuplicateDetector creates a new duplicate detector instance
func NewDuplicateDetector(configManager *ConfigManager) *DuplicateDetector {
	return &DuplicateDetector{
		configManager: configManager,
	}
}

// getTraefikFetcher gets the Traefik fetcher from the config manager
func (d *DuplicateDetector) getTraefikFetcher() *TraefikFetcher {
	if d.configManager == nil {
		return nil
	}

	// Try to get Traefik config from data sources
	sources := d.configManager.GetDataSources()
	traefikConfig, ok := sources["traefik"]
	if !ok {
		// Check if active source is Traefik
		activeConfig, err := d.configManager.GetActiveDataSourceConfig()
		if err != nil || activeConfig.Type != models.TraefikAPI {
			return nil
		}
		traefikConfig = activeConfig
	}

	return NewTraefikFetcher(traefikConfig)
}

// CheckDuplicates checks if a middleware name already exists in Traefik
func (d *DuplicateDetector) CheckDuplicates(name, pluginName string) *models.DuplicateCheckResult {
	result := &models.DuplicateCheckResult{
		HasDuplicates: false,
		Duplicates:    []models.Duplicate{},
		APIAvailable:  true,
	}

	// Get TraefikFetcher
	fetcher := d.getTraefikFetcher()
	if fetcher == nil {
		result.APIAvailable = false
		result.WarningMessage = "Traefik API not configured. Cannot check for duplicates."
		return result
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fetch existing middlewares from Traefik
	middlewares, err := fetcher.GetTraefikMiddlewares(ctx)
	if err != nil {
		log.Printf("Failed to fetch Traefik middlewares for duplicate check: %v", err)
		result.APIAvailable = false
		result.WarningMessage = "Could not connect to Traefik API: " + err.Error()
		return result
	}

	// Normalize the name for comparison
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	normalizedPluginName := strings.ToLower(strings.TrimSpace(pluginName))

	// Check each middleware for duplicates
	for _, mw := range middlewares {
		mwName := strings.ToLower(mw.Name)

		// Check for exact name match
		if mwName == normalizedName {
			result.HasDuplicates = true
			result.Duplicates = append(result.Duplicates, models.Duplicate{
				Name:     mw.Name,
				Provider: mw.Provider,
				Type:     mw.Type,
			})
			continue
		}

		// For plugins, also check if the plugin name appears in the middleware config
		if normalizedPluginName != "" && strings.Contains(mwName, normalizedPluginName) {
			// Check the first 5 fields for plugin name match
			if d.containsPluginName(mw, normalizedPluginName) {
				result.HasDuplicates = true
				result.Duplicates = append(result.Duplicates, models.Duplicate{
					Name:     mw.Name,
					Provider: mw.Provider,
					Type:     mw.Type,
				})
			}
		}
	}

	return result
}

// containsPluginName checks if a middleware contains the plugin name in its configuration
func (d *DuplicateDetector) containsPluginName(mw models.TraefikMiddleware, pluginName string) bool {
	// Check if it's a plugin type middleware and contains the plugin name
	if strings.Contains(strings.ToLower(mw.Type), "plugin") {
		return true
	}

	// Check the name itself
	if strings.Contains(strings.ToLower(mw.Name), pluginName) {
		return true
	}

	return false
}
