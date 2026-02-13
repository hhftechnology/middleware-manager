package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

// TestNewSecurityHandler tests security handler creation
func TestNewSecurityHandler(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewSecurityHandler(db.DB, cm)

	if handler == nil {
		t.Fatal("NewSecurityHandler() returned nil")
	}
	if handler.DB == nil {
		t.Error("handler.DB is nil")
	}
	if handler.DuplicateDetector == nil {
		t.Error("handler.DuplicateDetector is nil")
	}
}

// TestSecurityHandler_GetConfig tests fetching security config
func TestSecurityHandler_GetConfig(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewSecurityHandler(db.DB, cm)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/security/config", nil)
	handler.GetConfig(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var config map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &config); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Should have security config fields
	if config["tls_hardening_enabled"] == nil {
		t.Error("expected tls_hardening_enabled in response")
	}
	if config["secure_headers_enabled"] == nil {
		t.Error("expected secure_headers_enabled in response")
	}
}

// TestSecurityHandler_GetConfig_WithExistingData tests fetching existing config
func TestSecurityHandler_GetConfig_WithExistingData(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewSecurityHandler(db.DB, cm)

	// Update the security config
	testutil.MustExec(t, db, `
		UPDATE security_config SET tls_hardening_enabled = 1, secure_headers_enabled = 1 WHERE id = 1
	`)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/security/config", nil)
	handler.GetConfig(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var config map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &config)

	if config["tls_hardening_enabled"] != true {
		t.Errorf("expected tls_hardening_enabled true, got %v", config["tls_hardening_enabled"])
	}
	if config["secure_headers_enabled"] != true {
		t.Errorf("expected secure_headers_enabled true, got %v", config["secure_headers_enabled"])
	}
}

// TestSecurityHandler_EnableTLSHardening tests enabling TLS hardening
func TestSecurityHandler_EnableTLSHardening(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewSecurityHandler(db.DB, cm)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/security/tls/enable", nil)
	handler.EnableTLSHardening(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["enabled"] != true {
		t.Errorf("expected enabled true, got %v", response["enabled"])
	}

	// Verify database was updated
	var enabled int
	db.DB.QueryRow("SELECT tls_hardening_enabled FROM security_config WHERE id = 1").Scan(&enabled)
	if enabled != 1 {
		t.Errorf("expected db tls_hardening_enabled 1, got %d", enabled)
	}
}

// TestSecurityHandler_DisableTLSHardening tests disabling TLS hardening
func TestSecurityHandler_DisableTLSHardening(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewSecurityHandler(db.DB, cm)

	// First enable it
	testutil.MustExec(t, db, `UPDATE security_config SET tls_hardening_enabled = 1 WHERE id = 1`)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/security/tls/disable", nil)
	handler.DisableTLSHardening(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["enabled"] != false {
		t.Errorf("expected enabled false, got %v", response["enabled"])
	}

	// Verify database was updated
	var enabled int
	db.DB.QueryRow("SELECT tls_hardening_enabled FROM security_config WHERE id = 1").Scan(&enabled)
	if enabled != 0 {
		t.Errorf("expected db tls_hardening_enabled 0, got %d", enabled)
	}
}

// TestSecurityHandler_EnableSecureHeaders tests enabling secure headers
func TestSecurityHandler_EnableSecureHeaders(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewSecurityHandler(db.DB, cm)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/security/headers/enable", nil)
	handler.EnableSecureHeaders(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["enabled"] != true {
		t.Errorf("expected enabled true, got %v", response["enabled"])
	}

	// Verify database was updated
	var enabled int
	db.DB.QueryRow("SELECT secure_headers_enabled FROM security_config WHERE id = 1").Scan(&enabled)
	if enabled != 1 {
		t.Errorf("expected db secure_headers_enabled 1, got %d", enabled)
	}
}

// TestSecurityHandler_DisableSecureHeaders tests disabling secure headers
func TestSecurityHandler_DisableSecureHeaders(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewSecurityHandler(db.DB, cm)

	// First enable it
	testutil.MustExec(t, db, `UPDATE security_config SET secure_headers_enabled = 1 WHERE id = 1`)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/security/headers/disable", nil)
	handler.DisableSecureHeaders(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["enabled"] != false {
		t.Errorf("expected enabled false, got %v", response["enabled"])
	}
}

// TestSecurityHandler_UpdateSecureHeadersConfig tests updating secure headers config
func TestSecurityHandler_UpdateSecureHeadersConfig(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewSecurityHandler(db.DB, cm)

	body := bytes.NewBufferString(`{
		"x_content_type_options": "nosniff",
		"x_frame_options": "DENY",
		"x_xss_protection": "1; mode=block",
		"hsts": "max-age=63072000; includeSubDomains; preload",
		"referrer_policy": "no-referrer",
		"csp": "default-src 'self'",
		"permissions_policy": "geolocation=()"
	}`)

	c, rec := testutil.NewContext(t, http.MethodPut, "/api/security/headers/config", body)
	handler.UpdateSecureHeadersConfig(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify database was updated
	var xFrameOptions string
	db.DB.QueryRow("SELECT secure_headers_x_frame_options FROM security_config WHERE id = 1").Scan(&xFrameOptions)
	if xFrameOptions != "DENY" {
		t.Errorf("expected x_frame_options 'DENY', got %q", xFrameOptions)
	}
}

// TestSecurityHandler_UpdateSecureHeadersConfig_InvalidJSON tests invalid request
func TestSecurityHandler_UpdateSecureHeadersConfig_InvalidJSON(t *testing.T) {
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	handler := NewSecurityHandler(db.DB, cm)

	body := bytes.NewBufferString(`{invalid}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/security/headers/config", body)
	handler.UpdateSecureHeadersConfig(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
