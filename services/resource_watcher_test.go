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

// mockResourceFetcher implements ResourceFetcher for testing
type mockResourceFetcher struct {
	resources *models.ResourceCollection
	err       error
}

func (m *mockResourceFetcher) FetchResources(ctx context.Context) (*models.ResourceCollection, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.resources, nil
}

// setActiveDataSource updates the config manager to point to a test data source
func setActiveDataSource(t *testing.T, cm *ConfigManager, name string, url string, username string, password string) {
	t.Helper()
	cfg := models.DataSourceConfig{
		Type: models.DataSourceType(name),
		URL:  url,
	}
	if username != "" || password != "" {
		cfg.BasicAuth.Username = username
		cfg.BasicAuth.Password = password
	}

	if err := cm.UpdateDataSource(name, cfg); err != nil {
		t.Fatalf("failed to update data source: %v", err)
	}
	if err := cm.SetActiveDataSource(name); err != nil {
		t.Fatalf("failed to set active data source: %v", err)
	}
}

// TestNewResourceWatcher tests resource watcher creation
func TestNewResourceWatcher(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	// Create a mock server for the data source
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	// Update config manager with test URL
	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	watcher, err := NewResourceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewResourceWatcher() error = %v", err)
	}

	if watcher == nil {
		t.Fatal("NewResourceWatcher() returned nil")
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
	if watcher.httpClient == nil {
		t.Error("watcher.httpClient is nil")
	}
}

// TestResourceWatcher_Stop tests stopping when not running
func TestResourceWatcher_Stop(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	watcher, err := NewResourceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewResourceWatcher() error = %v", err)
	}

	// Should not panic when stopping a non-running watcher
	watcher.Stop()

	if watcher.isRunning {
		t.Error("watcher.isRunning should be false after Stop()")
	}
}

// TestResourceWatcher_RefreshFetcher tests fetcher refresh
func TestResourceWatcher_RefreshFetcher(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	watcher, err := NewResourceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewResourceWatcher() error = %v", err)
	}

	// Should not error on refresh
	err = watcher.refreshFetcher()
	if err != nil {
		t.Errorf("refreshFetcher() error = %v", err)
	}
}

// TestResourceWatcher_StartStop tests start/stop lifecycle
func TestResourceWatcher_StartStop(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	watcher, err := NewResourceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewResourceWatcher() error = %v", err)
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

// TestResourceWatcher_CheckResources tests resource checking
func TestResourceWatcher_CheckResources(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	// Create a mock server that returns resources via routers
	config := models.PangolinTraefikConfig{
		HTTP: models.PangolinHTTP{
			Routers: map[string]models.PangolinRouter{
				"test-router": {
					Rule:        "Host(`example.com`)",
					Service:     "test-service",
					Entrypoints: []string{"websecure"},
				},
			},
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
		json.NewEncoder(w).Encode(config)
	}))
	defer server.Close()

	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	watcher, err := NewResourceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewResourceWatcher() error = %v", err)
	}

	// Manually call checkResources
	err = watcher.checkResources()
	if err != nil {
		t.Fatalf("checkResources() error = %v", err)
	}

	// Verify resource was created in database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM resources WHERE host = 'example.com'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query resources: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 resource, got %d", count)
	}
}

// TestResourceWatcher_CheckResources_EmptyResult tests handling empty results
func TestResourceWatcher_CheckResources_EmptyResult(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	// Create a resource that will be disabled
	_, err := db.Exec(`
		INSERT INTO resources (id, host, service_id, status, created_at, updated_at)
		VALUES ('test-id', 'old.example.com', 'old-service', 'active', ?, ?)
	`, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("failed to create test resource: %v", err)
	}

	// Create a mock server that returns empty config
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	watcher, err := NewResourceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewResourceWatcher() error = %v", err)
	}

	// Should not error on empty result
	err = watcher.checkResources()
	if err != nil {
		t.Errorf("checkResources() should not error on empty result: %v", err)
	}

	// Verify old resource is now disabled
	var status string
	err = db.QueryRow("SELECT status FROM resources WHERE id = 'test-id'").Scan(&status)
	if err != nil {
		t.Fatalf("failed to query resource status: %v", err)
	}
	if status != "disabled" {
		t.Errorf("expected status 'disabled', got %q", status)
	}
}

// TestIsSystemRouter tests system router detection
func TestIsSystemRouter(t *testing.T) {
	tests := []struct {
		routerID string
		expected bool
	}{
		{"api@internal", true},
		{"dashboard@internal", true},
		{"acme-http@internal", true},
		{"noop@internal", true},
		{"api@file", true},              // Starts with api@
		{"dashboard@docker", true},       // Starts with dashboard@
		{"traefik@file", true},           // Starts with traefik@
		{"my-router@file", false},        // User router
		{"web-service@docker", false},    // User router
		{"api-router@file", false},       // Allowed user pattern
		{"next-router@file", false},      // Allowed user pattern
		{"ws-router@file", false},        // Allowed user pattern
	}

	for _, tt := range tests {
		t.Run(tt.routerID, func(t *testing.T) {
			got := isSystemRouter(tt.routerID)
			if got != tt.expected {
				t.Errorf("isSystemRouter(%q) = %v, want %v", tt.routerID, got, tt.expected)
			}
		})
	}
}

