package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

// TestNewDataSourceHandler tests data source handler creation
func TestNewDataSourceHandler(t *testing.T) {
	cm := testutil.NewTestConfigManager(t)
	handler := NewDataSourceHandler(cm)

	if handler == nil {
		t.Fatal("NewDataSourceHandler() returned nil")
	}
	if handler.ConfigManager == nil {
		t.Error("handler.ConfigManager is nil")
	}
}

// TestDataSourceHandler_GetDataSources tests fetching all data sources
func TestDataSourceHandler_GetDataSources(t *testing.T) {
	cm := testutil.NewTestConfigManager(t)
	handler := NewDataSourceHandler(cm)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/datasource", nil)
	handler.GetDataSources(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Should have active_source and sources
	if response["active_source"] == nil {
		t.Error("expected active_source in response")
	}
	if response["sources"] == nil {
		t.Error("expected sources in response")
	}
}

// TestDataSourceHandler_GetActiveDataSource tests fetching active data source
func TestDataSourceHandler_GetActiveDataSource(t *testing.T) {
	cm := testutil.NewTestConfigManager(t)
	handler := NewDataSourceHandler(cm)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/datasource/active", nil)
	handler.GetActiveDataSource(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["name"] == nil {
		t.Error("expected name in response")
	}
	if response["config"] == nil {
		t.Error("expected config in response")
	}
}

// TestDataSourceHandler_SetActiveDataSource tests setting active data source
func TestDataSourceHandler_SetActiveDataSource(t *testing.T) {
	cm := testutil.NewTestConfigManager(t)
	handler := NewDataSourceHandler(cm)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "switch to traefik",
			body:       `{"name": "traefik"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "switch to pangolin",
			body:       `{"name": "pangolin"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid source",
			body:       `{"name": "invalid-source"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing name",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := testutil.NewContext(t, http.MethodPut, "/api/datasource/active", bytes.NewBufferString(tt.body))
			handler.SetActiveDataSource(c)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

// TestDataSourceHandler_SetActiveDataSource_Response tests response structure
func TestDataSourceHandler_SetActiveDataSource_Response(t *testing.T) {
	cm := testutil.NewTestConfigManager(t)
	handler := NewDataSourceHandler(cm)

	body := bytes.NewBufferString(`{"name": "traefik"}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/datasource/active", body)
	handler.SetActiveDataSource(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["message"] == nil {
		t.Error("expected message in response")
	}
	if response["name"] != "traefik" {
		t.Errorf("expected name traefik, got %v", response["name"])
	}
	if response["refresh_needed"] == nil {
		t.Error("expected refresh_needed in response")
	}
}

// TestDataSourceHandler_UpdateDataSource tests updating a data source
func TestDataSourceHandler_UpdateDataSource(t *testing.T) {
	cm := testutil.NewTestConfigManager(t)
	handler := NewDataSourceHandler(cm)

	body := bytes.NewBufferString(`{
		"type": "traefik",
		"url": "http://new-traefik:8080",
		"skip_tls_verify": true
	}`)

	c, rec := testutil.NewContext(t, http.MethodPut, "/api/datasource/traefik", body)
	c.Params = gin.Params{{Key: "name", Value: "traefik"}}
	handler.UpdateDataSource(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["message"] == nil {
		t.Error("expected message in response")
	}
}

// TestDataSourceHandler_UpdateDataSource_EmptyName tests missing name parameter
func TestDataSourceHandler_UpdateDataSource_EmptyName(t *testing.T) {
	cm := testutil.NewTestConfigManager(t)
	handler := NewDataSourceHandler(cm)

	body := bytes.NewBufferString(`{"type": "traefik", "url": "http://localhost:8080"}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/datasource/", body)
	c.Params = gin.Params{{Key: "name", Value: ""}}
	handler.UpdateDataSource(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestDataSourceHandler_UpdateDataSource_InvalidJSON tests invalid JSON body
func TestDataSourceHandler_UpdateDataSource_InvalidJSON(t *testing.T) {
	cm := testutil.NewTestConfigManager(t)
	handler := NewDataSourceHandler(cm)

	body := bytes.NewBufferString(`{invalid json}`)
	c, rec := testutil.NewContext(t, http.MethodPut, "/api/datasource/traefik", body)
	c.Params = gin.Params{{Key: "name", Value: "traefik"}}
	handler.UpdateDataSource(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestDataSourceHandler_TestDataSourceConnection_EmptyName tests missing name
func TestDataSourceHandler_TestDataSourceConnection_EmptyName(t *testing.T) {
	cm := testutil.NewTestConfigManager(t)
	handler := NewDataSourceHandler(cm)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/datasource/test/", nil)
	c.Params = gin.Params{{Key: "name", Value: ""}}
	handler.TestDataSourceConnection(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestDataSourceHandler_TestDataSourceConnection_NotFound tests non-existent source
func TestDataSourceHandler_TestDataSourceConnection_NotFound(t *testing.T) {
	cm := testutil.NewTestConfigManager(t)
	handler := NewDataSourceHandler(cm)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/datasource/test/nonexistent", nil)
	c.Params = gin.Params{{Key: "name", Value: "nonexistent"}}
	handler.TestDataSourceConnection(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
