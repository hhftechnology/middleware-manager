package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tmtypes "github.com/hhftechnology/middleware-manager/internal/traefikmanager/types"
	"gopkg.in/yaml.v3"
)

type SettingsStore struct {
	path string
	cfg  tmtypes.RuntimeConfig
}

func NewSettingsStore(cfg tmtypes.RuntimeConfig) *SettingsStore {
	return &SettingsStore{path: cfg.SettingsPath, cfg: cfg}
}

func (s *SettingsStore) Defaults() tmtypes.Settings {
	visibleTabs := make(map[string]bool, len(tmtypes.OptionalTabs))
	for _, tab := range tmtypes.OptionalTabs {
		visibleTabs[tab] = false
	}
	return tmtypes.Settings{
		Domains:        []string{"example.com"},
		CertResolver:   "cloudflare",
		TraefikAPIURL:  s.cfg.TraefikAPIURL,
		VisibleTabs:    visibleTabs,
		DisabledRoutes: map[string]tmtypes.DisabledRoute{},
		SelfRoute:      tmtypes.SelfRoute{},
	}
}

func (s *SettingsStore) Load() (tmtypes.Settings, map[string]any, error) {
	settings := s.Defaults()
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return settings, map[string]any{}, nil
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return settings, nil, fmt.Errorf("read settings: %w", err)
	}
	raw := map[string]any{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return settings, nil, fmt.Errorf("parse settings yaml: %w", err)
	}

	if domains := StringSliceFromAny(raw["domains"]); len(domains) > 0 {
		settings.Domains = domains
	}
	if value := StringFromAny(raw["cert_resolver"]); value != "" {
		settings.CertResolver = value
	}
	if value := StringFromAny(raw["traefik_api_url"]); value != "" {
		settings.TraefikAPIURL = value
	}
	if tabs := MapFromAny(raw["visible_tabs"]); len(tabs) > 0 {
		for _, tab := range tmtypes.OptionalTabs {
			if value, ok := tabs[tab].(bool); ok {
				settings.VisibleTabs[tab] = value
			}
		}
	}
	if disabled := MapFromAny(raw["disabled_routes"]); len(disabled) > 0 {
		settings.DisabledRoutes = decodeDisabledRoutes(disabled)
	}
	if selfRoute := MapFromAny(raw["self_route"]); len(selfRoute) > 0 {
		settings.SelfRoute = tmtypes.SelfRoute{
			Domain:     StringFromAny(selfRoute["domain"]),
			ServiceURL: StringFromAny(selfRoute["service_url"]),
			RouterName: StringFromAny(selfRoute["router_name"]),
		}
	}
	settings.AcmeJSONPath = StringFromAny(raw["acme_json_path"])
	settings.AccessLogPath = StringFromAny(raw["access_log_path"])
	settings.StaticConfig = StringFromAny(raw["static_config_path"])

	return settings, raw, nil
}

func (s *SettingsStore) Save(settings tmtypes.Settings) error {
	_, raw, err := s.Load()
	if err != nil {
		return err
	}
	if raw == nil {
		raw = map[string]any{}
	}
	raw["domains"] = settings.Domains
	raw["cert_resolver"] = settings.CertResolver
	raw["traefik_api_url"] = settings.TraefikAPIURL
	raw["visible_tabs"] = settings.VisibleTabs
	raw["disabled_routes"] = settings.DisabledRoutes
	raw["self_route"] = settings.SelfRoute
	raw["acme_json_path"] = settings.AcmeJSONPath
	raw["access_log_path"] = settings.AccessLogPath
	raw["static_config_path"] = settings.StaticConfig

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create settings dir: %w", err)
	}
	encoded, err := yaml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("encode settings yaml: %w", err)
	}
	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, encoded, 0o644); err != nil {
		return fmt.Errorf("write temp settings: %w", err)
	}
	return os.Rename(tmpPath, s.path)
}

func (s *SettingsStore) EffectiveAcmePath(settings tmtypes.Settings) string {
	if strings.TrimSpace(settings.AcmeJSONPath) != "" {
		return strings.TrimSpace(settings.AcmeJSONPath)
	}
	return s.cfg.AcmeJSONPath
}

func (s *SettingsStore) EffectiveAccessLogPath(settings tmtypes.Settings) string {
	if strings.TrimSpace(settings.AccessLogPath) != "" {
		return strings.TrimSpace(settings.AccessLogPath)
	}
	return s.cfg.AccessLogPath
}

func (s *SettingsStore) EffectiveStaticConfigPath(settings tmtypes.Settings) string {
	if strings.TrimSpace(settings.StaticConfig) != "" {
		return strings.TrimSpace(settings.StaticConfig)
	}
	return s.cfg.StaticConfigPath
}

func decodeDisabledRoutes(raw map[string]any) map[string]tmtypes.DisabledRoute {
	out := make(map[string]tmtypes.DisabledRoute, len(raw))
	for key, value := range raw {
		entry := MapFromAny(value)
		out[key] = tmtypes.DisabledRoute{
			Protocol:   StringFromAny(entry["protocol"]),
			Router:     MapFromAny(entry["router"]),
			Service:    MapFromAny(entry["service"]),
			ConfigFile: StringFromAny(entry["configFile"]),
		}
	}
	return out
}
