package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

// TestNewServiceHandler tests service handler creation
func TestNewServiceHandler(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	if handler == nil {
		t.Fatal("NewServiceHandler() returned nil")
	}
	if handler.DB == nil {
		t.Error("handler.DB is nil")
	}
}

// TestServiceHandler_GetServices tests fetching all services
func TestServiceHandler_GetServices(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	// Insert test services
	testutil.MustExec(t, db, `
		INSERT INTO services (id, name, type, config, status, source_type)
		VALUES ('svc-1', 'backend-1', 'loadBalancer', '{"servers":[{"url":"http://localhost:8080"}]}', 'active', 'manual')
	`)
	testutil.MustExec(t, db, `
		INSERT INTO services (id, name, type, config, status, source_type)
		VALUES ('svc-2', 'backend-2', 'loadBalancer', '{"servers":[{"url":"http://localhost:9090"}]}', 'active', 'pangolin')
	`)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/services", nil)
	handler.GetServices(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var services []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &services); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(services) != 2 {
		t.Errorf("expected 2 services, got %d", len(services))
	}
}

// TestServiceHandler_GetServices_StatusFilter tests filtering by status
func TestServiceHandler_GetServices_StatusFilter(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO services (id, name, type, config, status, source_type)
		VALUES ('svc-active', 'active-service', 'loadBalancer', '{}', 'active', 'manual')
	`)
	testutil.MustExec(t, db, `
		INSERT INTO services (id, name, type, config, status, source_type)
		VALUES ('svc-disabled', 'disabled-service', 'loadBalancer', '{}', 'disabled', 'pangolin')
	`)

	tests := []struct {
		name          string
		status        string
		expectedCount int
	}{
		{"default (active)", "", 1},
		{"active only", "active", 1},
		{"disabled only", "disabled", 1},
		{"all", "all", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/services"
			if tt.status != "" {
				path += "?status=" + tt.status
			}
			c, rec := testutil.NewContext(t, http.MethodGet, path, nil)
			handler.GetServices(c)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}

			var services []map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &services)

			if len(services) != tt.expectedCount {
				t.Errorf("expected %d services, got %d", tt.expectedCount, len(services))
			}
		})
	}
}

// TestServiceHandler_GetServices_Pagination tests paginated results
func TestServiceHandler_GetServices_Pagination(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	// Insert 5 test services
	for i := 1; i <= 5; i++ {
		testutil.MustExec(t, db, `
			INSERT INTO services (id, name, type, config, status, source_type)
			VALUES (?, ?, 'loadBalancer', '{}', 'active', 'manual')
		`, "svc-"+string(rune('0'+i)), "service-"+string(rune('0'+i)))
	}

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/services?page=1&page_size=2", nil)
	handler.GetServices(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	// Check pagination metadata
	if response["total"] == nil {
		t.Error("expected total in paginated response")
	}
	if response["page"] == nil {
		t.Error("expected page in paginated response")
	}
	if response["data"] == nil {
		t.Error("expected data array in paginated response")
	}
}

// TestServiceHandler_GetService tests fetching a single service
func TestServiceHandler_GetService(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO services (id, name, type, config, status, source_type)
		VALUES ('test-svc-id', 'test-service', 'loadBalancer', '{"servers":[{"url":"http://backend:8080"}]}', 'active', 'manual')
	`)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/services/test-svc-id", nil)
	c.Params = gin.Params{{Key: "id", Value: "test-svc-id"}}
	handler.GetService(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var service map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &service)

	if service["name"] != "test-service" {
		t.Errorf("expected name test-service, got %v", service["name"])
	}
	if service["type"] != "loadBalancer" {
		t.Errorf("expected type loadBalancer, got %v", service["type"])
	}
}

