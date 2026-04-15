package util

import (
	"regexp"
	"strings"
	"sync"
)

const maxCacheSize = 1000

var (
	// Regular expression to match cascading auth suffixes
	authCascadeRegex = regexp.MustCompile(`(-auth)+$`)

	// Regular expression for router suffix with auth patterns
	routerAuthRegex = regexp.MustCompile(`-router(-auth)*$`)

	// Bounded memoization cache for normalized IDs
	cacheMu           sync.RWMutex
	normalizedIDCache = make(map[string]string, maxCacheSize)
)

// NormalizeID provides a standard way to normalize any ID across the application
// It removes provider suffixes and handles special cases like auth cascades
// Uses memoization for improved performance on repeated calls
func NormalizeID(id string) string {
	// Check cache first (read lock)
	cacheMu.RLock()
	if cached, ok := normalizedIDCache[id]; ok {
		cacheMu.RUnlock()
		return cached
	}
	cacheMu.RUnlock()

	// Perform normalization
	normalized := normalizeIDInternal(id)

	// Store in cache (write lock), flush if full
	cacheMu.Lock()
	if len(normalizedIDCache) >= maxCacheSize {
		normalizedIDCache = make(map[string]string, maxCacheSize)
	}
	normalizedIDCache[id] = normalized
	cacheMu.Unlock()

	return normalized
}

// normalizeIDInternal performs the actual normalization logic
func normalizeIDInternal(id string) string {
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
			baseName = strings.TrimSuffix(baseName, "-auth")
		}
	}

	return baseName
}

// ClearNormalizationCache clears the ID normalization cache
// Useful for testing or when IDs change
func ClearNormalizationCache() {
	cacheMu.Lock()
	normalizedIDCache = make(map[string]string, maxCacheSize)
	cacheMu.Unlock()
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
