package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hhftechnology/middleware-manager/models"
)

// mockServiceFetcher implements ServiceFetcher for testing
type mockServiceFetcher struct {
	services *models.ServiceCollection
	err      error
}

func (m *mockServiceFetcher) FetchServices(ctx context.Context) (*models.ServiceCollection, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.services, nil
}

// TestNewServiceWatcher tests service watcher creation
func TestNewServiceWatcher(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	// Create a mock server for the data source
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	// Update config manager with test URL
	cm.db.Exec("UPDATE data_sources SET url = ?, is_active = 1 WHERE id = 1", server.URL)

	watcher, err := NewServiceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewServiceWatcher() error = %v", err)
	}

	if watcher == nil {
		t.Fatal("NewServiceWatcher() returned nil")
	}
	if watcher.db == nil {
		t.Error("watcher.db is nil")
	}
	if watcher.configManager == nil {
		t.Error("watcher.configManager is nil")
	}
	if watcher.isRunning {
		t.Error("watcher.isRunning should be false initially")
	}
}

// TestServiceWatcher_Stop tests stopping when not running
func TestServiceWatcher_Stop(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	cm.db.Exec("UPDATE data_sources SET url = ?, is_active = 1 WHERE id = 1", server.URL)

	watcher, err := NewServiceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewServiceWatcher() error = %v", err)
	}

	// Should not panic when stopping a non-running watcher
	watcher.Stop()

	if watcher.isRunning {
		t.Error("watcher.isRunning should be false after Stop()")
	}
}

