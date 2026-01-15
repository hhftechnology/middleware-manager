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

	url := strings.TrimSuffix(pangolinURL, "/") + "/api/v1/traefik-config"

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

// mergeMiddlewareManagerConfig merges all MW-manager additions into the config
func (cp *ConfigProxy) mergeMiddlewareManagerConfig(config *ProxiedTraefikConfig) error {
	// Apply custom middlewares from database
	if err := cp.applyMiddlewares(config); err != nil {
		return fmt.Errorf("failed to apply middlewares: %w", err)
	}

	// Apply custom services from database
	if err := cp.applyServices(config); err != nil {
		return fmt.Errorf("failed to apply services: %w", err)
	}

	// Apply resource overrides (middleware assignments, headers, priority)
	if err := cp.applyResourceOverrides(config); err != nil {
		return fmt.Errorf("failed to apply resource overrides: %w", err)
	}

	// Apply mTLS configuration
	if err := cp.applyMTLSConfig(config); err != nil {
		return fmt.Errorf("failed to apply mTLS config: %w", err)
	}

	return nil
}

// applyMiddlewares adds custom middlewares from the database
func (cp *ConfigProxy) applyMiddlewares(config *ProxiedTraefikConfig) error {
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
func (cp *ConfigProxy) applyResourceOverrides(config *ProxiedTraefikConfig) error {
	query := `
		SELECT r.id, r.host, r.service_id, r.entrypoints, r.tls_domains,
		       r.custom_headers, r.router_priority, r.source_type, r.mtls_enabled,
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
		return fmt.Errorf("failed to fetch resources: %w", err)
	}
	defer rows.Close()

	type middlewareWithPriority struct {
		ID       string
		Priority int
	}

	type resourceData struct {
		ID              string
		Host            string
		ServiceID       string
		Entrypoints     string
		TLSDomains      string
		CustomHeaders   string
		RouterPriority  int
		SourceType      string
		MTLSEnabled     bool
		Middlewares     []middlewareWithPriority
		CustomServiceID sql.NullString
	}

	resourceMap := make(map[string]*resourceData)

	for rows.Next() {
		var rID, host, serviceID, entrypoints, tlsDomains, customHeaders, sourceType string
		var routerPriority sql.NullInt64
		var mtlsEnabled int
		var middlewareID sql.NullString
		var middlewarePriority sql.NullInt64
		var customServiceID sql.NullString

		err := rows.Scan(
			&rID, &host, &serviceID, &entrypoints, &tlsDomains,
			&customHeaders, &routerPriority, &sourceType, &mtlsEnabled,
			&middlewareID, &middlewarePriority, &customServiceID,
		)
		if err != nil {
			log.Printf("Failed to scan resource: %v", err)
			continue
		}

		// Get or create resource data
		data, exists := resourceMap[rID]
		if !exists {
			priority := 200
			if routerPriority.Valid {
				priority = int(routerPriority.Int64)
			}
			data = &resourceData{
				ID:              rID,
				Host:            host,
				ServiceID:       serviceID,
				Entrypoints:     entrypoints,
				TLSDomains:      tlsDomains,
				CustomHeaders:   customHeaders,
				RouterPriority:  priority,
				SourceType:      sourceType,
				MTLSEnabled:     mtlsEnabled == 1,
				CustomServiceID: customServiceID,
			}
			resourceMap[rID] = data
		}

		// Add middleware if present
		if middlewareID.Valid {
			mwPriority := 200
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
		return fmt.Errorf("error iterating resources: %w", err)
	}

	// Now apply overrides to matching routers
	for _, resource := range resourceMap {
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

		// Build middleware list
		var newMiddlewares []string

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

		// Add mTLS middleware if enabled for this resource
		if resource.MTLSEnabled {
			// Prepend mTLS middleware to run first
			newMiddlewares = append([]string{"mtls-auth"}, newMiddlewares...)
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
		if resource.RouterPriority != 200 {
			router["priority"] = resource.RouterPriority
		}

		// Add mTLS TLS options if enabled
		if resource.MTLSEnabled {
			if tlsConfig, ok := router["tls"].(map[string]interface{}); ok {
				tlsConfig["options"] = "mtls-verify"
			} else {
				router["tls"] = map[string]interface{}{
					"options": "mtls-verify",
				}
			}
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

// applyMTLSConfig adds TLS options and mTLS middleware if globally enabled
func (cp *ConfigProxy) applyMTLSConfig(config *ProxiedTraefikConfig) error {
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
			// No mTLS config, skip
			return nil
		}
		return fmt.Errorf("failed to check mTLS config: %w", err)
	}

	// If mTLS is not enabled, don't add config
	if enabled != 1 {
		return nil
	}

	// If no CA cert path configured, skip
	if caCertPath == "" {
		log.Printf("Warning: mTLS enabled but no CA certificate path configured")
		return nil
	}

	// Add TLS options with VerifyClientCertIfGiven
	config.TLS.Options["mtls-verify"] = map[string]interface{}{
		"clientAuth": map[string]interface{}{
			"caFiles":        []string{caCertPath},
			"clientAuthType": "VerifyClientCertIfGiven",
		},
		"minVersion": "VersionTLS12",
		"sniStrict":  true,
	}

	// Add the mtls-auth middleware using mtlswhitelist plugin
	pluginConfig := map[string]interface{}{
		"caFiles": []string{caCertPath},
	}

	// Add optional plugin configuration if set
	if middlewareRules.Valid && middlewareRules.String != "" {
		var rules []interface{}
		if err := json.Unmarshal([]byte(middlewareRules.String), &rules); err == nil && len(rules) > 0 {
			pluginConfig["rules"] = rules
		}
	}

	if middlewareRequestHeaders.Valid && middlewareRequestHeaders.String != "" {
		var headers map[string]interface{}
		if err := json.Unmarshal([]byte(middlewareRequestHeaders.String), &headers); err == nil && len(headers) > 0 {
			pluginConfig["requestHeaders"] = headers
		}
	}

	if middlewareRejectMessage.Valid && middlewareRejectMessage.String != "" {
		pluginConfig["rejectMessage"] = middlewareRejectMessage.String
	}

	if middlewareRefreshInterval.Valid && middlewareRefreshInterval.Int64 > 0 {
		pluginConfig["refreshInterval"] = middlewareRefreshInterval.Int64
	}

	config.HTTP.Middlewares["mtls-auth"] = map[string]interface{}{
		"plugin": map[string]interface{}{
			"mtlswhitelist": pluginConfig,
		},
	}

	if shouldLog() {
		log.Printf("Added mTLS TLS options and mtls-auth middleware with CA cert: %s", caCertPath)
	}

	return nil
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

// SetPangolinURL updates the Pangolin API URL
func (cp *ConfigProxy) SetPangolinURL(url string) {
	cp.pangolinURL = url
	cp.InvalidateCache()
}

// SetCacheDuration updates the cache duration
func (cp *ConfigProxy) SetCacheDuration(duration time.Duration) {
	cp.cacheDuration = duration
}
