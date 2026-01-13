package models

import (
	"time"
)

// Resource represents a Pangolin resource
type Resource struct {
	ID             string    `json:"id"`
	Host           string    `json:"host"`
	ServiceID      string    `json:"service_id"`
	OrgID          string    `json:"org_id"`
	SiteID         string    `json:"site_id"`
	Status         string    `json:"status"`
	
	// HTTP router configuration
	Entrypoints    string    `json:"entrypoints"`
	
	// TLS certificate configuration
	TLSDomains     string    `json:"tls_domains"`
	
	// TCP SNI routing configuration
	TCPEnabled     bool      `json:"tcp_enabled"`
	TCPEntrypoints string    `json:"tcp_entrypoints"`
	TCPSNIRule     string    `json:"tcp_sni_rule"`
	
	// Custom headers configuration
	CustomHeaders  string    `json:"custom_headers"`
	
	// Router priority configuration
	RouterPriority int       `json:"router_priority"`

	// Source type for tracking data origin
	SourceType     string    `json:"source_type"`

	// mTLS configuration
	MTLSEnabled    bool      `json:"mtls_enabled"`

	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PangolinResource represents the format of a resource from Pangolin API
type PangolinResource struct {
	ID     string `json:"id"`
	Host   string `json:"host"`
	OrgID  string `json:"org_id"`
	SiteID string `json:"site_id"`
}

// PangolinTraefikConfig represents the Traefik configuration from Pangolin API
type PangolinTraefikConfig struct {
	HTTP struct {
		Routers     map[string]PangolinRouter         `json:"routers"`
		Services    map[string]PangolinService        `json:"services"`
		Middlewares map[string]map[string]interface{} `json:"middlewares"`
	} `json:"http"`
}

// PangolinRouter represents a router configuration from Pangolin API
type PangolinRouter struct {
	Rule        string   `json:"rule"`
	Service     string   `json:"service"`
	EntryPoints []string `json:"entryPoints"`
	Middlewares []string `json:"middlewares"`
	TLS         struct {
		CertResolver string `json:"certResolver"`
	} `json:"tls"`
}

type PangolinService struct {
	LoadBalancer interface{} `json:"loadBalancer,omitempty"`
	Weighted     interface{} `json:"weighted,omitempty"`
	Mirroring    interface{} `json:"mirroring,omitempty"`
	Failover     interface{} `json:"failover,omitempty"`
}

// PangolinServiceConfig represents a service configuration from Pangolin API
type PangolinServiceConfig struct {
	LoadBalancer struct {
		Servers []struct {
			URL string `json:"url"`
		} `json:"servers"`
	} `json:"loadBalancer"`
}

// TraefikService represents a service configuration from Traefik API
type TraefikService struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	
	// Service types - only one will be populated based on service type
	LoadBalancer *struct {
		Servers []struct {
			URL     string `json:"url,omitempty"`
			Address string `json:"address,omitempty"`
			Weight  *int   `json:"weight,omitempty"`
		} `json:"servers,omitempty"`
		PassHostHeader *bool       `json:"passHostHeader,omitempty"`
		Sticky         interface{} `json:"sticky,omitempty"`
		HealthCheck    interface{} `json:"healthCheck,omitempty"`
	} `json:"loadBalancer,omitempty"`
	
	Weighted *struct {
		Services []struct {
			Name   string `json:"name"`
			Weight int    `json:"weight"`
		} `json:"services,omitempty"`
		Sticky      interface{} `json:"sticky,omitempty"`
		HealthCheck interface{} `json:"healthCheck,omitempty"`
	} `json:"weighted,omitempty"`
	
	Mirroring *struct {
		Service     string `json:"service"`
		Mirrors     []struct {
			Name    string `json:"name"`
			Percent int    `json:"percent"`
		} `json:"mirrors,omitempty"`
		MaxBodySize *int        `json:"maxBodySize,omitempty"`
		MirrorBody  *bool       `json:"mirrorBody,omitempty"`
		HealthCheck interface{} `json:"healthCheck,omitempty"`
	} `json:"mirroring,omitempty"`
	
	Failover *struct {
		Service     string      `json:"service"`
		Fallback    string      `json:"fallback"`
		HealthCheck interface{} `json:"healthCheck,omitempty"`
	} `json:"failover,omitempty"`
}