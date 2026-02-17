package models

import (
	"testing"
)

func TestFormatBasicAuth(t *testing.T) {
	t.Run("masks non-empty password", func(t *testing.T) {
		dc := DataSourceConfig{}
		dc.BasicAuth.Username = "admin"
		dc.BasicAuth.Password = "secret123"

		dc.FormatBasicAuth()

		if dc.BasicAuth.Password != "••••••••" {
			t.Errorf("password = %q, want masked", dc.BasicAuth.Password)
		}
		if dc.BasicAuth.Username != "admin" {
			t.Errorf("username should be unchanged, got %q", dc.BasicAuth.Username)
		}
	})

	t.Run("no-op on empty password", func(t *testing.T) {
		dc := DataSourceConfig{}
		dc.BasicAuth.Username = "admin"
		dc.BasicAuth.Password = ""

		dc.FormatBasicAuth()

		if dc.BasicAuth.Password != "" {
			t.Errorf("empty password should remain empty, got %q", dc.BasicAuth.Password)
		}
	})
}

func TestJoinTLSDomains(t *testing.T) {
	tests := []struct {
		name     string
		domains  []TraefikTLSDomain
		expected string
	}{
		{
			name:     "single domain",
			domains:  []TraefikTLSDomain{{Main: "example.com"}},
			expected: "example.com",
		},
		{
			name:     "domain with SANs",
			domains:  []TraefikTLSDomain{{Main: "example.com", Sans: []string{"www.example.com", "api.example.com"}}},
			expected: "example.com,www.example.com,api.example.com",
		},
		{
			name: "multiple domains",
			domains: []TraefikTLSDomain{
				{Main: "a.com"},
				{Main: "b.com"},
			},
			expected: "a.com,b.com",
		},
		{
			name:     "empty slice",
			domains:  []TraefikTLSDomain{},
			expected: "",
		},
		{
			name:     "domain with empty Main",
			domains:  []TraefikTLSDomain{{Main: "", Sans: []string{"alt.com"}}},
			expected: "alt.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JoinTLSDomains(tt.domains)
			if got != tt.expected {
				t.Errorf("JoinTLSDomains() = %q, want %q", got, tt.expected)
			}
		})
	}
}
