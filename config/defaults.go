package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hhftechnology/middleware-manager/database"
	"gopkg.in/yaml.v3"
)

// DefaultMiddleware represents a default middleware template
type DefaultMiddleware struct {
	ID     string                 `yaml:"id"`
	Name   string                 `yaml:"name"`
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

// DefaultTemplates represents the structure of the templates.yaml file
type DefaultTemplates struct {
	Middlewares []DefaultMiddleware `yaml:"middlewares"`
}

// LoadDefaultTemplates loads the default middleware templates
func LoadDefaultTemplates(db *database.DB) error {
	// Determine the path to the templates file
	templatesFile := "config/templates.yaml"
	
	// Check if the file exists in the current directory
	if _, err := os.Stat(templatesFile); os.IsNotExist(err) {
		// Try to find it in different locations
		possiblePaths := []string{
			"/app/config/templates.yaml",  // Docker container path
			"templates.yaml",              // Current directory
		}
		
		found := false
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				templatesFile = path
				found = true
				break
			}
		}
		
		if !found {
			log.Printf("Warning: templates.yaml not found, skipping default templates")
			return nil
		}
	}
	
	// Read the templates file
	data, err := os.ReadFile(templatesFile)
	if err != nil {
		return err
	}
	
	// Parse the YAML
	var templates DefaultTemplates
	if err := yaml.Unmarshal(data, &templates); err != nil {
		return err
	}
	
	// Process templates to ensure proper value preservation based on middleware type
	for i := range templates.Middlewares {
		// Apply middleware-specific processing based on type
		switch templates.Middlewares[i].Type {
		case "headers":
			processHeadersMiddleware(&templates.Middlewares[i].Config)
		case "redirectRegex", "redirectScheme", "replacePath", "replacePathRegex", "stripPrefix", "stripPrefixRegex":
			processPathMiddleware(&templates.Middlewares[i].Config, templates.Middlewares[i].Type)
		case "basicAuth", "digestAuth", "forwardAuth":
			processAuthMiddleware(&templates.Middlewares[i].Config, templates.Middlewares[i].Type)
		case "plugin":
			processPluginMiddleware(&templates.Middlewares[i].Config)
		case "chain":
			processChainingMiddleware(&templates.Middlewares[i].Config)
		default:
			// General processing for other middleware types
			templates.Middlewares[i].Config = preserveTraefikValues(templates.Middlewares[i].Config).(map[string]interface{})
		}
	}

	// Add templates to the database if they don't exist
	for _, middleware := range templates.Middlewares {
		// Check if the middleware already exists
		var exists int
		err := db.QueryRow("SELECT 1 FROM middlewares WHERE id = ?", middleware.ID).Scan(&exists)
		if err == nil {
			// Middleware exists, skip
			continue
		}
		
		// Convert config to JSON string
		configJSON, err := json.Marshal(middleware.Config)
		if err != nil {
			log.Printf("Failed to marshal config for %s: %v", middleware.Name, err)
			continue
		}
		
		// Insert into database
		_, err = db.Exec(
			"INSERT INTO middlewares (id, name, type, config) VALUES (?, ?, ?, ?)",
			middleware.ID, middleware.Name, middleware.Type, string(configJSON),
		)
		
		if err != nil {
			log.Printf("Failed to insert middleware %s: %v", middleware.Name, err)
			continue
		}
		
		log.Printf("Added default middleware: %s", middleware.Name)
	}
	
	return nil
}

