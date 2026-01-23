package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

// TestNewMTLSHandler tests mTLS handler creation
func TestNewMTLSHandler(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	if handler == nil {
		t.Fatal("NewMTLSHandler() returned nil")
	}
	if handler.DB == nil {
		t.Error("handler.DB is nil")
	}
	if handler.CertGenerator == nil {
		t.Error("handler.CertGenerator is nil")
	}
}

// TestMTLSHandler_SetTraefikConfigPath tests setting config path
func TestMTLSHandler_SetTraefikConfigPath(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	testPath := "/etc/traefik/traefik.yml"
	handler.SetTraefikConfigPath(testPath)

	if handler.TraefikStaticConfigPath != testPath {
		t.Errorf("TraefikStaticConfigPath = %q, want %q", handler.TraefikStaticConfigPath, testPath)
	}
}

// TestMTLSHandler_GetConfig tests fetching mTLS config
func TestMTLSHandler_GetConfig(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/mtls/config", nil)
	handler.GetConfig(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var config map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &config); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Should have mTLS config fields
	if config["enabled"] == nil {
		t.Error("expected 'enabled' in response")
	}
	if config["client_count"] == nil {
		t.Error("expected 'client_count' in response")
	}
}

// TestMTLSHandler_EnableMTLS_NoCA tests enabling mTLS without a CA
func TestMTLSHandler_EnableMTLS_NoCA(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/mtls/enable", nil)
	handler.EnableMTLS(c)

	// Should fail without a CA configured
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 (no CA), got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestMTLSHandler_DisableMTLS tests disabling mTLS
func TestMTLSHandler_DisableMTLS(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/mtls/disable", nil)
	handler.DisableMTLS(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["enabled"] != false {
		t.Errorf("expected enabled false, got %v", response["enabled"])
	}
}

// TestMTLSHandler_CreateCA_InvalidRequest tests invalid CA creation request
func TestMTLSHandler_CreateCA_InvalidRequest(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	body := bytes.NewBufferString(`{invalid}`)
	c, rec := testutil.NewContext(t, http.MethodPost, "/api/mtls/ca", body)
	handler.CreateCA(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestMTLSHandler_GetClients tests fetching client certificates
func TestMTLSHandler_GetClients(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/mtls/clients", nil)
	handler.GetClients(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var clients []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &clients); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Should return empty array
	if clients == nil {
		t.Error("expected empty array, got nil")
	}
}

// TestMTLSHandler_GetClient_NotFound tests fetching non-existent client
func TestMTLSHandler_GetClient_NotFound(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/mtls/clients/nonexistent", nil)
	c.Params = gin.Params{{Key: "id", Value: "nonexistent"}}
	handler.GetClient(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestMTLSHandler_CreateClient_NoCA tests creating client without CA
func TestMTLSHandler_CreateClient_NoCA(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	body := bytes.NewBufferString(`{"name": "test-client"}`)
	c, rec := testutil.NewContext(t, http.MethodPost, "/api/mtls/clients", body)
	handler.CreateClient(c)

	// Should fail without CA
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 (no CA), got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestMTLSHandler_CreateClient_InvalidRequest tests invalid client creation
func TestMTLSHandler_CreateClient_InvalidRequest(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	body := bytes.NewBufferString(`{invalid}`)
	c, rec := testutil.NewContext(t, http.MethodPost, "/api/mtls/clients", body)
	handler.CreateClient(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestMTLSHandler_DeleteClient_NotFound tests deleting non-existent client
func TestMTLSHandler_DeleteClient_NotFound(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodDelete, "/api/mtls/clients/nonexistent", nil)
	c.Params = gin.Params{{Key: "id", Value: "nonexistent"}}
	handler.DeleteClient(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestMTLSHandler_RevokeClient_NotFound tests revoking non-existent client
func TestMTLSHandler_RevokeClient_NotFound(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/mtls/clients/nonexistent/revoke", nil)
	c.Params = gin.Params{{Key: "id", Value: "nonexistent"}}
	handler.RevokeClient(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestMTLSHandler_UpdateMiddlewareConfig tests updating middleware config
func TestMTLSHandler_UpdateMiddlewareConfig(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	body := bytes.NewBufferString(`{
		"rules": "*.example.com",
		"request_headers": "X-Client-CN,X-Client-Serial",
		"reject_message": "Access denied",
		"refresh_interval": 300
	}`)

	c, rec := testutil.NewContext(t, http.MethodPut, "/api/mtls/middleware", body)
	handler.UpdateMiddlewareConfig(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify database was updated
	var rules string
	db.DB.QueryRow("SELECT middleware_rules FROM mtls_config WHERE id = 1").Scan(&rules)
	if rules != "*.example.com" {
		t.Errorf("expected rules '*.example.com', got %q", rules)
	}
}

// TestMTLSHandler_UpdateMiddlewareConfig_InvalidJSON tests invalid request
func TestMTLSHandler_UpdateMiddlewareConfig_InvalidJSON(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	body := bytes.NewBufferString(`{invalid}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/mtls/middleware", body)
	handler.UpdateMiddlewareConfig(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestMTLSHandler_GetCACert_NoCA tests getting CA cert when none exists
func TestMTLSHandler_GetCACert_NoCA(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMTLSHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/mtls/ca/cert", nil)
	handler.GetCACert(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 (no CA), got %d", rec.Code)
	}
}
