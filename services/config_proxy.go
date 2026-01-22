package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/hhftechnology/middleware-manager/database"
	"github.com/hhftechnology/middleware-manager/models"
)

// ProxiedTraefikConfig represents the full Traefik config structure (JSON format)
type ProxiedTraefikConfig struct {
	HTTP *HTTPConfig `json:"http,omitempty"`
	TCP  *TCPConfig  `json:"tcp,omitempty"`
	UDP  *UDPConfig  `json:"udp,omitempty"`
	TLS  *TLSConfig  `json:"tls,omitempty"`
}

// HTTPConfig represents HTTP configuration section
type HTTPConfig struct {
	Middlewares map[string]interface{} `json:"middlewares,omitempty"`
	Routers     map[string]interface{} `json:"routers,omitempty"`
	Services    map[string]interface{} `json:"services,omitempty"`
}

// TCPConfig represents TCP configuration section
type TCPConfig struct {
	Routers  map[string]interface{} `json:"routers,omitempty"`
	Services map[string]interface{} `json:"services,omitempty"`
}

// UDPConfig represents UDP configuration section
type UDPConfig struct {
	Routers  map[string]interface{} `json:"routers,omitempty"`
	Services map[string]interface{} `json:"services,omitempty"`
}

// TLSConfig represents TLS configuration section
type TLSConfig struct {
	Options map[string]interface{} `json:"options,omitempty"`
}

// OrderedRouter represents a Traefik HTTP router with fields in Pangolin's order.
// The JSON field order matches Pangolin API output for consistency.
type OrderedRouter struct {
	EntryPoints []string               `json:"entryPoints,omitempty"`
	Middlewares []string               `json:"middlewares,omitempty"`
	Service     string                 `json:"service,omitempty"`
	Rule        string                 `json:"rule,omitempty"`
	Priority    int                    `json:"priority,omitempty"`
	TLS         *OrderedTLSConfig      `json:"tls,omitempty"`
}

// OrderedTLSConfig represents TLS config for a router with Pangolin's field order.
type OrderedTLSConfig struct {
	CertResolver string   `json:"certResolver,omitempty"`
	Domains      []string `json:"domains,omitempty"`
	Options      string   `json:"options,omitempty"`
}

// OrderedMiddleware represents a middleware with Pangolin's field order.
type OrderedMiddleware struct {
	RedirectScheme map[string]interface{} `json:"redirectScheme,omitempty"`
	Plugin         map[string]interface{} `json:"plugin,omitempty"`
	Headers        map[string]interface{} `json:"headers,omitempty"`
}

type middlewareWithPriority struct {
	ID       string
	Priority int
}

type mtlsConfigData struct {
	CACertPath      string
	Rules           []interface{}
	RequestHeaders  map[string]string
	RejectMessage   string
	RejectCode      int
	RefreshInterval string
}

type resourceData struct {
	ID                   string
	Host                 string
	ServiceID            string
	Entrypoints          string
	TLSDomains           string
	CustomHeaders        string
	RouterPriority       int
	SourceType           string
	MTLSEnabled          bool
	MTLSRules            sql.NullString
	MTLSRequestHdrs      sql.NullString
	MTLSRejectMsg        sql.NullString
	MTLSRejectCode       sql.NullInt64
	MTLSRefresh          sql.NullString
	MTLSExternal         sql.NullString
	TLSHardeningEnabled  bool
	SecureHeadersEnabled bool
	Middlewares          []middlewareWithPriority
	CustomServiceID      sql.NullString
}

// securityConfigData holds global security settings from the database
type securityConfigData struct {
	TLSHardeningEnabled  bool
	SecureHeadersEnabled bool
	SecureHeaders        models.SecureHeadersConfig
}

// ConfigProxy fetches config from Pangolin and merges MW-manager additions
type ConfigProxy struct {
	db            *database.DB
	configManager *ConfigManager
	pangolinURL   string
	httpClient    *http.Client

	// Caching
	cache         *ProxiedTraefikConfig
	cacheExpiry   time.Time
	cacheDuration time.Duration
	cacheMutex    sync.RWMutex
}

// NewConfigProxy creates a new config proxy instance
func NewConfigProxy(db *database.DB, configManager *ConfigManager, pangolinURL string) *ConfigProxy {
	return &ConfigProxy{
		db:            db,
		configManager: configManager,
		pangolinURL:   pangolinURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cacheDuration: 5 * time.Second, // Match typical Traefik poll interval
	}
}

// GetMergedConfig returns the merged Pangolin + MW-manager configuration
func (cp *ConfigProxy) GetMergedConfig() (*ProxiedTraefikConfig, error) {
	// Try to use cached config
	cp.cacheMutex.RLock()
	if cp.cache != nil && time.Now().Before(cp.cacheExpiry) {
		defer cp.cacheMutex.RUnlock()
		return cp.cache, nil
	}
	cp.cacheMutex.RUnlock()

	// Acquire write lock for cache update
	cp.cacheMutex.Lock()
	defer cp.cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if cp.cache != nil && time.Now().Before(cp.cacheExpiry) {
		return cp.cache, nil
	}

	// Fetch fresh config from Pangolin
	config, err := cp.fetchPangolinConfig()
	if err != nil {
		// Return stale cache on error if available
		if cp.cache != nil {
			log.Printf("Warning: Pangolin fetch failed, using stale cache: %v", err)
			return cp.cache, nil
		}
		return nil, fmt.Errorf("failed to fetch Pangolin config: %w", err)
	}

	// Merge MW-manager additions
	if err := cp.mergeMiddlewareManagerConfig(config); err != nil {
		return nil, fmt.Errorf("failed to merge MW-manager config: %w", err)
	}

	// Remove empty protocol sections so Traefik doesn't reject blank configs
	cp.pruneEmptySections(config)

	// Normalize router field ordering to match Pangolin's JSON format
	cp.normalizeRouterOrder(config)

	// Normalize middleware field ordering to match Pangolin's JSON format
	cp.normalizeMiddlewareOrder(config)

	// Update cache
	cp.cache = config
	cp.cacheExpiry = time.Now().Add(cp.cacheDuration)

	return config, nil
}

