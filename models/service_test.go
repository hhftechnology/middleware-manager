package models

import (
	"testing"
)

func TestIsValidServiceType(t *testing.T) {
	validTypes := []string{"loadBalancer", "weighted", "mirroring", "failover"}
	for _, typ := range validTypes {
		t.Run(typ+"_valid", func(t *testing.T) {
			if !IsValidServiceType(typ) {
				t.Errorf("IsValidServiceType(%q) = false, want true", typ)
			}
		})
	}

	t.Run("invalid", func(t *testing.T) {
		if IsValidServiceType("invalidType") {
			t.Error("IsValidServiceType(invalidType) = true, want false")
		}
	})

	t.Run("empty", func(t *testing.T) {
		if IsValidServiceType("") {
			t.Error("IsValidServiceType('') = true, want false")
		}
	})
}

func TestService_ConfigMap(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		s := Service{Config: `{"servers":[{"url":"http://localhost:8080"}]}`}
		cfg, err := s.ConfigMap()
		if err != nil {
			t.Fatalf("ConfigMap() error = %v", err)
		}
		if cfg["servers"] == nil {
			t.Error("cfg[servers] should not be nil")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		s := Service{Config: `not json`}
		_, err := s.ConfigMap()
		if err == nil {
			t.Error("ConfigMap() should return error for invalid JSON")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		s := Service{Config: ""}
		_, err := s.ConfigMap()
		if err == nil {
			t.Error("ConfigMap() should return error for empty string")
		}
	})
}

func TestGetServiceProcessor(t *testing.T) {
	p := GetServiceProcessor("loadBalancer")
	if p == nil {
		t.Fatal("GetServiceProcessor returned nil")
	}
	if _, ok := p.(*DefaultServiceProcessor); !ok {
		t.Error("expected DefaultServiceProcessor")
	}
}

func TestProcessServiceConfig(t *testing.T) {
	config := map[string]interface{}{
		"servers": []interface{}{
			map[string]interface{}{"url": "http://localhost:8080"},
		},
	}

	result := ProcessServiceConfig("loadBalancer", config)
	if result == nil {
		t.Error("ProcessServiceConfig returned nil")
	}
}