// TestServiceHandler_GetService_NotFound tests fetching non-existent service
func TestServiceHandler_GetService_NotFound(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/services/non-existent", nil)
	c.Params = gin.Params{{Key: "id", Value: "non-existent"}}
	handler.GetService(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestServiceHandler_CreateService tests creating a new service
func TestServiceHandler_CreateService(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	body := bytes.NewBufferString(`{
		"name": "new-backend",
		"type": "loadBalancer",
		"config": {
			"servers": [{"url": "http://localhost:3000"}]
		}
	}`)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/services", body)
	handler.CreateService(c)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var created map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &created)

	if created["id"] == nil || created["id"] == "" {
		t.Error("expected generated ID")
	}
	if created["name"] != "new-backend" {
		t.Errorf("expected name new-backend, got %v", created["name"])
	}
}

// TestServiceHandler_CreateService_InvalidType tests invalid service type
func TestServiceHandler_CreateService_InvalidType(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	body := bytes.NewBufferString(`{
		"name": "invalid-service",
		"type": "invalidType",
		"config": {}
	}`)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/services", body)
	handler.CreateService(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestServiceHandler_CreateService_ValidationError tests missing required fields
func TestServiceHandler_CreateService_ValidationError(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	// Missing name field
	body := bytes.NewBufferString(`{
		"type": "loadBalancer",
		"config": {}
	}`)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/services", body)
	handler.CreateService(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestServiceHandler_UpdateService tests updating a service
func TestServiceHandler_UpdateService(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	// Create service first
	testutil.MustExec(t, db, `
		INSERT INTO services (id, name, type, config, status, source_type)
		VALUES ('update-test', 'old-name', 'loadBalancer', '{}', 'active', 'manual')
	`)

	body := bytes.NewBufferString(`{
		"name": "updated-name",
		"type": "loadBalancer",
		"config": {"servers": [{"url": "http://new-backend:8080"}]}
	}`)

	c, rec := testutil.NewContext(t, http.MethodPut, "/api/services/update-test", body)
	c.Params = gin.Params{{Key: "id", Value: "update-test"}}
	handler.UpdateService(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var updated map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &updated)

	if updated["name"] != "updated-name" {
		t.Errorf("expected updated name, got %v", updated["name"])
	}
}

// TestServiceHandler_DeleteService tests deleting a service
func TestServiceHandler_DeleteService(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO services (id, name, type, config, status, source_type)
		VALUES ('delete-test', 'delete-me', 'loadBalancer', '{}', 'active', 'manual')
	`)

	c, rec := testutil.NewContext(t, http.MethodDelete, "/api/services/delete-test", nil)
	c.Params = gin.Params{{Key: "id", Value: "delete-test"}}
	handler.DeleteService(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify service is deleted
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM services WHERE id = 'delete-test'").Scan(&count)
	if count != 0 {
		t.Error("service was not deleted")
	}
}

// TestServiceHandler_DeleteService_NotFound tests deleting non-existent service
func TestServiceHandler_DeleteService_NotFound(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodDelete, "/api/services/non-existent", nil)
	c.Params = gin.Params{{Key: "id", Value: "non-existent"}}
	handler.DeleteService(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestServiceHandler_GetServices_Empty tests empty database
func TestServiceHandler_GetServices_Empty(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/services", nil)
	handler.GetServices(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var services []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &services)

	// Should return empty array, not null
	if services == nil {
		t.Error("expected empty array, got nil")
	}
}

// TestServiceHandler_GetServices_ConfigParsing tests config JSON parsing
func TestServiceHandler_GetServices_ConfigParsing(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO services (id, name, type, config, status, source_type)
		VALUES ('svc-config', 'config-test', 'loadBalancer', '{"servers":[{"url":"http://a:80"},{"url":"http://b:80"}],"healthCheck":{"interval":"10s"}}', 'active', 'manual')
	`)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/services", nil)
	handler.GetServices(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var services []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &services)

	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}

	config, ok := services[0]["config"].(map[string]interface{})
	if !ok {
		t.Fatal("config should be a map")
	}

	servers, ok := config["servers"].([]interface{})
	if !ok {
		t.Fatal("servers should be an array")
	}

	if len(servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(servers))
	}
}

// TestServiceHandler_CreateService_SetsSourceTypeManual tests source_type is set to manual
func TestServiceHandler_CreateService_SetsSourceTypeManual(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewServiceHandler(db.DB)

	body := bytes.NewBufferString(`{
		"name": "manual-service",
		"type": "loadBalancer",
		"config": {}
	}`)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/services", body)
	handler.CreateService(c)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	var created map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &created)

	// Query the database to verify source_type
	var sourceType string
	id := created["id"].(string)
	db.DB.QueryRow("SELECT source_type FROM services WHERE id = ?", id).Scan(&sourceType)

	if sourceType != "manual" {
		t.Errorf("expected source_type 'manual', got %q", sourceType)
	}
}