// InvalidateCache forces the next GetMergedConfig call to fetch fresh data
func (cp *ConfigProxy) InvalidateCache() {
	cp.cacheMutex.Lock()
	defer cp.cacheMutex.Unlock()
	cp.cacheExpiry = time.Now().Add(-1 * time.Second) // Expire immediately
}

// fetchPangolinConfig fetches the Traefik configuration from Pangolin API
func (cp *ConfigProxy) fetchPangolinConfig() (*ProxiedTraefikConfig, error) {
	// Use configured Pangolin URL or get from config manager
	pangolinURL := cp.pangolinURL
	if pangolinURL == "" {
		// Try to get from environment or config
		pangolinURL = os.Getenv("PANGOLIN_URL")
		if pangolinURL == "" {
			pangolinURL = "http://pangolin:3001"
		}
	}

	// Build the correct URL - handle both base URL and URLs that include /api/v1
	pangolinURL = strings.TrimSuffix(pangolinURL, "/")
	var url string
	if strings.HasSuffix(pangolinURL, "/api/v1") {
		// URL already includes /api/v1, just append traefik-config
		url = pangolinURL + "/traefik-config"
	} else {
		// Base URL, append full path
		url = pangolinURL + "/api/v1/traefik-config"
	}

	if shouldLogInfo() {
		log.Printf("Fetching Pangolin config from: %s", url)
	}

	resp, err := cp.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Pangolin returned status %d: %s", resp.StatusCode, string(body))
	}

	var config ProxiedTraefikConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode Pangolin response: %w", err)
	}

	// Initialize nil maps
	cp.initializeConfigMaps(&config)

	return &config, nil
}

// initializeConfigMaps ensures all config maps are initialized
func (cp *ConfigProxy) initializeConfigMaps(config *ProxiedTraefikConfig) {
	if config.HTTP == nil {
		config.HTTP = &HTTPConfig{}
	}
	if config.HTTP.Middlewares == nil {
		config.HTTP.Middlewares = make(map[string]interface{})
	}
	if config.HTTP.Routers == nil {
		config.HTTP.Routers = make(map[string]interface{})
	}
	if config.HTTP.Services == nil {
		config.HTTP.Services = make(map[string]interface{})
	}

	if config.TCP == nil {
		config.TCP = &TCPConfig{}
	}
	if config.TCP.Routers == nil {
		config.TCP.Routers = make(map[string]interface{})
	}
	if config.TCP.Services == nil {
		config.TCP.Services = make(map[string]interface{})
	}

	if config.UDP == nil {
		config.UDP = &UDPConfig{}
	}
	if config.UDP.Routers == nil {
		config.UDP.Routers = make(map[string]interface{})
	}
	if config.UDP.Services == nil {
		config.UDP.Services = make(map[string]interface{})
	}

	if config.TLS == nil {
		config.TLS = &TLSConfig{}
	}
	if config.TLS.Options == nil {
		config.TLS.Options = make(map[string]interface{})
	}
}

// pruneEmptySections removes protocol blocks that contain no routers or services.
// Traefik rejects empty protocol sections from the HTTP provider response.
func (cp *ConfigProxy) pruneEmptySections(config *ProxiedTraefikConfig) {
	if config == nil {
		return
	}

	if config.TCP != nil {
		if len(config.TCP.Routers) == 0 && len(config.TCP.Services) == 0 {
			config.TCP = nil
		}
	}

	if config.UDP != nil {
		if len(config.UDP.Routers) == 0 && len(config.UDP.Services) == 0 {
			config.UDP = nil
		}
	}

	if config.TLS != nil {
		if len(config.TLS.Options) == 0 {
			config.TLS = nil
		}
	}
}

