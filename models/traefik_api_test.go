package models

import (
	"encoding/json"
	"testing"
)

func TestTraefikOverview_ScanValue(t *testing.T) {
	t.Run("round-trip", func(t *testing.T) {
		original := TraefikOverview{
			HTTP: ProtocolOverview{
				Routers:  StatusCount{Total: 5, Warnings: 1},
				Services: StatusCount{Total: 3},
			},
			Providers: []string{"docker", "file"},
		}

		data, err := original.Value()
		if err != nil {
			t.Fatalf("Value() error = %v", err)
		}

		var scanned TraefikOverview
		if err := scanned.Scan(data); err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if scanned.HTTP.Routers.Total != 5 {
			t.Errorf("HTTP.Routers.Total = %d, want 5", scanned.HTTP.Routers.Total)
		}
		if len(scanned.Providers) != 2 {
			t.Errorf("len(Providers) = %d, want 2", len(scanned.Providers))
		}
	})

	t.Run("nil input", func(t *testing.T) {
		var o TraefikOverview
		if err := o.Scan(nil); err != nil {
			t.Errorf("Scan(nil) error = %v", err)
		}
	})

	t.Run("[]byte input", func(t *testing.T) {
		data, _ := json.Marshal(TraefikOverview{Providers: []string{"test"}})
		var o TraefikOverview
		if err := o.Scan(data); err != nil {
			t.Errorf("Scan([]byte) error = %v", err)
		}
		if len(o.Providers) != 1 {
			t.Errorf("len(Providers) = %d, want 1", len(o.Providers))
		}
	})

	t.Run("string input", func(t *testing.T) {
		data, _ := json.Marshal(TraefikOverview{Providers: []string{"str"}})
		var o TraefikOverview
		if err := o.Scan(string(data)); err != nil {
			t.Errorf("Scan(string) error = %v", err)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		var o TraefikOverview
		if err := o.Scan([]byte("{invalid")); err == nil {
			t.Error("Scan(invalid) should return error")
		}
	})
}

func TestTraefikEntrypoint_ScanValue(t *testing.T) {
	original := TraefikEntrypoint{
		Name:    "web",
		Address: ":80",
	}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	var scanned TraefikEntrypoint
	if err := scanned.Scan(data); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if scanned.Name != "web" {
		t.Errorf("Name = %q, want web", scanned.Name)
	}

	t.Run("nil", func(t *testing.T) {
		var e TraefikEntrypoint
		if err := e.Scan(nil); err != nil {
			t.Errorf("Scan(nil) error = %v", err)
		}
	})
}

func TestTraefikVersion_ScanValue(t *testing.T) {
	original := TraefikVersion{
		Version:  "3.0.0",
		Codename: "beaufort",
	}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	var scanned TraefikVersion
	if err := scanned.Scan(data); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if scanned.Version != "3.0.0" {
		t.Errorf("Version = %q, want 3.0.0", scanned.Version)
	}

	t.Run("nil", func(t *testing.T) {
		var v TraefikVersion
		if err := v.Scan(nil); err != nil {
			t.Errorf("Scan(nil) error = %v", err)
		}
	})
}

func TestFullTraefikData_Counts(t *testing.T) {
	t.Run("populated", func(t *testing.T) {
		d := FullTraefikData{
			HTTPRouters:     make([]TraefikRouter, 3),
			HTTPServices:    make([]TraefikService, 2),
			HTTPMiddlewares: make([]TraefikMiddleware, 4),
			TCPRouters:      make([]TCPRouter, 1),
			TCPServices:     make([]TCPService, 1),
			TCPMiddlewares:  make([]TCPMiddleware, 1),
			UDPRouters:      make([]UDPRouter, 2),
			UDPServices:     make([]UDPService, 1),
		}

		if d.GetHTTPRouterCount() != 3 {
			t.Errorf("GetHTTPRouterCount = %d", d.GetHTTPRouterCount())
		}
		if d.GetTCPRouterCount() != 1 {
			t.Errorf("GetTCPRouterCount = %d", d.GetTCPRouterCount())
		}
		if d.GetUDPRouterCount() != 2 {
			t.Errorf("GetUDPRouterCount = %d", d.GetUDPRouterCount())
		}
		if d.GetTotalRouterCount() != 6 {
			t.Errorf("GetTotalRouterCount = %d", d.GetTotalRouterCount())
		}
		if d.GetHTTPServiceCount() != 2 {
			t.Errorf("GetHTTPServiceCount = %d", d.GetHTTPServiceCount())
		}
		if d.GetTCPServiceCount() != 1 {
			t.Errorf("GetTCPServiceCount = %d", d.GetTCPServiceCount())
		}
		if d.GetUDPServiceCount() != 1 {
			t.Errorf("GetUDPServiceCount = %d", d.GetUDPServiceCount())
		}
		if d.GetTotalServiceCount() != 4 {
			t.Errorf("GetTotalServiceCount = %d", d.GetTotalServiceCount())
		}
		if d.GetHTTPMiddlewareCount() != 4 {
			t.Errorf("GetHTTPMiddlewareCount = %d", d.GetHTTPMiddlewareCount())
		}
		if d.GetTCPMiddlewareCount() != 1 {
			t.Errorf("GetTCPMiddlewareCount = %d", d.GetTCPMiddlewareCount())
		}
		if d.GetTotalMiddlewareCount() != 5 {
			t.Errorf("GetTotalMiddlewareCount = %d", d.GetTotalMiddlewareCount())
		}
	})

	t.Run("empty", func(t *testing.T) {
		d := FullTraefikData{}
		if d.GetTotalRouterCount() != 0 {
			t.Errorf("empty GetTotalRouterCount = %d", d.GetTotalRouterCount())
		}
		if d.GetTotalServiceCount() != 0 {
			t.Errorf("empty GetTotalServiceCount = %d", d.GetTotalServiceCount())
		}
		if d.GetTotalMiddlewareCount() != 0 {
			t.Errorf("empty GetTotalMiddlewareCount = %d", d.GetTotalMiddlewareCount())
		}
	})
}
