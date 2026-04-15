package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	tmtypes "github.com/hhftechnology/middleware-manager/internal/traefikmanager/types"
)

// Env var names defined below are mirrored in docs/traefik-manager.md. When
// renaming or adding any variable, update that file in the same commit.
func LoadRuntimeConfig(debug bool) tmtypes.RuntimeConfig {
	settingsPath := firstNonEmpty(
		os.Getenv("TM_SETTINGS_PATH"),
		os.Getenv("SETTINGS_PATH"),
		"/app/config/manager.yml",
	)
	settingsDir := filepath.Dir(settingsPath)
	backupDir := firstNonEmpty(
		os.Getenv("TM_BACKUP_DIR"),
		os.Getenv("BACKUP_DIR"),
		"/app/backups",
	)

	configPaths := make([]string, 0)
	if rawPaths := strings.TrimSpace(os.Getenv("TM_CONFIG_PATHS")); rawPaths != "" {
		for _, path := range strings.Split(rawPaths, ",") {
			if trimmed := strings.TrimSpace(path); trimmed != "" {
				configPaths = append(configPaths, trimmed)
			}
		}
	}

	allowCORS := strings.EqualFold(strings.TrimSpace(os.Getenv("ALLOW_CORS")), "true")
	var trustedProxies []string
	if rawTrustedProxies := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES")); rawTrustedProxies != "" {
		for _, proxy := range strings.Split(rawTrustedProxies, ",") {
			if trimmed := strings.TrimSpace(proxy); trimmed != "" {
				trustedProxies = append(trustedProxies, trimmed)
			}
		}
	}

	return tmtypes.RuntimeConfig{
		Port:              firstNonEmpty(os.Getenv("PORT"), "3456"),
		UIPath:            firstNonEmpty(os.Getenv("TM_UI_PATH"), "/app/traefik-ui/dist"),
		SettingsPath:      settingsPath,
		BackupDir:         backupDir,
		ConfigDir:         strings.TrimSpace(os.Getenv("TM_CONFIG_DIR")),
		ConfigPath:        strings.TrimSpace(os.Getenv("TM_CONFIG_PATH")),
		ConfigPaths:       configPaths,
		TraefikAPIURL:     firstNonEmpty(os.Getenv("TRAEFIK_API_URL"), "http://traefik:8080"),
		AcmeJSONPath:      firstNonEmpty(os.Getenv("TM_ACME_JSON_PATH"), "/app/acme.json"),
		AccessLogPath:     firstNonEmpty(os.Getenv("TM_ACCESS_LOG_PATH"), "/app/logs/access.log"),
		StaticConfigPath:  firstNonEmpty(os.Getenv("TM_STATIC_CONFIG_PATH"), "/app/traefik.yml"),
		AllowCORS:         allowCORS,
		CORSOrigin:        strings.TrimSpace(os.Getenv("CORS_ORIGIN")),
		TrustedProxies:    trustedProxies,
		Debug:             debug,
		GitHubRepo:        "hhftechnology/middleware-manager",
		SettingsDir:       settingsDir,
		GroupsConfigFile:  filepath.Join(settingsDir, "dashboard.yml"),
		GroupsCacheDir:    filepath.Join(settingsDir, "cache"),
		HTTPClientTimeout: 5 * time.Second,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
