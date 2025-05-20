package util

import (
	"strings"
	"regexp"
)

var (
	// Regular expression to match cascading auth suffixes
	authCascadeRegex = regexp.MustCompile(`(-auth)+$`)
	
	// Regular expression for router suffix with auth patterns
	routerAuthRegex = regexp.MustCompile(`-router(-auth)*$`)
)

// NormalizeID provides a standard way to normalize any ID across the application
// It removes provider suffixes and handles special cases like auth cascades
func NormalizeID(id string) string {
	// First, remove any provider suffix (if present)
	baseName := id
	if idx := strings.Index(baseName, "@"); idx > 0 {
		baseName = baseName[:idx]
	}
	
	// Handle cascading auth patterns
	baseName = authCascadeRegex.ReplaceAllString(baseName, "-auth")
	
	// Special handling for router resources
	if strings.Contains(baseName, "-router") {
		// For router-auth, router-auth-auth patterns, normalize to router-auth
		baseName = routerAuthRegex.ReplaceAllString(baseName, "-router-auth")
		
		// Handle redirect suffixes in routers
		if strings.Contains(baseName, "-redirect") {
			// Normalize router-redirect-auth to router-redirect
			if strings.HasSuffix(baseName, "-auth") {
				baseName = strings.TrimSuffix(baseName, "-auth")
			}
		}
	}
	
	return baseName
}

// GetProviderSuffix extracts the provider suffix from an ID
func GetProviderSuffix(id string) string {
	if idx := strings.Index(id, "@"); idx > 0 {
		return id[idx:]
	}
	return ""
}

// AddProviderSuffix adds a provider suffix if one doesn't exist
// If the ID already has a suffix, it returns the original ID
func AddProviderSuffix(id string, suffix string) string {
	if suffix == "" || strings.Contains(id, "@") {
		return id
	}
	
	// Ensure suffix starts with @
	if !strings.HasPrefix(suffix, "@") {
		suffix = "@" + suffix
	}
	
	return id + suffix
}

// DetermineProviderSuffix returns the appropriate provider suffix based on context
func DetermineProviderSuffix(sourceType string, activeDataSourceType string) string {
	// Use file provider for custom services
	if sourceType == "file" {
		return "@file"
	}
	
	// For Traefik API, prefer docker provider for matching source types
	if activeDataSourceType == "traefik" && sourceType == "traefik" {
		return "@docker"
	}
	
	// Default to http provider
	return "@http"
}