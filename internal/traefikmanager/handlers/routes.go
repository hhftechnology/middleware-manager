package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	tmconfig "github.com/hhftechnology/middleware-manager/internal/traefikmanager/config"
	tmtypes "github.com/hhftechnology/middleware-manager/internal/traefikmanager/types"
)

type RouteHandler struct {
	files    *tmconfig.FileStore
	settings *tmconfig.SettingsStore
}

func NewRouteHandler(files *tmconfig.FileStore, settings *tmconfig.SettingsStore) *RouteHandler {
	return &RouteHandler{files: files, settings: settings}
}

func (h *RouteHandler) List(c *gin.Context) {
	settings, _, err := h.settings.Load()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load settings", err)
		return
	}
	apps, middlewares, err := h.loadAll(settings)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load routes", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"apps": apps, "middlewares": middlewares})
}

func (h *RouteHandler) ListAll(c *gin.Context) {
	h.List(c)
}

func (h *RouteHandler) Configs(c *gin.Context) {
	paths := h.files.ConfigPaths()
	items := make([]gin.H, 0, len(paths))
	for _, path := range paths {
		items = append(items, gin.H{"label": filepath.Base(path), "path": path})
	}
	respondJSON(c, gin.H{"files": items, "configDirSet": h.files.ActiveConfigDir() != ""})
}

func (h *RouteHandler) RouterNames(c *gin.Context) {
	names := make([]string, 0)
	seen := map[string]struct{}{}
	for _, path := range h.files.ConfigPaths() {
		config, err := h.files.LoadConfig(path)
		if err != nil {
			continue
		}
		for _, proto := range []string{"http", "tcp", "udp"} {
			for name := range tmconfig.MapFromAny(tmconfig.MapFromAny(config[proto])["routers"]) {
				if _, ok := seen[name]; ok {
					continue
				}
				seen[name] = struct{}{}
				names = append(names, name)
			}
		}
	}
	sort.Strings(names)
	respondJSON(c, names)
}

func (h *RouteHandler) Create(c *gin.Context) {
	var request tmtypes.RouteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request", err)
		return
	}
	if err := h.saveRoute("", request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func (h *RouteHandler) Update(c *gin.Context) {
	var request tmtypes.RouteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request", err)
		return
	}
	if err := h.saveRoute(c.Param("id"), request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	respondJSON(c, gin.H{"ok": true})
}

func (h *RouteHandler) Delete(c *gin.Context) {
	if err := h.deleteRoute(c.Param("id")); err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		respondError(c, status, err.Error(), err)
		return
	}
	respondJSON(c, gin.H{"ok": true})
}

func (h *RouteHandler) Toggle(c *gin.Context) {
	var request tmtypes.ToggleRouteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request", err)
		return
	}
	if err := h.toggleRoute(c.Param("id"), request.Enable); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	respondJSON(c, gin.H{"ok": true})
}

func (h *RouteHandler) loadAll(settings tmtypes.Settings) ([]tmtypes.App, []tmtypes.MiddlewareEntry, error) {
	apps := make([]tmtypes.App, 0)
	middlewares := make([]tmtypes.MiddlewareEntry, 0)
	multi := h.files.MultiConfig() || h.files.ActiveConfigDir() != ""
	for _, path := range h.files.ConfigPaths() {
		config, err := h.files.LoadConfig(path)
		if err != nil {
			return nil, nil, err
		}
		configFile := ""
		if multi {
			configFile = filepath.Base(path)
		}
		apps = append(apps, tmconfig.BuildApps(config, configFile, multi)...)
		items, err := tmconfig.BuildMiddlewares(config, configFile)
		if err != nil {
			return nil, nil, err
		}
		middlewares = append(middlewares, items...)
	}
	apps = tmconfig.ApplyDisabledRoutes(settings, apps)
	sort.Slice(apps, func(i, j int) bool { return apps[i].Name < apps[j].Name })
	return apps, middlewares, nil
}