// preserveTraefikValues ensures all values in Traefik configurations are properly handled
// This handles special cases in different middleware types and ensures precise value preservation
func preserveTraefikValues(data interface{}) interface{} {
	if data == nil {
		return nil
	}

	switch v := data.(type) {
	case map[string]interface{}:
		// Process each key-value pair in the map
		for key, val := range v {
			// Process values based on key names that might need special handling
			switch {
			// URL or path related fields
			case key == "path" || key == "url" || key == "address" || strings.HasSuffix(key, "Path"):
				// Ensure path strings keep their exact format
				if strVal, ok := val.(string); ok && strVal != "" {
					// Keep exact string formatting
					v[key] = strVal
				} else {
					v[key] = preserveTraefikValues(val)
				}
			
			// Regex and replacement patterns
			case key == "regex" || key == "replacement" || strings.HasSuffix(key, "Regex"):
				// Ensure regex patterns are preserved exactly
				if strVal, ok := val.(string); ok && strVal != "" {
					// Keep exact string formatting with special characters
					v[key] = strVal
				} else {
					v[key] = preserveTraefikValues(val)
				}
			
			// API keys and security tokens
			case key == "key" || key == "token" || key == "secret" || 
				 strings.Contains(key, "Key") || strings.Contains(key, "Token") || 
				 strings.Contains(key, "Secret") || strings.Contains(key, "Password"):
				// Ensure API keys and tokens are preserved exactly
				if strVal, ok := val.(string); ok {
					// Always preserve keys/tokens exactly as-is, even if empty
					v[key] = strVal
				} else {
					v[key] = preserveTraefikValues(val)
				}
			
			// Empty header values (common in security headers middleware)
			case key == "Server" || key == "X-Powered-By" || strings.HasPrefix(key, "X-"):
				// Empty string values are often used to remove headers
				if strVal, ok := val.(string); ok {
					// Preserve empty strings exactly
					v[key] = strVal
				} else {
					v[key] = preserveTraefikValues(val)
				}
			
			// IP addresses and networks
			case key == "ip" || key == "clientIP" || strings.Contains(key, "IP") ||
				 key == "sourceRange" || key == "excludedIPs":
				// IP addresses often need exact formatting
				v[key] = preserveTraefikValues(val)
			
			// Boolean flags that control behavior
			case strings.HasPrefix(key, "is") || strings.HasPrefix(key, "has") || 
				 strings.HasPrefix(key, "enable") || strings.HasSuffix(key, "enabled") ||
				 strings.HasSuffix(key, "Enabled") || key == "permanent" || key == "forceSlash":
				// Ensure boolean values are preserved as actual booleans
				if boolVal, ok := val.(bool); ok {
					v[key] = boolVal
				} else if strVal, ok := val.(string); ok {
					// Convert string "true"/"false" to actual boolean if needed
					if strVal == "true" {
						v[key] = true
					} else if strVal == "false" {
						v[key] = false
					} else {
						v[key] = strVal // Keep as is if not a boolean string
					}
				} else {
					v[key] = preserveTraefikValues(val)
				}
			
			// Integer values like timeouts, ports, limits
			case key == "amount" || key == "burst" || key == "port" || 
				 strings.HasSuffix(key, "Seconds") || strings.HasSuffix(key, "Limit") || 
				 strings.HasSuffix(key, "Timeout") || strings.HasSuffix(key, "Size") ||
				 key == "depth" || key == "priority" || key == "statusCode" || 
				 key == "attempts" || key == "responseCode":
				// Ensure numeric values are preserved as numbers
				switch numVal := val.(type) {
				case int:
					v[key] = numVal
				case float64:
					// Keep as float if it has decimal part, otherwise convert to int
					if numVal == float64(int(numVal)) {
						v[key] = int(numVal)
					} else {
						v[key] = numVal
					}
				case string:
					// Try to convert string to number if it looks like one
					if i, err := strconv.Atoi(numVal); err == nil {
						v[key] = i
					} else if f, err := strconv.ParseFloat(numVal, 64); err == nil {
						v[key] = f
					} else {
						v[key] = numVal // Keep as string if not numeric
					}
				default:
					v[key] = preserveTraefikValues(val)
				}
			
			// Default handling for other keys
			default:
				v[key] = preserveTraefikValues(val)
			}
		}
		return v
	
	case []interface{}:
		// Process each element in the array
		for i, item := range v {
			v[i] = preserveTraefikValues(item)
		}
		return v
	
	case string:
		// Preserve all string values exactly as they are
		return v
	
	case int, float64, bool:
		// Preserve primitive types as they are
		return v
	
	default:
		// For any other type, return as is
		return v
	}
}