// mergeMiddlewareManagerConfig merges MW-manager middlewares into the config
// NOTE: Routers and services come from Pangolin API and are NOT modified here.
func (cp *ConfigProxy) mergeMiddlewareManagerConfig(config *ProxiedTraefikConfig) error {
	// Load resources and their middleware assignments
	resources, err := cp.fetchResourceData()
	if err != nil {
		return fmt.Errorf("failed to fetch resources: %w", err)
	}

	// Load global security config
	securityCfg, err := cp.loadSecurityConfig()
	if err != nil {
		log.Printf("Warning: failed to load security config: %v", err)
		securityCfg = nil
	}

	assignedMiddlewareIDs := make(map[string]struct{})
	hasMTLSResources := false
	hasTLSHardeningResources := false

	for _, res := range resources {
		if res.MTLSEnabled {
			hasMTLSResources = true
		}
		if res.TLSHardeningEnabled && !res.MTLSEnabled {
			hasTLSHardeningResources = true
		}
		for _, mw := range res.Middlewares {
			assignedMiddlewareIDs[mw.ID] = struct{}{}
		}
	}

	var mtlsCfg *mtlsConfigData
	if hasMTLSResources {
		cfg, err := cp.loadGlobalMTLSConfig()
		if err != nil {
			return fmt.Errorf("failed to load global mTLS config: %w", err)
		}
		mtlsCfg = cfg
		if mtlsCfg != nil {
			cp.applyTLSOptions(config, mtlsCfg)
		}
	}

	// Apply TLS hardening options if any resource has it enabled (and not mTLS)
	if hasTLSHardeningResources {
		cp.applyTLSHardeningOptions(config)
	}

	// Only add MW-manager middlewares that are assigned to resources/routers
	if len(assignedMiddlewareIDs) > 0 {
		if err := cp.applyMiddlewares(config, assignedMiddlewareIDs); err != nil {
			return fmt.Errorf("failed to apply middlewares: %w", err)
		}
	}

	// Apply resource-specific overrides (middleware attachments, priorities, headers, mtls, security)
	if len(resources) > 0 {
		if err := cp.applyResourceOverrides(config, resources, mtlsCfg, securityCfg); err != nil {
			return fmt.Errorf("failed to apply resource overrides: %w", err)
		}
	}

	// Sanitize mtlswhitelist requestHeaders to ensure map type (Traefik plugin is strict)
	cp.sanitizeMTLSWhitelist(config)

	return nil
}

