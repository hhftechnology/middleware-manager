package services

import (
	"testing"

	"github.com/hhftechnology/middleware-manager/models"
)

// TestNewDuplicateDetector tests detector creation
func TestNewDuplicateDetector(t *testing.T) {
	t.Run("with config manager", func(t *testing.T) {
		cm := newTestConfigManager(t)
		detector := NewDuplicateDetector(cm)
		if detector == nil {
			t.Fatal("NewDuplicateDetector() returned nil")
		}
		if detector.configManager != cm {
			t.Error("configManager not set correctly")
		}
	})

	t.Run("with nil config manager", func(t *testing.T) {
		detector := NewDuplicateDetector(nil)
		if detector == nil {
			t.Fatal("NewDuplicateDetector(nil) returned nil")
		}
		if detector.configManager != nil {
			t.Error("configManager should be nil")
		}
	})
}

// TestContainsPluginName tests plugin name detection in middlewares
func TestContainsPluginName(t *testing.T) {
	detector := &DuplicateDetector{}

	tests := []struct {
		name       string
		middleware models.TraefikMiddleware
		pluginName string
		want       bool
	}{
		{
			name: "plugin type middleware",
			middleware: models.TraefikMiddleware{
				Name: "my-middleware",
				Type: "plugin-somePlugin",
			},
			pluginName: "other",
			want:       true, // Any plugin type returns true
		},
		{
			name: "name contains plugin name",
			middleware: models.TraefikMiddleware{
				Name: "mtls-whitelist-middleware",
				Type: "headers",
			},
			pluginName: "mtls-whitelist",
			want:       true,
		},
		{
			name: "no match",
			middleware: models.TraefikMiddleware{
				Name: "rate-limiter",
				Type: "rateLimit",
			},
			pluginName: "mtls",
			want:       false,
		},
		{
			name: "case insensitive plugin name match",
			middleware: models.TraefikMiddleware{
				Name: "MTLS-Whitelist-Prod",
				Type: "headers",
			},
			pluginName: "mtls-whitelist",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.containsPluginName(tt.middleware, tt.pluginName)
			if got != tt.want {
				t.Errorf("containsPluginName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCheckDuplicates_NoConfigManager tests duplicate check when config manager is nil
func TestCheckDuplicates_NoConfigManager(t *testing.T) {
	detector := NewDuplicateDetector(nil)

	result := detector.CheckDuplicates("test-middleware", "test-plugin")

	if result.APIAvailable {
		t.Error("APIAvailable should be false when configManager is nil")
	}
	if result.HasDuplicates {
		t.Error("HasDuplicates should be false when API is unavailable")
	}
	if result.WarningMessage == "" {
		t.Error("WarningMessage should be set")
	}
}

// TestCheckDuplicates_NoTraefikConfig tests duplicate check when traefik is not configured
func TestCheckDuplicates_NoTraefikConfig(t *testing.T) {
	cm := newTestConfigManager(t)
	detector := NewDuplicateDetector(cm)

	// Config manager exists but no Traefik data source configured
	result := detector.CheckDuplicates("test-middleware", "test-plugin")

	if result.APIAvailable {
		t.Error("APIAvailable should be false when Traefik is not configured")
	}
	if result.HasDuplicates {
		t.Error("HasDuplicates should be false when API is unavailable")
	}
}

// TestDuplicateCheckResult tests result structure defaults
func TestDuplicateCheckResult(t *testing.T) {
	result := &models.DuplicateCheckResult{
		HasDuplicates: false,
		Duplicates:    []models.Duplicate{},
		APIAvailable:  true,
	}

	if result.HasDuplicates {
		t.Error("HasDuplicates default should be false")
	}
	if len(result.Duplicates) != 0 {
		t.Error("Duplicates should be empty by default")
	}
	if !result.APIAvailable {
		t.Error("APIAvailable should be true when set")
	}
}

// TestGetTraefikFetcher tests fetcher retrieval from config manager
func TestGetTraefikFetcher(t *testing.T) {
	t.Skip("skipping outdated fetcher expectations")
	t.Run("nil config manager", func(t *testing.T) {
		detector := &DuplicateDetector{configManager: nil}
		fetcher := detector.getTraefikFetcher()
		if fetcher != nil {
			t.Error("Expected nil fetcher for nil configManager")
		}
	})

	t.Run("config manager without traefik source", func(t *testing.T) {
		cm := newTestConfigManager(t)
		detector := NewDuplicateDetector(cm)
		fetcher := detector.getTraefikFetcher()
		// Without setting up a traefik data source, should return nil
		if fetcher != nil {
			t.Error("Expected nil fetcher when no traefik source configured")
		}
	})
}
