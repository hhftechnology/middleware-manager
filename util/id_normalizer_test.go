package util

import (
	"fmt"
	"testing"
)

func TestNormalizeID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"service@http", "service"},
		{"service@docker", "service"},
		{"my-app-auth-auth", "my-app-auth"},
		{"my-app-router-auth-auth", "my-app-router-auth"},
		{"my-app-router-redirect-auth", "my-app-router-redirect"},
		{"simple-id", "simple-id"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ClearNormalizationCache()
			result := NormalizeID(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeIDCacheHit(t *testing.T) {
	ClearNormalizationCache()
	first := NormalizeID("test@http")
	second := NormalizeID("test@http")
	if first != second {
		t.Errorf("cache returned different results: %q vs %q", first, second)
	}
}

func TestCacheBoundedness(t *testing.T) {
	ClearNormalizationCache()

	// Fill cache beyond maxCacheSize
	for i := 0; i < maxCacheSize+100; i++ {
		NormalizeID(fmt.Sprintf("id-%d@http", i))
	}

	// Cache should have been flushed, so size <= maxCacheSize
	cacheMu.RLock()
	size := len(normalizedIDCache)
	cacheMu.RUnlock()

	if size > maxCacheSize {
		t.Errorf("cache size %d exceeds max %d", size, maxCacheSize)
	}
}

func TestClearNormalizationCache(t *testing.T) {
	NormalizeID("something@http")

	ClearNormalizationCache()

	cacheMu.RLock()
	size := len(normalizedIDCache)
	cacheMu.RUnlock()

	if size != 0 {
		t.Errorf("cache not empty after clear: %d entries", size)
	}
}

func TestGetProviderSuffix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"service@http", "@http"},
		{"service@docker", "@docker"},
		{"no-suffix", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GetProviderSuffix(tt.input)
			if result != tt.expected {
				t.Errorf("GetProviderSuffix(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func BenchmarkNormalizeIDUnique(b *testing.B) {
	ClearNormalizationCache()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NormalizeID(fmt.Sprintf("unique-id-%d@http", i))
	}
}

func BenchmarkNormalizeIDRepeated(b *testing.B) {
	ClearNormalizationCache()
	NormalizeID("repeated-id@http")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NormalizeID("repeated-id@http")
	}
}