func (h *RouteHandler) saveRoute(originalID string, request tmtypes.RouteRequest) error {
	serviceNameInput := strings.TrimSpace(request.ServiceName)
	if serviceNameInput == "" {
		return fmt.Errorf("serviceName is required")
	}
	protocol := strings.ToLower(strings.TrimSpace(request.Protocol))
	if protocol == "" {
		protocol = "http"
	}
	if protocol != "http" && protocol != "tcp" && protocol != "udp" {
		return fmt.Errorf("invalid protocol")
	}

	settings, _, err := h.settings.Load()
	if err != nil {
		return err
	}

	targetPath := h.files.ResolveConfigPath(request.ConfigFile)
	if targetPath == "" {
		targetPath = h.files.ResolveConfigPath("")
	}
	config, err := h.files.LoadConfig(targetPath)
	if err != nil {
		return err
	}
	if _, err := h.files.CreateBackup(targetPath); err != nil {
		return err
	}

	routerName := serviceNameInput
	serviceName := routerName + "-service"
	oldName := tmconfig.RouteNameFromID(originalID)
	if oldName != "" {
		removeRoute(config, oldName)
	}

	switch protocol {
	case "http":
		h.saveHTTP(config, settings, request, routerName, serviceName)
	case "tcp":
		h.saveTCP(config, settings, request, routerName, serviceName)
	case "udp":
		h.saveUDP(config, request, routerName, serviceName)
	}
	return h.files.SaveConfig(targetPath, config)
}

func (h *RouteHandler) saveHTTP(config map[string]any, settings tmtypes.Settings, request tmtypes.RouteRequest, routerName, serviceName string) {
	httpSection := tmconfig.MapFromAny(config["http"])
	routers := tmconfig.MapFromAny(httpSection["routers"])
	services := tmconfig.MapFromAny(httpSection["services"])
	transports := tmconfig.MapFromAny(httpSection["serversTransports"])

	domains := request.Domains
	if len(domains) == 0 {
		domains = settings.Domains
	}
	rule := strings.TrimSpace(request.Rule)
	if rule == "" {
		rule = buildHTTPRule(strings.TrimSpace(request.Subdomain), domains)
	}

	target := strings.TrimSpace(request.Target)
	if !strings.Contains(target, "://") {
		scheme := strings.TrimSpace(request.Scheme)
		if scheme == "" {
			scheme = "http"
		}
		target = scheme + "://" + strings.TrimSpace(request.Target)
		if strings.TrimSpace(request.TargetPort) != "" {
			target += ":" + strings.TrimSpace(request.TargetPort)
		}
	}

	router := map[string]any{
		"rule":        rule,
		"entryPoints": defaultEntryPoints(request.EntryPoints, []string{"websecure"}),
		"service":     serviceName,
	}
	certResolver := strings.TrimSpace(request.CertResolver)
	if certResolver == "" {
		certResolver = settings.CertResolver
	}
	if certResolver != "__none__" && certResolver != "" {
		router["tls"] = map[string]any{"certResolver": certResolver}
	}
	if len(request.Middlewares) > 0 {
		router["middlewares"] = request.Middlewares
	}

	passHostHeader := true
	if request.PassHostHeader != nil {
		passHostHeader = *request.PassHostHeader
	}
	loadBalancer := map[string]any{"servers": []map[string]any{{"url": target}}}
	if !passHostHeader {
		loadBalancer["passHostHeader"] = false
	}
	if request.InsecureSkipVerify {
		transportName := routerName + "-transport"
		loadBalancer["serversTransport"] = transportName
		transports[transportName] = map[string]any{"insecureSkipVerify": true}
	} else {
		delete(transports, routerName+"-transport")
	}

	routers[routerName] = router
	services[serviceName] = map[string]any{"loadBalancer": loadBalancer}
	httpSection["routers"] = routers
	httpSection["services"] = services
	if len(transports) > 0 {
		httpSection["serversTransports"] = transports
	} else {
		delete(httpSection, "serversTransports")
	}
	config["http"] = httpSection
}

