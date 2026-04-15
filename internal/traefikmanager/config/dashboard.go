package config

import (
	"fmt"
	"os"
	"path/filepath"

	tmtypes "github.com/hhftechnology/middleware-manager/internal/traefikmanager/types"
	"gopkg.in/yaml.v3"
)

type DashboardStore struct {
	configPath string
	cacheDir   string
}

func NewDashboardStore(cfg tmtypes.RuntimeConfig) *DashboardStore {
	return &DashboardStore{configPath: cfg.GroupsConfigFile, cacheDir: cfg.GroupsCacheDir}
}

func (d *DashboardStore) Load() (tmtypes.DashboardConfig, error) {
	if _, err := os.Stat(d.configPath); os.IsNotExist(err) {
		return tmtypes.DashboardConfig{CustomGroups: []map[string]any{}, RouteOverrides: map[string]any{}}, nil
	}
	data, err := os.ReadFile(d.configPath)
	if err != nil {
		return tmtypes.DashboardConfig{}, fmt.Errorf("read dashboard config: %w", err)
	}
	raw := map[string]any{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return tmtypes.DashboardConfig{}, fmt.Errorf("parse dashboard config: %w", err)
	}
	customGroups := make([]map[string]any, 0)
	if items, ok := raw["custom_groups"].([]any); ok {
		for _, item := range items {
			customGroups = append(customGroups, MapFromAny(item))
		}
	}
	return tmtypes.DashboardConfig{
		CustomGroups:   customGroups,
		RouteOverrides: MapFromAny(raw["route_overrides"]),
	}, nil
}

func (d *DashboardStore) Save(config tmtypes.DashboardConfig) error {
	if err := os.MkdirAll(filepath.Dir(d.configPath), 0o755); err != nil {
		return fmt.Errorf("create dashboard dir: %w", err)
	}
	encoded, err := yaml.Marshal(map[string]any{
		"custom_groups":   config.CustomGroups,
		"route_overrides": config.RouteOverrides,
	})
	if err != nil {
		return fmt.Errorf("encode dashboard config: %w", err)
	}
	return os.WriteFile(d.configPath, encoded, 0o644)
}

func (d *DashboardStore) EnsureCacheDir() (string, error) {
	if err := os.MkdirAll(d.cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}
	return d.cacheDir, nil
}
