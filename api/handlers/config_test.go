package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

// TestNewConfigHandler tests config handler creation
func TestNewConfigHandler(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewConfigHandler(db.DB)

	if handler == nil {
		t.Fatal("NewConfigHandler() returned nil")
	}
	if handler.DB == nil {
		t.Error("handler.DB is nil")
	}
}

// TestConfigHandler_UpdateRouterPriority tests updating router priority
func TestConfigHandler_UpdateRouterPriority(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewConfigHandler(db.DB)

	// Create a test resource
	testutil.MustExec(t, db, `
		INSERT INTO resources (id, host, service_id, org_id, site_id, status, router_priority, router_priority_manual)
		VALUES ('test-res', 'test.example.com', 'svc-1', 'org-1', 'site-1', 'active', 100, 0)
	`)

	body := bytes.NewBufferString(`{"router_priority": 500}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/config/resources/test-res/priority", body)
	c.Params = gin.Params{{Key: "id", Value: "test-res"}}
	handler.UpdateRouterPriority(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["router_priority"] != float64(500) {
		t.Errorf("expected router_priority 500, got %v", response["router_priority"])
	}

	// Verify the database was updated and manual flag is set
	var priority, manual int
	db.DB.QueryRow("SELECT router_priority, router_priority_manual FROM resources WHERE id = 'test-res'").Scan(&priority, &manual)

	if priority != 500 {
		t.Errorf("expected db router_priority 500, got %d", priority)
	}
	if manual != 1 {
		t.Errorf("expected router_priority_manual 1, got %d", manual)
	}
}

// TestConfigHandler_UpdateRouterPriority_NotFound tests non-existent resource
func TestConfigHandler_UpdateRouterPriority_NotFound(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewConfigHandler(db.DB)

	body := bytes.NewBufferString(`{"router_priority": 100}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/config/resources/nonexistent/priority", body)
	c.Params = gin.Params{{Key: "id", Value: "nonexistent"}}
	handler.UpdateRouterPriority(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestConfigHandler_UpdateRouterPriority_DisabledResource tests updating disabled resource
func TestConfigHandler_UpdateRouterPriority_DisabledResource(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewConfigHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, host, service_id, org_id, site_id, status, router_priority)
		VALUES ('disabled-res', 'disabled.example.com', 'svc-1', 'org-1', 'site-1', 'disabled', 100)
	`)

	body := bytes.NewBufferString(`{"router_priority": 200}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/config/resources/disabled-res/priority", body)
	c.Params = gin.Params{{Key: "id", Value: "disabled-res"}}
	handler.UpdateRouterPriority(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestConfigHandler_UpdateRouterPriority_EmptyID tests missing resource ID
func TestConfigHandler_UpdateRouterPriority_EmptyID(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewConfigHandler(db.DB)

	body := bytes.NewBufferString(`{"router_priority": 100}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/config/resources//priority", body)
	c.Params = gin.Params{{Key: "id", Value: ""}}
	handler.UpdateRouterPriority(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestConfigHandler_UpdateRouterPriority_InvalidJSON tests invalid request body
func TestConfigHandler_UpdateRouterPriority_InvalidJSON(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewConfigHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, host, service_id, org_id, site_id, status)
		VALUES ('test-res', 'test.example.com', 'svc-1', 'org-1', 'site-1', 'active')
	`)

	body := bytes.NewBufferString(`{invalid}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/config/resources/test-res/priority", body)
	c.Params = gin.Params{{Key: "id", Value: "test-res"}}
	handler.UpdateRouterPriority(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestConfigHandler_UpdateRouterPriority_MissingPriority tests missing required field
func TestConfigHandler_UpdateRouterPriority_MissingPriority(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewConfigHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, host, service_id, org_id, site_id, status)
		VALUES ('test-res', 'test.example.com', 'svc-1', 'org-1', 'site-1', 'active')
	`)

	body := bytes.NewBufferString(`{}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/config/resources/test-res/priority", body)
	c.Params = gin.Params{{Key: "id", Value: "test-res"}}
	handler.UpdateRouterPriority(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestConfigHandler_UpdateHTTPConfig tests updating HTTP config
func TestConfigHandler_UpdateHTTPConfig(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewConfigHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, host, service_id, org_id, site_id, status, entrypoints)
		VALUES ('test-res', 'test.example.com', 'svc-1', 'org-1', 'site-1', 'active', 'websecure')
	`)

	body := bytes.NewBufferString(`{"entrypoints": "web,websecure"}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/config/resources/test-res/http", body)
	c.Params = gin.Params{{Key: "id", Value: "test-res"}}
	handler.UpdateHTTPConfig(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify database was updated
	var entrypoints string
	db.DB.QueryRow("SELECT entrypoints FROM resources WHERE id = 'test-res'").Scan(&entrypoints)

	if entrypoints != "web,websecure" {
		t.Errorf("expected entrypoints 'web,websecure', got %q", entrypoints)
	}
}

// TestConfigHandler_UpdateHTTPConfig_DefaultEntrypoints tests default entrypoints
func TestConfigHandler_UpdateHTTPConfig_DefaultEntrypoints(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewConfigHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, host, service_id, org_id, site_id, status, entrypoints)
		VALUES ('test-res', 'test.example.com', 'svc-1', 'org-1', 'site-1', 'active', 'custom')
	`)

	// Empty entrypoints should default to websecure
	body := bytes.NewBufferString(`{"entrypoints": ""}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/config/resources/test-res/http", body)
	c.Params = gin.Params{{Key: "id", Value: "test-res"}}
	handler.UpdateHTTPConfig(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var entrypoints string
	db.DB.QueryRow("SELECT entrypoints FROM resources WHERE id = 'test-res'").Scan(&entrypoints)

	if entrypoints != "websecure" {
		t.Errorf("expected default entrypoints 'websecure', got %q", entrypoints)
	}
}

// TestConfigHandler_UpdateHTTPConfig_NotFound tests non-existent resource
func TestConfigHandler_UpdateHTTPConfig_NotFound(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewConfigHandler(db.DB)

	body := bytes.NewBufferString(`{"entrypoints": "web"}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/config/resources/nonexistent/http", body)
	c.Params = gin.Params{{Key: "id", Value: "nonexistent"}}
	handler.UpdateHTTPConfig(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestConfigHandler_UpdateHTTPConfig_DisabledResource tests updating disabled resource
func TestConfigHandler_UpdateHTTPConfig_DisabledResource(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewConfigHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO resources (id, host, service_id, org_id, site_id, status)
		VALUES ('disabled-res', 'disabled.example.com', 'svc-1', 'org-1', 'site-1', 'disabled')
	`)

	body := bytes.NewBufferString(`{"entrypoints": "web"}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/config/resources/disabled-res/http", body)
	c.Params = gin.Params{{Key: "id", Value: "disabled-res"}}
	handler.UpdateHTTPConfig(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
