package services

import (
	"testing"

	"github.com/hhftechnology/middleware-manager/models"
)

// TestExtractHostFromRule tests extracting host from various Traefik rule formats
func TestExtractHostFromRule(t *testing.T) {
	tests := []struct {
		name string
		rule string
		want string
	}{
		{
			name: "standard Host rule",
			rule: "Host(`example.com`)",
			want: "example.com",
		},
		{
			name: "Host rule with subdomain",
			rule: "Host(`api.example.com`)",
			want: "api.example.com",
		},
		{
			name: "Host rule with path",
			rule: "Host(`example.com`) && PathPrefix(`/api`)",
			want: "example.com",
		},
		{
			name: "HostRegexp with wildcard",
			rule: "HostRegexp(`.+`)",
			want: "any-host",
		},
		{
			name: "HostRegexp with specific pattern",
			rule: "HostRegexp(`.*\\.development\\.hhf\\.technology`)",
			want: "x.development.hhf.technology",
		},
		{
			name: "legacy Host:domain format",
			rule: "Host:example.com",
			want: "example.com",
		},
		{
			name: "legacy Host format with space terminator",
			rule: "Host:api.test.com PathPrefix:/v1",
			want: "api.test.com",
		},
		{
			name: "complex rule with && operators",
			rule: "PathPrefix(`/api`) && Host(`backend.example.com`) && Headers(`X-Custom`, `true`)",
			want: "backend.example.com",
		},
		{
			name: "empty rule",
			rule: "",
			want: "",
		},
		{
			name: "rule without host",
			rule: "PathPrefix(`/api`)",
			want: "",
		},
		{
			name: "multiple hosts (first match)",
			rule: "Host(`first.com`) || Host(`second.com`)",
			want: "first.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHostFromRule(tt.rule)
			if got != tt.want {
				t.Errorf("extractHostFromRule(%q) = %q, want %q", tt.rule, got, tt.want)
			}
		})
	}
}

// TestExtractHostFromRegexp tests regex pattern to host extraction
func TestExtractHostFromRegexp(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    string
	}{
		{
			name:    "development hhf technology pattern",
			pattern: `app\.development\.hhf\.technology`,
			want:    "app.development.hhf.technology",
		},
		{
			name:    "subdomain wildcard pattern",
			pattern: `[a-z]+\.development\.hhf\.technology`,
			want:    "x.development.hhf.technology",
		},
		{
			name:    "simple domain pattern",
			pattern: `example\.com`,
			want:    "example.com",
		},
		{
			name:    "pattern without dots",
			pattern: `localhost`,
			want:    "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHostFromRegexp(tt.pattern)
			if got != tt.want {
				t.Errorf("extractHostFromRegexp(%q) = %q, want %q", tt.pattern, got, tt.want)
			}
		})
	}
}

