package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tmtypes "github.com/hhftechnology/middleware-manager/internal/traefikmanager/types"
)

func TestSettingsStorePreservesUnknownFields(t *testing.T) {
	cfg := tmtypes.RuntimeConfig{SettingsPath: filepath.Join(t.TempDir(), "manager.yml"), TraefikAPIURL: "http://traefik:8080"}
	store := NewSettingsStore(cfg)
	if err := os.WriteFile(cfg.SettingsPath, []byte("domains:\n  - example.com\nlegacy_auth:\n  enabled: true\n"), 0o644); err != nil {
		t.Fatalf("write settings fixture: %v", err)
	}

	settings, _, err := store.Load()
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	settings.CertResolver = "letsencrypt"
	if err := store.Save(settings); err != nil {
		t.Fatalf("save settings: %v", err)
	}

	data, err := os.ReadFile(cfg.SettingsPath)
	if err != nil {
		t.Fatalf("read saved settings: %v", err)
	}
	contents := string(data)
	if !containsAll(contents, "legacy_auth", "enabled: true", "cert_resolver: letsencrypt") {
		t.Fatalf("expected saved settings to preserve unknown fields, got: %s", contents)
	}
}

func containsAll(haystack string, needles ...string) bool {
	for _, needle := range needles {
		if !strings.Contains(haystack, needle) {
			return false
		}
	}
	return true
}
