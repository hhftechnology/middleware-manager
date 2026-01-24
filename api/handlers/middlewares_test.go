package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/database"
	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

// TestNewMiddlewareHandler tests middleware handler creation
func TestNewMiddlewareHandler(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	if handler == nil {
		t.Fatal("NewMiddlewareHandler() returned nil")
	}
	if handler.DB == nil {
		t.Error("handler.DB is nil")
	}
}

// TestMiddlewareHandler_GetMiddlewares tests fetching all middlewares
func TestMiddlewareHandler_GetMiddlewares(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	// Insert test middlewares
	testutil.MustExec(t, db, `
		INSERT INTO middlewares (id, name, type, config)
		VALUES ('mw-1', 'rate-limiter', 'rateLimit', '{"average":100,"burst":50}')
	`)
	testutil.MustExec(t, db, `
		INSERT INTO middlewares (id, name, type, config)
		VALUES ('mw-2', 'custom-headers', 'headers', '{"customRequestHeaders":{"X-Custom":"value"}}')
	`)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/middlewares", nil)
	handler.GetMiddlewares(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var middlewares []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &middlewares); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(middlewares) != 2 {
		t.Errorf("expected 2 middlewares, got %d", len(middlewares))
	}
}

// TestMiddlewareHandler_GetMiddlewares_Pagination tests paginated results
func TestMiddlewareHandler_GetMiddlewares_Pagination(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	// Insert 5 test middlewares
	for i := 1; i <= 5; i++ {
		testutil.MustExec(t, db, `
			INSERT INTO middlewares (id, name, type, config)
			VALUES (?, ?, 'headers', '{}')
		`, "mw-"+string(rune('0'+i)), "middleware-"+string(rune('0'+i)))
	}

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/middlewares?page=1&page_size=2", nil)
	handler.GetMiddlewares(c)

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

// TestMiddlewareHandler_GetMiddleware tests fetching a single middleware
func TestMiddlewareHandler_GetMiddleware(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO middlewares (id, name, type, config)
		VALUES ('test-mw-id', 'test-middleware', 'headers', '{"customRequestHeaders":{"X-Test":"1"}}')
	`)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/middlewares/test-mw-id", nil)
	c.Params = gin.Params{{Key: "id", Value: "test-mw-id"}}
	handler.GetMiddleware(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var middleware map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &middleware)

	if middleware["name"] != "test-middleware" {
		t.Errorf("expected name test-middleware, got %v", middleware["name"])
	}
	if middleware["type"] != "headers" {
		t.Errorf("expected type headers, got %v", middleware["type"])
	}
}

// TestMiddlewareHandler_GetMiddleware_NotFound tests fetching non-existent middleware
func TestMiddlewareHandler_GetMiddleware_NotFound(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/middlewares/non-existent", nil)
	c.Params = gin.Params{{Key: "id", Value: "non-existent"}}
	handler.GetMiddleware(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestMiddlewareHandler_GetMiddleware_EmptyID tests missing middleware ID
func TestMiddlewareHandler_GetMiddleware_EmptyID(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/middlewares/", nil)
	c.Params = gin.Params{{Key: "id", Value: ""}}
	handler.GetMiddleware(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestMiddlewareHandler_CreateMiddleware tests creating a new middleware
func TestMiddlewareHandler_CreateMiddleware(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	body := bytes.NewBufferString(`{
		"name": "new-middleware",
		"type": "headers",
		"config": {
			"customRequestHeaders": {"X-New": "value"}
		}
	}`)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/middlewares", body)
	handler.CreateMiddleware(c)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var created map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &created)

	if created["id"] == nil || created["id"] == "" {
		t.Error("expected generated ID")
	}
	if created["name"] != "new-middleware" {
		t.Errorf("expected name new-middleware, got %v", created["name"])
	}
}

// TestMiddlewareHandler_CreateMiddleware_InvalidType tests invalid middleware type
func TestMiddlewareHandler_CreateMiddleware_InvalidType(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	body := bytes.NewBufferString(`{
		"name": "invalid-middleware",
		"type": "invalidType",
		"config": {}
	}`)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/middlewares", body)
	handler.CreateMiddleware(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestMiddlewareHandler_CreateMiddleware_ValidationError tests missing required fields
func TestMiddlewareHandler_CreateMiddleware_ValidationError(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	// Missing name field
	body := bytes.NewBufferString(`{
		"type": "headers",
		"config": {}
	}`)

	c, rec := testutil.NewContext(t, http.MethodPost, "/api/middlewares", body)
	handler.CreateMiddleware(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestMiddlewareHandler_UpdateMiddleware tests updating a middleware
func TestMiddlewareHandler_UpdateMiddleware(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	// Create middleware first
	testutil.MustExec(t, db, `
		INSERT INTO middlewares (id, name, type, config)
		VALUES ('update-test', 'old-name', 'headers', '{}')
	`)

	body := bytes.NewBufferString(`{
		"name": "updated-name",
		"type": "headers",
		"config": {"customRequestHeaders": {"X-Updated": "true"}}
	}`)

	c, rec := testutil.NewContext(t, http.MethodPut, "/api/middlewares/update-test", body)
	c.Params = gin.Params{{Key: "id", Value: "update-test"}}
	handler.UpdateMiddleware(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var updated map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &updated)

	if updated["name"] != "updated-name" {
		t.Errorf("expected updated name, got %v", updated["name"])
	}
}

// TestMiddlewareHandler_DeleteMiddleware tests deleting a middleware
func TestMiddlewareHandler_DeleteMiddleware(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO middlewares (id, name, type, config)
		VALUES ('delete-test', 'delete-me', 'headers', '{}')
	`)

	c, rec := testutil.NewContext(t, http.MethodDelete, "/api/middlewares/delete-test", nil)
	c.Params = gin.Params{{Key: "id", Value: "delete-test"}}
	handler.DeleteMiddleware(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify middleware is deleted
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM middlewares WHERE id = 'delete-test'").Scan(&count)
	if count != 0 {
		t.Error("middleware was not deleted")
	}

	// Verify deleted_templates entry was created
	var templateCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM deleted_templates WHERE id = 'delete-test' AND type = 'middleware'").Scan(&templateCount)
	if templateCount != 1 {
		t.Error("deleted_templates entry was not created")
	}
}

// TestMiddlewareHandler_DeleteMiddleware_NotFound tests deleting non-existent middleware
func TestMiddlewareHandler_DeleteMiddleware_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "middleware-delete-notfound")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init temp db: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.RemoveAll(tmpDir)
	})

	handler := NewMiddlewareHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodDelete, "/api/middlewares/non-existent", nil)
	c.Params = gin.Params{{Key: "id", Value: "non-existent"}}
	handler.DeleteMiddleware(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestMiddlewareHandler_GetMiddlewares_Empty tests empty database
func TestMiddlewareHandler_GetMiddlewares_Empty(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/middlewares", nil)
	handler.GetMiddlewares(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var middlewares []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &middlewares)

	// Should return empty array, not null
	if middlewares == nil {
		t.Error("expected empty array, got nil")
	}
}

// TestMiddlewareHandler_GetMiddlewares_ConfigParsing tests config JSON parsing
func TestMiddlewareHandler_GetMiddlewares_ConfigParsing(t *testing.T) {
	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	testutil.MustExec(t, db, `
		INSERT INTO middlewares (id, name, type, config)
		VALUES ('mw-config', 'config-test', 'rateLimit', '{"average":100,"burst":50,"period":"1s"}')
	`)

	c, rec := testutil.NewContext(t, http.MethodGet, "/api/middlewares", nil)
	handler.GetMiddlewares(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var middlewares []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &middlewares)

	if len(middlewares) != 1 {
		t.Fatalf("expected 1 middleware, got %d", len(middlewares))
	}

	config, ok := middlewares[0]["config"].(map[string]interface{})
	if !ok {
		t.Fatal("config should be a map")
	}

	if config["average"] == nil {
		t.Error("config should contain 'average' field")
	}
	if config["burst"] == nil {
		t.Error("config should contain 'burst' field")
	}
}

// TestMiddlewareHandler_ValidMiddlewareTypes tests all valid middleware types
func TestMiddlewareHandler_ValidMiddlewareTypes(t *testing.T) {
	validTypes := []string{
		"basicAuth",
		"digestAuth",
		"forwardAuth",
		"ipAllowList",
		"rateLimit",
		"headers",
		"stripPrefix",
		"stripPrefixRegex",
		"addPrefix",
		"redirectRegex",
		"redirectScheme",
		"replacePath",
		"replacePathRegex",
		"buffering",
		"circuitBreaker",
		"compress",
		"contentType",
		"retry",
		"chain",
		"plugin",
		"errors",
		"grpcWeb",
		"inFlightReq",
		"passTLSClientCert",
	}

	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	for _, mwType := range validTypes {
		t.Run(mwType, func(t *testing.T) {
			body := bytes.NewBufferString(`{
				"name": "test-` + mwType + `",
				"type": "` + mwType + `",
				"config": {}
			}`)

			c, rec := testutil.NewContext(t, http.MethodPost, "/api/middlewares", body)
			handler.CreateMiddleware(c)

			if rec.Code != http.StatusCreated {
				t.Errorf("expected 201 for type %s, got %d: %s", mwType, rec.Code, rec.Body.String())
			}
		})
	}
}
