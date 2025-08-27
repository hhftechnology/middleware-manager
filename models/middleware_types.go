package models

import (
	"strings"
)

// MiddlewareProcessor interface for type-specific processing
type MiddlewareProcessor interface {
	Process(config map[string]interface{}) map[string]interface{}
}
type BufferingProcessor struct{}

// Process implements special handling for buffering middleware
func (p *BufferingProcessor) Process(config map[string]interface{}) map[string]interface{} {
	// Convert byte size fields from float64 to int
	byteFields := []string{
		"maxRequestBodyBytes", 
		"memRequestBodyBytes", 
		"maxResponseBodyBytes", 
		"memResponseBodyBytes",
	}
	
	for _, field := range byteFields {
		if val, ok := config[field].(float64); ok {
			// Convert to int if it's a whole number
			if val == float64(int64(val)) {
				config[field] = int64(val)
			} else {
				config[field] = val
			}
		}
	}
	
	// Process other buffering configuration values with general processor
	return preserveTraefikValues(config).(map[string]interface{})
}

// Registry of middleware processors
var middlewareProcessors = map[string]MiddlewareProcessor{
	"headers":         &HeadersProcessor{},
	"basicAuth":       &AuthProcessor{},
	"forwardAuth":     &AuthProcessor{},
	"digestAuth":      &AuthProcessor{},
	"redirectRegex":   &PathProcessor{},
	"redirectScheme":  &PathProcessor{},
	"replacePath":     &PathProcessor{},
	"replacePathRegex": &PathProcessor{},
	"stripPrefix":     &PathProcessor{},
	"stripPrefixRegex": &PathProcessor{},
	"chain":           &ChainProcessor{},
	"plugin":          &PluginProcessor{},
	"rateLimit":       &RateLimitProcessor{},
	"inFlightReq":     &RateLimitProcessor{},
	"ipWhiteList":     &IPFilterProcessor{},
	"ipAllowList":     &IPFilterProcessor{},
	"buffering":       &BufferingProcessor{}, // ADD THIS LINE
	// Add more middleware types as needed
}

// GetProcessor returns the appropriate processor for a middleware type
func GetProcessor(middlewareType string) MiddlewareProcessor {
	if processor, exists := middlewareProcessors[middlewareType]; exists {
		return processor
	}
	return &DefaultProcessor{} // Fallback processor
}

// ProcessMiddlewareConfig processes a middleware configuration based on its type
func ProcessMiddlewareConfig(middlewareType string, config map[string]interface{}) map[string]interface{} {
	processor := GetProcessor(middlewareType)
	return processor.Process(config)
}

// DefaultProcessor is the fallback processor for middleware types without a specific processor
type DefaultProcessor struct{}

// Process handles general middleware configuration processing
func (p *DefaultProcessor) Process(config map[string]interface{}) map[string]interface{} {
	return preserveTraefikValues(config).(map[string]interface{})
}

// HeadersProcessor handles headers middleware specific processing
type HeadersProcessor struct{}

// Process implements special handling for headers middleware
func (p *HeadersProcessor) Process(config map[string]interface{}) map[string]interface{} {
	// Special handling for response headers which might contain empty strings
	if customResponseHeaders, ok := config["customResponseHeaders"].(map[string]interface{}); ok {
		for key, value := range customResponseHeaders {
			// Ensure empty strings are preserved exactly
			if strValue, ok := value.(string); ok {
				customResponseHeaders[key] = strValue
			}
		}
	}
	
	// Special handling for request headers which might contain empty strings
	if customRequestHeaders, ok := config["customRequestHeaders"].(map[string]interface{}); ok {
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
		if value, ok := config[field].(string); ok {
			// Preserve string exactly, especially if it contains quotes
			config[field] = value
		}
	}
	
	// Process other header configuration values with general processor
	return preserveTraefikValues(config).(map[string]interface{})
}

// AuthProcessor handles authentication middlewares specific processing
type AuthProcessor struct{}

// Process implements special handling for authentication middlewares
func (p *AuthProcessor) Process(config map[string]interface{}) map[string]interface{} {
	// ForwardAuth middleware special handling
	if address, ok := config["address"].(string); ok {
		// Preserve address URL exactly
		config["address"] = address
	}
	
	// Process trust settings
	if trustForward, ok := config["trustForwardHeader"].(bool); ok {
		config["trustForwardHeader"] = trustForward
	}
	
	// Process headers array
	if headers, ok := config["authResponseHeaders"].([]interface{}); ok {
		for i, header := range headers {
			if headerStr, ok := header.(string); ok {
				headers[i] = headerStr
			}
		}
	}
	
	// BasicAuth/DigestAuth middleware special handling
	// Preserve exact format of users array
	if users, ok := config["users"].([]interface{}); ok {
		for i, user := range users {
			if userStr, ok := user.(string); ok {
				users[i] = userStr
			}
		}
	}
	
	// Process other auth configuration values with general processor
	return preserveTraefikValues(config).(map[string]interface{})
}

// PathProcessor handles path manipulation middlewares specific processing
type PathProcessor struct{}

