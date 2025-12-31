package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// Plugin struct remains the same
type Plugin struct {
	DisplayName string `json:"displayName"`
	Type        string `json:"type"`
	IconPath    string `json:"iconPath"`
	Import      string `json:"import"`
	Summary     string `json:"summary"`
	Author      string `json:"author,omitempty"`
	Version     string `json:"version,omitempty"`
	TestedWith  string `json:"tested_with,omitempty"`
	Stars       int    `json:"stars,omitempty"`
	Homepage    string `json:"homepage,omitempty"`
	Docs        string `json:"docs,omitempty"`
}

// PluginHandler handles plugin-related requests
type PluginHandler struct {
	DB                      *sql.DB
	TraefikStaticConfigPath string
	PluginsJSONURL          string
}

// NewPluginHandler creates a new plugin handler
func NewPluginHandler(db *sql.DB, traefikStaticConfigPath string, pluginsJSONURL string) *PluginHandler {
	return &PluginHandler{
		DB:                      db,
		TraefikStaticConfigPath: traefikStaticConfigPath,
		PluginsJSONURL:          pluginsJSONURL,
	}
}

// GetPlugins fetches the list of plugins from the configured JSON URL
func (h *PluginHandler) GetPlugins(c *gin.Context) {
	if h.PluginsJSONURL == "" {
		ResponseWithError(c, http.StatusInternalServerError, "Plugins JSON URL is not configured in Middleware Manager.")
		return
	}

	resp, err := http.Get(h.PluginsJSONURL)
	if err != nil {
		LogError("fetching plugins JSON", err)
		ResponseWithError(c, http.StatusServiceUnavailable, "Failed to fetch plugins list from external source.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		LogError("fetching plugins JSON status", fmt.Errorf("received status code %d. Body: %s", resp.StatusCode, string(bodyBytes)))
		ResponseWithError(c, http.StatusServiceUnavailable, fmt.Sprintf("Failed to fetch plugins list: External source returned status %d.", resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		LogError("reading plugins JSON response body", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to read plugins list data from the external source.")
		return
	}

	var plugins []Plugin
	if err := json.Unmarshal(body, &plugins); err != nil {
		LogError("unmarshaling plugins JSON", fmt.Errorf("%w. Body received for unmarshaling: %s", err, string(body)))
		ResponseWithError(c, http.StatusInternalServerError, "Failed to parse plugins list data from the external source. Ensure it's valid JSON.")
		return
	}

	// Check local Traefik config to mark installed plugins
	installedPlugins, err := h.getLocalInstalledPlugins()
	if err != nil {
		// Log the error but don't fail the entire request, frontend can still show plugins
		LogInfo(fmt.Sprintf("Could not read local Traefik config to determine installed plugins: %v", err))
	}

	type PluginWithStatus struct {
		Plugin
		IsInstalled bool   `json:"isInstalled"`
		InstalledVersion string `json:"installedVersion,omitempty"`
	}

	pluginsWithStatus := make([]PluginWithStatus, len(plugins))
	for i, p := range plugins {
		status := PluginWithStatus{Plugin: p, IsInstalled: false}
		if localPlugin, ok := installedPlugins[getPluginKey(p.Import)]; ok {
			status.IsInstalled = true
			if version, vOk := localPlugin["version"].(string); vOk {
				status.InstalledVersion = version
			}
		}
		pluginsWithStatus[i] = status
	}

	c.JSON(http.StatusOK, pluginsWithStatus)
}

// InstallPluginBody defines the expected request body for installing a plugin
type InstallPluginBody struct {
	ModuleName string `json:"moduleName" binding:"required"`
	Version    string `json:"version,omitempty"`
}

// readTraefikStaticConfig is a helper to read and unmarshal the static config
func (h *PluginHandler) readTraefikStaticConfig(filePath string) (map[string]interface{}, error) {
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err // Error will be handled by the caller
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, fmt.Errorf("failed to parse Traefik static configuration (YAML format error): %w", err)
	}
	return config, nil
}

// writeTraefikStaticConfig is a helper to marshal and write the static config
func (h *PluginHandler) writeTraefikStaticConfig(filePath string, config map[string]interface{}) error {
	updatedYaml, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to prepare updated Traefik configuration for saving: %w", err)
	}

	backupPath := filePath + ".bak." + time.Now().Format("20060102150405")
	if err := copyFile(filePath, backupPath); err != nil {
		LogInfo(fmt.Sprintf("Warning: Could not create backup of %s to %s: %v. Proceeding with writing the main file.", filePath, backupPath, err))
	} else {
		LogInfo(fmt.Sprintf("Created backup of Traefik static config at %s", backupPath))
	}

	tempFile := filePath + ".tmp"
	if err := os.WriteFile(tempFile, updatedYaml, 0644); err != nil {
		_ = os.Remove(tempFile)
		return fmt.Errorf("failed to write updated Traefik configuration to a temporary file: %w", err)
	}

	if err := os.Rename(tempFile, filePath); err != nil {
		return fmt.Errorf("failed to finalize updated Traefik configuration file. Check file permissions and if a temporary file '.tmp' exists: %w", err)
	}
	return nil
}

// getLocalInstalledPlugins reads the Traefik static config and returns a map of installed plugin configurations.
func (h *PluginHandler) getLocalInstalledPlugins() (map[string]map[string]interface{}, error) {
	if h.TraefikStaticConfigPath == "" {
		return nil, fmt.Errorf("Traefik static configuration path is not set")
	}
	cleanPath := filepath.Clean(h.TraefikStaticConfigPath)

	config, err := h.readTraefikStaticConfig(cleanPath)
	if err != nil {
		if os.IsNotExist(err) { // If file doesn't exist, no plugins are installed
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


// InstallPlugin adds a plugin to the Traefik static configuration
func (h *PluginHandler) InstallPlugin(c *gin.Context) {
	var body InstallPluginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	if h.TraefikStaticConfigPath == "" {
		ResponseWithError(c, http.StatusInternalServerError, "Traefik static configuration file path is not configured in Middleware Manager. Please set it in settings.")
		return
	}
	cleanPath := filepath.Clean(h.TraefikStaticConfigPath)

	traefikStaticConfig, err := h.readTraefikStaticConfig(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, create a new config structure
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

	if _, exists := pluginsConfig[pluginKey]; exists {
		// LogInfo(fmt.Sprintf("Plugin '%s' (key: '%s') already exists in configuration. Overwriting.", body.ModuleName, pluginKey))
		// Allow overwrite for update purposes, or return a conflict error:
		// ResponseWithError(c, http.StatusConflict, fmt.Sprintf("Plugin '%s' is already configured.", body.ModuleName))
		// return
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
		ResponseWithError(c, http.StatusInternalServerError, err.Error()) // Provide more specific error from write
		return
	}

	log.Printf("Successfully configured plugin '%s' (key: '%s') in %s", body.ModuleName, pluginKey, cleanPath)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Plugin %s configured in %s. A Traefik restart is required to load the plugin.", body.ModuleName, filepath.Base(cleanPath))})
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
			ResponseWithError(c, http.StatusNotFound, fmt.Sprintf("Traefik static configuration file not found at: %s. Cannot remove plugin.", cleanPath))
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
				// If pluginsConfig becomes empty, optionally remove it
				if len(pluginsConfig) == 0 {
					delete(experimentalSection, "plugins")
				}
				// If experimentalSection becomes empty, optionally remove it
				if len(experimentalSection) == 0 {
					delete(traefikStaticConfig, "experimental")
				}
			}
		}
	}

	if !pluginRemoved {
		LogInfo(fmt.Sprintf("Plugin '%s' (key: '%s') not found in Traefik static configuration. Nothing to remove.", body.ModuleName, pluginKey))
		ResponseWithError(c, http.StatusNotFound, fmt.Sprintf("Plugin '%s' not found in configuration.", body.ModuleName))
		return
	}

	if err := h.writeTraefikStaticConfig(cleanPath, traefikStaticConfig); err != nil {
		LogError("writing traefik static config after removal", err)
		ResponseWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("Successfully removed plugin '%s' (key: '%s') from %s", body.ModuleName, pluginKey, cleanPath)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Plugin %s removed from configuration in %s. A Traefik restart is required for changes to take effect.", body.ModuleName, filepath.Base(cleanPath))})
}


// getPluginKey function remains the same
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

// GetTraefikStaticConfigPath function remains the same
func (h *PluginHandler) GetTraefikStaticConfigPath(c *gin.Context) {
	if h.TraefikStaticConfigPath == "" {
		c.JSON(http.StatusOK, gin.H{"path": "", "message": "Traefik static config path is not currently set in Middleware Manager's environment configuration."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"path": h.TraefikStaticConfigPath})
}

// UpdatePathBody struct remains the same
type UpdatePathBody struct {
	Path string `json:"path" binding:"required"`
}

// UpdateTraefikStaticConfigPath function remains the same (with persistence TODO)
func (h *PluginHandler) UpdateTraefikStaticConfigPath(c *gin.Context) {
	var body UpdatePathBody
	if err := c.ShouldBindJSON(&body); err != nil {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	cleanPath := filepath.Clean(body.Path)
	if cleanPath == "" || cleanPath == "." || cleanPath == "/" || strings.HasSuffix(cleanPath, "/") {
		ResponseWithError(c, http.StatusBadRequest, "Invalid configuration path provided. Path cannot be empty, root, relative, or end with a slash.")
		return
	}
	
	// TODO: PERSISTENCE OF TRAEFIK_STATIC_CONFIG_PATH (Critical for this endpoint to be useful across restarts)
	oldPath := h.TraefikStaticConfigPath
	h.TraefikStaticConfigPath = cleanPath // In-memory update
	log.Printf("Traefik static config path updated in memory for this session from '%s' to: '%s'. Persistence requires saving this to the main application configuration.", oldPath, cleanPath)

	c.JSON(http.StatusOK, gin.H{"message": "Traefik static config path updated in memory for the current session. Ensure this change is made persistent in the application's startup configuration (e.g., environment variable or main config file) to survive restarts.", "path": cleanPath})
}


// copyFile function remains the same
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
		return fmt.Errorf("could not copy content from %s to %s: %w", src, dst, err)
	}
	return destinationFile.Sync()
}

// LogInfo function remains the same
func LogInfo(message string) {
	log.Println("INFO:", message)
}