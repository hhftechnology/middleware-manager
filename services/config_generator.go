package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"net/http"

	"github.com/hhftechnology/middleware-manager/database"
	"github.com/hhftechnology/middleware-manager/models"
	"gopkg.in/yaml.v3"
)

// ConfigGenerator generates separate Traefik configuration files
type ConfigGenerator struct {
	db            *database.DB
	confDir       string
	configManager *ConfigManager
	stopChan      chan struct{}
	isRunning     bool
	mutex         sync.Mutex
	lastConfigs   map[string][]byte // Track individual file configs
}

// TraefikConfig represents the structure of the Traefik configuration
type TraefikConfig struct {
	HTTP struct {
		Middlewares map[string]interface{} `yaml:"middlewares,omitempty"`
		Routers     map[string]interface{} `yaml:"routers,omitempty"`
		Services    map[string]interface{} `yaml:"services,omitempty"`
	} `yaml:"http"`

	TCP struct {
		Routers  map[string]interface{} `yaml:"routers,omitempty"`
		Services map[string]interface{} `yaml:"services,omitempty"`
	} `yaml:"tcp,omitempty"`

	UDP struct {
		Services map[string]interface{} `yaml:"services,omitempty"`
	} `yaml:"udp,omitempty"`
}

// Separate config structures for individual files
type MiddlewareConfig struct {
	HTTP struct {
		Middlewares map[string]interface{} `yaml:"middlewares"`
	} `yaml:"http"`
}

type RouterConfig struct {
	HTTP struct {
		Routers map[string]interface{} `yaml:"routers"`
	} `yaml:"http,omitempty"`
	TCP struct {
		Routers map[string]interface{} `yaml:"routers,omitempty"`
	} `yaml:"tcp,omitempty"`
}

type ServiceConfig struct {
	HTTP struct {
		Services map[string]interface{} `yaml:"services"`
	} `yaml:"http,omitempty"`
	TCP struct {
		Services map[string]interface{} `yaml:"services,omitempty"`
	} `yaml:"tcp,omitempty"`
	UDP struct {
		Services map[string]interface{} `yaml:"services,omitempty"`
	} `yaml:"udp,omitempty"`
}

func shouldLog() bool {
    logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
    return logLevel == "debug" || logLevel == ""
}

func shouldLogInfo() bool {
    logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
    return logLevel == "debug" || logLevel == "info" || logLevel == ""
}

// NewConfigGenerator creates a new config generator
func NewConfigGenerator(db *database.DB, confDir string, configManager *ConfigManager) *ConfigGenerator {
	return &ConfigGenerator{
		db:            db,
		confDir:       confDir,
		configManager: configManager,
		stopChan:      make(chan struct{}),
		isRunning:     false,
		lastConfigs:   make(map[string][]byte),
	}
}

// Start begins generating configuration files
func (cg *ConfigGenerator) Start(interval time.Duration) {
	cg.mutex.Lock()
	if cg.isRunning {
		cg.mutex.Unlock()
		return
	}
	cg.isRunning = true
	cg.mutex.Unlock()

	log.Printf("Config generator started, checking every %v", interval)

	if err := os.MkdirAll(cg.confDir, 0755); err != nil {
		log.Printf("Failed to create conf directory: %v", err)
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if err := cg.generateConfig(); err != nil {
		log.Printf("Initial config generation failed: %v", err)
	}

    for {
        select {
        case <-ticker.C:
            if err := cg.generateConfigWithRetry(); err != nil {
                log.Printf("Config generation failed: %v", err)
            }
        case <-cg.stopChan:
            log.Println("Config generator stopped")
            return
        }
    }
}

func normalizeServiceID(id string) string {
    baseName := id
    if idx := strings.Index(id, "@"); idx > 0 {
        baseName = id[:idx]
    }
    return baseName
}

// Stop stops the config generator
func (cg *ConfigGenerator) Stop() {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()

	if !cg.isRunning {
		return
	}
	close(cg.stopChan)
	cg.isRunning = false
}

func (cg *ConfigGenerator) generateConfigWithRetry() error {
    maxRetries := 3
    baseDelay := 1 * time.Second
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        err := cg.generateConfig()
        if err == nil {
            return nil
        }
        
        if strings.Contains(strings.ToLower(err.Error()), "database is locked") {
            if attempt < maxRetries-1 {
                delay := baseDelay * time.Duration(1<<attempt)
                log.Printf("⚠️  Database locked on attempt %d, retrying in %v", attempt+1, delay)
                time.Sleep(delay)
                continue
            }
        }
        
        return err
    }
    
    return fmt.Errorf("config generation failed after %d attempts", maxRetries)
}