// TestResourceWatcher_UpdateOrCreateResource_New tests creating new resource
func TestResourceWatcher_UpdateOrCreateResource_New(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	watcher, err := NewResourceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewResourceWatcher() error = %v", err)
	}

	// Create a new resource
	resource := models.Resource{
		ID:         "new-router@file",
		Host:       "new.example.com",
		ServiceID:  "new-service",
		SourceType: "pangolin",
	}

	internalID, err := watcher.updateOrCreateResource(resource)
	if err != nil {
		t.Fatalf("updateOrCreateResource() error = %v", err)
	}

	if internalID == "" {
		t.Error("expected non-empty internal ID")
	}

	// Verify it was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM resources WHERE host = 'new.example.com'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query resources: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 resource, got %d", count)
	}
}

// TestResourceWatcher_UpdateOrCreateResource_Update tests updating existing resource
func TestResourceWatcher_UpdateOrCreateResource_Update(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	// Create existing resource
	existingID := "existing-uuid-123"
	_, err := db.Exec(`
		INSERT INTO resources (id, pangolin_router_id, host, service_id, status, source_type, created_at, updated_at)
		VALUES (?, 'existing-router', 'existing.example.com', 'old-service', 'active', 'pangolin', ?, ?)
	`, existingID, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("failed to create existing resource: %v", err)
	}

	watcher, err := NewResourceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewResourceWatcher() error = %v", err)
	}

	// Update the resource (by pangolin_router_id)
	resource := models.Resource{
		ID:         "existing-router",
		Host:       "existing.example.com",
		ServiceID:  "updated-service",
		SourceType: "pangolin",
	}

	returnedID, err := watcher.updateOrCreateResource(resource)
	if err != nil {
		t.Fatalf("updateOrCreateResource() error = %v", err)
	}

	if returnedID != existingID {
		t.Errorf("expected internal ID %q, got %q", existingID, returnedID)
	}

	// Verify service was updated
	var serviceID string
	err = db.QueryRow("SELECT service_id FROM resources WHERE id = ?", existingID).Scan(&serviceID)
	if err != nil {
		t.Fatalf("failed to query resource: %v", err)
	}
	if serviceID != "updated-service" {
		t.Errorf("expected service_id 'updated-service', got %q", serviceID)
	}
}

// TestResourceWatcher_UpdateOrCreateResource_ByHost tests finding by host
func TestResourceWatcher_UpdateOrCreateResource_ByHost(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	// Create existing resource with old pangolin_router_id
	existingID := "existing-uuid-456"
	_, err := db.Exec(`
		INSERT INTO resources (id, pangolin_router_id, host, service_id, status, source_type, created_at, updated_at)
		VALUES (?, 'old-router-id', 'findme.example.com', 'service', 'active', 'pangolin', ?, ?)
	`, existingID, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("failed to create existing resource: %v", err)
	}

	watcher, err := NewResourceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewResourceWatcher() error = %v", err)
	}

	// Update the resource with a NEW pangolin_router_id but same host
	resource := models.Resource{
		ID:         "new-router-id", // Pangolin changed the router ID
		Host:       "findme.example.com",
		ServiceID:  "service",
		SourceType: "pangolin",
	}

	returnedID, err := watcher.updateOrCreateResource(resource)
	if err != nil {
		t.Fatalf("updateOrCreateResource() error = %v", err)
	}

	// Should have found the existing resource by host and returned its internal ID
	if returnedID != existingID {
		t.Errorf("expected to find existing resource by host, got new ID %q", returnedID)
	}

	// Verify pangolin_router_id was updated
	var pangolinID string
	err = db.QueryRow("SELECT pangolin_router_id FROM resources WHERE id = ?", existingID).Scan(&pangolinID)
	if err != nil {
		t.Fatalf("failed to query resource: %v", err)
	}
	if pangolinID != "new-router-id" {
		t.Errorf("expected pangolin_router_id 'new-router-id', got %q", pangolinID)
	}
}

// TestResourceWatcher_DisablesRemovedResources tests marking removed resources as disabled
func TestResourceWatcher_DisablesRemovedResources(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	// Create a resource that will be "removed"
	_, err := db.Exec(`
		INSERT INTO resources (id, pangolin_router_id, host, service_id, status, created_at, updated_at)
		VALUES ('old-resource', 'old-router', 'old.example.com', 'old-service', 'active', ?, ?)
	`, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("failed to create test resource: %v", err)
	}

	// Create a mock server that returns empty routers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{
			HTTP: models.PangolinHTTP{
				Routers: map[string]models.PangolinRouter{},
			},
		})
	}))
	defer server.Close()

	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	watcher, err := NewResourceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewResourceWatcher() error = %v", err)
	}

	// Check resources (old-resource should be marked disabled)
	err = watcher.checkResources()
	if err != nil {
		t.Fatalf("checkResources() error = %v", err)
	}

	// Verify old-resource is now disabled
	var status string
	err = db.QueryRow("SELECT status FROM resources WHERE id = 'old-resource'").Scan(&status)
	if err != nil {
		t.Fatalf("failed to query resource status: %v", err)
	}
	if status != "disabled" {
		t.Errorf("expected status 'disabled', got %q", status)
	}
}

