package services

import (
	"testing"

	"github.com/hhftechnology/middleware-manager/models"
)

func TestConfigManagerEnsureDefaults(t *testing.T) {
	cm := newTestConfigManager(t)

	if err := cm.EnsureDefaultDataSources("http://pangolin.local", "http://traefik.local"); err != nil {
		t.Fatalf("ensure defaults failed: %v", err)
	}

	sources := cm.GetDataSources()
	if len(sources) < 2 {
		t.Fatalf("expected default data sources to be present, got %d", len(sources))
	}

	if ds := cm.GetActiveSourceName(); ds != "pangolin" {
		t.Fatalf("expected pangolin to remain active, got %s", ds)
	}

	pangolin := sources["pangolin"]
	if pangolin.Type != models.PangolinAPI {
		t.Fatalf("expected pangolin type set, got %s", pangolin.Type)
	}
}

func TestConfigManagerSetActiveDataSource(t *testing.T) {
	cm := newTestConfigManager(t)
	if err := cm.SetActiveDataSource("traefik"); err != nil {
		t.Fatalf("set active failed: %v", err)
	}
	if cm.GetActiveSourceName() != "traefik" {
		t.Fatalf("expected active source to update to traefik")
	}

	if err := cm.SetActiveDataSource("unknown"); err == nil {
		t.Fatalf("expected error for unknown data source")
	}
}
