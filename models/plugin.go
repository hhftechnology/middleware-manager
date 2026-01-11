package models

// TraefikPlugin represents a plugin from Traefik's API
// This is used for plugins that are currently loaded/active in Traefik
type TraefikPlugin struct {
	Name        string                 `json:"name"`
	ModuleName  string                 `json:"moduleName"`
	Version     string                 `json:"version"`
	Type        string                 `json:"type,omitempty"`
	Description string                 `json:"description,omitempty"`
	Author      string                 `json:"author,omitempty"`
	Homepage    string                 `json:"homepage,omitempty"`
	Status      string                 `json:"status,omitempty"` // enabled, disabled, error
	Error       string                 `json:"error,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// TraefikPluginOverview represents plugin info from Traefik's /api/overview
type TraefikPluginOverview struct {
	Plugins struct {
		Enabled    []string `json:"enabled,omitempty"`
		Disabled   []string `json:"disabled,omitempty"`
		WithErrors []string `json:"withErrors,omitempty"`
	} `json:"plugins,omitempty"`
}

// TraefikPluginMiddleware represents a plugin used as middleware
// This is extracted from the middlewares API when the middleware type is "plugin"
type TraefikPluginMiddleware struct {
	Name        string                 `json:"name"`
	PluginName  string                 `json:"pluginName"`
	Provider    string                 `json:"provider,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	UsedBy      []string               `json:"usedBy,omitempty"`
}

// PluginResponse represents the combined plugin information for the API response
type PluginResponse struct {
	// Basic info
	Name        string `json:"name"`
	ModuleName  string `json:"moduleName"`
	Version     string `json:"version"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Author      string `json:"author,omitempty"`
	Homepage    string `json:"homepage,omitempty"`

	// Status from Traefik
	Status   string `json:"status"` // enabled, disabled, error, not_loaded
	Error    string `json:"error,omitempty"`
	Provider string `json:"provider,omitempty"`

	// Installation info
	IsInstalled      bool   `json:"isInstalled"`
	InstalledVersion string `json:"installedVersion,omitempty"`

	// Usage info
	UsageCount int      `json:"usageCount"` // Number of middlewares using this plugin
	UsedBy     []string `json:"usedBy,omitempty"`

	// Config
	Config map[string]interface{} `json:"config,omitempty"`
}

// PluginInstallRequest represents a request to install a plugin
type PluginInstallRequest struct {
	ModuleName string `json:"moduleName" binding:"required"`
	Version    string `json:"version,omitempty"`
}

// PluginRemoveRequest represents a request to remove a plugin
type PluginRemoveRequest struct {
	ModuleName string `json:"moduleName" binding:"required"`
}