// TestResourceWatcher_FetchTraefikConfig tests fetching Traefik config
func TestResourceWatcher_FetchTraefikConfig(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	expectedConfig := models.PangolinTraefikConfig{
		HTTP: models.PangolinHTTP{
			Routers: map[string]models.PangolinRouter{
				"test-router": {
					Rule:    "Host(`test.example.com`)",
					Service: "test-service",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/traefik-config" {
			json.NewEncoder(w).Encode(expectedConfig)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	watcher, err := NewResourceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewResourceWatcher() error = %v", err)
	}

	ctx := context.Background()
	config, err := watcher.fetchTraefikConfig(ctx)
	if err != nil {
		t.Fatalf("fetchTraefikConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("fetchTraefikConfig() returned nil")
	}

	if len(config.HTTP.Routers) != 1 {
		t.Errorf("expected 1 router, got %d", len(config.HTTP.Routers))
	}

	if _, exists := config.HTTP.Routers["test-router"]; !exists {
		t.Error("expected 'test-router' in routers")
	}
}

// TestResourceWatcher_FetchTraefikConfig_WithAuth tests auth header
func TestResourceWatcher_FetchTraefikConfig_WithAuth(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "secret" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	// Update data source with auth
	setActiveDataSource(t, cm, "pangolin", server.URL, "admin", "secret")

	watcher, err := NewResourceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewResourceWatcher() error = %v", err)
	}

	ctx := context.Background()
	_, err = watcher.fetchTraefikConfig(ctx)
	if err != nil {
		t.Fatalf("fetchTraefikConfig() with auth error = %v", err)
	}
}

// TestResourceWatcher_PreservesRouterPriorityManual tests manual priority preservation
func TestResourceWatcher_PreservesRouterPriorityManual(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	// Create existing resource with manual priority
	existingID := "priority-test-uuid"
	_, err := db.Exec(`
		INSERT INTO resources (id, pangolin_router_id, host, service_id, status, router_priority, router_priority_manual, created_at, updated_at)
		VALUES (?, 'priority-router', 'priority.example.com', 'service', 'active', 500, 1, ?, ?)
	`, existingID, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("failed to create existing resource: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	watcher, err := NewResourceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewResourceWatcher() error = %v", err)
	}

	// Try to update with different priority from Pangolin
	resource := models.Resource{
		ID:             "priority-router",
		Host:           "priority.example.com",
		ServiceID:      "service",
		RouterPriority: 100, // Different priority
		SourceType:     "pangolin",
	}

	_, err = watcher.updateOrCreateResource(resource)
	if err != nil {
		t.Fatalf("updateOrCreateResource() error = %v", err)
	}

	// Verify manual priority was preserved
	var priority int
	err = db.QueryRow("SELECT router_priority FROM resources WHERE id = ?", existingID).Scan(&priority)
	if err != nil {
		t.Fatalf("failed to query resource: %v", err)
	}
	if priority != 500 {
		t.Errorf("expected router_priority 500 (manual), got %d", priority)
	}
}

// TestResourceWatcher_CreateWithDefaults tests default values for new resources
func TestResourceWatcher_CreateWithDefaults(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.PangolinTraefikConfig{})
	}))
	defer server.Close()

	setActiveDataSource(t, cm, "pangolin", server.URL, "", "")

	watcher, err := NewResourceWatcher(db, cm)
	if err != nil {
		t.Fatalf("NewResourceWatcher() error = %v", err)
	}

	// Create resource without optional fields
	resource := models.Resource{
		ID:        "default-router",
		Host:      "default.example.com",
		ServiceID: "default-service",
		// No OrgID, SiteID, Entrypoints, or RouterPriority
	}

	internalID, err := watcher.updateOrCreateResource(resource)
	if err != nil {
		t.Fatalf("updateOrCreateResource() error = %v", err)
	}

	// Verify defaults were applied
	var entrypoints, orgID string
	var priority int
	err = db.QueryRow(`
		SELECT entrypoints, org_id, router_priority
		FROM resources WHERE id = ?
	`, internalID).Scan(&entrypoints, &orgID, &priority)
	if err != nil {
		t.Fatalf("failed to query resource: %v", err)
	}

	if entrypoints != "websecure" {
		t.Errorf("expected entrypoints 'websecure', got %q", entrypoints)
	}
	if orgID != "unknown" {
		t.Errorf("expected org_id 'unknown', got %q", orgID)
	}
	if priority != 100 {
		t.Errorf("expected router_priority 100, got %d", priority)
	}
}
