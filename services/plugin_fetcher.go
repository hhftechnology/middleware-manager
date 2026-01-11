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

// PluginFetcher fetches plugin information from Traefik API
type PluginFetcher struct {
	config       models.DataSourceConfig
	httpClient   *http.Client
	singleflight singleflight.Group
	lastFetch    time.Time
	lastFetchMu  sync.RWMutex
	minInterval  time.Duration

	// Cached data
	cachedPlugins   []models.PluginResponse
	cachedPluginsMu sync.RWMutex
}

// NewPluginFetcher creates a new plugin fetcher
func NewPluginFetcher(config models.DataSourceConfig) *PluginFetcher {
	httpClient := GetHTTPClient()

	return &PluginFetcher{
		config:      config,
		httpClient:  httpClient,
		minInterval: 5 * time.Second,
	}
}

// FetchPlugins fetches all plugin information from Traefik API
func (f *PluginFetcher) FetchPlugins(ctx context.Context) ([]models.PluginResponse, error) {
	result, err, _ := f.singleflight.Do("fetch-plugins", func() (interface{}, error) {
		return f.fetchPluginsInternal(ctx)
	})

	if err != nil {
		return nil, err
	}

	return result.([]models.PluginResponse), nil
}

// fetchPluginsInternal performs the actual fetch
func (f *PluginFetcher) fetchPluginsInternal(ctx context.Context) ([]models.PluginResponse, error) {
	// Check rate limiting
	f.lastFetchMu.RLock()
	timeSinceLastFetch := time.Since(f.lastFetch)
	f.lastFetchMu.RUnlock()

	if timeSinceLastFetch < f.minInterval {
		f.cachedPluginsMu.RLock()
		cached := f.cachedPlugins
		f.cachedPluginsMu.RUnlock()
		if cached != nil {
			return cached, nil
		}
	}

	log.Println("Fetching plugins from Traefik API...")

	// Get middlewares to find plugin-based middlewares
	middlewares, err := f.fetchMiddlewares(ctx)
	if err != nil {
		log.Printf("Warning: Failed to fetch middlewares for plugin detection: %v", err)
		middlewares = []models.TraefikMiddleware{}
	}

	// Get overview for plugin status
	overview, err := f.fetchOverview(ctx)
	if err != nil {
		log.Printf("Warning: Failed to fetch overview for plugin status: %v", err)
	}

	// Build plugin responses from middleware data
	plugins := f.buildPluginResponses(middlewares, overview)

	// Update cache
	f.cachedPluginsMu.Lock()
	f.cachedPlugins = plugins
	f.cachedPluginsMu.Unlock()

	f.lastFetchMu.Lock()
	f.lastFetch = time.Now()
	f.lastFetchMu.Unlock()

	log.Printf("Fetched %d plugins from Traefik API", len(plugins))
	return plugins, nil
}

// fetchMiddlewares fetches all middlewares from Traefik API
func (f *PluginFetcher) fetchMiddlewares(ctx context.Context) ([]models.TraefikMiddleware, error) {
	url := f.config.URL + "/api/http/middlewares"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

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

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Try to decode as array or map
	return DecodeArrayOrMap[models.TraefikMiddleware](body, func(m *models.TraefikMiddleware, name string) {
		m.Name = name
	})
}

// fetchOverview fetches Traefik overview for plugin status
func (f *PluginFetcher) fetchOverview(ctx context.Context) (*models.TraefikPluginOverview, error) {
	url := f.config.URL + "/api/overview"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

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

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var overview models.TraefikPluginOverview
	if err := json.Unmarshal(body, &overview); err != nil {
		return nil, fmt.Errorf("failed to parse overview: %w", err)
	}

	return &overview, nil
}