// generateConfig generates separate Traefik configuration files
func (cg *ConfigGenerator) generateConfig() error {
    if shouldLog() {
        log.Println("Generating Traefik configuration files...")
    }

    config := TraefikConfig{}
    config.HTTP.Middlewares = make(map[string]interface{})
    config.HTTP.Routers = make(map[string]interface{})
    config.HTTP.Services = make(map[string]interface{})
    config.TCP.Routers = make(map[string]interface{})
    config.TCP.Services = make(map[string]interface{})
    config.UDP.Services = make(map[string]interface{})

    if err := cg.processMiddlewares(&config); err != nil {
        return fmt.Errorf("failed to process middlewares: %w", err)
    }
    if err := cg.processServices(&config); err != nil {
        return fmt.Errorf("failed to process services: %w", err)
    }
    if err := cg.processResourcesWithServices(&config); err != nil {
        return fmt.Errorf("failed to process HTTP resources with services: %w", err)
    }
    if err := cg.processTCPRouters(&config); err != nil {
        return fmt.Errorf("failed to process TCP resources: %w", err)
    }

    // Write separate files
    if err := cg.writeSeparateConfigFiles(&config); err != nil {
        return fmt.Errorf("failed to write separate config files: %w", err)
    }

    return nil
}

// writeSeparateConfigFiles writes individual configuration files
func (cg *ConfigGenerator) writeSeparateConfigFiles(config *TraefikConfig) error {
    var errCount int
    
    // 1. Write individual middleware files
    if err := cg.writeMiddlewareFiles(config); err != nil {
        log.Printf("Failed to write middleware files: %v", err)
        errCount++
    }

    // 2. Write router files
    if err := cg.writeRouterFiles(config); err != nil {
        log.Printf("Failed to write router files: %v", err)
        errCount++
    }

    // 3. Write service files
    if err := cg.writeServiceFiles(config); err != nil {
        log.Printf("Failed to write service files: %v", err)
        errCount++
    }

    if errCount > 0 {
        return fmt.Errorf("failed to write %d file types", errCount)
    }

    return nil
}

// writeMiddlewareFiles writes individual middleware files
func (cg *ConfigGenerator) writeMiddlewareFiles(config *TraefikConfig) error {
    middlewareDir := filepath.Join(cg.confDir, "middlewares")
    if err := os.MkdirAll(middlewareDir, 0755); err != nil {
        return fmt.Errorf("failed to create middlewares directory: %w", err)
    }

    // Remove old middleware files first
    if err := cg.cleanupDirectory(middlewareDir, ".yml"); err != nil {
        log.Printf("Warning: failed to cleanup middleware directory: %v", err)
    }

    for middlewareID, middlewareConfig := range config.HTTP.Middlewares {
        middlewareFile := MiddlewareConfig{}
        middlewareFile.HTTP.Middlewares = make(map[string]interface{})
        middlewareFile.HTTP.Middlewares[middlewareID] = middlewareConfig

        processedConfig := preserveTraefikValues(middlewareFile)

        yamlNode := &yaml.Node{}
        if err := yamlNode.Encode(processedConfig); err != nil {
            log.Printf("Failed to encode middleware %s to YAML: %v", middlewareID, err)
            continue
        }
        preserveStringsInYamlNode(yamlNode)
        
        yamlData, err := yaml.Marshal(yamlNode)
        if err != nil {
            log.Printf("Failed to marshal middleware %s: %v", middlewareID, err)
            continue
        }

        filename := fmt.Sprintf("%s.yml", middlewareID)
        filepath := filepath.Join(middlewareDir, filename)
        
        if cg.hasFileChanged(filename, yamlData) {
            if err := cg.writeFileAtomic(filepath, yamlData); err != nil {
                log.Printf("Failed to write middleware file %s: %v", filename, err)
                continue
            }
            if shouldLogInfo() {
                log.Printf("Generated middleware file: %s", filename)
            }
        }
    }

    return nil
}

