package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestNewResourceHandler tests resource handler creation
func TestNewResourceHandler(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	if handler == nil {
		t.Fatal("NewResourceHandler() returned nil")
	}
	if handler.DB == nil {
		t.Error("handler.DB is nil")
	}
}

// TestResourceHandler_GetResources tests fetching all resources
func TestResourceHandler_GetResources(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	// Insert test resources
	testutil.MustExec(t, db, `
		INSERT INTO resources (id, pangolin_router_id, host, service_id, org_id, site_id, status, source_type, router_priority)
		VALUES ('res-1', 'pangolin-1', 'test1.example.com', 'svc-1', 'org-1', 'site-1', 'active', 'pangolin', 100)
	`)
	testutil.MustExec(t, db, `
		INSERT INTO resources (id, pangolin_router_id, host, service_id, org_id, site_id, status, source_type, router_priority)
		VALUES ('res-2', 'pangolin-2', 'test2.example.com', 'svc-2', 'org-1', 'site-1', 'active', 'traefik', 200)
	`)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/resources", nil)
	handler.GetResources(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resources []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resources); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resources))
	}
}

// TestResourceHandler_GetResources_StatusFilter tests filtering by status
func TestResourceHandler_GetResources_StatusFilter(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	// Insert test resources with different statuses
	testutil.MustExec(t, db, `
		INSERT INTO resources (id, host, service_id, org_id, site_id, status, source_type)
		VALUES ('res-active', 'active.example.com', 'svc-1', 'org-1', 'site-1', 'active', 'pangolin')
	`)
	testutil.MustExec(t, db, `
		INSERT INTO resources (id, host, service_id, org_id, site_id, status, source_type)
		VALUES ('res-disabled', 'disabled.example.com', 'svc-2', 'org-1', 'site-1', 'disabled', 'pangolin')
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
			path := "/api/resources"
			if tt.status != "" {
				path += "?status=" + tt.status
			}
			c, rec := testutil.NewContext(t, http.MethodGet, path, nil)
			handler.GetResources(c)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}

			var resources []map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &resources)

			if len(resources) != tt.expectedCount {
				t.Errorf("expected %d resources, got %d", tt.expectedCount, len(resources))
			}
		})
	}
}

// TestResourceHandler_GetResources_SourceTypeFilter tests filtering by source_type
func TestResourceHandler_GetResources_SourceTypeFilter(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, host, service_id, org_id, site_id, status, source_type)
		VALUES ('res-pangolin', 'pangolin.example.com', 'svc-1', 'org-1', 'site-1', 'active', 'pangolin')
	`)
	testutil.MustExec(t, db, `
		INSERT INTO resources (id, host, service_id, org_id, site_id, status, source_type)
		VALUES ('res-traefik', 'traefik.example.com', 'svc-2', 'org-1', 'site-1', 'active', 'traefik')
	`)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/resources?source_type=pangolin", nil)
	handler.GetResources(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resources []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resources)

	if len(resources) != 1 {
		t.Errorf("expected 1 pangolin resource, got %d", len(resources))
	}

	if len(resources) > 0 && resources[0]["source_type"] != "pangolin" {
		t.Errorf("expected source_type pangolin, got %v", resources[0]["source_type"])
	}
}