// Process implements special handling for path manipulation middlewares
func (p *PathProcessor) Process(config map[string]interface{}) map[string]interface{} {
	// Special handling for regex patterns - these need exact preservation
	if regex, ok := config["regex"].(string); ok {
		// Preserve regex pattern exactly
		config["regex"] = regex
	} else if regexList, ok := config["regex"].([]interface{}); ok {
		// Handle regex arrays (like in stripPrefixRegex)
		for i, r := range regexList {
			if regexStr, ok := r.(string); ok {
				regexList[i] = regexStr
			}
		}
	}
	
	// Special handling for replacement patterns
	if replacement, ok := config["replacement"].(string); ok {
		// Preserve replacement pattern exactly
		config["replacement"] = replacement
	}
	
	// Special handling for path values
	if path, ok := config["path"].(string); ok {
		// Preserve path exactly
		config["path"] = path
	}
	
	// Special handling for prefixes arrays
	if prefixes, ok := config["prefixes"].([]interface{}); ok {
		for i, prefix := range prefixes {
			if prefixStr, ok := prefix.(string); ok {
				prefixes[i] = prefixStr
			}
		}
	}
	
	// Special handling for scheme
	if scheme, ok := config["scheme"].(string); ok {
		// Preserve scheme exactly
		config["scheme"] = scheme
	}
	
	// Process boolean options
	if forceSlash, ok := config["forceSlash"].(bool); ok {
		config["forceSlash"] = forceSlash
	}
	
	if permanent, ok := config["permanent"].(bool); ok {
		config["permanent"] = permanent
	}
	
	// Process other path manipulation configuration values with general processor
	return preserveTraefikValues(config).(map[string]interface{})
}

// ChainProcessor handles chain middleware specific processing
type ChainProcessor struct{}

// Process implements special handling for chain middleware
func (p *ChainProcessor) Process(config map[string]interface{}) map[string]interface{} {
	// Process middlewares array
	if middlewares, ok := config["middlewares"].([]interface{}); ok {
		for i, middleware := range middlewares {
			if middlewareStr, ok := middleware.(string); ok {
				// If this is not already a fully qualified middleware reference
				if !strings.Contains(middlewareStr, "@") {
					// Assume it's from our file provider
					middlewares[i] = middlewareStr
				}
			}
		}
	}
	
	// Process other chain configuration values with general processor
	return preserveTraefikValues(config).(map[string]interface{})
}

// PluginProcessor handles plugin middleware specific processing
type PluginProcessor struct{}

// Process implements special handling for plugin middleware
func (p *PluginProcessor) Process(config map[string]interface{}) map[string]interface{} {
	// Process plugins (including CrowdSec)
	for _, pluginCfg := range config {
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
			
			// Process remaining fields with general processor
			preserveTraefikValues(pluginConfig)
		}
	}
	
	// Process the entire config with general processor
	return preserveTraefikValues(config).(map[string]interface{})
}

// RateLimitProcessor handles rate limiting middlewares specific processing
type RateLimitProcessor struct{}

// Process implements special handling for rate limiting middlewares
func (p *RateLimitProcessor) Process(config map[string]interface{}) map[string]interface{} {
	// Process numeric values
	if average, ok := config["average"].(float64); ok {
		// Convert to int if it's a whole number
		if average == float64(int(average)) {
			config["average"] = int(average)
		} else {
			config["average"] = average
		}
	}
	
	if burst, ok := config["burst"].(float64); ok {
		// Convert to int if it's a whole number
		if burst == float64(int(burst)) {
			config["burst"] = int(burst)
		} else {
			config["burst"] = burst
		}
	}
	
	if amount, ok := config["amount"].(float64); ok {
		// Convert to int if it's a whole number
		if amount == float64(int(amount)) {
			config["amount"] = int(amount)
		} else {
			config["amount"] = amount
		}
	}
	
	// Process sourceCriterion for inFlightReq
	if sourceCriterion, ok := config["sourceCriterion"].(map[string]interface{}); ok {
		// Process IP strategy
		if ipStrategy, ok := sourceCriterion["ipStrategy"].(map[string]interface{}); ok {
			// Process depth
			if depth, ok := ipStrategy["depth"].(float64); ok {
				ipStrategy["depth"] = int(depth)
			}
			
			// Process excluded IPs
			if excludedIPs, ok := ipStrategy["excludedIPs"].([]interface{}); ok {
				for i, ip := range excludedIPs {
					if ipStr, ok := ip.(string); ok {
						excludedIPs[i] = ipStr
					}
				}
			}
		}
		
		// Process requestHost boolean
		if requestHost, ok := sourceCriterion["requestHost"].(bool); ok {
			sourceCriterion["requestHost"] = requestHost
		}
	}
	
	// Process other rate limiting configuration values with general processor
	return preserveTraefikValues(config).(map[string]interface{})
}

// IPFilterProcessor handles IP filtering middlewares specific processing
type IPFilterProcessor struct{}

// Process implements special handling for IP filtering middlewares
func (p *IPFilterProcessor) Process(config map[string]interface{}) map[string]interface{} {
	// Process sourceRange IPs
	if sourceRange, ok := config["sourceRange"].([]interface{}); ok {
		for i, range_ := range sourceRange {
			if rangeStr, ok := range_.(string); ok {
				// Preserve IP CIDR notation exactly
				sourceRange[i] = rangeStr
			}
		}
	}
	
	// Process other IP filtering configuration values with general processor
	return preserveTraefikValues(config).(map[string]interface{})
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
				// Handle float64 to int conversion for whole numbers, common in JSON unmarshaling
				if f, ok := val.(float64); ok && f == float64(int(f)) {
					v[key] = int(f)
				} else {
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
	
	case string, int, float64, bool:
		// Preserve primitive types as they are
		return v
	
	default:
		// For any other type, return as is
		return v
	}
}