// writeRouterFiles writes router configuration files
func (cg *ConfigGenerator) writeRouterFiles(config *TraefikConfig) error {
    // HTTP Routers
    if len(config.HTTP.Routers) > 0 {
        httpRouterFile := RouterConfig{}
        httpRouterFile.HTTP.Routers = config.HTTP.Routers
        
        if err := cg.writeConfigFile("http-routers.yml", httpRouterFile); err != nil {
            return fmt.Errorf("failed to write HTTP routers: %w", err)
        }
    }

    // TCP Routers
    if len(config.TCP.Routers) > 0 {
        tcpRouterFile := RouterConfig{}
        tcpRouterFile.TCP.Routers = config.TCP.Routers
        
        if err := cg.writeConfigFile("tcp-routers.yml", tcpRouterFile); err != nil {
            return fmt.Errorf("failed to write TCP routers: %w", err)
        }
    }

    return nil
}

// writeServiceFiles writes service configuration files
func (cg *ConfigGenerator) writeServiceFiles(config *TraefikConfig) error {
    // HTTP Services
    if len(config.HTTP.Services) > 0 {
        httpServiceFile := ServiceConfig{}
        httpServiceFile.HTTP.Services = config.HTTP.Services
        
        if err := cg.writeConfigFile("http-services.yml", httpServiceFile); err != nil {
            return fmt.Errorf("failed to write HTTP services: %w", err)
        }
    }

    // TCP Services
    if len(config.TCP.Services) > 0 {
        tcpServiceFile := ServiceConfig{}
        tcpServiceFile.TCP.Services = config.TCP.Services
        
        if err := cg.writeConfigFile("tcp-services.yml", tcpServiceFile); err != nil {
            return fmt.Errorf("failed to write TCP services: %w", err)
        }
    }

    // UDP Services
    if len(config.UDP.Services) > 0 {
        udpServiceFile := ServiceConfig{}
        udpServiceFile.UDP.Services = config.UDP.Services
        
        if err := cg.writeConfigFile("udp-services.yml", udpServiceFile); err != nil {
            return fmt.Errorf("failed to write UDP services: %w", err)
        }
    }

    return nil
}

// writeConfigFile writes a configuration file with change detection
func (cg *ConfigGenerator) writeConfigFile(filename string, configData interface{}) error {
    processedConfig := preserveTraefikValues(configData)

    yamlNode := &yaml.Node{}
    if err := yamlNode.Encode(processedConfig); err != nil {
        return fmt.Errorf("failed to encode %s to YAML: %w", filename, err)
    }
    preserveStringsInYamlNode(yamlNode)
    
    yamlData, err := yaml.Marshal(yamlNode)
    if err != nil {
        return fmt.Errorf("failed to marshal %s: %w", filename, err)
    }

    filepath := filepath.Join(cg.confDir, filename)
    
    if cg.hasFileChanged(filename, yamlData) {
        if err := cg.writeFileAtomic(filepath, yamlData); err != nil {
            return fmt.Errorf("failed to write %s: %w", filename, err)
        }
        if shouldLogInfo() {
            log.Printf("Generated configuration file: %s", filename)
        }
    }

    return nil
}

// hasFileChanged checks if a specific file's content has changed
func (cg *ConfigGenerator) hasFileChanged(filename string, newConfig []byte) bool {
    lastConfig, exists := cg.lastConfigs[filename]
    if !exists || len(lastConfig) != len(newConfig) || string(lastConfig) != string(newConfig) {
        cg.lastConfigs[filename] = make([]byte, len(newConfig))
        copy(cg.lastConfigs[filename], newConfig)
        return true
    }
    return false
}

// writeFileAtomic writes a file atomically using a temporary file
func (cg *ConfigGenerator) writeFileAtomic(filepath string, data []byte) error {
    tempFile := filepath + ".tmp"
    if err := os.WriteFile(tempFile, data, 0644); err != nil {
        return fmt.Errorf("failed to write temp file: %w", err)
    }
    return os.Rename(tempFile, filepath)
}

// cleanupDirectory removes old files with the specified extension from a directory
func (cg *ConfigGenerator) cleanupDirectory(dir, ext string) error {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return err
    }

    for _, entry := range entries {
        if !entry.IsDir() && strings.HasSuffix(entry.Name(), ext) {
            filepath := filepath.Join(dir, entry.Name())
            if err := os.Remove(filepath); err != nil {
                log.Printf("Warning: failed to remove old file %s: %v", entry.Name(), err)
            }
        }
    }
    
    return nil
}

