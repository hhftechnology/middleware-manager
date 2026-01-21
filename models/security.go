package models

import (
	"time"
)

// SecurityConfig represents the global security configuration (singleton)
type SecurityConfig struct {
	ID                   int       `json:"id"`
	TLSHardeningEnabled  bool      `json:"tls_hardening_enabled"`
	SecureHeadersEnabled bool      `json:"secure_headers_enabled"`
	SecureHeaders        SecureHeadersConfig `json:"secure_headers"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// SecureHeadersConfig represents configurable security headers
type SecureHeadersConfig struct {
	XContentTypeOptions string `json:"x_content_type_options"`
	XFrameOptions       string `json:"x_frame_options"`
	XXSSProtection      string `json:"x_xss_protection"`
	HSTS                string `json:"hsts"`
	ReferrerPolicy      string `json:"referrer_policy"`
	CSP                 string `json:"csp"`
	PermissionsPolicy   string `json:"permissions_policy"`
}

// DefaultSecureHeaders returns the default security headers configuration
func DefaultSecureHeaders() SecureHeadersConfig {
	return SecureHeadersConfig{
		XContentTypeOptions: "nosniff",
		XFrameOptions:       "SAMEORIGIN",
		XXSSProtection:      "1; mode=block",
		HSTS:                "max-age=31536000; includeSubDomains",
		ReferrerPolicy:      "strict-origin-when-cross-origin",
		CSP:                 "",
		PermissionsPolicy:   "",
	}
}

// DuplicateCheckResult represents the result of middleware duplicate detection
type DuplicateCheckResult struct {
	HasDuplicates  bool       `json:"has_duplicates"`
	Duplicates     []Duplicate `json:"duplicates"`
	APIAvailable   bool       `json:"api_available"`
	WarningMessage string     `json:"warning_message,omitempty"`
}

// Duplicate represents a detected duplicate middleware
type Duplicate struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Type     string `json:"type"`
}

// DuplicateCheckRequest represents the request to check for duplicates
type DuplicateCheckRequest struct {
	Name       string `json:"name" binding:"required"`
	PluginName string `json:"plugin_name,omitempty"`
}

// UpdateSecurityConfigRequest represents request to update security config
type UpdateSecurityConfigRequest struct {
	TLSHardeningEnabled  *bool                `json:"tls_hardening_enabled,omitempty"`
	SecureHeadersEnabled *bool                `json:"secure_headers_enabled,omitempty"`
	SecureHeaders        *SecureHeadersConfig `json:"secure_headers,omitempty"`
}

// UpdateResourceSecurityRequest represents request to update per-resource security settings
type UpdateResourceSecurityRequest struct {
	Enabled bool `json:"enabled"`
}

// TLSHardeningOptions returns the TLS options for hardened security
func TLSHardeningOptions() map[string]interface{} {
	return map[string]interface{}{
		"minVersion": "VersionTLS12",
		"maxVersion": "VersionTLS13",
		"sniStrict":  true,
		"cipherSuites": []string{
			"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
			"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
			"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
			"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
		},
		"curvePreferences": []string{
			"X25519",
			"CurveP384",
			"CurveP521",
		},
	}
}
