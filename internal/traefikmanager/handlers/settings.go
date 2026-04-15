package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	tmconfig "github.com/hhftechnology/middleware-manager/internal/traefikmanager/config"
	tmtypes "github.com/hhftechnology/middleware-manager/internal/traefikmanager/types"
)

type SettingsHandler struct {
	settings *tmconfig.SettingsStore
	files    *tmconfig.FileStore
	client   *http.Client
}

func NewSettingsHandler(settings *tmconfig.SettingsStore, files *tmconfig.FileStore, client *http.Client) *SettingsHandler {
	return &SettingsHandler{settings: settings, files: files, client: client}
}

func (h *SettingsHandler) Get(c *gin.Context) {
	settings, _, err := h.settings.Load()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load settings", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"domains":            settings.Domains,
		"cert_resolver":      settings.CertResolver,
		"traefik_api_url":    settings.TraefikAPIURL,
		"visible_tabs":       settings.VisibleTabs,
		"disabled_routes":    settings.DisabledRoutes,
		"self_route":         settings.SelfRoute,
		"acme_json_path":     settings.AcmeJSONPath,
		"access_log_path":    settings.AccessLogPath,
		"static_config_path": settings.StaticConfig,
		"auth_enabled":       false,
		"has_password":       false,
		"config_dir_set":     h.files.ActiveConfigDir() != "",
	})
}

func (h *SettingsHandler) Save(c *gin.Context) {
	var request tmtypes.SettingsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request", err)
		return
	}
	settings, _, err := h.settings.Load()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load settings", err)
		return
	}
	if len(request.Domains) > 0 {
		settings.Domains = sanitizeStrings(request.Domains)
	}
	if strings.TrimSpace(request.CertResolver) != "" {
		settings.CertResolver = strings.TrimSpace(request.CertResolver)
	}
	if strings.TrimSpace(request.TraefikAPIURL) != "" {
		settings.TraefikAPIURL = strings.TrimSpace(request.TraefikAPIURL)
	}
	if request.VisibleTabs != nil {
		for key, value := range request.VisibleTabs {
			if _, ok := settings.VisibleTabs[key]; ok {
				settings.VisibleTabs[key] = value
			}
		}
	}
	settings.AcmeJSONPath = strings.TrimSpace(request.AcmeJSONPath)
	settings.AccessLogPath = strings.TrimSpace(request.AccessLogPath)
	settings.StaticConfig = strings.TrimSpace(request.StaticConfigPath)
	if request.SelfRoute != nil {
		settings.SelfRoute = *request.SelfRoute
	}
	if err := h.settings.Save(settings); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save settings", err)
		return
	}
	respondJSON(c, gin.H{"success": true, "settings": settings})
}

func (h *SettingsHandler) SaveTabs(c *gin.Context) {
	var request map[string]bool
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request", err)
		return
	}
	settings, _, err := h.settings.Load()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load settings", err)
		return
	}
	for key, value := range request {
		if _, ok := settings.VisibleTabs[key]; ok {
			settings.VisibleTabs[key] = value
		}
	}
	if err := h.settings.Save(settings); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save tabs", err)
		return
	}
	respondJSON(c, gin.H{"success": true, "visible_tabs": settings.VisibleTabs})
}

func (h *SettingsHandler) GetSelfRoute(c *gin.Context) {
	settings, _, err := h.settings.Load()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load settings", err)
		return
	}
	respondJSON(c, settings.SelfRoute)
}