// buildPluginResponses builds plugin responses from middleware and overview data
func (f *PluginFetcher) buildPluginResponses(middlewares []models.TraefikMiddleware, overview *models.TraefikPluginOverview) []models.PluginResponse {
	// Map to track unique plugins and their usage
	pluginMap := make(map[string]*models.PluginResponse)

	// Extract plugins from middlewares
	for _, mw := range middlewares {
		// Check if this middleware is a plugin type
		if mw.Type != "plugin" && !isPluginMiddleware(mw) {
			continue
		}

		pluginName := extractPluginName(mw)
		if pluginName == "" {
			continue
		}

		// Get or create plugin entry
		plugin, exists := pluginMap[pluginName]
		if !exists {
			plugin = &models.PluginResponse{
				Name:        pluginName,
				ModuleName:  extractModuleName(mw, pluginName),
				Version:     extractVersion(mw),
				Type:        "middleware",
				Provider:    mw.Provider,
				Status:      "enabled",
				IsInstalled: true,
				UsageCount:  0,
				UsedBy:      []string{},
				Config:      mw.Config,
			}
			pluginMap[pluginName] = plugin
		}

		// Track usage
		plugin.UsageCount++
		plugin.UsedBy = append(plugin.UsedBy, mw.Name)
	}

	// Update status from overview if available
	if overview != nil {
		for _, name := range overview.Plugins.Enabled {
			if plugin, ok := pluginMap[name]; ok {
				plugin.Status = "enabled"
			} else {
				// Plugin is enabled but not used in any middleware
				pluginMap[name] = &models.PluginResponse{
					Name:        name,
					ModuleName:  name,
					Type:        "middleware",
					Status:      "enabled",
					IsInstalled: true,
					UsageCount:  0,
					UsedBy:      []string{},
				}
			}
		}

		for _, name := range overview.Plugins.Disabled {
			if plugin, ok := pluginMap[name]; ok {
				plugin.Status = "disabled"
			} else {
				pluginMap[name] = &models.PluginResponse{
					Name:        name,
					ModuleName:  name,
					Type:        "middleware",
					Status:      "disabled",
					IsInstalled: true,
					UsageCount:  0,
					UsedBy:      []string{},
				}
			}
		}

		for _, name := range overview.Plugins.WithErrors {
			if plugin, ok := pluginMap[name]; ok {
				plugin.Status = "error"
			} else {
				pluginMap[name] = &models.PluginResponse{
					Name:        name,
					ModuleName:  name,
					Type:        "middleware",
					Status:      "error",
					IsInstalled: true,
					UsageCount:  0,
					UsedBy:      []string{},
				}
			}
		}
	}

	// Convert map to slice
	plugins := make([]models.PluginResponse, 0, len(pluginMap))
	for _, plugin := range pluginMap {
		plugins = append(plugins, *plugin)
	}

	return plugins
}

// isPluginMiddleware checks if a middleware is plugin-based
func isPluginMiddleware(mw models.TraefikMiddleware) bool {
	// Check if config contains a "plugin" key
	if mw.Config != nil {
		if _, ok := mw.Config["plugin"]; ok {
			return true
		}
	}

	// Check if type indicates plugin
	if strings.Contains(strings.ToLower(mw.Type), "plugin") {
		return true
	}

	return false
}

// extractPluginName extracts the plugin name from a middleware
func extractPluginName(mw models.TraefikMiddleware) string {
	if mw.Config != nil {
		// Check for plugin config structure: {"plugin": {"pluginName": {...}}}
		if pluginConfig, ok := mw.Config["plugin"].(map[string]interface{}); ok {
			for name := range pluginConfig {
				return name
			}
		}
	}

	// Try to extract from middleware name (e.g., "badger@file" -> "badger")
	name := mw.Name
	if idx := strings.Index(name, "@"); idx != -1 {
		name = name[:idx]
	}

	// Remove common suffixes
	name = strings.TrimSuffix(name, "-middleware")
	name = strings.TrimSuffix(name, "-plugin")

	return name
}

// extractModuleName extracts the module name from middleware config
func extractModuleName(mw models.TraefikMiddleware, fallback string) string {
	if mw.Config != nil {
		if pluginConfig, ok := mw.Config["plugin"].(map[string]interface{}); ok {
			for _, config := range pluginConfig {
				if configMap, ok := config.(map[string]interface{}); ok {
					if moduleName, ok := configMap["moduleName"].(string); ok {
						return moduleName
					}
				}
			}
		}
	}
	return fallback
}

// extractVersion extracts version from middleware config
func extractVersion(mw models.TraefikMiddleware) string {
	if mw.Config != nil {
		if pluginConfig, ok := mw.Config["plugin"].(map[string]interface{}); ok {
			for _, config := range pluginConfig {
				if configMap, ok := config.(map[string]interface{}); ok {
					if version, ok := configMap["version"].(string); ok {
						return version
					}
				}
			}
		}
	}
	return ""
}

// GetCachedPlugins returns cached plugins without fetching
func (f *PluginFetcher) GetCachedPlugins() []models.PluginResponse {
	f.cachedPluginsMu.RLock()
	defer f.cachedPluginsMu.RUnlock()
	return f.cachedPlugins
}

// InvalidateCache clears the cached plugins
func (f *PluginFetcher) InvalidateCache() {
	f.cachedPluginsMu.Lock()
	f.cachedPlugins = nil
	f.cachedPluginsMu.Unlock()
}
