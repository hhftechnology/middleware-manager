package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIError represents a standardized error response
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ResponseWithError sends a standardized error response
func ResponseWithError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, APIError{
		Code:    statusCode,
		Message: message,
	})
}

// generateID generates a random 16-character hex string
func generateID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// isValidMiddlewareType checks if a middleware type is valid
func isValidMiddlewareType(typ string) bool {
	validTypes := map[string]bool{
		"basicAuth":       true,
		"forwardAuth":     true,
		"ipWhiteList":     true,
		"rateLimit":       true,
		"headers":         true,
		"stripPrefix":     true,
		"addPrefix":       true,
		"redirectRegex":   true,
		"redirectScheme":  true,
		"chain":           true,
		"replacepathregex": true,
		"plugin":          true,
	}
	
	return validTypes[typ]
}

// sanitizeMiddlewareConfig ensures proper formatting of duration values and strings
func sanitizeMiddlewareConfig(config map[string]interface{}) {
	// List of keys that should be treated as duration values
	durationKeys := map[string]bool{
		"checkPeriod":      true,
		"fallbackDuration": true,
		"recoveryDuration": true,
		"initialInterval":  true,
		"retryTimeout":     true,
		"gracePeriod":      true,
	}

	// Process the configuration recursively
	sanitizeConfigRecursive(config, durationKeys)
}

// sanitizeConfigRecursive processes config values recursively
func sanitizeConfigRecursive(data interface{}, durationKeys map[string]bool) {
	// Process based on data type
	switch v := data.(type) {
	case map[string]interface{}:
		// Process each key-value pair in the map
		for key, value := range v {
			// Handle different value types
			switch innerVal := value.(type) {
			case string:
				// Check if this is a duration field and ensure proper format
				if durationKeys[key] {
					// Check if the string has extra quotes
					if len(innerVal) > 2 && strings.HasPrefix(innerVal, "\"") && strings.HasSuffix(innerVal, "\"") {
						// Remove the extra quotes
						v[key] = strings.Trim(innerVal, "\"")
					}
				}
			case map[string]interface{}, []interface{}:
				// Recursively process nested structures
				sanitizeConfigRecursive(innerVal, durationKeys)
			}
		}
	case []interface{}:
		// Process each item in the array
		for i, item := range v {
			switch innerVal := item.(type) {
			case map[string]interface{}, []interface{}:
				// Recursively process nested structures
				sanitizeConfigRecursive(innerVal, durationKeys)
			case string:
				// Check if string has unnecessary quotes
				if len(innerVal) > 2 && strings.HasPrefix(innerVal, "\"") && strings.HasSuffix(innerVal, "\"") {
					v[i] = strings.Trim(innerVal, "\"")
				}
			}
		}
	}
}

// LogError logs an error with context information
func LogError(context string, err error) {
	if err != nil {
		log.Printf("Error %s: %v", context, err)
	}
}