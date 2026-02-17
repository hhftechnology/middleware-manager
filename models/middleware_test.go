package models

import (
	"testing"
)

func TestMiddleware_ConfigMap(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		m := Middleware{Config: `{"key":"value","num":42}`}
		cfg, err := m.ConfigMap()
		if err != nil {
			t.Fatalf("ConfigMap() error = %v", err)
		}
		if cfg["key"] != "value" {
			t.Errorf("cfg[key] = %v, want value", cfg["key"])
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		m := Middleware{Config: `{invalid`}
		_, err := m.ConfigMap()
		if err == nil {
			t.Error("ConfigMap() should return error for invalid JSON")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		m := Middleware{Config: ""}
		_, err := m.ConfigMap()
		if err == nil {
			t.Error("ConfigMap() should return error for empty string")
		}
	})
}