// TestCleanupRegexChars tests regex character cleanup
func TestCleanupRegexChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "digit sequence",
			input: `user\d+`,
			want:  "userN",
		},
		{
			name:  "digit class",
			input: `item[0-9]+`,
			want:  "itemN",
		},
		{
			name:  "alphanumeric class",
			input: `[a-z0-9]+.example.com`,
			want:  "x.example.com",
		},
		{
			name:  "word characters",
			input: `prefix-\w+-suffix`,
			want:  "prefix-x-suffix",
		},
		{
			name:  "anchored pattern",
			input: `^example\.com$`,
			want:  "example.com",
		},
		{
			name:  "any char sequences",
			input: `.*.example.com`,
			want:  "x.example.com",
		},
		{
			name:  "alternation",
			input: `app|api`,
			want:  "app-api",
		},
		{
			name:  "clean string (no changes)",
			input: `example.com`,
			want:  "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanupRegexChars(tt.input)
			if got != tt.want {
				t.Errorf("cleanupRegexChars(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestExtractHostSNI tests SNI host extraction
func TestExtractHostSNI(t *testing.T) {
	tests := []struct {
		name string
		rule string
		want string
	}{
		{
			name: "simple HostSNI",
			rule: "HostSNI(`tcp.example.com`)",
			want: "tcp.example.com",
		},
		{
			name: "HostSNI with wildcard",
			rule: "HostSNI(`*`)",
			want: "*",
		},
		{
			name: "no HostSNI",
			rule: "Host(`example.com`)",
			want: "",
		},
		{
			name: "empty rule",
			rule: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHostSNI(tt.rule)
			if got != tt.want {
				t.Errorf("extractHostSNI(%q) = %q, want %q", tt.rule, got, tt.want)
			}
		})
	}
}

// TestExtractHostSNIRegexp tests SNI regexp host extraction
func TestExtractHostSNIRegexp(t *testing.T) {
	tests := []struct {
		name string
		rule string
		want string
	}{
		{
			name: "HostSNIRegexp with pattern",
			rule: "HostSNIRegexp(`.*\\.tcp\\.example\\.com`)",
			want: "x.tcp.example.com",
		},
		{
			name: "no HostSNIRegexp",
			rule: "HostSNI(`example.com`)",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHostSNIRegexp(tt.rule)
			if got != tt.want {
				t.Errorf("extractHostSNIRegexp(%q) = %q, want %q", tt.rule, got, tt.want)
			}
		})
	}
}

// TestJoinEntrypoints tests entrypoint joining
func TestJoinEntrypoints(t *testing.T) {
	tests := []struct {
		name        string
		entrypoints []string
		want        string
	}{
		{
			name:        "single entrypoint",
			entrypoints: []string{"web"},
			want:        "web",
		},
		{
			name:        "multiple entrypoints",
			entrypoints: []string{"web", "websecure"},
			want:        "web,websecure",
		},
		{
			name:        "empty slice",
			entrypoints: []string{},
			want:        "",
		},
		{
			name:        "nil slice",
			entrypoints: nil,
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinEntrypoints(tt.entrypoints)
			if got != tt.want {
				t.Errorf("joinEntrypoints(%v) = %q, want %q", tt.entrypoints, got, tt.want)
			}
		})
	}
}

// TestNewResourceFetcher tests fetcher factory function
func TestNewResourceFetcher(t *testing.T) {
	tests := []struct {
		name    string
		config  models.DataSourceConfig
		wantErr bool
	}{
		{
			name: "pangolin fetcher",
			config: models.DataSourceConfig{
				Type: models.PangolinAPI,
				URL:  "http://localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "traefik fetcher",
			config: models.DataSourceConfig{
				Type: models.TraefikAPI,
				URL:  "http://localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "unknown type",
			config: models.DataSourceConfig{
				Type: "unknown",
				URL:  "http://localhost:8080",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher, err := NewResourceFetcher(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewResourceFetcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && fetcher == nil {
				t.Error("NewResourceFetcher() returned nil fetcher without error")
			}
		})
	}
}

// TestNewDataFetcher tests data fetcher factory function
func TestNewDataFetcher(t *testing.T) {
	tests := []struct {
		name    string
		config  models.DataSourceConfig
		wantErr bool
	}{
		{
			name: "pangolin data fetcher",
			config: models.DataSourceConfig{
				Type: models.PangolinAPI,
				URL:  "http://localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "traefik data fetcher",
			config: models.DataSourceConfig{
				Type: models.TraefikAPI,
				URL:  "http://localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "unknown type",
			config: models.DataSourceConfig{
				Type: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher, err := NewDataFetcher(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDataFetcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && fetcher == nil {
				t.Error("NewDataFetcher() returned nil fetcher without error")
			}
		})
	}
}

// TestNewFullDataFetcher tests full data fetcher factory function
func TestNewFullDataFetcher(t *testing.T) {
	tests := []struct {
		name    string
		config  models.DataSourceConfig
		wantErr bool
	}{
		{
			name: "pangolin full fetcher",
			config: models.DataSourceConfig{
				Type: models.PangolinAPI,
				URL:  "http://localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "traefik full fetcher",
			config: models.DataSourceConfig{
				Type: models.TraefikAPI,
				URL:  "http://localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "unknown type",
			config: models.DataSourceConfig{
				Type: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher, err := NewFullDataFetcher(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFullDataFetcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && fetcher == nil {
				t.Error("NewFullDataFetcher() returned nil fetcher without error")
			}
		})
	}
}
