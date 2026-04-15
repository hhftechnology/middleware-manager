package services

import (
	"os"
	"strings"
)

// TraefikFallbackURLsEnv is the env var used to opt-in to fallback URL probing
// when the primary Traefik API URL fails. Comma-separated list.
const TraefikFallbackURLsEnv = "TRAEFIK_API_FALLBACK_URLS"

// defaultTraefikURL returns the seed URL for the traefik data source when no
// settings file exists. Prefers TRAEFIK_API_URL so users running the container
// with a custom Traefik host (e.g. gerbil:8080) don't get silently overridden.
func defaultTraefikURL() string {
	if u := strings.TrimSpace(os.Getenv("TRAEFIK_API_URL")); u != "" {
		return u
	}
	return "http://traefik:8080"
}

// fallbackTraefikURLsFromEnv returns user-configured fallback URLs for the
// Traefik API, or nil when the env var is unset. The primary URL is excluded
// so it is not retried during the fallback pass.
func fallbackTraefikURLsFromEnv(primaryURL string) []string {
	raw := strings.TrimSpace(os.Getenv(TraefikFallbackURLsEnv))
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	urls := make([]string, 0, len(parts))
	for _, p := range parts {
		u := strings.TrimSpace(p)
		if u == "" || u == primaryURL {
			continue
		}
		urls = append(urls, u)
	}
	return urls
}