// MiddlewareWithPriority represents a middleware with its priority value
type MiddlewareWithPriority struct {
    ID       string
    Priority int
}

func (cg *ConfigGenerator) processMiddlewares(config *TraefikConfig) error {
    rows, err := cg.db.Query("SELECT id, name, type, config FROM middlewares")
    if err != nil {
        return fmt.Errorf("failed to fetch middlewares: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var id, name, typ, configStr string
        if err := rows.Scan(&id, &name, &typ, &configStr); err != nil {
            if shouldLog() {
                log.Printf("Failed to scan middleware: %v", err)
            }
            continue
        }
        var middlewareConfig map[string]interface{}
        if err := json.Unmarshal([]byte(configStr), &middlewareConfig); err != nil {
            if shouldLog() {
                log.Printf("Failed to parse middleware config for %s: %v", name, err)
            }
            continue
        }
        
        middlewareConfig = models.ProcessMiddlewareConfig(typ, middlewareConfig)

        config.HTTP.Middlewares[id] = map[string]interface{}{
            typ: middlewareConfig,
        }
    }
    return rows.Err()
}

func (cg *ConfigGenerator) processServices(config *TraefikConfig) error {
    rows, err := cg.db.Query("SELECT id, name, type, config FROM services")
    if err != nil {
        return fmt.Errorf("failed to fetch services: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var id, name, typ, configStr string
        if err := rows.Scan(&id, &name, &typ, &configStr); err != nil {
            if shouldLog() {
                log.Printf("Failed to scan service: %v", err)
            }
            continue
        }
        
        var serviceConfig map[string]interface{}
        if err := json.Unmarshal([]byte(configStr), &serviceConfig); err != nil {
            if shouldLog() {
                log.Printf("Failed to parse service config for %s: %v", name, err)
            }
            continue
        }

        serviceConfig = models.ProcessServiceConfig(typ, serviceConfig)
        
        protocol := determineServiceProtocol(typ, serviceConfig)
        switch protocol {
        case "tcp":
            config.TCP.Services[id] = serviceConfig
        case "udp":
            config.UDP.Services[id] = serviceConfig
        default: // http
            config.HTTP.Services[id] = serviceConfig
        }
    }
    return rows.Err()
}

func stringSliceContains(slice []string, str string) bool {
    for _, s := range slice {
        if s == str {
            return true
        }
    }
    return false
}

func determineServiceProtocol(serviceType string, config map[string]interface{}) string {
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

// processResourcesWithServices processes resources with their assigned services
func (cg *ConfigGenerator) processResourcesWithServices(config *TraefikConfig) error {
    activeDSConfig, err := cg.configManager.GetActiveDataSourceConfig()
    if err != nil {
        if shouldLog() {
            log.Printf("Warning: Could not get active data source config in ConfigGenerator: %v. Defaulting to Pangolin logic.", err)
        }
        activeDSConfig = models.DataSourceConfig{Type: "pangolin"}
    }

    type ResourceData struct {
        Info            models.Resource
        Middlewares     []MiddlewareWithPriority
        CustomServiceID sql.NullString
    }

    resourceDataMap := make(map[string]ResourceData)

    // FIXED QUERY: Use correct table alias 'rs' consistently
    query := `
        SELECT DISTINCT 
            r.id, r.host, r.service_id, r.entrypoints, r.tls_domains, r.custom_headers, r.source_type, r.router_priority,
            rm.middleware_id, rm.priority as middleware_priority,
            rs.service_id as custom_service_id
        FROM resources r
        LEFT JOIN resource_middlewares rm ON r.id = rm.resource_id
        LEFT JOIN resource_services rs ON r.id = rs.resource_id
        ORDER BY r.id, rm.priority DESC`

    rows, err := cg.db.Query(query)
    if err != nil {
        return fmt.Errorf("failed to query resources with middlewares: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var (
            rID_db                  string
            host_db                 string
            serviceID_db            string
            entrypoints_db          string
            tlsDomains_db           string
            customHeadersStr_db     string
            sourceType_db           string
            routerPriority_db       sql.NullInt64
            middlewareID_db         sql.NullString
            middlewarePriority_db   sql.NullInt64
            customServiceID_db      sql.NullString
        )

        err := rows.Scan(
            &rID_db, &host_db, &serviceID_db, &entrypoints_db, &tlsDomains_db, 
            &customHeadersStr_db, &sourceType_db, &routerPriority_db, &middlewareID_db, 
            &middlewarePriority_db, &customServiceID_db,
        )
        if err != nil {
            log.Printf("Failed to scan resource data for HTTP router: %v", err)
            continue
        }
        
        data, exists := resourceDataMap[rID_db]
        if !exists {
            data.Info = models.Resource{
                ID:            rID_db,
                Host:          host_db,
                ServiceID:     serviceID_db,
                Entrypoints:   entrypoints_db,
                TLSDomains:    tlsDomains_db,
                CustomHeaders: customHeadersStr_db,
                SourceType:    sourceType_db,
            }
            if routerPriority_db.Valid {
                data.Info.RouterPriority = int(routerPriority_db.Int64)
            } else {
                data.Info.RouterPriority = 200
            }
            data.CustomServiceID = customServiceID_db
        }

        if middlewareID_db.Valid {
            mwPriority := 200 
            if middlewarePriority_db.Valid {
                mwPriority = int(middlewarePriority_db.Int64)
            }
            data.Middlewares = append(data.Middlewares, MiddlewareWithPriority{
                ID:       middlewareID_db.String,
                Priority: mwPriority,
            })
        }
        resourceDataMap[rID_db] = data
    }
    if err = rows.Err(); err != nil {
        return fmt.Errorf("error iterating resource rows for HTTP: %w", err)
    }
    
    for _, mapValueDataEntry := range resourceDataMap {
        info := mapValueDataEntry.Info
        assignedMiddlewares := mapValueDataEntry.Middlewares
        
        sort.SliceStable(assignedMiddlewares, func(i, j int) bool {
            return assignedMiddlewares[i].Priority > assignedMiddlewares[j].Priority
        })

        routerEntryPoints := strings.Split(strings.TrimSpace(info.Entrypoints), ",")
        if len(routerEntryPoints) == 0 || (len(routerEntryPoints) == 1 && routerEntryPoints[0] == "") {
            routerEntryPoints = []string{"websecure"}
        }

        var customHeadersMiddlewareID string
        if info.CustomHeaders != "" && info.CustomHeaders != "{}" && info.CustomHeaders != "null" {
            var headersMap map[string]string 
            if err := json.Unmarshal([]byte(info.CustomHeaders), &headersMap); err == nil && len(headersMap) > 0 {
                middlewareName := fmt.Sprintf("%s-customheaders", info.ID) 
                customRequestHeadersMap := make(map[string]string)
                for k,v := range headersMap {
                    customRequestHeadersMap[k] = v
                }
                config.HTTP.Middlewares[middlewareName] = map[string]interface{}{
                    "headers": map[string]interface{}{"customRequestHeaders": customRequestHeadersMap},
                }
                customHeadersMiddlewareID = fmt.Sprintf("%s@file", middlewareName)
            } else if err != nil {
                log.Printf("Failed to parse custom headers for resource %s: %v.", info.ID, err)
            }
        }

        var finalMiddlewares []string
        if customHeadersMiddlewareID != "" {
            finalMiddlewares = append(finalMiddlewares, customHeadersMiddlewareID)
        }
        for _, mw := range assignedMiddlewares {
            finalMiddlewares = append(finalMiddlewares, fmt.Sprintf("%s@file", mw.ID))
        }

        var serviceReference string
        if mapValueDataEntry.CustomServiceID.Valid && mapValueDataEntry.CustomServiceID.String != "" {
            serviceReference = fmt.Sprintf("%s@file", mapValueDataEntry.CustomServiceID.String)
        } else {
            switch activeDSConfig.Type {
            case "traefik":
                serviceReference = cg.fetchTraefikServiceReference(info.ServiceID)
            default:
                serviceReference = fmt.Sprintf("%s@file", info.ServiceID)
            }
        }

        if shouldLog() {
            log.Printf("Processing resource %s -> service %s (SourceType: %s, ActiveDS: %s, CustomSvc: %s)",
                info.ID,
                serviceReference,
                info.SourceType,
                activeDSConfig.Type,
                mapValueDataEntry.CustomServiceID.String)
        }

        routerIDBase := extractBaseName(info.ID)
        routerIDForTraefik := fmt.Sprintf("%s-auth", routerIDBase) 
        
        routerConfig := map[string]interface{}{
            "rule":        fmt.Sprintf("Host(`%s`)", info.Host),
            "service":     serviceReference,
            "entryPoints": routerEntryPoints,
            "priority":    info.RouterPriority, 
        }
        if len(finalMiddlewares) > 0 {
            routerConfig["middlewares"] = finalMiddlewares
        }

        tlsConfig := map[string]interface{}{"certResolver": "letsencrypt"}
        if info.TLSDomains != "" {
            sans := strings.Split(strings.TrimSpace(info.TLSDomains), ",")
            var cleanSans []string
            for _, s := range sans {
                if trimmed := strings.TrimSpace(s); trimmed != "" {
                    cleanSans = append(cleanSans, trimmed)
                }
            }
            if len(cleanSans) > 0 {
                tlsConfig["domains"] = []map[string]interface{}{{"main": info.Host, "sans": cleanSans}}
            }
        }
        routerConfig["tls"] = tlsConfig
        config.HTTP.Routers[routerIDForTraefik] = routerConfig
    }
    return nil
}

// processTCPRouters processes TCP routers
func (cg *ConfigGenerator) processTCPRouters(config *TraefikConfig) error {
    activeDSConfig, err := cg.configManager.GetActiveDataSourceConfig()
    if err != nil {
        if shouldLog() {
            log.Printf("Warning: Could not get active data source config: %v", err)
        }
        activeDSConfig = models.DataSourceConfig{Type: "pangolin"}
    }

    // FIXED QUERY: Use correct table alias 'rs' consistently  
    query := `
        SELECT DISTINCT 
            r.id, r.host, r.service_id, r.entrypoints, r.source_type, r.router_priority,
            rs.service_id as custom_service_id
        FROM resources r
        LEFT JOIN resource_services rs ON r.id = rs.resource_id
        WHERE r.tcp_enabled = 1
        ORDER BY r.id`

    rows, err := cg.db.Query(query)
    if err != nil {
        return fmt.Errorf("failed to query TCP resources: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var (
            id                  string
            host                string
            serviceID           string
            entrypoints         string
            sourceType          string
            routerPriority      sql.NullInt64
            customServiceID     sql.NullString
        )

        err := rows.Scan(&id, &host, &serviceID, &entrypoints, &sourceType, &routerPriority, &customServiceID)
        if err != nil {
            log.Printf("Failed to scan TCP resource: %v", err)
            continue
        }

        priority := 200
        if routerPriority.Valid {
            priority = int(routerPriority.Int64)
        }

        var tcpServiceReference string
        if customServiceID.Valid && customServiceID.String != "" {
            tcpServiceReference = fmt.Sprintf("%s@file", customServiceID.String)
        } else {
            switch activeDSConfig.Type {
            case "traefik":
                tcpServiceReference = cg.fetchTraefikServiceReference(serviceID)
            default:
                tcpServiceReference = fmt.Sprintf("%s@file", serviceID)
            }
        }

        rule := fmt.Sprintf("HostSNI(`%s`)", host)
        entrypointsList := strings.Split(strings.TrimSpace(entrypoints), ",")
        if len(entrypointsList) == 0 || (len(entrypointsList) == 1 && entrypointsList[0] == "") {
            entrypointsList = []string{"tcpsecure"}
        }

        if shouldLog() {
            log.Printf("Processing TCP resource %s -> service %s (SourceType: %s, ActiveDS: %s, CustomSvc: %s)", 
                id, tcpServiceReference, sourceType, activeDSConfig.Type, customServiceID.String)
        }
        
        routerIDBase := extractBaseName(id)
        tcpRouterID := fmt.Sprintf("%s-tcp", routerIDBase)
        
        config.TCP.Routers[tcpRouterID] = map[string]interface{}{
            "rule":        rule,
            "service":     tcpServiceReference,
            "entryPoints": entrypointsList,
            "priority":    priority,
            "tls":         map[string]interface{}{},
        }
    }
    return rows.Err()
}

// fetchTraefikServiceReference fetches service reference from Traefik API
func (cg *ConfigGenerator) fetchTraefikServiceReference(serviceID string) string {
    serviceMap := cg.fetchTraefikServiceNames()
    if serviceName, exists := serviceMap[serviceID]; exists {
        return serviceName
    }
    return fmt.Sprintf("%s@file", serviceID)
}

// fetchTraefikServiceNames fetches service names from Traefik API
func (cg *ConfigGenerator) fetchTraefikServiceNames() map[string]string {
    serviceMap := make(map[string]string)
    client := &http.Client{Timeout: 5 * time.Second}
    
    dsConfig, err := cg.configManager.GetActiveDataSourceConfig()
    if err != nil {
        log.Printf("Warning: Failed to get active data source config: %v", err)
        return serviceMap
    }
    
    apiURL := dsConfig.URL
    
    resp, err := client.Get(apiURL + "/api/http/services")
    if err != nil {
        log.Printf("Warning: Failed to fetch Traefik services: %v", err)
        return serviceMap
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Printf("Warning: Traefik API returned status %d", resp.StatusCode)
        return serviceMap
    }

    var services []interface{}
    if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
        log.Printf("Warning: Failed to decode Traefik services: %v", err)
        return serviceMap
    }

    for _, s := range services {
        if serviceObj, ok := s.(map[string]interface{}); ok {
            if name, ok := serviceObj["name"].(string); ok {
                normalizedID := normalizeServiceID(name)
                serviceMap[normalizedID] = name
            }
        }
    }

    return serviceMap
}

// extractBaseName extracts the base name without provider suffixes
func extractBaseName(id string) string {
    if idx := strings.Index(id, "@"); idx > 0 {
        return id[:idx]
    }
    return id
}

// preserveTraefikValues preserves Traefik configuration values
func preserveTraefikValues(data interface{}) interface{} {
    switch v := data.(type) {
    case map[string]interface{}:
        result := make(map[string]interface{})
        for key, value := range v {
            result[key] = preserveTraefikValues(value)
        }
        return result
    case []interface{}:
        result := make([]interface{}, len(v))
        for i, item := range v {
            result[i] = preserveTraefikValues(item)
        }
        return result
    default:
        return v
    }
}

// preserveStringsInYamlNode preserves string formatting in YAML nodes
func preserveStringsInYamlNode(node *yaml.Node) {
    if node == nil { 
        return 
    }
    
    switch node.Kind {
    case yaml.DocumentNode, yaml.SequenceNode:
        for i := range node.Content {
            preserveStringsInYamlNode(node.Content[i])
        }
    case yaml.MappingNode:
        for i := 0; i < len(node.Content); i += 2 {
            keyNode := node.Content[i]
            valueNode := node.Content[i+1]
            if (keyNode.Value == "Server" || keyNode.Value == "X-Powered-By" || strings.HasPrefix(keyNode.Value, "X-")) &&
                valueNode.Kind == yaml.ScalarNode && valueNode.Value == "" {
                valueNode.Style = yaml.DoubleQuotedStyle
            }
            if containsSpecialStringField(keyNode.Value) && valueNode.Kind == yaml.ScalarNode {
                valueNode.Style = yaml.DoubleQuotedStyle
            }
            preserveStringsInYamlNode(keyNode)
            preserveStringsInYamlNode(valueNode)
        }
    case yaml.ScalarNode:
        if node.Value == "" {
            node.Style = yaml.DoubleQuotedStyle
        } else if isNumericString(node.Value) && len(node.Value) > 5 {
            node.Tag = "!!str"
        }
    }
}

// isNumericString checks if a string is numeric
func isNumericString(s string) bool {
    _, err := strconv.ParseFloat(s, 64)
    return err == nil
}

// containsSpecialStringField checks for special string fields
func containsSpecialStringField(fieldName string) bool {
    specialFields := []string{
        "key", "token", "secret", "apiKey", "Key", "Token", "Secret", "Password", "Pass", "User", "Users",
        "regex", "replacement", "Regex", "Path", "path", "scheme", "url", "address",
        "prefix", "prefixes", "expression", "rule", "certResolver", "address", "authResponseHeaders",
        "customRequestHeaders", "customResponseHeaders", "customFrameOptionsValue", "contentSecurityPolicy",
        "referrerPolicy", "permissionsPolicy", "stsSeconds", "excludedIPs", "sourceRange",
        "query", "service", "fallback", "flushInterval", "interval", "timeout",
    }
    for _, field := range specialFields {
        if strings.EqualFold(fieldName, field) || strings.Contains(strings.ToLower(fieldName), strings.ToLower(field)) {
            return true
        }
    }
    return false
}