// applyMiddlewares adds custom middlewares from the database
func (cp *ConfigProxy) applyMiddlewares(config *ProxiedTraefikConfig, allowedIDs map[string]struct{}) error {
	rows, err := cp.db.Query("SELECT id, name, type, config FROM middlewares")
	if err != nil {
		return fmt.Errorf("failed to fetch middlewares: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, name, typ, configStr string
		if err := rows.Scan(&id, &name, &typ, &configStr); err != nil {
			log.Printf("Failed to scan middleware: %v", err)
			continue
		}

		// Skip middlewares not assigned to any resource/router
		if allowedIDs != nil {
			if _, ok := allowedIDs[id]; !ok {
				continue
			}
		}

		var middlewareConfig map[string]interface{}
		if err := json.Unmarshal([]byte(configStr), &middlewareConfig); err != nil {
			log.Printf("Failed to parse middleware config for %s: %v", name, err)
			continue
		}

		// Use the centralized processing logic from models package
		middlewareConfig = models.ProcessMiddlewareConfig(typ, middlewareConfig)

		// Add middleware using its ID as the key
		config.HTTP.Middlewares[id] = map[string]interface{}{
			typ: middlewareConfig,
		}

		if shouldLog() {
			log.Printf("Added middleware %s (%s) to config", id, typ)
		}
	}

	return rows.Err()
}

// applyServices adds custom services from the database
func (cp *ConfigProxy) applyServices(config *ProxiedTraefikConfig) error {
	rows, err := cp.db.Query("SELECT id, name, type, config FROM services")
	if err != nil {
		return fmt.Errorf("failed to fetch services: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, name, typ, configStr string
		if err := rows.Scan(&id, &name, &typ, &configStr); err != nil {
			log.Printf("Failed to scan service: %v", err)
			continue
		}

		var serviceConfig map[string]interface{}
		if err := json.Unmarshal([]byte(configStr), &serviceConfig); err != nil {
			log.Printf("Failed to parse service config for %s: %v", name, err)
			continue
		}

		// Use the centralized processing logic from models package
		serviceConfig = models.ProcessServiceConfig(typ, serviceConfig)

		// Determine protocol based on service type and config
		protocol := cp.determineServiceProtocol(typ, serviceConfig)

		serviceEntry := map[string]interface{}{typ: serviceConfig}

		switch protocol {
		case "http":
			config.HTTP.Services[id] = serviceEntry
		case "tcp":
			config.TCP.Services[id] = serviceEntry
		case "udp":
			config.UDP.Services[id] = serviceEntry
		}

		if shouldLog() {
			log.Printf("Added service %s (%s, %s) to config", id, typ, protocol)
		}
	}

	return rows.Err()
}

// applyResourceOverrides applies middleware assignments and other overrides to routers
func (cp *ConfigProxy) applyResourceOverrides(config *ProxiedTraefikConfig, resources []*resourceData, mtlsCfg *mtlsConfigData, securityCfg *securityConfigData) error {
	for _, resource := range resources {
		// Find matching router by host
		routerKey, router := cp.findMatchingRouter(config.HTTP.Routers, resource.Host)

		if routerKey == "" {
			// No matching router found, might need to create one
			if shouldLog() {
				log.Printf("No matching router found for resource %s (host: %s)", resource.ID, resource.Host)
			}
			continue
		}

		// Sort middlewares by priority (highest first)
		sort.SliceStable(resource.Middlewares, func(i, j int) bool {
			return resource.Middlewares[i].Priority > resource.Middlewares[j].Priority
		})

		// Build middleware list (mTLS first, then secure headers, then custom headers, then assigned)
		var newMiddlewares []string

		if resource.MTLSEnabled && mtlsCfg != nil {
			mtlsMiddlewareName, err := cp.ensureResourceMTLSMiddleware(config, resource, mtlsCfg)
			if err != nil {
				log.Printf("Failed to build mTLS middleware for resource %s: %v", resource.ID, err)
			} else if mtlsMiddlewareName != "" {
				newMiddlewares = append(newMiddlewares, mtlsMiddlewareName)

				// Add mTLS TLS options on the router
				if tlsConfig, ok := router["tls"].(map[string]interface{}); ok {
					if _, ok := tlsConfig["options"]; !ok {
						tlsConfig["options"] = "mtls-verify"
					}
				} else {
					router["tls"] = map[string]interface{}{
						"options": "mtls-verify",
					}
				}
			}
		}

		// Apply TLS hardening if enabled for this resource AND mTLS is NOT enabled
		// (mTLS already includes TLS hardening via mtls-verify options)
		if resource.TLSHardeningEnabled && !resource.MTLSEnabled {
			if tlsConfig, ok := router["tls"].(map[string]interface{}); ok {
				tlsConfig["options"] = "tls-hardened"
			} else {
				router["tls"] = map[string]interface{}{
					"options": "tls-hardened",
				}
			}
		}

		// Add secure headers middleware if enabled for this resource
		if resource.SecureHeadersEnabled && securityCfg != nil && securityCfg.SecureHeadersEnabled {
			secureHeadersMiddlewareName := cp.ensureSecureHeadersMiddleware(config, resource, securityCfg)
			if secureHeadersMiddlewareName != "" {
				newMiddlewares = append(newMiddlewares, secureHeadersMiddlewareName)
			}
		}

		// Add custom headers middleware if configured
		if resource.CustomHeaders != "" && resource.CustomHeaders != "{}" && resource.CustomHeaders != "null" {
			var headersMap map[string]string
			if err := json.Unmarshal([]byte(resource.CustomHeaders), &headersMap); err == nil && len(headersMap) > 0 {
				middlewareName := fmt.Sprintf("%s-customheaders", resource.ID)
				config.HTTP.Middlewares[middlewareName] = map[string]interface{}{
					"headers": map[string]interface{}{"customRequestHeaders": headersMap},
				}
				newMiddlewares = append(newMiddlewares, middlewareName)
			}
		}

		// Add assigned middlewares
		for _, mw := range resource.Middlewares {
			newMiddlewares = append(newMiddlewares, mw.ID)
		}

		// Get existing middlewares from router
		existingMiddlewares := cp.getRouterMiddlewares(router)

		// Merge middlewares (MW-manager additions first, then existing)
		finalMiddlewares := newMiddlewares
		for _, em := range existingMiddlewares {
			// Avoid duplicates
			found := false
			for _, nm := range newMiddlewares {
				if em == nm {
					found = true
					break
				}
			}
			if !found {
				finalMiddlewares = append(finalMiddlewares, em)
			}
		}

		// Update router
		if len(finalMiddlewares) > 0 {
			router["middlewares"] = finalMiddlewares
		}

		// Update priority if customized
		if resource.RouterPriority != 100 {
			router["priority"] = resource.RouterPriority
		}

		// Update custom service if configured
		if resource.CustomServiceID.Valid && resource.CustomServiceID.String != "" {
			router["service"] = resource.CustomServiceID.String
		}

		config.HTTP.Routers[routerKey] = router

		if shouldLog() {
			log.Printf("Applied overrides to router %s (resource: %s)", routerKey, resource.ID)
		}
	}

	return nil
}

// ensureResourceMTLSMiddleware builds and registers a per-resource mtlswhitelist middleware
func (cp *ConfigProxy) ensureResourceMTLSMiddleware(config *ProxiedTraefikConfig, resource *resourceData, mtlsCfg *mtlsConfigData) (string, error) {
	if mtlsCfg == nil || mtlsCfg.CACertPath == "" {
		return "", fmt.Errorf("mTLS enabled for resource %s but no CA certificate configured", resource.ID)
	}

	pluginConfig := map[string]interface{}{
		"caFiles": []string{mtlsCfg.CACertPath},
	}

	if len(mtlsCfg.Rules) > 0 {
		pluginConfig["rules"] = append([]interface{}{}, mtlsCfg.Rules...)
	}
	if len(mtlsCfg.RequestHeaders) > 0 {
		pluginConfig["requestHeaders"] = mtlsCfg.RequestHeaders
	}
	if mtlsCfg.RejectMessage != "" || mtlsCfg.RejectCode > 0 {
		code := mtlsCfg.RejectCode
		if code == 0 {
			code = 403
		}
		entry := map[string]interface{}{
			"code": code,
		}
		if mtlsCfg.RejectMessage != "" {
			entry["message"] = mtlsCfg.RejectMessage
		}
		pluginConfig["rejectMessage"] = entry
	}
	if mtlsCfg.RefreshInterval != "" {
		pluginConfig["refreshInterval"] = mtlsCfg.RefreshInterval
	}

	// Resource-level overrides
	if resource.MTLSRules.Valid && strings.TrimSpace(resource.MTLSRules.String) != "" {
		var rules []interface{}
		if err := json.Unmarshal([]byte(resource.MTLSRules.String), &rules); err == nil {
			pluginConfig["rules"] = rules
		} else {
			log.Printf("Failed to parse mtls_rules for resource %s: %v", resource.ID, err)
		}
	}

	if resource.MTLSRequestHdrs.Valid && strings.TrimSpace(resource.MTLSRequestHdrs.String) != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(resource.MTLSRequestHdrs.String), &headers); err == nil && len(headers) > 0 {
			pluginConfig["requestHeaders"] = headers
		} else if err != nil {
			log.Printf("Failed to parse mtls_request_headers for resource %s: %v", resource.ID, err)
		}
	}

	// Reject message + code
	if (resource.MTLSRejectMsg.Valid && strings.TrimSpace(resource.MTLSRejectMsg.String) != "") || resource.MTLSRejectCode.Valid {
		code := mtlsCfg.RejectCode
		if resource.MTLSRejectCode.Valid {
			code = int(resource.MTLSRejectCode.Int64)
		}
		if code == 0 {
			code = 403
		}
		entry := map[string]interface{}{
			"code": code,
		}
		if resource.MTLSRejectMsg.Valid && strings.TrimSpace(resource.MTLSRejectMsg.String) != "" {
			entry["message"] = resource.MTLSRejectMsg.String
		}
		pluginConfig["rejectMessage"] = entry
	}

	if resource.MTLSRefresh.Valid && strings.TrimSpace(resource.MTLSRefresh.String) != "" {
		pluginConfig["refreshInterval"] = resource.MTLSRefresh.String
	}

	if resource.MTLSExternal.Valid && strings.TrimSpace(resource.MTLSExternal.String) != "" {
		var external map[string]interface{}
		if err := json.Unmarshal([]byte(resource.MTLSExternal.String), &external); err == nil {
			pluginConfig["externalData"] = external
		} else {
			log.Printf("Failed to parse mtls_external_data for resource %s: %v", resource.ID, err)
		}
	}

	middlewareName := fmt.Sprintf("%s-mtlsauth", resource.ID)
	config.HTTP.Middlewares[middlewareName] = map[string]interface{}{
		"plugin": map[string]interface{}{
			"mtlswhitelist": pluginConfig,
		},
	}

	return middlewareName, nil
}

