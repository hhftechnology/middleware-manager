package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/models"
	"github.com/hhftechnology/middleware-manager/services"
	"gopkg.in/yaml.v3"
)

// PluginHandler handles plugin-related requests
type PluginHandler struct {
	DB                      *sql.DB
	TraefikStaticConfigPath string
	ConfigManager           *services.ConfigManager
	pluginFetcher           *services.PluginFetcher
}

// NewPluginHandler creates a new plugin handler
func NewPluginHandler(db *sql.DB, traefikStaticConfigPath string, configManager *services.ConfigManager) *PluginHandler {
	handler := &PluginHandler{
		DB:                      db,
		TraefikStaticConfigPath: traefikStaticConfigPath,
		ConfigManager:           configManager,
	}

	// Initialize plugin fetcher with data source config
	if configManager != nil {
		dsConfig, err := configManager.GetActiveDataSourceConfig()
		if err == nil && dsConfig.Type == models.TraefikAPI {
			handler.pluginFetcher = services.NewPluginFetcher(dsConfig)
		}
	}

	return handler
}

// RefreshPluginFetcher refreshes the plugin fetcher with current config
func (h *PluginHandler) RefreshPluginFetcher() error {
	if h.ConfigManager == nil {
		return fmt.Errorf("config manager not initialized")
	}

	dsConfig, err := h.ConfigManager.GetActiveDataSourceConfig()
	if err != nil {
		return fmt.Errorf("failed to get data source config: %w", err)
	}

	// Only create fetcher for Traefik API type
	if dsConfig.Type == models.TraefikAPI {
		h.pluginFetcher = services.NewPluginFetcher(dsConfig)
	} else {
		// For Pangolin, we can also fetch plugins from the config
		h.pluginFetcher = services.NewPluginFetcher(dsConfig)
	}

	return nil
}

// GetPlugins fetches plugins from Traefik API
func (h *PluginHandler) GetPlugins(c *gin.Context) {
	// Refresh fetcher if needed
	if h.pluginFetcher == nil {
		if err := h.RefreshPluginFetcher(); err != nil {
			log.Printf("Warning: Failed to refresh plugin fetcher: %v", err)
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var plugins []models.PluginResponse

	// Try to fetch from Traefik API
	if h.pluginFetcher != nil {
		var err error
		plugins, err = h.pluginFetcher.FetchPlugins(ctx)
		if err != nil {
			log.Printf("Error fetching plugins from Traefik API: %v", err)
			// Fall back to local config check
			plugins = h.getPluginsFromLocalConfig()
		}
	} else {
		// No fetcher available, check local config
		plugins = h.getPluginsFromLocalConfig()
	}

	// Merge with local config to get installation status
	localPlugins, err := h.getLocalInstalledPlugins()
	if err != nil {
		log.Printf("Warning: Could not read local Traefik config: %v", err)
	} else {
		plugins = h.mergeWithLocalConfig(plugins, localPlugins)
	}

	c.JSON(http.StatusOK, plugins)
}

// getPluginsFromLocalConfig reads plugins from local Traefik static config
func (h *PluginHandler) getPluginsFromLocalConfig() []models.PluginResponse {
	plugins := []models.PluginResponse{}

	localPlugins, err := h.getLocalInstalledPlugins()
	if err != nil {
		return plugins
	}

	for key, config := range localPlugins {
		plugin := models.PluginResponse{
			Name:        key,
			Type:        "middleware",
			Status:      "configured",
			IsInstalled: true,
		}

		if moduleName, ok := config["moduleName"].(string); ok {
			plugin.ModuleName = moduleName
		} else {
			plugin.ModuleName = key
		}

		if version, ok := config["version"].(string); ok {
			plugin.Version = version
			plugin.InstalledVersion = version
		}

		plugins = append(plugins, plugin)
	}

	return plugins
}

// mergeWithLocalConfig merges API plugins with local config info
func (h *PluginHandler) mergeWithLocalConfig(apiPlugins []models.PluginResponse, localPlugins map[string]map[string]interface{}) []models.PluginResponse {
	// Create a map for quick lookup by name
	apiPluginMap := make(map[string]*models.PluginResponse)
	for i := range apiPlugins {
		apiPluginMap[apiPlugins[i].Name] = &apiPlugins[i]
	}
	
	// Also track which plugins are confirmed enabled (from API with status=enabled)
	enabledPlugins := make(map[string]bool)
	for _, plugin := range apiPlugins {
		if plugin.Status == "enabled" {
			enabledPlugins[plugin.Name] = true
		}
	}

	// Update API plugins with local config info
	for key, localConfig := range localPlugins {
		if plugin, exists := apiPluginMap[key]; exists {
			plugin.IsInstalled = true
			if version, ok := localConfig["version"].(string); ok {
				plugin.InstalledVersion = version
			}
		} else {
			// Plugin exists in local config but not directly in API
			// Check if it might be enabled under a different detection method
			// If plugin is in enabledPlugins map, it's actually running
			newPlugin := models.PluginResponse{
				Name:        key,
				Type:        "middleware",
				IsInstalled: true,
			}
			
			// Check if this plugin is actually enabled (might be detected via middleware usage)
			if enabledPlugins[key] {
				newPlugin.Status = "enabled"
			} else {
				// Plugin is configured but not detected as loaded - needs restart
				newPlugin.Status = "not_loaded"
			}

			if moduleName, ok := localConfig["moduleName"].(string); ok {
				newPlugin.ModuleName = moduleName
			} else {
				newPlugin.ModuleName = key
			}

			if version, ok := localConfig["version"].(string); ok {
				newPlugin.Version = version
				newPlugin.InstalledVersion = version
			}

			apiPlugins = append(apiPlugins, newPlugin)
		}
	}

	return apiPlugins
}

// getLocalInstalledPlugins reads the Traefik static config and returns installed plugins
func (h *PluginHandler) getLocalInstalledPlugins() (map[string]map[string]interface{}, error) {
	if h.TraefikStaticConfigPath == "" {
		return nil, fmt.Errorf("Traefik static configuration path is not set")
	}

	cleanPath := filepath.Clean(h.TraefikStaticConfigPath)

	config, err := h.readTraefikStaticConfig(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]map[string]interface{}), nil
		}
		return nil, fmt.Errorf("reading traefik static config: %w", err)
	}

	installedPlugins := make(map[string]map[string]interface{})
	if experimentalSection, ok := config["experimental"].(map[string]interface{}); ok {
		if pluginsConfig, ok := experimentalSection["plugins"].(map[string]interface{}); ok {
			for key, pluginData := range pluginsConfig {
				if pluginEntry, okData := pluginData.(map[string]interface{}); okData {
					installedPlugins[key] = pluginEntry
				}
			}
		}
	}

	return installedPlugins, nil
}