func (h *RouteHandler) saveTCP(config map[string]any, settings tmtypes.Settings, request tmtypes.RouteRequest, routerName, serviceName string) {
	section := tmconfig.MapFromAny(config["tcp"])
	routers := tmconfig.MapFromAny(section["routers"])
	services := tmconfig.MapFromAny(section["services"])
	rule := strings.TrimSpace(request.Rule)
	if rule == "" {
		domain := ""
		if len(request.Domains) > 0 {
			domain = request.Domains[0]
		} else if len(settings.Domains) > 0 {
			domain = settings.Domains[0]
		}
		if strings.TrimSpace(request.Subdomain) != "" && domain != "" {
			rule = fmt.Sprintf("HostSNI(`%s.%s`)", strings.TrimSpace(request.Subdomain), domain)
		} else {
			rule = "HostSNI(`*`)"
		}
	}
	certResolver := strings.TrimSpace(request.CertResolver)
	if certResolver == "" {
		certResolver = settings.CertResolver
	}
	routers[routerName] = map[string]any{
		"rule":        rule,
		"entryPoints": defaultEntryPoints(request.EntryPoints, []string{"websecure"}),
		"tls":         map[string]any{"certResolver": certResolver},
		"service":     serviceName,
	}
	services[serviceName] = map[string]any{"loadBalancer": map[string]any{"servers": []map[string]any{{"address": joinTarget(request.Target, request.TargetPort)}}}}
	section["routers"] = routers
	section["services"] = services
	config["tcp"] = section
}

func (h *RouteHandler) saveUDP(config map[string]any, request tmtypes.RouteRequest, routerName, serviceName string) {
	section := tmconfig.MapFromAny(config["udp"])
	routers := tmconfig.MapFromAny(section["routers"])
	services := tmconfig.MapFromAny(section["services"])
	routers[routerName] = map[string]any{
		"entryPoints": defaultEntryPoints(request.EntryPoints, []string{}),
		"service":     serviceName,
	}
	services[serviceName] = map[string]any{"loadBalancer": map[string]any{"servers": []map[string]any{{"address": joinTarget(request.Target, request.TargetPort)}}}}
	section["routers"] = routers
	section["services"] = services
	config["udp"] = section
}

func (h *RouteHandler) deleteRoute(id string) error {
	plainID := tmconfig.RouteNameFromID(id)
	targets := h.files.ConfigPaths()
	if strings.Contains(id, "::") {
		if path := h.files.ResolveConfigPath(strings.SplitN(id, "::", 2)[0]); path != "" {
			targets = []string{path}
		}
	}
	for _, path := range targets {
		config, err := h.files.LoadConfig(path)
		if err != nil {
			return err
		}
		if !routeExists(config, plainID) {
			continue
		}
		if _, err := h.files.CreateBackup(path); err != nil {
			return err
		}
		removeRoute(config, plainID)
		return h.files.SaveConfig(path, config)
	}
	return fmt.Errorf("route %q not found", plainID)
}

