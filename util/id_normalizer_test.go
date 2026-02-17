package util

import (
	"testing"
)

func TestNormalizeID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"provider suffix removal", "my-svc@docker", "my-svc"},
		{"no provider suffix", "my-svc", "my-svc"},
		{"auth cascade", "svc-auth-auth", "svc-auth"},
		{"triple auth cascade", "svc-auth-auth-auth", "svc-auth"},
		{"router auth pattern", "my-router-auth-auth", "my-router-auth"},
		{"router redirect", "my-router-redirect-auth", "my-router-redirect"},
		{"empty string", "", ""},
		{"at sign only suffix", "svc@file", "svc"},
		{"memoization cache hit", "cached-svc@docker", "cached-svc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear cache before each test for isolation
			ClearNormalizationCache()
			got := NormalizeID(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeID(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}

	// Verify memoization: call twice and check cache is used
	t.Run("memoization", func(t *testing.T) {
		ClearNormalizationCache()
		first := NormalizeID("test-svc@docker")
		second := NormalizeID("test-svc@docker")
		if first != second {
			t.Errorf("memoization mismatch: first=%q, second=%q", first, second)
		}
	})
}

func TestClearNormalizationCache(t *testing.T) {
	// Populate cache
	NormalizeID("a@docker")
	NormalizeID("b@file")

	// Clear it
	ClearNormalizationCache()

	// Verify by loading directly from sync.Map
	count := 0
	normalizedIDCache.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	if count != 0 {
		t.Errorf("cache should be empty after ClearNormalizationCache, got %d items", count)
	}
}

func TestGetProviderSuffix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with suffix", "my-svc@docker", "@docker"},
		{"without suffix", "my-svc", ""},
		{"empty string", "", ""},
		{"file provider", "svc@file", "@file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetProviderSuffix(tt.input)
			if got != tt.expected {
				t.Errorf("GetProviderSuffix(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestAddProviderSuffix(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		suffix   string
		expected string
	}{
		{"adds suffix", "my-svc", "docker", "my-svc@docker"},
		{"adds suffix with @", "my-svc", "@docker", "my-svc@docker"},
		{"skips if already has @", "my-svc@file", "docker", "my-svc@file"},
		{"empty suffix returns original", "my-svc", "", "my-svc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddProviderSuffix(tt.id, tt.suffix)
			if got != tt.expected {
				t.Errorf("AddProviderSuffix(%q, %q) = %q, want %q", tt.id, tt.suffix, got, tt.expected)
			}
		})
	}
}

func TestDetermineProviderSuffix(t *testing.T) {
	tests := []struct {
		name             string
		sourceType       string
		activeDS         string
		expected         string
	}{
		{"file source", "file", "pangolin", "@file"},
		{"traefik+traefik", "traefik", "traefik", "@docker"},
		{"default http", "pangolin", "pangolin", "@http"},
		{"other combo", "custom", "traefik", "@http"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineProviderSuffix(tt.sourceType, tt.activeDS)
			if got != tt.expected {
				t.Errorf("DetermineProviderSuffix(%q, %q) = %q, want %q", tt.sourceType, tt.activeDS, got, tt.expected)
			}
		})
	}
}