// processHeadersMiddleware handles the headers middleware special processing
func processHeadersMiddleware(config *map[string]interface{}) {
	// Special handling for response headers which might contain empty strings
	if customResponseHeaders, ok := (*config)["customResponseHeaders"].(map[string]interface{}); ok {
		for key, value := range customResponseHeaders {
			// Ensure empty strings are preserved exactly
			if strValue, ok := value.(string); ok {
				customResponseHeaders[key] = strValue
			}
		}
	}
	
	// Special handling for request headers which might contain empty strings
	if customRequestHeaders, ok := (*config)["customRequestHeaders"].(map[string]interface{}); ok {
		for key, value := range customRequestHeaders {
			// Ensure empty strings are preserved exactly
			if strValue, ok := value.(string); ok {
				customRequestHeaders[key] = strValue
			}
		}
	}
	
	// Process header fields that are often quoted strings
	specialStringFields := []string{
		"customFrameOptionsValue", "contentSecurityPolicy", 
		"referrerPolicy", "permissionsPolicy",
	}
	
	for _, field := range specialStringFields {
		if value, ok := (*config)[field].(string); ok {
			// Preserve string exactly, especially if it contains quotes
			(*config)[field] = value
		}
	}
	
	// Process other header configuration values
	*config = preserveTraefikValues(*config).(map[string]interface{})
}

// processChainingMiddleware handles chain middleware special processing
func processChainingMiddleware(config *map[string]interface{}) {
	if middlewares, ok := (*config)["middlewares"].([]interface{}); ok {
		for i, middleware := range middlewares {
			if middlewareStr, ok := middleware.(string); ok {
				// If this is not already a fully qualified middleware reference
				if !strings.Contains(middlewareStr, "@") {
					// Assume it's from our file provider
					middlewares[i] = fmt.Sprintf("%s@file", middlewareStr)
				}
			}
		}
		(*config)["middlewares"] = middlewares
	}
	
	// Process other chain configuration values
	*config = preserveTraefikValues(*config).(map[string]interface{})
}

// processPathMiddleware handles path manipulation middlewares
func processPathMiddleware(config *map[string]interface{}, middlewareType string) {
	// Special handling for regex patterns - these need exact preservation
	if regex, ok := (*config)["regex"].(string); ok {
		// Preserve regex pattern exactly
		(*config)["regex"] = regex
	} else if regexList, ok := (*config)["regex"].([]interface{}); ok {
		// Handle regex arrays (like in stripPrefixRegex)
		for i, r := range regexList {
			if regexStr, ok := r.(string); ok {
				regexList[i] = regexStr
			}
		}
	}
	
	// Special handling for replacement patterns
	if replacement, ok := (*config)["replacement"].(string); ok {
		// Preserve replacement pattern exactly
		(*config)["replacement"] = replacement
	}
	
	// Special handling for path values
	if path, ok := (*config)["path"].(string); ok {
		// Preserve path exactly
		(*config)["path"] = path
	}
	
	// Special handling for prefixes arrays
	if prefixes, ok := (*config)["prefixes"].([]interface{}); ok {
		for i, prefix := range prefixes {
			if prefixStr, ok := prefix.(string); ok {
				prefixes[i] = prefixStr
			}
		}
	}
	
	// Special handling for scheme
	if scheme, ok := (*config)["scheme"].(string); ok {
		// Preserve scheme exactly
		(*config)["scheme"] = scheme
	}
	
	// Process boolean options
	if forceSlash, ok := (*config)["forceSlash"].(bool); ok {
		(*config)["forceSlash"] = forceSlash
	}
	
	if permanent, ok := (*config)["permanent"].(bool); ok {
		(*config)["permanent"] = permanent
	}
	
	// Process other path manipulation configuration values
	*config = preserveTraefikValues(*config).(map[string]interface{})
}

// processAuthMiddleware handles authentication middleware special processing
func processAuthMiddleware(config *map[string]interface{}, middlewareType string) {
	// ForwardAuth middleware special handling
	if middlewareType == "forwardAuth" {
		if address, ok := (*config)["address"].(string); ok {
			// Preserve address URL exactly
			(*config)["address"] = address
		}
		
		// Process trust settings
		if trustForward, ok := (*config)["trustForwardHeader"].(bool); ok {
			(*config)["trustForwardHeader"] = trustForward
		}
		
		// Process headers array
		if headers, ok := (*config)["authResponseHeaders"].([]interface{}); ok {
			for i, header := range headers {
				if headerStr, ok := header.(string); ok {
					headers[i] = headerStr
				}
			}
		}
	}
	
	// BasicAuth/DigestAuth middleware special handling
	if middlewareType == "basicAuth" || middlewareType == "digestAuth" {
		// Preserve exact format of users array
		if users, ok := (*config)["users"].([]interface{}); ok {
			for i, user := range users {
				if userStr, ok := user.(string); ok {
					users[i] = userStr
				}
			}
		}
	}
	
	// Process other auth configuration values
	*config = preserveTraefikValues(*config).(map[string]interface{})
}

