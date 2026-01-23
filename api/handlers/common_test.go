package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

// TestGenerateID tests ID generation
func TestGenerateID(t *testing.T) {
	// Generate multiple IDs and check uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := generateID()
		if err != nil {
			t.Fatalf("generateID() error = %v", err)
		}

		// Should be 16 characters (8 bytes = 16 hex chars)
		if len(id) != 16 {
			t.Errorf("generateID() length = %d, want 16", len(id))
		}

		// Should be unique
		if ids[id] {
			t.Errorf("generateID() generated duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

// TestIsValidMiddlewareType tests middleware type validation
func TestIsValidMiddlewareType(t *testing.T) {
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
		"chain",
		"plugin",
		"buffering",
		"circuitBreaker",
		"compress",
		"contentType",
		"errors",
		"grpcWeb",
		"inFlightReq",
		"passTLSClientCert",
		"retry",
	}

	for _, typ := range validTypes {
		t.Run("valid_"+typ, func(t *testing.T) {
			if !isValidMiddlewareType(typ) {
				t.Errorf("isValidMiddlewareType(%q) = false, want true", typ)
			}
		})
	}

	invalidTypes := []string{
		"invalid",
		"unknown",
		"",
		"HEADERS", // case sensitive
		"rate-limit",
		"basic_auth",
	}

	for _, typ := range invalidTypes {
		t.Run("invalid_"+typ, func(t *testing.T) {
			if isValidMiddlewareType(typ) {
				t.Errorf("isValidMiddlewareType(%q) = true, want false", typ)
			}
		})
	}
}

// TestSanitizeMiddlewareConfig tests config sanitization
func TestSanitizeMiddlewareConfig(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]interface{}
		check  func(t *testing.T, config map[string]interface{})
	}{
		{
			name: "remove extra quotes from duration",
			input: map[string]interface{}{
				"checkPeriod": "\"10s\"",
			},
			check: func(t *testing.T, config map[string]interface{}) {
				if config["checkPeriod"] != "10s" {
					t.Errorf("checkPeriod = %q, want \"10s\"", config["checkPeriod"])
				}
			},
		},
		{
			name: "nested config with duration",
			input: map[string]interface{}{
				"circuitBreaker": map[string]interface{}{
					"checkPeriod":      "\"5s\"",
					"fallbackDuration": "\"30s\"",
				},
			},
			check: func(t *testing.T, config map[string]interface{}) {
				nested := config["circuitBreaker"].(map[string]interface{})
				if nested["checkPeriod"] != "5s" {
					t.Errorf("checkPeriod = %q, want \"5s\"", nested["checkPeriod"])
				}
				if nested["fallbackDuration"] != "30s" {
					t.Errorf("fallbackDuration = %q, want \"30s\"", nested["fallbackDuration"])
				}
			},
		},
		{
			name: "array with quoted strings",
			input: map[string]interface{}{
				"items": []interface{}{"\"value1\"", "\"value2\""},
			},
			check: func(t *testing.T, config map[string]interface{}) {
				items := config["items"].([]interface{})
				if items[0] != "value1" {
					t.Errorf("items[0] = %q, want \"value1\"", items[0])
				}
			},
		},
		{
			name: "non-duration field unchanged",
			input: map[string]interface{}{
				"name":        "test",
				"checkPeriod": "10s", // Already clean
			},
			check: func(t *testing.T, config map[string]interface{}) {
				if config["name"] != "test" {
					t.Errorf("name = %q, want \"test\"", config["name"])
				}
				if config["checkPeriod"] != "10s" {
					t.Errorf("checkPeriod = %q, want \"10s\"", config["checkPeriod"])
				}
			},
		},
		{
			name:  "empty config",
			input: map[string]interface{}{},
			check: func(t *testing.T, config map[string]interface{}) {
				if len(config) != 0 {
					t.Errorf("config should be empty, got %v", config)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizeMiddlewareConfig(tt.input)
			tt.check(t, tt.input)
		})
	}
}

// TestResponseWithError tests error response formatting
func TestResponseWithError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
	}{
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			message:    "Invalid input",
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			message:    "Resource not found",
		},
		{
			name:       "internal error",
			statusCode: http.StatusInternalServerError,
			message:    "Database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := testutil.NewContext(t, http.MethodGet, "/test", nil)
			ResponseWithError(c, tt.statusCode, tt.message)

			if rec.Code != tt.statusCode {
				t.Errorf("status code = %d, want %d", rec.Code, tt.statusCode)
			}

			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)

			if response["error"] == nil {
				t.Error("response should contain 'error' field")
			}
		})
	}
}

// TestLogError tests error logging (mainly for coverage)
func TestLogError(t *testing.T) {
	// Should not panic with nil error
	LogError("test context", nil)

	// Should not panic with actual error
	LogError("test context", http.ErrBodyNotAllowed)
}