// InstallPluginBody defines the expected request body for installing a plugin
type InstallPluginBody struct {
	ModuleName string `json:"moduleName" binding:"required"`
	Version    string `json:"version,omitempty"`
}

// InstallPlugin adds a plugin to the Traefik static configuration
func (h *PluginHandler) InstallPlugin(c *gin.Context) {
	var body InstallPluginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	if h.TraefikStaticConfigPath == "" {
		ResponseWithError(c, http.StatusInternalServerError, "Traefik static configuration file path is not configured. Please set it in settings.")
		return
	}

	cleanPath := filepath.Clean(h.TraefikStaticConfigPath)

	traefikStaticConfig, err := h.readTraefikStaticConfig(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			traefikStaticConfig = make(map[string]interface{})
			LogInfo(fmt.Sprintf("Traefik static config file not found at %s, will create a new one.", cleanPath))
		} else {
			LogError(fmt.Sprintf("reading traefik static config file %s", cleanPath), err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to read Traefik static configuration file.")
			return
		}
	}

	experimentalSection, ok := traefikStaticConfig["experimental"].(map[string]interface{})
	if !ok {
		if traefikStaticConfig["experimental"] != nil {
			ResponseWithError(c, http.StatusInternalServerError, "Traefik static configuration 'experimental' section has an unexpected format.")
			return
		}
		experimentalSection = make(map[string]interface{})
		traefikStaticConfig["experimental"] = experimentalSection
	}

	pluginsConfig, ok := experimentalSection["plugins"].(map[string]interface{})
	if !ok {
		if experimentalSection["plugins"] != nil {
			ResponseWithError(c, http.StatusInternalServerError, "Traefik static configuration 'plugins' section has an unexpected format.")
			return
		}
		pluginsConfig = make(map[string]interface{})
		experimentalSection["plugins"] = pluginsConfig
	}

	pluginKey := getPluginKey(body.ModuleName)
	if pluginKey == "" {
		ResponseWithError(c, http.StatusBadRequest, "Invalid plugin module name, could not derive a configuration key.")
		return
	}

	pluginEntry := map[string]interface{}{
		"moduleName": body.ModuleName,
	}
	if body.Version != "" {
		pluginEntry["version"] = body.Version
	}
	pluginsConfig[pluginKey] = pluginEntry

	if err := h.writeTraefikStaticConfig(cleanPath, traefikStaticConfig); err != nil {
		LogError("writing traefik static config", err)
		ResponseWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Invalidate plugin cache
	if h.pluginFetcher != nil {
		h.pluginFetcher.InvalidateCache()
	}

	log.Printf("Successfully configured plugin '%s' (key: '%s') in %s", body.ModuleName, pluginKey, cleanPath)
	c.JSON(http.StatusOK, gin.H{
		"message":    fmt.Sprintf("Plugin %s configured. A Traefik restart is required to load the plugin.", body.ModuleName),
		"pluginKey":  pluginKey,
		"moduleName": body.ModuleName,
		"version":    body.Version,
	})
}

// RemovePluginBody defines the expected request body for removing a plugin
type RemovePluginBody struct {
	ModuleName string `json:"moduleName" binding:"required"`
}

// RemovePlugin removes a plugin from the Traefik static configuration
func (h *PluginHandler) RemovePlugin(c *gin.Context) {
	var body RemovePluginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	if h.TraefikStaticConfigPath == "" {
		ResponseWithError(c, http.StatusInternalServerError, "Traefik static configuration file path is not configured.")
		return
	}

	cleanPath := filepath.Clean(h.TraefikStaticConfigPath)

	traefikStaticConfig, err := h.readTraefikStaticConfig(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			ResponseWithError(c, http.StatusNotFound, fmt.Sprintf("Traefik static configuration file not found at: %s", cleanPath))
		} else {
			LogError(fmt.Sprintf("reading traefik static config file %s for removal", cleanPath), err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to read Traefik static configuration file.")
		}
		return
	}

	pluginKey := getPluginKey(body.ModuleName)
	pluginRemoved := false

	if experimentalSection, ok := traefikStaticConfig["experimental"].(map[string]interface{}); ok {
		if pluginsConfig, ok := experimentalSection["plugins"].(map[string]interface{}); ok {
			if _, exists := pluginsConfig[pluginKey]; exists {
				delete(pluginsConfig, pluginKey)
				pluginRemoved = true

				if len(pluginsConfig) == 0 {
					delete(experimentalSection, "plugins")
				}
				if len(experimentalSection) == 0 {
					delete(traefikStaticConfig, "experimental")
				}
			}
		}
	}

	if !pluginRemoved {
		LogInfo(fmt.Sprintf("Plugin '%s' (key: '%s') not found in Traefik static configuration.", body.ModuleName, pluginKey))
		ResponseWithError(c, http.StatusNotFound, fmt.Sprintf("Plugin '%s' not found in configuration.", body.ModuleName))
		return
	}

	if err := h.writeTraefikStaticConfig(cleanPath, traefikStaticConfig); err != nil {
		LogError("writing traefik static config after removal", err)
		ResponseWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Invalidate plugin cache
	if h.pluginFetcher != nil {
		h.pluginFetcher.InvalidateCache()
	}

	log.Printf("Successfully removed plugin '%s' (key: '%s') from %s", body.ModuleName, pluginKey, cleanPath)
	c.JSON(http.StatusOK, gin.H{
		"message":    fmt.Sprintf("Plugin %s removed. A Traefik restart is required for changes to take effect.", body.ModuleName),
		"pluginKey":  pluginKey,
		"moduleName": body.ModuleName,
	})
}

// GetTraefikStaticConfigPath returns the current Traefik static config path
func (h *PluginHandler) GetTraefikStaticConfigPath(c *gin.Context) {
	if h.TraefikStaticConfigPath == "" {
		c.JSON(http.StatusOK, gin.H{
			"path":    "",
			"message": "Traefik static config path is not currently set.",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"path": h.TraefikStaticConfigPath})
}

// UpdatePathBody defines the request body for updating config path
type UpdatePathBody struct {
	Path string `json:"path" binding:"required"`
}

// UpdateTraefikStaticConfigPath updates the Traefik static config path
func (h *PluginHandler) UpdateTraefikStaticConfigPath(c *gin.Context) {
	var body UpdatePathBody
	if err := c.ShouldBindJSON(&body); err != nil {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	cleanPath := filepath.Clean(body.Path)
	if cleanPath == "" || cleanPath == "." || cleanPath == "/" || strings.HasSuffix(cleanPath, "/") {
		ResponseWithError(c, http.StatusBadRequest, "Invalid configuration path provided.")
		return
	}

	oldPath := h.TraefikStaticConfigPath
	h.TraefikStaticConfigPath = cleanPath
	log.Printf("Traefik static config path updated from '%s' to: '%s'", oldPath, cleanPath)

	c.JSON(http.StatusOK, gin.H{
		"message": "Traefik static config path updated.",
		"path":    cleanPath,
	})
}

// GetPluginUsage returns usage information for a specific plugin
func (h *PluginHandler) GetPluginUsage(c *gin.Context) {
	pluginName := c.Param("name")
	if pluginName == "" {
		ResponseWithError(c, http.StatusBadRequest, "Plugin name is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if h.pluginFetcher == nil {
		if err := h.RefreshPluginFetcher(); err != nil {
			ResponseWithError(c, http.StatusInternalServerError, "Failed to initialize plugin fetcher")
			return
		}
	}

	plugins, err := h.pluginFetcher.FetchPlugins(ctx)
	if err != nil {
		ResponseWithError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch plugins: %v", err))
		return
	}

	for _, plugin := range plugins {
		if plugin.Name == pluginName {
			c.JSON(http.StatusOK, gin.H{
				"name":       plugin.Name,
				"usageCount": plugin.UsageCount,
				"usedBy":     plugin.UsedBy,
				"status":     plugin.Status,
			})
			return
		}
	}

	ResponseWithError(c, http.StatusNotFound, fmt.Sprintf("Plugin '%s' not found", pluginName))
}

// Helper functions

func (h *PluginHandler) readTraefikStaticConfig(filePath string) (map[string]interface{}, error) {
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, fmt.Errorf("failed to parse Traefik static configuration: %w", err)
	}
	return config, nil
}

func (h *PluginHandler) writeTraefikStaticConfig(filePath string, config map[string]interface{}) error {
	updatedYaml, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to prepare updated Traefik configuration: %w", err)
	}

	// Create backup
	backupPath := filePath + ".bak." + time.Now().Format("20060102150405")
	if err := copyFile(filePath, backupPath); err != nil {
		LogInfo(fmt.Sprintf("Warning: Could not create backup: %v", err))
	} else {
		LogInfo(fmt.Sprintf("Created backup at %s", backupPath))
	}

	// Write to temp file first
	tempFile := filePath + ".tmp"
	if err := os.WriteFile(tempFile, updatedYaml, 0644); err != nil {
		_ = os.Remove(tempFile)
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, filePath); err != nil {
		return fmt.Errorf("failed to finalize configuration: %w", err)
	}

	return nil
}

func getPluginKey(moduleName string) string {
	if moduleName == "" {
		return ""
	}
	parts := strings.Split(moduleName, "/")
	if len(parts) > 0 {
		lastKeyPart := parts[len(parts)-1]
		lastKeyPart = strings.Split(lastKeyPart, "@")[0]
		lastKeyPart = strings.TrimSuffix(lastKeyPart, ".git")
		lastKeyPart = strings.TrimSuffix(lastKeyPart, "-plugin")
		return strings.ToLower(lastKeyPart)
	}
	key := strings.Split(moduleName, "@")[0]
	key = strings.TrimSuffix(key, ".git")
	key = strings.TrimSuffix(key, "-plugin")
	return strings.ToLower(key)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("could not open source file %s: %w", src, err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("could not create destination file %s: %w", dst, err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("could not copy content: %w", err)
	}
	return destinationFile.Sync()
}

func LogInfo(message string) {
	log.Println("INFO:", message)
}

// GetPluginCatalogue fetches the full plugin catalogue from plugins.traefik.io
func (h *PluginHandler) GetPluginCatalogue(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	plugins, err := services.FetchPluginCatalogue(ctx)
	if err != nil {
		log.Printf("Error fetching plugin catalogue: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch plugin catalogue: %v", err))
		return
	}

	// Also get installed plugins to mark them
	installedPlugins, _ := h.getLocalInstalledPlugins()

	// Mark installed plugins in the catalogue
	result := make([]map[string]interface{}, len(plugins))
	for i, plugin := range plugins {
		entry := map[string]interface{}{
			"id":            plugin.ID,
			"name":          plugin.Name,
			"displayName":   plugin.DisplayName,
			"author":        plugin.Author,
			"type":          plugin.Type,
			"import":        plugin.Import,
			"summary":       plugin.Summary,
			"iconUrl":       plugin.IconURL,
			"bannerUrl":     plugin.BannerURL,
			"latestVersion": plugin.LatestVersion,
			"versions":      plugin.Versions,
			"stars":         plugin.Stars,
			"snippet":       plugin.Snippet,
			"isInstalled":   false,
		}

		// Check if plugin is installed by looking up the module name in local plugins
		pluginKey := getPluginKey(plugin.Import)
		if _, installed := installedPlugins[pluginKey]; installed {
			entry["isInstalled"] = true
		}

		result[i] = entry
	}

	c.JSON(http.StatusOK, result)
}