// processPluginMiddleware handles plugin middleware special processing
func processPluginMiddleware(config *map[string]interface{}) {
	// Process plugins (including CrowdSec)
	for _, pluginCfg := range *config {
		if pluginConfig, ok := pluginCfg.(map[string]interface{}); ok {
			// Process special fields in plugin configurations
			
			// Process API keys and secrets - must be preserved exactly
			keyFields := []string{
				"crowdsecLapiKey", "apiKey", "token", "secret", "password", 
				"key", "accessKey", "secretKey", "captchaSiteKey", "captchaSecretKey",
			}
			
			for _, field := range keyFields {
				if val, exists := pluginConfig[field]; exists {
					if strVal, ok := val.(string); ok {
						// Ensure keys are preserved exactly as-is
						pluginConfig[field] = strVal
					}
				}
			}
			
			// Process boolean options
			boolFields := []string{
				"enabled", "failureBlock", "unreachableBlock", "insecureVerify",
				"allowLocalRequests", "logLocalRequests", "logAllowedRequests",
				"logApiRequests", "silentStartUp", "forceMonthlyUpdate",
				"allowUnknownCountries", "blackListMode", "addCountryHeader",
			}
			
			for _, field := range boolFields {
				for configKey, val := range pluginConfig {
					if strings.Contains(configKey, field) {
						if boolVal, ok := val.(bool); ok {
							pluginConfig[configKey] = boolVal
						}
					}
				}
			}
			
			// Process arrays
			arrayFields := []string{
				"forwardedHeadersTrustedIPs", "clientTrustedIPs", "countries",
			}
			
			for _, field := range arrayFields {
				for configKey, val := range pluginConfig {
					if strings.Contains(configKey, field) {
						if arrayVal, ok := val.([]interface{}); ok {
							for i, item := range arrayVal {
								if strItem, ok := item.(string); ok {
									arrayVal[i] = strItem
								}
							}
						}
					}
				}
			}
			
			// Process remaining fields
			pluginConfig = preserveTraefikValues(pluginConfig).(map[string]interface{})
		}
	}

	// Process the entire config
	*config = preserveTraefikValues(*config).(map[string]interface{})
}

