package models

import (
	"time"
)

// MTLSConfig represents the global mTLS configuration (singleton)
type MTLSConfig struct {
	ID            int        `json:"id"`
	Enabled       bool       `json:"enabled"`
	CACert        string     `json:"ca_cert,omitempty"`
	CAKey         string     `json:"-"` // Never expose private key via API
	CACertPath    string     `json:"ca_cert_path"`
	CASubject     string     `json:"ca_subject"`
	CAExpiry      *time.Time `json:"ca_expiry,omitempty"`
	CertsBasePath string     `json:"certs_base_path"`
	HasCA         bool       `json:"has_ca"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// MTLSClient represents a client certificate
type MTLSClient struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Cert            string     `json:"cert,omitempty"`
	Key             string     `json:"-"` // Never expose private key via API
	P12             []byte     `json:"-"` // Binary data, downloaded separately
	P12PasswordHint string     `json:"p12_password_hint,omitempty"`
	Subject         string     `json:"subject"`
	Expiry          *time.Time `json:"expiry,omitempty"`
	Revoked         bool       `json:"revoked"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CreateCARequest represents the request to create a new Certificate Authority
type CreateCARequest struct {
	CommonName   string `json:"common_name" binding:"required"`
	Organization string `json:"organization"`
	Country      string `json:"country"`
	ValidityDays int    `json:"validity_days"` // Default: 1825 (5 years)
}

// CreateClientRequest represents the request to create a new client certificate
type CreateClientRequest struct {
	Name         string `json:"name" binding:"required"`
	ValidityDays int    `json:"validity_days"` // Default: 730 (2 years)
	P12Password  string `json:"p12_password" binding:"required"`
}

// UpdateMTLSConfigRequest represents the request to update resource mTLS settings
type UpdateMTLSConfigRequest struct {
	MTLSEnabled bool `json:"mtls_enabled"`
}

// MTLSConfigResponse is the API response for mTLS configuration
type MTLSConfigResponse struct {
	MTLSConfig
	ClientCount int `json:"client_count"`
}

// MTLSMiddlewareConfig represents the mtlswhitelist plugin configuration
type MTLSMiddlewareConfig struct {
	// Rules for certificate validation (JSON array of rules)
	Rules string `json:"rules"`
	// RequestHeaders to add to requests with cert info (JSON object)
	RequestHeaders string `json:"request_headers"`
	// RejectMessage to return when certificate validation fails
	RejectMessage string `json:"reject_message"`
	// RefreshInterval in seconds for external data
	RefreshInterval int `json:"refresh_interval"`
}