// TestFormatServiceName tests service name formatting
func TestFormatServiceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with provider suffix",
			input:    "my-service@docker",
			expected: "My Service",
		},
		{
			name:     "with underscores",
			input:    "api_gateway_service",
			expected: "Api Gateway Service",
		},
		{
			name:     "with dashes",
			input:    "web-frontend",
			expected: "Web Frontend",
		},
		{
			name:     "simple name",
			input:    "backend",
			expected: "Backend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatServiceName(tt.input)
			if got != tt.expected {
				t.Errorf("formatServiceName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestShouldUpdateService tests service update detection
func TestShouldUpdateService(t *testing.T) {
	db := newTestDB(t)

	// Create a test service in the database
	_, err := db.Exec(`
		INSERT INTO services (id, name, type, config, status, created_at, updated_at)
		VALUES ('test-svc', 'Test Service', 'loadBalancer', '{"servers":[{"url":"http://old:8080"}]}', 'active', ?, ?)
	`, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("failed to create test service: %v", err)
	}

	tests := []struct {
		name       string
		newService models.Service
		shouldUpd  bool
	}{
		{
			name: "same config",
			newService: models.Service{
				ID:     "test-svc",
				Type:   "loadBalancer",
				Config: `{"servers":[{"url":"http://old:8080"}]}`,
			},
			shouldUpd: false,
		},
		{
			name: "different type",
			newService: models.Service{
				ID:     "test-svc",
				Type:   "weighted",
				Config: `{"servers":[{"url":"http://old:8080"}]}`,
			},
			shouldUpd: true,
		},
		{
			name: "different server URL",
			newService: models.Service{
				ID:     "test-svc",
				Type:   "loadBalancer",
				Config: `{"servers":[{"url":"http://new:8080"}]}`,
			},
			shouldUpd: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldUpdateService(db, tt.newService, "test-svc")
			if got != tt.shouldUpd {
				t.Errorf("shouldUpdateService() = %v, want %v", got, tt.shouldUpd)
			}
		})
	}
}

// TestConfigsAreDifferent tests config comparison
func TestConfigsAreDifferent(t *testing.T) {
	tests := []struct {
		name     string
		config1  map[string]interface{}
		config2  map[string]interface{}
		expected bool
	}{
		{
			name:     "identical configs",
			config1:  map[string]interface{}{"key": "value"},
			config2:  map[string]interface{}{"key": "value"},
			expected: false,
		},
		{
			name:     "different values",
			config1:  map[string]interface{}{"key": "value1"},
			config2:  map[string]interface{}{"key": "value2"},
			expected: true,
		},
		{
			name:     "missing key in config2",
			config1:  map[string]interface{}{"key1": "value", "key2": "value"},
			config2:  map[string]interface{}{"key1": "value"},
			expected: true,
		},
		{
			name: "same servers",
			config1: map[string]interface{}{
				"servers": []interface{}{
					map[string]interface{}{"url": "http://localhost:8080"},
				},
			},
			config2: map[string]interface{}{
				"servers": []interface{}{
					map[string]interface{}{"url": "http://localhost:8080"},
				},
			},
			expected: false,
		},
		{
			name: "different servers",
			config1: map[string]interface{}{
				"servers": []interface{}{
					map[string]interface{}{"url": "http://server1:8080"},
				},
			},
			config2: map[string]interface{}{
				"servers": []interface{}{
					map[string]interface{}{"url": "http://server2:8080"},
				},
			},
			expected: true,
		},
		{
			name: "different number of servers",
			config1: map[string]interface{}{
				"servers": []interface{}{
					map[string]interface{}{"url": "http://server1:8080"},
				},
			},
			config2: map[string]interface{}{
				"servers": []interface{}{
					map[string]interface{}{"url": "http://server1:8080"},
					map[string]interface{}{"url": "http://server2:8080"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := configsAreDifferent(tt.config1, tt.config2)
			if got != tt.expected {
				t.Errorf("configsAreDifferent() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestServiceWatcher_CheckServices tests service checking
func TestServiceWatcher_CheckServices(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	// Create a mock server that returns services
	services := models.PangolinTraefikConfig{
		HTTP: models.PangolinHTTP{
			Services: map[string]models.PangolinService{
				"test-service": {
					LoadBalancer: map[string]interface{}{
						"servers": []map[string]interface{}{
							{"url": "http://backend:8080"},
						},
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(services)
	}))
	defer server.Close()

	cm.db.Exec("UPDATE data_sources SET url = ?, is_active = 1 WHERE id = 1", server.URL)

	watcher, err := NewServiceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewServiceWatcher() error = %v", err)
	}

	// Manually call checkServices
	err = watcher.checkServices()
	if err != nil {
		t.Fatalf("checkServices() error = %v", err)
	}

	// Verify service was created in database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM services WHERE id = 'test-service'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query services: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 service, got %d", count)
	}
}

// TestServiceWatcher_CheckServices_EmptyResult tests handling empty results
func TestServiceWatcher_CheckServices_EmptyResult(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	// Create a mock server that returns empty services
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	cm.db.Exec("UPDATE data_sources SET url = ?, is_active = 1 WHERE id = 1", server.URL)

	watcher, err := NewServiceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewServiceWatcher() error = %v", err)
	}

	// Should not error on empty result
	err = watcher.checkServices()
	if err != nil {
		t.Errorf("checkServices() should not error on empty result: %v", err)
	}
}

// TestServiceWatcher_RefreshFetcher tests fetcher refresh
func TestServiceWatcher_RefreshFetcher(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	cm.db.Exec("UPDATE data_sources SET url = ?, is_active = 1 WHERE id = 1", server.URL)

	watcher, err := NewServiceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewServiceWatcher() error = %v", err)
	}

	// Should not error on refresh
	err = watcher.refreshFetcher()
	if err != nil {
		t.Errorf("refreshFetcher() error = %v", err)
	}
}

// TestServiceWatcher_StartStop tests start/stop lifecycle
func TestServiceWatcher_StartStop(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	cm.db.Exec("UPDATE data_sources SET url = ?, is_active = 1 WHERE id = 1", server.URL)

	watcher, err := NewServiceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewServiceWatcher() error = %v", err)
	}

	// Start in goroutine
	go watcher.Start(100 * time.Millisecond)

	// Wait a bit for it to start
	time.Sleep(50 * time.Millisecond)

	if !watcher.isRunning {
		t.Error("watcher should be running after Start()")
	}

	// Stop it
	watcher.Stop()

	// Wait for stop to complete
	time.Sleep(50 * time.Millisecond)

	if watcher.isRunning {
		t.Error("watcher should not be running after Stop()")
	}
}

// TestServiceWatcher_UpdateOrCreateService tests service upsert
func TestServiceWatcher_UpdateOrCreateService(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	cm.db.Exec("UPDATE data_sources SET url = ?, is_active = 1 WHERE id = 1", server.URL)

	watcher, err := NewServiceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewServiceWatcher() error = %v", err)
	}

	// Create a new service
	service := models.Service{
		ID:     "new-service@file",
		Name:   "New Service",
		Type:   string(models.LoadBalancerType),
		Config: `{"servers":[{"url":"http://backend:8080"}]}`,
	}

	err = watcher.updateOrCreateService(service)
	if err != nil {
		t.Fatalf("updateOrCreateService() error = %v", err)
	}

	// Verify it was created (with normalized ID)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM services WHERE id = 'new-service'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query services: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 service, got %d", count)
	}

	// Update the same service
	service.Config = `{"servers":[{"url":"http://updated:9090"}]}`
	err = watcher.updateOrCreateService(service)
	if err != nil {
		t.Fatalf("updateOrCreateService() update error = %v", err)
	}

	// Should still be 1 service
	err = db.QueryRow("SELECT COUNT(*) FROM services WHERE id = 'new-service'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query services: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 service after update, got %d", count)
	}
}

// TestServiceWatcher_DisablesRemovedServices tests marking removed services as disabled
func TestServiceWatcher_DisablesRemovedServices(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	// First, create a service with source_type='pangolin' that will be "removed"
	_, err := db.Exec(`
		INSERT INTO services (id, name, type, config, status, source_type, created_at, updated_at)
		VALUES ('old-service', 'Old Service', 'loadBalancer', '{}', 'active', 'pangolin', ?, ?)
	`, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("failed to create test service: %v", err)
	}

	// Create a mock server that returns empty services
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{
			HTTP: models.PangolinHTTP{
				Services: map[string]models.PangolinService{},
			},
		})
	}))
	defer server.Close()

	cm.db.Exec("UPDATE data_sources SET url = ?, is_active = 1 WHERE id = 1", server.URL)

	watcher, err := NewServiceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewServiceWatcher() error = %v", err)
	}

	// Check services (old-service should be marked disabled)
	err = watcher.checkServices()
	if err != nil {
		t.Fatalf("checkServices() error = %v", err)
	}

	// Verify old-service is now disabled
	var status string
	err = db.QueryRow("SELECT status FROM services WHERE id = 'old-service'").Scan(&status)
	if err != nil {
		t.Fatalf("failed to query service status: %v", err)
	}
	if status != "disabled" {
		t.Errorf("expected status 'disabled', got %q", status)
	}
}