// EnsureConfigDirectory ensures the configuration directory exists
func EnsureConfigDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// SaveTemplateFile saves the default templates file if it doesn't exist
func SaveTemplateFile(templatesDir string) error {
	templatesFile := filepath.Join(templatesDir, "templates.yaml")
	
	// Check if file already exists
	if _, err := os.Stat(templatesFile); err == nil {
		// File exists, skip
		return nil
	}
	
	// Create default templates
	templates := DefaultTemplates{
		Middlewares: []DefaultMiddleware{
			// Authentication middlewares
			{
				ID:   "authelia",
				Name: "Authelia",
				Type: "forwardAuth",
				Config: map[string]interface{}{
					"address":            "http://authelia:9091/api/authz/forward-auth",
					"trustForwardHeader": true,
					"authResponseHeaders": []string{
						"Remote-User",
						"Remote-Groups",
						"Remote-Name",
						"Remote-Email",
					},
				},
			},
			{
				ID:   "authentik",
				Name: "Authentik",
				Type: "forwardAuth",
				Config: map[string]interface{}{
					"address":            "http://authentik:9000/outpost.goauthentik.io/auth/traefik",
					"trustForwardHeader": true,
					"authResponseHeaders": []string{
						"X-authentik-username",
						"X-authentik-groups",
						"X-authentik-email",
						"X-authentik-name",
						"X-authentik-uid",
					},
				},
			},
			{
				ID:   "basic-auth",
				Name: "Basic Auth",
				Type: "basicAuth",
				Config: map[string]interface{}{
					"users": []string{
						"admin:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
					},
				},
			},
			{
				ID:   "digest-auth",
				Name: "Digest Auth",
				Type: "digestAuth",
				Config: map[string]interface{}{
					"users": []string{
						"test:traefik:a2688e031edb4be6a3797f3882655c05",
					},
				},
			},
			{
				ID:   "jwt-auth",
				Name: "JWT Authentication",
				Type: "forwardAuth",
				Config: map[string]interface{}{
					"address":            "http://jwt-auth:8080/verify",
					"trustForwardHeader": true,
					"authResponseHeaders": []string{
						"X-JWT-Sub",
						"X-JWT-Name",
						"X-JWT-Email",
					},
				},
			},
			
			// Security middlewares
			{
				ID:   "ip-whitelist",
				Name: "IP Whitelist",
				Type: "ipWhiteList",
				Config: map[string]interface{}{
					"sourceRange": []string{
						"127.0.0.1/32",
						"192.168.1.0/24",
						"10.0.0.0/8",
					},
				},
			},
			{
				ID:   "ip-allowlist",
				Name: "IP Allow List",
				Type: "ipAllowList",
				Config: map[string]interface{}{
					"sourceRange": []string{
						"127.0.0.1/32",
						"192.168.1.0/24",
						"10.0.0.0/8",
					},
				},
			},
			{
				ID:   "rate-limit",
				Name: "Rate Limit",
				Type: "rateLimit",
				Config: map[string]interface{}{
					"average": int(100),
					"burst":   int(50),
				},
			},
			{
				ID:   "headers-standard",
				Name: "Standard Security Headers",
				Type: "headers",
				Config: map[string]interface{}{
					"accessControlAllowMethods": []string{
						"GET",
						"OPTIONS",
						"PUT",
					},
					"browserXssFilter":        true,
					"contentTypeNosniff":      true,
					"customFrameOptionsValue": "SAMEORIGIN",
					"customResponseHeaders": map[string]string{
						"X-Forwarded-Proto": "https",
						"X-Robots-Tag":      "none,noarchive,nosnippet,notranslate,noimageindex",
						"Server":            "", // Empty string to remove header
					},
					"forceSTSHeader": true,
					"hostsProxyHeaders": []string{
						"X-Forwarded-Host",
					},
					"permissionsPolicy": "camera=(), microphone=(), geolocation=(), payment=(), usb=(), vr=()",
					"referrerPolicy":    "same-origin",
					"sslProxyHeaders": map[string]string{
						"X-Forwarded-Proto": "https",
					},
					"stsIncludeSubdomains": true,
					"stsPreload":           true,
					"stsSeconds":           int(63072000),
				},
			},
			{
				ID:   "in-flight-req",
				Name: "In-Flight Request Limiter",
				Type: "inFlightReq",
				Config: map[string]interface{}{
					"amount": int(10),
					"sourceCriterion": map[string]interface{}{
						"ipStrategy": map[string]interface{}{
							"depth": int(2),
							"excludedIPs": []string{
								"127.0.0.1/32",
							},
						},
					},
				},
			},
			{
				ID:   "pass-tls-cert",
				Name: "Pass TLS Client Certificate",
				Type: "passTLSClientCert",
				Config: map[string]interface{}{
					"pem": true,
				},
			},
			
			// Path manipulation middlewares - with properly formatted regex patterns
			{
				ID:   "add-prefix",
				Name: "Add Prefix",
				Type: "addPrefix",
				Config: map[string]interface{}{
					"prefix": "/api",
				},
			},
			{
				ID:   "strip-prefix",
				Name: "Strip Prefix",
				Type: "stripPrefix",
				Config: map[string]interface{}{
					"prefixes": []string{
						"/api",
						"/v1",
					},
					"forceSlash": true,
				},
			},
			{
				ID:   "strip-prefix-regex",
				Name: "Strip Prefix Regex",
				Type: "stripPrefixRegex",
				Config: map[string]interface{}{
					"regex": []string{
						"/foo/[a-z0-9]+/[0-9]+/",
					},
				},
			},
			{
				ID:   "replace-path",
				Name: "Replace Path",
				Type: "replacePath",
				Config: map[string]interface{}{
					"path": "/api",
				},
			},
			{
				ID:   "replace-path-regex",
				Name: "Replace Path Regex",
				Type: "replacePathRegex",
				Config: map[string]interface{}{
					"regex":       "^/foo/(.*)",
					"replacement": "/bar/$1",
				},
			},
			
			// Redirect middlewares - with properly formatted regex patterns
			{
				ID:   "redirect-regex",
				Name: "Redirect Regex",
				Type: "redirectRegex",
				Config: map[string]interface{}{
					"regex":       "^http://localhost/(.*)",
					"replacement": "https://example.com/${1}",
					"permanent":   true,
				},
			},
			{
				ID:   "redirect-scheme",
				Name: "Redirect to HTTPS",
				Type: "redirectScheme",
				Config: map[string]interface{}{
					"scheme":    "https",
					"port":      "443",
					"permanent": true,
				},
			},
			
			// Content processing middlewares
			{
				ID:   "compress",
				Name: "Compress Response",
				Type: "compress",
				Config: map[string]interface{}{
					"excludedContentTypes": []string{
						"text/event-stream",
					},
					"includedContentTypes": []string{
						"text/html",
						"text/plain",
						"application/json",
					},
					"minResponseBodyBytes": int(1024),
					"encodings": []string{
						"gzip",
						"br",
					},
				},
			},
			{
				ID:   "buffering",
				Name: "Request/Response Buffering",
				Type: "buffering",
				Config: map[string]interface{}{
					"maxRequestBodyBytes":  int(5000000),
					"memRequestBodyBytes":  int(2000000),
					"maxResponseBodyBytes": int(5000000),
					"memResponseBodyBytes": int(2000000),
					"retryExpression":      "IsNetworkError() && Attempts() < 2",
				},
			},
			{
				ID:   "content-type",
				Name: "Content Type Auto-Detector",
				Type: "contentType",
				Config: map[string]interface{}{},
			},
			
			// Error handling and reliability middlewares
			{
				ID:   "circuit-breaker",
				Name: "Circuit Breaker",
				Type: "circuitBreaker",
				Config: map[string]interface{}{
					"expression":        "NetworkErrorRatio() > 0.20 || ResponseCodeRatio(500, 600, 0, 600) > 0.25",
					"checkPeriod":       "10s",
					"fallbackDuration":  "30s",
					"recoveryDuration":  "60s",
					"responseCode":      int(503),
				},
			},
			{
				ID:   "retry",
				Name: "Retry Failed Requests",
				Type: "retry",
				Config: map[string]interface{}{
					"attempts":        int(3),
					"initialInterval": "100ms",
				},
			},
			{
				ID:   "error-pages",
				Name: "Custom Error Pages",
				Type: "errors",
				Config: map[string]interface{}{
					"status": []string{
						"500-599",
					},
					"service": "error-handler-service",
					"query":   "/{status}.html",
				},
			},
			{
				ID:   "grpc-web",
				Name: "gRPC Web",
				Type: "grpcWeb",
				Config: map[string]interface{}{
					"allowOrigins": []string{
						"*",
					},
				},
			},
			
			// Chain middlewares
			{
				ID:   "security-chain",
				Name: "Security Chain",
				Type: "chain",
				Config: map[string]interface{}{
					"middlewares": []string{
						"rate-limit",
						"ip-whitelist",
					},
				},
			},
			
			// Crowdsec plugin middleware with proper API key handling
			{
				ID:   "crowdsec",
				Name: "Crowdsec",
				Type: "plugin",
				Config: map[string]interface{}{
					"crowdsec": map[string]interface{}{
						"enabled":                     true,
						"logLevel":                    "INFO",
						"updateIntervalSeconds":       int(15),
						"updateMaxFailure":            int(0),
						"defaultDecisionSeconds":      int(15),
						"httpTimeoutSeconds":          int(10),
						"crowdsecMode":                "live",
						"crowdsecAppsecEnabled":       true,
						"crowdsecAppsecHost":          "crowdsec:7422",
						"crowdsecAppsecFailureBlock":  true,
						"crowdsecAppsecUnreachableBlock": true,
						"crowdsecAppsecBodyLimit":     int(10485760),  // Use int instead of integer to avoid scientific notation
						"crowdsecLapiKey":             "PUT_YOUR_BOUNCER_KEY_HERE_OR_IT_WILL_NOT_WORK",
						"crowdsecLapiHost":            "crowdsec:8080",
						"crowdsecLapiScheme":          "http",
						"forwardedHeadersTrustedIPs": []string{
							"0.0.0.0/0",
						},
						"clientTrustedIPs": []string{
							"10.0.0.0/8",
							"172.16.0.0/12",
							"192.168.0.0/16",
						},
					},
				},
			},
			
			// Special use case middlewares - with properly formatted regex pattern
			{
				ID:   "nextcloud-dav",
				Name: "Nextcloud WebDAV Redirect",
				Type: "replacePathRegex",
				Config: map[string]interface{}{
					"regex":       "^/.well-known/ca(l|rd)dav",
					"replacement": "/remote.php/dav/",
				},
			},
			
			// Custom headers example with empty string preservation
			{
				ID:   "custom-headers-example",
				Name: "Custom Headers Example",
				Type: "headers",
				Config: map[string]interface{}{
					"customRequestHeaders": map[string]string{
						"X-Script-Name":        "test",
						"X-Custom-Value":       "value with spaces",
						"X-Custom-Request-Header": "",  // Empty string to remove header
					},
					"customResponseHeaders": map[string]string{
						"X-Custom-Response-Header": "value",
						"Server":                  "",  // Empty string to remove header
					},
				},
			},
		},
	}
	
	// Process all templates to ensure proper value preservation
	for i := range templates.Middlewares {
		// Apply middleware-specific processing based on type
		switch templates.Middlewares[i].Type {
		case "headers":
			processHeadersMiddleware(&templates.Middlewares[i].Config)
		case "redirectRegex", "redirectScheme", "replacePath", "replacePathRegex", "stripPrefix", "stripPrefixRegex":
			processPathMiddleware(&templates.Middlewares[i].Config, templates.Middlewares[i].Type)
		case "basicAuth", "digestAuth", "forwardAuth":
			processAuthMiddleware(&templates.Middlewares[i].Config, templates.Middlewares[i].Type)
		case "plugin":
			processPluginMiddleware(&templates.Middlewares[i].Config)
		case "chain":
			processChainingMiddleware(&templates.Middlewares[i].Config)
		default:
			// General processing for other middleware types
			templates.Middlewares[i].Config = preserveTraefikValues(templates.Middlewares[i].Config).(map[string]interface{})
		}
	}
	
	// Create a custom YAML encoder that preserves string formatting
	yamlNode := &yaml.Node{}
	err := yamlNode.Encode(templates)
	if err != nil {
		return fmt.Errorf("failed to encode templates to YAML node: %w", err)
	}
	
	// Apply additional string preservation to the YAML node
	preserveStringsInYamlNode(yamlNode)
	
	// Marshal the processed node
	data, err := yaml.Marshal(yamlNode)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML node: %w", err)
	}
	
	// Write to file
	return os.WriteFile(templatesFile, data, 0644)
}

