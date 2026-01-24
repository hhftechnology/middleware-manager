package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hhftechnology/middleware-manager/models"
)

// ConfigManager manages system configuration
type ConfigManager struct {
	configPath string
	config     models.SystemConfig
	mu         sync.RWMutex
}

// NewConfigManager creates a new config manager
func NewConfigManager(configPath string) (*ConfigManager, error) {
	cm := &ConfigManager{
		configPath: configPath,
	}

	if err := cm.loadConfig(); err != nil {
		return nil, err
	}

	return cm, nil
}

// loadConfig loads configuration from file
func (cm *ConfigManager) loadConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if config file exists
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// Create default config
		cm.config = models.SystemConfig{
			ActiveDataSource: "pangolin",
			DataSources: map[string]models.DataSourceConfig{
				"pangolin": {
					Type: models.PangolinAPI,
					URL:  "http://pangolin:3001/api/v1",
				},
				"traefik": {
					Type: models.TraefikAPI,
					URL:  "http://host.docker.internal:8080",
				},
			},
		}

		// Save default config
		return cm.saveConfig()
	}

	// Read config file
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse config
	if err := json.Unmarshal(data, &cm.config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return nil
}

// EnsureDefaultDataSources ensures default data sources are configured
func (cm *ConfigManager) EnsureDefaultDataSources(pangolinURL, traefikURL string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Ensure data sources map exists
	if cm.config.DataSources == nil {
		cm.config.DataSources = make(map[string]models.DataSourceConfig)
	}

	// Add default Pangolin data source if not present
	if _, exists := cm.config.DataSources["pangolin"]; !exists {
		cm.config.DataSources["pangolin"] = models.DataSourceConfig{
			Type: models.PangolinAPI,
			URL:  pangolinURL,
		}
	} else {
		// Ensure Type is set for existing Pangolin config (fix for old configs)
		pConfig := cm.config.DataSources["pangolin"]
		if pConfig.Type == "" {
			pConfig.Type = models.PangolinAPI
			if pConfig.URL == "" {
				pConfig.URL = pangolinURL
			}
			cm.config.DataSources["pangolin"] = pConfig
			log.Printf("Fixed missing Type for pangolin data source")
		}
	}

	// Add default Traefik data source if not present
	if _, exists := cm.config.DataSources["traefik"]; !exists {
		cm.config.DataSources["traefik"] = models.DataSourceConfig{
			Type: models.TraefikAPI,
			URL:  traefikURL,
		}
	} else {
		// Ensure Type is set for existing Traefik config (fix for old configs)
		tConfig := cm.config.DataSources["traefik"]
		if tConfig.Type == "" {
			tConfig.Type = models.TraefikAPI
			log.Printf("Fixed missing Type for traefik data source")
		}
		// Update Traefik URL if provided (could be auto-discovered)
		if traefikURL != "" && tConfig.URL != traefikURL {
			log.Printf("Updating Traefik URL from %s to %s", tConfig.URL, traefikURL)
			tConfig.URL = traefikURL
		}
		cm.config.DataSources["traefik"] = tConfig
	}

	// Ensure there's an active data source
	if cm.config.ActiveDataSource == "" {
		cm.config.ActiveDataSource = "pangolin"
	}

	// Try to determine if Traefik is available
	if cm.config.ActiveDataSource == "pangolin" {
		client := &http.Client{Timeout: 2 * time.Second}
		traefikConfig := cm.config.DataSources["traefik"]

		// Try the Traefik URL
		resp, err := client.Get(traefikConfig.URL + "/api/version")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			// Traefik is available, but not active - log a message
			log.Printf("Note: Traefik API appears to be available at %s but is not the active source", traefikConfig.URL)
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Save the updated configuration
	return cm.saveConfig()
}

// saveConfig saves configuration to file
func (cm *ConfigManager) saveConfig() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetActiveDataSourceConfig returns the active data source configuration
func (cm *ConfigManager) GetActiveDataSourceConfig() (models.DataSourceConfig, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	dsName := cm.config.ActiveDataSource
	ds, ok := cm.config.DataSources[dsName]
	if !ok {
		return models.DataSourceConfig{}, fmt.Errorf("active data source not found: %s", dsName)
	}

	// Fallback: infer Type from name if empty (for old configs)
	if ds.Type == "" {
		switch dsName {
		case "pangolin":
			ds.Type = models.PangolinAPI
		case "traefik":
			ds.Type = models.TraefikAPI
		default:
			return models.DataSourceConfig{}, fmt.Errorf("unknown data source type for: %s", dsName)
		}
	}

	return ds, nil
}

// GetActiveSourceName returns the name of the active data source
func (cm *ConfigManager) GetActiveSourceName() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.config.ActiveDataSource
}

// SetActiveDataSource sets the active data source
func (cm *ConfigManager) SetActiveDataSource(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, ok := cm.config.DataSources[name]; !ok {
		return fmt.Errorf("data source not found: %s", name)
	}

	// Skip if already active
	if cm.config.ActiveDataSource == name {
		return nil
	}

	// Store the previous active source for logging
	oldSource := cm.config.ActiveDataSource

	// Update active source
	cm.config.ActiveDataSource = name

	// Log the change
	log.Printf("Changed active data source from %s to %s", oldSource, name)

	return cm.saveConfig()
}

// GetDataSources returns all configured data sources
func (cm *ConfigManager) GetDataSources() map[string]models.DataSourceConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Return a copy to prevent map mutation
	sources := make(map[string]models.DataSourceConfig)
	for k, v := range cm.config.DataSources {
		sources[k] = v
	}

	return sources
}

// UpdateDataSource updates a data source configuration
func (cm *ConfigManager) UpdateDataSource(name string, config models.DataSourceConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Create a copy to avoid reference issues
	newConfig := config

	// Ensure URL doesn't end with a slash
	if newConfig.URL != "" && strings.HasSuffix(newConfig.URL, "/") {
		newConfig.URL = strings.TrimSuffix(newConfig.URL, "/")
	}

	// Test the connection before saving
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cm.testDataSourceConnection(ctx, newConfig); err != nil {
		log.Printf("Warning: Data source connection test failed: %v", err)
		// Continue anyway but log the warning
	}

	// Update the config
	cm.config.DataSources[name] = newConfig

	// If this is the active data source, log a special message
	if cm.config.ActiveDataSource == name {
		log.Printf("Updated active data source '%s'", name)
	}

	return cm.saveConfig()
}

// testDataSourceConnection tests the connection to a data source
func (cm *ConfigManager) testDataSourceConnection(ctx context.Context, config models.DataSourceConfig) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	var url string
	switch config.Type {
	case models.PangolinAPI:
		// Use the same health-check endpoint as API handler; Pangolin does not expose /status
		url = config.URL + "/traefik-config"
	case models.TraefikAPI:
		url = config.URL + "/api/version"
	default:
		return fmt.Errorf("unsupported data source type: %s", config.Type)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth if configured
	if config.BasicAuth.Username != "" {
		req.SetBasicAuth(config.BasicAuth.Username, config.BasicAuth.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("connection test failed with status code: %d", resp.StatusCode)
	}

	return nil
}

// TestDataSourceConnection is a public method to test a connection
func (cm *ConfigManager) TestDataSourceConnection(config models.DataSourceConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return cm.testDataSourceConnection(ctx, config)
}