// TestServiceWatcher_PreservesManualServices tests that manual services are not affected
func TestServiceWatcher_PreservesManualServices(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	// Create a manual service
	_, err := db.Exec(`
		INSERT INTO services (id, name, type, config, status, source_type, created_at, updated_at)
		VALUES ('manual-service', 'Manual Service', 'loadBalancer', '{}', 'active', 'manual', ?, ?)
	`, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("failed to create manual service: %v", err)
	}

	// Create a mock server that returns empty services
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{
			HTTP: models.PangolinHTTP{
				Services: map[string]models.PangolinService{},
			},
		})
	}))
	defer server.Close()

	cm.db.Exec("UPDATE data_sources SET url = ?, is_active = 1 WHERE id = 1", server.URL)

	watcher, err := NewServiceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewServiceWatcher() error = %v", err)
	}

	// Check services
	err = watcher.checkServices()
	if err != nil {
		t.Fatalf("checkServices() error = %v", err)
	}

	// Manual service should still be active
	var status string
	err = db.QueryRow("SELECT status FROM services WHERE id = 'manual-service'").Scan(&status)
	if err != nil {
		t.Fatalf("failed to query service status: %v", err)
	}
	if status != "active" {
		t.Errorf("manual service should remain 'active', got %q", status)
	}
}