// TestResourceHandler_GetResources_Pagination tests paginated results
func TestResourceHandler_GetResources_Pagination(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	// Insert 5 test resources
	for i := 1; i <= 5; i++ {
		testutil.MustExec(t, db, `
			INSERT INTO resources (id, host, service_id, org_id, site_id, status, source_type)
			VALUES (?, ?, 'svc-1', 'org-1', 'site-1', 'active', 'pangolin')
		`, "res-"+string(rune('0'+i)), "test"+string(rune('0'+i))+".example.com")
	}

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/resources?page=1&page_size=2", nil)
	handler.GetResources(c)

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

// TestResourceHandler_GetResource tests fetching a single resource
func TestResourceHandler_GetResource(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, pangolin_router_id, host, service_id, org_id, site_id, status, source_type, router_priority)
		VALUES ('test-res-id', 'pangolin-123', 'test.example.com', 'test-service', 'org-1', 'site-1', 'active', 'pangolin', 150)
	`)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/resources/test-res-id", nil)
	c.Params = gin.Params{{Key: "id", Value: "test-res-id"}}
	handler.GetResource(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resource map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resource)

	if resource["host"] != "test.example.com" {
		t.Errorf("expected host test.example.com, got %v", resource["host"])
	}
	if resource["pangolin_router_id"] != "pangolin-123" {
		t.Errorf("expected pangolin_router_id pangolin-123, got %v", resource["pangolin_router_id"])
	}
}

// TestResourceHandler_GetResource_NotFound tests fetching non-existent resource
func TestResourceHandler_GetResource_NotFound(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/resources/non-existent", nil)
	c.Params = gin.Params{{Key: "id", Value: "non-existent"}}
	handler.GetResource(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestResourceHandler_GetResource_EmptyID tests missing resource ID
func TestResourceHandler_GetResource_EmptyID(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/resources/", nil)
	c.Params = gin.Params{{Key: "id", Value: ""}}
	handler.GetResource(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestResourceHandler_CreateResource tests creating a new resource
func TestResourceHandler_CreateResource(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	body := bytes.NewBufferString(`{
		"host": "new.example.com",
		"service_id": "new-service",
		"org_id": "org-1",
		"site_id": "site-1"
	}`)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/resources", body)
	handler.CreateResource(c)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var created map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &created)

	if created["id"] == nil || created["id"] == "" {
		t.Error("expected generated ID")
	}
	if created["host"] != "new.example.com" {
		t.Errorf("expected host new.example.com, got %v", created["host"])
	}
}

// TestResourceHandler_CreateResource_ValidationError tests validation
func TestResourceHandler_CreateResource_ValidationError(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	// Missing required field
	body := bytes.NewBufferString(`{"service_id": "svc-1"}`)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/resources", body)
	handler.CreateResource(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestResourceHandler_UpdateResource tests updating a resource
func TestResourceHandler_UpdateResource(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	// Create resource first
	testutil.MustExec(t, db, `
		INSERT INTO resources (id, host, service_id, org_id, site_id, status, source_type, router_priority)
		VALUES ('update-test', 'old.example.com', 'old-service', 'org-1', 'site-1', 'active', 'manual', 100)
	`)

	body := bytes.NewBufferString(`{
		"host": "updated.example.com",
		"service_id": "updated-service",
		"router_priority": 200
	}`)

	c, rec := testutil.NewContext(t, http.MethodPut, "/api/resources/update-test", body)
	c.Params = gin.Params{{Key: "id", Value: "update-test"}}
	handler.UpdateResource(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var updated map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &updated)

	if updated["host"] != "updated.example.com" {
		t.Errorf("expected updated host, got %v", updated["host"])
	}
}

// TestResourceHandler_DeleteResource tests deleting a resource
func TestResourceHandler_DeleteResource(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, host, service_id, org_id, site_id, status, source_type)
		VALUES ('delete-test', 'delete.example.com', 'svc-1', 'org-1', 'site-1', 'active', 'manual')
	`)

	c, rec := testutil.NewContext(t, http.MethodDelete, "/api/resources/delete-test", nil)
	c.Params = gin.Params{{Key: "id", Value: "delete-test"}}
	handler.DeleteResource(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify resource is deleted
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM resources WHERE id = 'delete-test'").Scan(&count)
	if count != 0 {
		t.Error("resource was not deleted")
	}
}

// TestResourceHandler_DeleteResource_NotFound tests deleting non-existent resource
func TestResourceHandler_DeleteResource_NotFound(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodDelete, "/api/resources/non-existent", nil)
	c.Params = gin.Params{{Key: "id", Value: "non-existent"}}
	handler.DeleteResource(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestResourceHandler_GetResources_Empty tests empty database
func TestResourceHandler_GetResources_Empty(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/resources", nil)
	handler.GetResources(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resources []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resources)

	// Should return empty array, not null
	if resources == nil {
		t.Error("expected empty array, got nil")
	}
}