func (h *SettingsHandler) SaveSelfRoute(c *gin.Context) {
	var request tmtypes.SelfRouteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request", err)
		return
	}
	settings, _, err := h.settings.Load()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load settings", err)
		return
	}
	if strings.TrimSpace(request.Domain) == "" {
		if err := h.deleteSelfRoute(); err != nil {
			respondError(c, http.StatusInternalServerError, "Failed to delete self route", err)
			return
		}
		settings.SelfRoute = tmtypes.SelfRoute{}
	} else {
		routerName := strings.TrimSpace(request.RouterName)
		if routerName == "" {
			routerName = "traefik-manager"
		}
		serviceURL := strings.TrimSpace(request.ServiceURL)
		if serviceURL == "" {
			serviceURL = "http://traefik-manager:5000"
		}
		if err := h.writeSelfRoute(strings.TrimSpace(request.Domain), serviceURL, settings.CertResolver, routerName); err != nil {
			respondError(c, http.StatusInternalServerError, "Failed to save self route", err)
			return
		}
		settings.SelfRoute = tmtypes.SelfRoute{Domain: strings.TrimSpace(request.Domain), ServiceURL: serviceURL, RouterName: routerName}
	}
	if err := h.settings.Save(settings); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save settings", err)
		return
	}
	respondJSON(c, gin.H{"ok": true})
}

func (h *SettingsHandler) TestConnection(c *gin.Context) {
	var request tmtypes.TestConnectionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request", err)
		return
	}
	url := strings.TrimRight(strings.TrimSpace(request.URL), "/")
	if url == "" {
		respondError(c, http.StatusBadRequest, "url is required", nil)
		return
	}
	resp, err := h.client.Get(url + "/api/version")
	if err != nil {
		respondJSON(c, gin.H{"ok": false, "error": "Connection failed"})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respondJSON(c, gin.H{"ok": false, "error": resp.Status})
		return
	}
	respondJSON(c, gin.H{"ok": true})
}

func (h *SettingsHandler) writeSelfRoute(domain, serviceURL, certResolver, routerName string) error {
	if h.files.ActiveConfigDir() != "" {
		content := map[string]any{
			"http": map[string]any{
				"routers": map[string]any{
					routerName: map[string]any{
						"rule":        "Host(`" + domain + "`)",
						"entryPoints": []string{"websecure"},
						"service":     routerName,
						"tls":         map[string]any{"certResolver": certResolver},
					},
				},
				"services": map[string]any{
					routerName: map[string]any{
						"loadBalancer": map[string]any{"servers": []map[string]any{{"url": serviceURL}}},
					},
				},
			},
		}
		return h.files.SaveConfig(filepath.Join(h.files.ActiveConfigDir(), tmconfig.SelfRouteFilename), content)
	}
	path := h.files.ResolveConfigPath("")
	config, err := h.files.LoadConfig(path)
	if err != nil {
		return err
	}
	httpSection := tmconfig.MapFromAny(config["http"])
	routers := tmconfig.MapFromAny(httpSection["routers"])
	services := tmconfig.MapFromAny(httpSection["services"])
	routers[routerName] = map[string]any{"rule": "Host(`" + domain + "`)", "entryPoints": []string{"websecure"}, "service": routerName, "tls": map[string]any{"certResolver": certResolver}}
	services[routerName] = map[string]any{"loadBalancer": map[string]any{"servers": []map[string]any{{"url": serviceURL}}}}
	httpSection["routers"] = routers
	httpSection["services"] = services
	config["http"] = httpSection
	return h.files.SaveConfig(path, config)
}

func (h *SettingsHandler) deleteSelfRoute() error {
	if h.files.ActiveConfigDir() != "" {
		path := filepath.Join(h.files.ActiveConfigDir(), tmconfig.SelfRouteFilename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil
		}
		return os.Remove(path)
	}
	path := h.files.ResolveConfigPath("")
	config, err := h.files.LoadConfig(path)
	if err != nil {
		return err
	}
	httpSection := tmconfig.MapFromAny(config["http"])
	routers := tmconfig.MapFromAny(httpSection["routers"])
	services := tmconfig.MapFromAny(httpSection["services"])
	delete(routers, "traefik-manager")
	delete(services, "traefik-manager")
	httpSection["routers"] = routers
	httpSection["services"] = services
	config["http"] = httpSection
	return h.files.SaveConfig(path, config)
}

func sanitizeStrings(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