// preserveStringsInYamlNode ensures that string values, especially empty strings,
// are preserved correctly in the YAML node structure before marshaling
func preserveStringsInYamlNode(node *yaml.Node) {
	if node == nil {
		return
	}
	
	// Process node based on its kind
	switch node.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		// Process all content/items
		for i := range node.Content {
			preserveStringsInYamlNode(node.Content[i])
		}
	
	case yaml.MappingNode:
		// Process all key-value pairs
		for i := 0; i < len(node.Content); i += 2 {
			// Get key and value
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			
			// Process based on key content
			if keyNode.Value == "Server" || keyNode.Value == "X-Powered-By" || 
			   strings.HasPrefix(keyNode.Value, "X-") {
				// These are likely header fields where empty strings are important
				if valueNode.Kind == yaml.ScalarNode && valueNode.Value == "" {
					// Ensure empty strings are properly encoded
					valueNode.Style = yaml.DoubleQuotedStyle
				}
			}
			
			// Special handling for known fields that need exact string preservation
			if containsSpecialField(keyNode.Value) && valueNode.Kind == yaml.ScalarNode {
				// Use double quotes for these fields to ensure proper encoding
				valueNode.Style = yaml.DoubleQuotedStyle
			}
			
			// Continue recursion
			preserveStringsInYamlNode(keyNode)
			preserveStringsInYamlNode(valueNode)
		}
	
	case yaml.ScalarNode:
		// For scalar nodes (including strings), ensure empty strings are properly quoted
		if node.Value == "" {
			node.Style = yaml.DoubleQuotedStyle
		}
	}
}

// containsSpecialField checks if a field name is one that needs special handling
// for correct string value preservation
func containsSpecialField(fieldName string) bool {
	specialFields := []string{
		"key", "token", "secret", "apiKey", "Key", "Token", "Secret", "Password",
		"regex", "replacement", "Regex", "path", "scheme", "url", "address", "Path",
		"prefix", "prefixes", "expression", "rule",
	}
	
	for _, field := range specialFields {
		if strings.Contains(fieldName, field) {
			return true
		}
	}
	
	return false
}