// fetchResourceData loads active resources and their middleware assignments
func (cp *ConfigProxy) fetchResourceData() ([]*resourceData, error) {
	query := `
		SELECT r.id, r.host, r.service_id, r.entrypoints, r.tls_domains,
		       r.custom_headers, r.router_priority, r.source_type, r.mtls_enabled,
		       r.mtls_rules, r.mtls_request_headers, r.mtls_reject_message, r.mtls_reject_code,
		       r.mtls_refresh_interval, r.mtls_external_data,
		       COALESCE(r.tls_hardening_enabled, 0), COALESCE(r.secure_headers_enabled, 0),
		       rm.middleware_id, rm.priority,
		       rs.service_id as custom_service_id
		FROM resources r
		LEFT JOIN resource_middlewares rm ON r.id = rm.resource_id
		LEFT JOIN resource_services rs ON r.id = rs.resource_id
		WHERE r.status = 'active'
		ORDER BY r.id, rm.priority DESC
	`
	rows, err := cp.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resourceMap := make(map[string]*resourceData)

	for rows.Next() {
		var rID, host, serviceID, entrypoints, tlsDomains, customHeaders, sourceType string
		var routerPriority sql.NullInt64
		var mtlsEnabled, tlsHardeningEnabled, secureHeadersEnabled int
		var middlewareID sql.NullString
		var middlewarePriority sql.NullInt64
		var customServiceID sql.NullString
		var mtlsRules, mtlsRequestHeaders, mtlsRejectMessage, mtlsRefreshInterval, mtlsExternalData sql.NullString
		var mtlsRejectCode sql.NullInt64

		err := rows.Scan(
			&rID, &host, &serviceID, &entrypoints, &tlsDomains,
			&customHeaders, &routerPriority, &sourceType, &mtlsEnabled,
			&mtlsRules, &mtlsRequestHeaders, &mtlsRejectMessage, &mtlsRejectCode,
			&mtlsRefreshInterval, &mtlsExternalData,
			&tlsHardeningEnabled, &secureHeadersEnabled,
			&middlewareID, &middlewarePriority, &customServiceID,
		)
		if err != nil {
			log.Printf("Failed to scan resource: %v", err)
			continue
		}

		data, exists := resourceMap[rID]
		if !exists {
			priority := 100
			if routerPriority.Valid {
				priority = int(routerPriority.Int64)
			}
			data = &resourceData{
				ID:                   rID,
				Host:                 host,
				ServiceID:            serviceID,
				Entrypoints:          entrypoints,
				TLSDomains:           tlsDomains,
				CustomHeaders:        customHeaders,
				RouterPriority:       priority,
				SourceType:           sourceType,
				MTLSEnabled:          mtlsEnabled == 1,
				TLSHardeningEnabled:  tlsHardeningEnabled == 1,
				SecureHeadersEnabled: secureHeadersEnabled == 1,
				CustomServiceID:      customServiceID,
				MTLSRules:            mtlsRules,
				MTLSRequestHdrs:      mtlsRequestHeaders,
				MTLSRejectMsg:        mtlsRejectMessage,
				MTLSRejectCode:       mtlsRejectCode,
				MTLSRefresh:          mtlsRefreshInterval,
				MTLSExternal:         mtlsExternalData,
			}
			resourceMap[rID] = data
		}

		if middlewareID.Valid {
			mwPriority := 100
			if middlewarePriority.Valid {
				mwPriority = int(middlewarePriority.Int64)
			}
			data.Middlewares = append(data.Middlewares, middlewareWithPriority{
				ID:       middlewareID.String,
				Priority: mwPriority,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	resources := make([]*resourceData, 0, len(resourceMap))
	for _, r := range resourceMap {
		resources = append(resources, r)
	}
	return resources, nil
}

// loadGlobalMTLSConfig retrieves global mTLS settings (including plugin defaults).
func (cp *ConfigProxy) loadGlobalMTLSConfig() (*mtlsConfigData, error) {
	var enabled int
	var caCertPath string
	var middlewareRules, middlewareRequestHeaders, middlewareRejectMessage sql.NullString
	var middlewareRefreshInterval sql.NullInt64

	err := cp.db.QueryRow(`
		SELECT enabled, ca_cert_path, middleware_rules, middleware_request_headers,
		       middleware_reject_message, middleware_refresh_interval
		FROM mtls_config WHERE id = 1
	`).Scan(&enabled, &caCertPath, &middlewareRules, &middlewareRequestHeaders,
		&middlewareRejectMessage, &middlewareRefreshInterval)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to check mTLS config: %w", err)
	}

	if enabled != 1 || caCertPath == "" {
		return nil, nil
	}

	cfg := &mtlsConfigData{
		CACertPath: caCertPath,
		RejectCode: 403,
	}

	if middlewareRules.Valid && middlewareRules.String != "" {
		var rules []interface{}
		if err := json.Unmarshal([]byte(middlewareRules.String), &rules); err == nil {
			cfg.Rules = rules
		}
	}

	if middlewareRequestHeaders.Valid && middlewareRequestHeaders.String != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(middlewareRequestHeaders.String), &headers); err == nil && len(headers) > 0 {
			cfg.RequestHeaders = headers
		}
	}

	if middlewareRejectMessage.Valid && middlewareRejectMessage.String != "" {
		cfg.RejectMessage = middlewareRejectMessage.String
	}

	if middlewareRefreshInterval.Valid && middlewareRefreshInterval.Int64 > 0 {
		cfg.RefreshInterval = fmt.Sprintf("%ds", middlewareRefreshInterval.Int64)
	}

	return cfg, nil
}

// applyTLSOptions adds TLS options for mTLS verification with hardened security settings
func (cp *ConfigProxy) applyTLSOptions(config *ProxiedTraefikConfig, mtlsCfg *mtlsConfigData) {
	if mtlsCfg == nil || mtlsCfg.CACertPath == "" {
		return
	}

	config.TLS.Options["mtls-verify"] = map[string]interface{}{
		"clientAuth": map[string]interface{}{
			"caFiles":        []string{mtlsCfg.CACertPath},
			"clientAuthType": "VerifyClientCertIfGiven",
		},
		"minVersion": "VersionTLS12",
		"maxVersion": "VersionTLS13",
		"sniStrict":  true,
		"cipherSuites": []string{
			"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
			"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
			"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
			"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
		},
		"curvePreferences": []string{
			"X25519",
			"CurveP384",
			"CurveP521",
		},
	}
}

// findMatchingRouter finds a router that matches the given host
func (cp *ConfigProxy) findMatchingRouter(routers map[string]interface{}, host string) (string, map[string]interface{}) {
	// Host matching regex
	hostRegex := regexp.MustCompile(`Host\(\x60([^` + "`" + `]+)\x60\)`)

	for routerName, routerConfig := range routers {
		router, ok := routerConfig.(map[string]interface{})
		if !ok {
			continue
		}

		rule, ok := router["rule"].(string)
		if !ok {
			continue
		}

		// Extract host from rule
		matches := hostRegex.FindStringSubmatch(rule)
		if len(matches) > 1 && matches[1] == host {
			return routerName, router
		}
	}

	return "", nil
}

// getRouterMiddlewares extracts the middleware list from a router config
func (cp *ConfigProxy) getRouterMiddlewares(router map[string]interface{}) []string {
	middlewares, ok := router["middlewares"]
	if !ok {
		return nil
	}

	switch v := middlewares.(type) {
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, m := range v {
			if s, ok := m.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case []string:
		return v
	default:
		return nil
	}
}

// determineServiceProtocol determines which protocol section a service belongs to
func (cp *ConfigProxy) determineServiceProtocol(serviceType string, config map[string]interface{}) string {
	if serviceType == string(models.LoadBalancerType) {
		if servers, ok := config["servers"].([]interface{}); ok {
			for _, s := range servers {
				if serverMap, ok := s.(map[string]interface{}); ok {
					if _, hasAddress := serverMap["address"]; hasAddress {
						return "tcp"
					}
					if _, hasURL := serverMap["url"]; hasURL {
						return "http"
					}
				}
			}
		}
	}
	return "http"
}

// sanitizeMTLSWhitelist ensures requestHeaders is a map for all mtlswhitelist middlewares
func (cp *ConfigProxy) sanitizeMTLSWhitelist(config *ProxiedTraefikConfig) {
	if config == nil || config.HTTP == nil {
		return
	}
	for key, mw := range config.HTTP.Middlewares {
		mwMap, ok := mw.(map[string]interface{})
		if !ok {
			continue
		}
		pluginVal, ok := mwMap["plugin"].(map[string]interface{})
		if !ok {
			continue
		}
		mtlsVal, ok := pluginVal["mtlswhitelist"].(map[string]interface{})
		if !ok {
			continue
		}
		if rh, exists := mtlsVal["requestHeaders"]; exists {
			switch v := rh.(type) {
			case map[string]interface{}:
				// ok
				if len(v) == 0 {
					delete(mtlsVal, "requestHeaders")
				}
			case map[string]string:
				if len(v) == 0 {
					delete(mtlsVal, "requestHeaders")
				} else {
					mtlsVal["requestHeaders"] = v
				}
			case string:
				// Traefik plugin expects a map; replace string with empty map
				delete(mtlsVal, "requestHeaders")
				if shouldLog() {
					log.Printf("Sanitized mtlswhitelist.requestHeaders for middleware %s (was string)", key)
				}
			default:
				delete(mtlsVal, "requestHeaders")
				if shouldLog() {
					log.Printf("Sanitized mtlswhitelist.requestHeaders for middleware %s (was %T)", key, v)
				}
			}
		}
	}
}

// SetPangolinURL updates the Pangolin API URL
func (cp *ConfigProxy) SetPangolinURL(url string) {
	cp.pangolinURL = url
	cp.InvalidateCache()
}

// SetCacheDuration updates the cache duration
func (cp *ConfigProxy) SetCacheDuration(duration time.Duration) {
	cp.cacheDuration = duration
}

// normalizeRouterOrder converts all HTTP routers to OrderedRouter structs
// to ensure consistent JSON field ordering matching Pangolin's output.
func (cp *ConfigProxy) normalizeRouterOrder(config *ProxiedTraefikConfig) {
	if config == nil || config.HTTP == nil || config.HTTP.Routers == nil {
		return
	}

	for routerKey, routerVal := range config.HTTP.Routers {
		router, ok := routerVal.(map[string]interface{})
		if !ok {
			continue
		}

		ordered := cp.mapToOrderedRouter(router)
		config.HTTP.Routers[routerKey] = ordered
	}
}

// mapToOrderedRouter converts a map[string]interface{} router to OrderedRouter
func (cp *ConfigProxy) mapToOrderedRouter(router map[string]interface{}) *OrderedRouter {
	ordered := &OrderedRouter{}

	// EntryPoints
	if eps, ok := router["entryPoints"]; ok {
		switch v := eps.(type) {
		case []interface{}:
			for _, ep := range v {
				if s, ok := ep.(string); ok {
					ordered.EntryPoints = append(ordered.EntryPoints, s)
				}
			}
		case []string:
			ordered.EntryPoints = v
		}
	}

	// Middlewares
	if mws, ok := router["middlewares"]; ok {
		switch v := mws.(type) {
		case []interface{}:
			for _, mw := range v {
				if s, ok := mw.(string); ok {
					ordered.Middlewares = append(ordered.Middlewares, s)
				}
			}
		case []string:
			ordered.Middlewares = v
		}
	}

	// Service
	if svc, ok := router["service"].(string); ok {
		ordered.Service = svc
	}

	// Rule
	if rule, ok := router["rule"].(string); ok {
		ordered.Rule = rule
	}

	// Priority
	if priority, ok := router["priority"]; ok {
		switch v := priority.(type) {
		case int:
			ordered.Priority = v
		case float64:
			ordered.Priority = int(v)
		case int64:
			ordered.Priority = int(v)
		}
	}

	// TLS
	if tlsVal, ok := router["tls"]; ok {
		switch tls := tlsVal.(type) {
		case map[string]interface{}:
			ordered.TLS = cp.mapToOrderedTLS(tls)
		}
	}

	return ordered
}

// mapToOrderedTLS converts a map[string]interface{} TLS config to OrderedTLSConfig
func (cp *ConfigProxy) mapToOrderedTLS(tls map[string]interface{}) *OrderedTLSConfig {
	ordered := &OrderedTLSConfig{}

	// CertResolver
	if cr, ok := tls["certResolver"].(string); ok {
		ordered.CertResolver = cr
	}

	// Domains
	if domains, ok := tls["domains"]; ok {
		switch v := domains.(type) {
		case []interface{}:
			for _, d := range v {
				if s, ok := d.(string); ok {
					ordered.Domains = append(ordered.Domains, s)
				}
			}
		case []string:
			ordered.Domains = v
		}
	}

	// Options
	if opts, ok := tls["options"].(string); ok {
		ordered.Options = opts
	}

	return ordered
}

// normalizeMiddlewareOrder converts HTTP middlewares to OrderedMiddleware structs
// to ensure consistent JSON field ordering matching Pangolin's output.
// Only converts middlewares with known field structures (redirectScheme, plugin, headers).
// Other middleware types are preserved as-is to avoid losing their configuration.
func (cp *ConfigProxy) normalizeMiddlewareOrder(config *ProxiedTraefikConfig) {
	if config == nil || config.HTTP == nil || config.HTTP.Middlewares == nil {
		return
	}

	for mwKey, mwVal := range config.HTTP.Middlewares {
		mw, ok := mwVal.(map[string]interface{})
		if !ok {
			continue
		}

		// Only convert middlewares that have fields we support in OrderedMiddleware
		// This preserves other middleware types (basicAuth, rateLimit, ipAllowList, etc.)
		if cp.isOrderableMiddleware(mw) {
			ordered := cp.mapToOrderedMiddleware(mw)
			config.HTTP.Middlewares[mwKey] = ordered
		}
		// Otherwise, keep the middleware as-is (map[string]interface{})
	}
}

// isOrderableMiddleware checks if a middleware has fields that can be converted
// to OrderedMiddleware without losing data.
func (cp *ConfigProxy) isOrderableMiddleware(mw map[string]interface{}) bool {
	// Only convert if it ONLY contains fields we support
	for key := range mw {
		switch key {
		case "redirectScheme", "plugin", "headers":
			// These are supported
		default:
			// Contains unsupported field, don't convert
			return false
		}
	}
	return true
}

// mapToOrderedMiddleware converts a map[string]interface{} middleware to OrderedMiddleware
func (cp *ConfigProxy) mapToOrderedMiddleware(mw map[string]interface{}) *OrderedMiddleware {
	ordered := &OrderedMiddleware{}

	// RedirectScheme (comes first in Pangolin output)
	if rs, ok := mw["redirectScheme"].(map[string]interface{}); ok {
		ordered.RedirectScheme = rs
	}

	// Plugin (comes second in Pangolin output)
	if plugin, ok := mw["plugin"].(map[string]interface{}); ok {
		ordered.Plugin = plugin
	}

	// Headers (for custom headers middleware)
	if headers, ok := mw["headers"].(map[string]interface{}); ok {
		ordered.Headers = headers
	}

	return ordered
}

// loadSecurityConfig loads global security configuration from the database
func (cp *ConfigProxy) loadSecurityConfig() (*securityConfigData, error) {
	var tlsHardeningEnabled, secureHeadersEnabled int
	var xContentTypeOptions, xFrameOptions, xXSSProtection, hsts, referrerPolicy, csp, permissionsPolicy string

	err := cp.db.QueryRow(`
		SELECT tls_hardening_enabled, secure_headers_enabled,
		       secure_headers_x_content_type_options, secure_headers_x_frame_options,
		       secure_headers_x_xss_protection, secure_headers_hsts,
		       secure_headers_referrer_policy, secure_headers_csp,
		       secure_headers_permissions_policy
		FROM security_config WHERE id = 1
	`).Scan(
		&tlsHardeningEnabled, &secureHeadersEnabled,
		&xContentTypeOptions, &xFrameOptions,
		&xXSSProtection, &hsts,
		&referrerPolicy, &csp,
		&permissionsPolicy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return defaults
			return &securityConfigData{
				TLSHardeningEnabled:  false,
				SecureHeadersEnabled: false,
				SecureHeaders:        models.DefaultSecureHeaders(),
			}, nil
		}
		return nil, fmt.Errorf("failed to load security config: %w", err)
	}

	return &securityConfigData{
		TLSHardeningEnabled:  tlsHardeningEnabled == 1,
		SecureHeadersEnabled: secureHeadersEnabled == 1,
		SecureHeaders: models.SecureHeadersConfig{
			XContentTypeOptions: xContentTypeOptions,
			XFrameOptions:       xFrameOptions,
			XXSSProtection:      xXSSProtection,
			HSTS:                hsts,
			ReferrerPolicy:      referrerPolicy,
			CSP:                 csp,
			PermissionsPolicy:   permissionsPolicy,
		},
	}, nil
}

// applyTLSHardeningOptions adds TLS options for hardened security (without client auth)
func (cp *ConfigProxy) applyTLSHardeningOptions(config *ProxiedTraefikConfig) {
	config.TLS.Options["tls-hardened"] = models.TLSHardeningOptions()
}

// ensureSecureHeadersMiddleware creates and registers a secure headers middleware for a resource
func (cp *ConfigProxy) ensureSecureHeadersMiddleware(config *ProxiedTraefikConfig, resource *resourceData, securityCfg *securityConfigData) string {
	if securityCfg == nil {
		return ""
	}

	customResponseHeaders := make(map[string]string)

	// Only add headers that have values configured
	if securityCfg.SecureHeaders.XContentTypeOptions != "" {
		customResponseHeaders["X-Content-Type-Options"] = securityCfg.SecureHeaders.XContentTypeOptions
	}
	if securityCfg.SecureHeaders.XFrameOptions != "" {
		customResponseHeaders["X-Frame-Options"] = securityCfg.SecureHeaders.XFrameOptions
	}
	if securityCfg.SecureHeaders.XXSSProtection != "" {
		customResponseHeaders["X-XSS-Protection"] = securityCfg.SecureHeaders.XXSSProtection
	}
	if securityCfg.SecureHeaders.HSTS != "" {
		customResponseHeaders["Strict-Transport-Security"] = securityCfg.SecureHeaders.HSTS
	}
	if securityCfg.SecureHeaders.ReferrerPolicy != "" {
		customResponseHeaders["Referrer-Policy"] = securityCfg.SecureHeaders.ReferrerPolicy
	}
	if securityCfg.SecureHeaders.CSP != "" {
		customResponseHeaders["Content-Security-Policy"] = securityCfg.SecureHeaders.CSP
	}
	if securityCfg.SecureHeaders.PermissionsPolicy != "" {
		customResponseHeaders["Permissions-Policy"] = securityCfg.SecureHeaders.PermissionsPolicy
	}

	// Skip if no headers configured
	if len(customResponseHeaders) == 0 {
		return ""
	}

	middlewareName := fmt.Sprintf("%s-secureheaders", resource.ID)
	config.HTTP.Middlewares[middlewareName] = map[string]interface{}{
		"headers": map[string]interface{}{
			"customResponseHeaders": customResponseHeaders,
		},
	}

	return middlewareName
}
