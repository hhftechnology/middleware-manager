package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
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

// TestResourceHandler_DeleteResource tests deleting a resource
func TestResourceHandler_DeleteResource(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, host, service_id, org_id, site_id, status, source_type)
		VALUES ('delete-test', 'delete.example.com', 'svc-1', 'org-1', 'site-1', 'disabled', 'manual')
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

	// Empty result is acceptable (nil or empty slice)
}

// TestResourceHandler_AssignExternalMiddleware tests assigning a Traefik-native middleware
func TestResourceHandler_AssignExternalMiddleware(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, pangolin_router_id, host, service_id, org_id, site_id, status, source_type)
		VALUES ('ext-res-1', 'pangolin-ext-1', 'ext.example.com', 'svc-1', 'org-1', 'site-1', 'active', 'pangolin')
	`)

	body := `{"middleware_name": "my-auth@file", "priority": 150, "provider": "file"}`
	c, rec := testutil.NewContext(t, http.MethodPost, "/api/resources/ext-res-1/external-middlewares", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "ext-res-1"}}
	handler.AssignExternalMiddleware(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["middleware_name"] != "my-auth@file" {
		t.Errorf("expected middleware_name 'my-auth@file', got %v", resp["middleware_name"])
	}
	if int(resp["priority"].(float64)) != 150 {
		t.Errorf("expected priority 150, got %v", resp["priority"])
	}
}

// TestResourceHandler_AssignExternalMiddleware_MissingName tests empty middleware name
func TestResourceHandler_AssignExternalMiddleware_MissingName(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, pangolin_router_id, host, service_id, org_id, site_id, status, source_type)
		VALUES ('ext-res-2', 'pangolin-ext-2', 'ext2.example.com', 'svc-1', 'org-1', 'site-1', 'active', 'pangolin')
	`)

	body := `{"priority": 100}`
	c, rec := testutil.NewContext(t, http.MethodPost, "/api/resources/ext-res-2/external-middlewares", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "ext-res-2"}}
	handler.AssignExternalMiddleware(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestResourceHandler_AssignExternalMiddleware_DisabledResource tests assigning to disabled resource
func TestResourceHandler_AssignExternalMiddleware_DisabledResource(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, pangolin_router_id, host, service_id, org_id, site_id, status, source_type)
		VALUES ('ext-res-3', 'pangolin-ext-3', 'ext3.example.com', 'svc-1', 'org-1', 'site-1', 'disabled', 'pangolin')
	`)

	body := `{"middleware_name": "test-mw@file", "priority": 100}`
	c, rec := testutil.NewContext(t, http.MethodPost, "/api/resources/ext-res-3/external-middlewares", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "ext-res-3"}}
	handler.AssignExternalMiddleware(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestResourceHandler_RemoveExternalMiddleware tests removing an external middleware
func TestResourceHandler_RemoveExternalMiddleware(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, pangolin_router_id, host, service_id, org_id, site_id, status, source_type)
		VALUES ('ext-res-4', 'pangolin-ext-4', 'ext4.example.com', 'svc-1', 'org-1', 'site-1', 'active', 'pangolin')
	`)
	testutil.MustExec(t, db, `
		INSERT INTO resource_external_middlewares (resource_id, middleware_name, priority, provider)
		VALUES ('ext-res-4', 'test-mw@file', 100, 'file')
	`)

	c, rec := testutil.NewContext(t, http.MethodDelete, "/api/resources/ext-res-4/external-middlewares/test-mw@file", nil)
	c.Params = gin.Params{
		{Key: "id", Value: "ext-res-4"},
		{Key: "name", Value: "test-mw@file"},
	}
	handler.RemoveExternalMiddleware(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify it was removed
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM resource_external_middlewares WHERE resource_id = 'ext-res-4'").Scan(&count)
	if count != 0 {
		t.Error("external middleware was not removed from database")
	}
}

// TestResourceHandler_RemoveExternalMiddleware_NotFound tests removing non-existent assignment
func TestResourceHandler_RemoveExternalMiddleware_NotFound(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodDelete, "/api/resources/ext-res-5/external-middlewares/nonexistent", nil)
	c.Params = gin.Params{
		{Key: "id", Value: "ext-res-5"},
		{Key: "name", Value: "nonexistent"},
	}
	handler.RemoveExternalMiddleware(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestResourceHandler_GetExternalMiddlewares tests listing external middlewares
func TestResourceHandler_GetExternalMiddlewares(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, pangolin_router_id, host, service_id, org_id, site_id, status, source_type)
		VALUES ('ext-res-6', 'pangolin-ext-6', 'ext6.example.com', 'svc-1', 'org-1', 'site-1', 'active', 'pangolin')
	`)
	testutil.MustExec(t, db, `
		INSERT INTO resource_external_middlewares (resource_id, middleware_name, priority, provider)
		VALUES ('ext-res-6', 'auth@file', 200, 'file')
	`)
	testutil.MustExec(t, db, `
		INSERT INTO resource_external_middlewares (resource_id, middleware_name, priority, provider)
		VALUES ('ext-res-6', 'rate-limit@docker', 100, 'docker')
	`)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/resources/ext-res-6/external-middlewares", nil)
	c.Params = gin.Params{{Key: "id", Value: "ext-res-6"}}
	handler.GetExternalMiddlewares(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var extMiddlewares []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &extMiddlewares); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(extMiddlewares) != 2 {
		t.Fatalf("expected 2 external middlewares, got %d", len(extMiddlewares))
	}

	// Should be sorted by priority DESC (200 first, then 100)
	if extMiddlewares[0]["middleware_name"] != "auth@file" {
		t.Errorf("expected first middleware to be 'auth@file', got %v", extMiddlewares[0]["middleware_name"])
	}
}

// TestResourceHandler_GetExternalMiddlewares_Empty tests empty result
func TestResourceHandler_GetExternalMiddlewares_Empty(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/resources/ext-res-empty/external-middlewares", nil)
	c.Params = gin.Params{{Key: "id", Value: "ext-res-empty"}}
	handler.GetExternalMiddlewares(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var extMiddlewares []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &extMiddlewares); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(extMiddlewares) != 0 {
		t.Errorf("expected 0 external middlewares, got %d", len(extMiddlewares))
	}
}

// TestResourceHandler_GetResource_IncludesExternalMiddlewares tests that GetResource includes external_middlewares
func TestResourceHandler_GetResource_IncludesExternalMiddlewares(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewResourceHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, pangolin_router_id, host, service_id, org_id, site_id, status, source_type, router_priority)
		VALUES ('ext-res-7', 'pangolin-ext-7', 'ext7.example.com', 'svc-1', 'org-1', 'site-1', 'active', 'pangolin', 100)
	`)
	testutil.MustExec(t, db, `
		INSERT INTO resource_external_middlewares (resource_id, middleware_name, priority, provider)
		VALUES ('ext-res-7', 'my-plugin@file', 150, 'file')
	`)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/resources/ext-res-7", nil)
	c.Params = gin.Params{{Key: "id", Value: "ext-res-7"}}
	handler.GetResource(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resource map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resource); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	extMws, ok := resource["external_middlewares"].(string)
	if !ok {
		t.Fatalf("external_middlewares field not found or not a string")
	}

	if extMws != "my-plugin@file:150:file" {
		t.Errorf("expected 'my-plugin@file:150:file', got '%s'", extMws)
	}
}
