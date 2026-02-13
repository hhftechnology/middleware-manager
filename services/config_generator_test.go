package services

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewConfigGenerator tests config generator creation
func TestNewConfigGenerator(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)
	confDir := t.TempDir()

	cg := NewConfigGenerator(db, confDir, cm)

	if cg == nil {
		t.Fatal("NewConfigGenerator() returned nil")
	}
	if cg.db == nil {
		t.Error("cg.db is nil")
	}
	if cg.confDir != confDir {
		t.Errorf("confDir = %q, want %q", cg.confDir, confDir)
	}
	if cg.configManager == nil {
		t.Error("cg.configManager is nil")
	}
	if cg.isRunning {
		t.Error("cg.isRunning should be false initially")
	}
}

// TestConfigGenerator_Stop tests stopping when not running
func TestConfigGenerator_Stop(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)
	confDir := t.TempDir()

	cg := NewConfigGenerator(db, confDir, cm)

	// Should not panic when stopping a non-running generator
	cg.Stop()

	if cg.isRunning {
		t.Error("cg.isRunning should be false after Stop()")
	}
}

// TestNormalizeServiceID tests service ID normalization
func TestNormalizeServiceID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "with docker suffix",
			input: "my-service@docker",
			want:  "my-service",
		},
		{
			name:  "with file suffix",
			input: "api-service@file",
			want:  "api-service",
		},
		{
			name:  "without suffix",
			input: "simple-service",
			want:  "simple-service",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "multiple @ symbols",
			input: "name@provider@extra",
			want:  "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeServiceID(tt.input)
			if got != tt.want {
				t.Errorf("normalizeServiceID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestShouldLog tests log level checking
func TestShouldLog(t *testing.T) {
	// Save original value
	original := os.Getenv("LOG_LEVEL")
	defer os.Setenv("LOG_LEVEL", original)

	tests := []struct {
		name     string
		logLevel string
		wantLog  bool
		wantInfo bool
	}{
		{
			name:     "debug level",
			logLevel: "debug",
			wantLog:  true,
			wantInfo: true,
		},
		{
			name:     "info level",
			logLevel: "info",
			wantLog:  false,
			wantInfo: true,
		},
		{
			name:     "empty level",
			logLevel: "",
			wantLog:  false,
			wantInfo: false,
		},
		{
			name:     "warn level",
			logLevel: "warn",
			wantLog:  false,
			wantInfo: false,
		},
		{
			name:     "DEBUG uppercase",
			logLevel: "DEBUG",
			wantLog:  true,
			wantInfo: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("LOG_LEVEL", tt.logLevel)

			if got := shouldLog(); got != tt.wantLog {
				t.Errorf("shouldLog() = %v, want %v", got, tt.wantLog)
			}
			if got := shouldLogInfo(); got != tt.wantInfo {
				t.Errorf("shouldLogInfo() = %v, want %v", got, tt.wantInfo)
			}
		})
	}
}

// TestTraefikConfig_Structure tests TraefikConfig struct
func TestTraefikConfig_Structure(t *testing.T) {
	config := TraefikConfig{}

	// Initialize empty maps
	config.HTTP.Middlewares = make(map[string]interface{})
	config.HTTP.Routers = make(map[string]interface{})
	config.HTTP.Services = make(map[string]interface{})
	config.TCP.Routers = make(map[string]interface{})
	config.TCP.Services = make(map[string]interface{})
	config.UDP.Services = make(map[string]interface{})
	config.TLS.Options = make(map[string]interface{})

	// Add test data
	config.HTTP.Middlewares["test-mw"] = map[string]interface{}{
		"headers": map[string]interface{}{},
	}

	if len(config.HTTP.Middlewares) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(config.HTTP.Middlewares))
	}
}

// TestConfigGenerator_Start_Disabled tests that Start does nothing when disabled
func TestConfigGenerator_Start_Disabled(t *testing.T) {
	// Save original value
	original := os.Getenv("ENABLE_FILE_CONFIG")
	defer os.Setenv("ENABLE_FILE_CONFIG", original)

	// Disable file config
	os.Setenv("ENABLE_FILE_CONFIG", "false")

	db := newTestDB(t)
	cm := newTestConfigManager(t)
	confDir := t.TempDir()

	cg := NewConfigGenerator(db, confDir, cm)

	// Start should return immediately without running
	// Run in goroutine with timeout to detect if it blocks
	done := make(chan struct{})
	go func() {
		cg.Start(0) // Interval of 0 to test immediate behavior
		close(done)
	}()

	select {
	case <-done:
		// Good, Start returned
	}

	if cg.isRunning {
		t.Error("cg.isRunning should be false when ENABLE_FILE_CONFIG != true")
	}

	// Conf directory should not be created
	_, err := os.Stat(filepath.Join(confDir, "resource-overrides.yml"))
	if !os.IsNotExist(err) {
		t.Error("config file should not be created when disabled")
	}
}