func (h *RouteHandler) toggleRoute(id string, enable bool) error {
	settings, _, err := h.settings.Load()
	if err != nil {
		return err
	}
	plainID := tmconfig.RouteNameFromID(id)
	if enable {
		disabled, ok := settings.DisabledRoutes[id]
		if !ok {
			return nil
		}
		targetPath := h.files.ResolveConfigPath(disabled.ConfigFile)
		if targetPath == "" {
			targetPath = h.files.ResolveConfigPath("")
		}
		config, err := h.files.LoadConfig(targetPath)
		if err != nil {
			return err
		}
		if _, err := h.files.CreateBackup(targetPath); err != nil {
			return err
		}
		section := tmconfig.MapFromAny(config[disabled.Protocol])
		routers := tmconfig.MapFromAny(section["routers"])
		services := tmconfig.MapFromAny(section["services"])
		routers[plainID] = disabled.Router
		serviceName := tmconfig.StringFromAny(disabled.Router["service"])
		if serviceName == "" {
			serviceName = plainID + "-service"
		}
		services[serviceName] = disabled.Service
		section["routers"] = routers
		section["services"] = services
		config[disabled.Protocol] = section
		if err := h.files.SaveConfig(targetPath, config); err != nil {
			return err
		}
		delete(settings.DisabledRoutes, id)
		return h.settings.Save(settings)
	}

	targets := h.files.ConfigPaths()
	if strings.Contains(id, "::") {
		if path := h.files.ResolveConfigPath(strings.SplitN(id, "::", 2)[0]); path != "" {
			targets = []string{path}
		}
	}
	for _, path := range targets {
		config, err := h.files.LoadConfig(path)
		if err != nil {
			return err
		}
		for _, proto := range []string{"http", "tcp", "udp"} {
			section := tmconfig.MapFromAny(config[proto])
			routers := tmconfig.MapFromAny(section["routers"])
			router := tmconfig.MapFromAny(routers[plainID])
			if len(router) == 0 {
				continue
			}
			services := tmconfig.MapFromAny(section["services"])
			serviceName := tmconfig.StringFromAny(router["service"])
			service := tmconfig.MapFromAny(services[serviceName])
			delete(routers, plainID)
			delete(services, serviceName)
			section["routers"] = routers
			section["services"] = services
			config[proto] = section
			if _, err := h.files.CreateBackup(path); err != nil {
				return err
			}
			if err := h.files.SaveConfig(path, config); err != nil {
				return err
			}
			configFile := ""
			if h.files.MultiConfig() || h.files.ActiveConfigDir() != "" {
				configFile = filepath.Base(path)
			}
			settings.DisabledRoutes[id] = tmtypes.DisabledRoute{Protocol: proto, Router: router, Service: service, ConfigFile: configFile}
			return h.settings.Save(settings)
		}
	}
	return nil
}

func routeExists(config map[string]any, name string) bool {
	for _, proto := range []string{"http", "tcp", "udp"} {
		if _, ok := tmconfig.MapFromAny(tmconfig.MapFromAny(config[proto])["routers"])[name]; ok {
			return true
		}
	}
	return false
}

func removeRoute(config map[string]any, name string) {
	for _, proto := range []string{"http", "tcp", "udp"} {
		section := tmconfig.MapFromAny(config[proto])
		routers := tmconfig.MapFromAny(section["routers"])
		router := tmconfig.MapFromAny(routers[name])
		serviceName := tmconfig.StringFromAny(router["service"])
		delete(routers, name)
		services := tmconfig.MapFromAny(section["services"])
		if serviceName != "" {
			delete(services, serviceName)
		}
		if proto == "http" {
			transports := tmconfig.MapFromAny(section["serversTransports"])
			delete(transports, name+"-transport")
			if len(transports) > 0 {
				section["serversTransports"] = transports
			} else {
				delete(section, "serversTransports")
			}
		}
		section["routers"] = routers
		section["services"] = services
		config[proto] = section
	}
}

func buildHTTPRule(subdomain string, domains []string) string {
	if subdomain != "" && strings.Contains(subdomain, ".") {
		return fmt.Sprintf("Host(`%s`)", subdomain)
	}
	if len(domains) == 0 {
		return ""
	}
	hosts := make([]string, 0, len(domains))
	for _, domain := range domains {
		if strings.TrimSpace(domain) == "" {
			continue
		}
		if subdomain == "" {
			hosts = append(hosts, fmt.Sprintf("Host(`%s`)", domain))
		} else {
			hosts = append(hosts, fmt.Sprintf("Host(`%s.%s`)", subdomain, domain))
		}
	}
	return strings.Join(hosts, " || ")
}

func joinTarget(host, port string) string {
	host = strings.TrimSpace(host)
	port = strings.TrimSpace(port)
	if port == "" {
		return host
	}
	return host + ":" + port
}

func defaultEntryPoints(entries []string, fallback []string) []string {
	if len(entries) == 0 {
		return fallback
	}
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		if trimmed := strings.TrimSpace(entry); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
