-- Middlewares table stores middleware definitions
CREATE TABLE IF NOT EXISTS middlewares (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    config TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Resources table stores Pangolin resources
-- Includes all configuration columns including the router_priority column
CREATE TABLE IF NOT EXISTS resources (
    id TEXT PRIMARY KEY,
    host TEXT NOT NULL,
    service_id TEXT NOT NULL,
    org_id TEXT NOT NULL,
    site_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    
    -- HTTP router configuration
    entrypoints TEXT DEFAULT 'websecure',
    
    -- TLS certificate configuration
    tls_domains TEXT DEFAULT '',
    
    -- TCP SNI routing configuration
    tcp_enabled INTEGER DEFAULT 0,
    tcp_entrypoints TEXT DEFAULT 'tcp',
    tcp_sni_rule TEXT DEFAULT '',
    
    -- Custom headers configuration
    custom_headers TEXT DEFAULT '',
    
    -- mTLS whitelist plugin configuration (per-resource overrides)
    mtls_rules TEXT DEFAULT '',
    mtls_request_headers TEXT DEFAULT '',
    mtls_reject_message TEXT DEFAULT '',
    mtls_reject_code INTEGER DEFAULT 403,
    mtls_refresh_interval TEXT DEFAULT '',
    mtls_external_data TEXT DEFAULT '',
    
    -- Router priority configuration
    router_priority INTEGER DEFAULT 100,
    
    -- Source type for tracking data origin
    source_type TEXT DEFAULT '',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Resource_middlewares table stores the relationship between resources and middlewares
CREATE TABLE IF NOT EXISTS resource_middlewares (
    resource_id TEXT NOT NULL,
    middleware_id TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (resource_id, middleware_id),
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE,
    FOREIGN KEY (middleware_id) REFERENCES middlewares(id) ON DELETE CASCADE
);

-- Note: Default middlewares are now loaded via Go code (config/defaults.go)
-- which respects the deleted_templates table to prevent re-creating user-deleted items

-- Services table stores service definitions
CREATE TABLE IF NOT EXISTS services (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    config TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Resource_services table stores the relationship between resources and services
CREATE TABLE IF NOT EXISTS resource_services (
    resource_id TEXT NOT NULL,
    service_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (resource_id, service_id),
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE,
    FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
);

-- Note: Default services are now loaded via Go code (config/service_default.go)
-- which respects the deleted_templates table to prevent re-creating user-deleted items

-- Deleted templates table tracks template IDs that users have explicitly deleted
-- This prevents templates from being re-created on application restart
CREATE TABLE IF NOT EXISTS deleted_templates (
    id TEXT NOT NULL,
    type TEXT NOT NULL, -- 'middleware' or 'service'
    deleted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, type)
);

-- mTLS Global Configuration (singleton table)
CREATE TABLE IF NOT EXISTS mtls_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    enabled INTEGER DEFAULT 0,
    ca_cert TEXT DEFAULT '',
    ca_key TEXT DEFAULT '',
    ca_cert_path TEXT DEFAULT '',
    ca_subject TEXT DEFAULT '',
    ca_expiry TIMESTAMP,
    certs_base_path TEXT DEFAULT '/etc/traefik/certs',
    -- Middleware plugin config
    middleware_rules TEXT DEFAULT '',
    middleware_request_headers TEXT DEFAULT '',
    middleware_reject_message TEXT DEFAULT 'Access denied: Valid client certificate required',
    middleware_refresh_interval INTEGER DEFAULT 300,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- mTLS Client Certificates
CREATE TABLE IF NOT EXISTS mtls_clients (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    cert TEXT NOT NULL,
    key TEXT NOT NULL,
    p12 BLOB,
    p12_password_hint TEXT DEFAULT '',
    subject TEXT DEFAULT '',
    expiry TIMESTAMP,
    revoked INTEGER DEFAULT 0,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Initialize mTLS config singleton row
INSERT OR IGNORE INTO mtls_config (id) VALUES (1);

-- Security Configuration (singleton table for TLS hardening and secure headers)
CREATE TABLE IF NOT EXISTS security_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    tls_hardening_enabled INTEGER DEFAULT 0,
    secure_headers_enabled INTEGER DEFAULT 0,
    secure_headers_x_content_type_options TEXT DEFAULT 'nosniff',
    secure_headers_x_frame_options TEXT DEFAULT 'SAMEORIGIN',
    secure_headers_x_xss_protection TEXT DEFAULT '1; mode=block',
    secure_headers_hsts TEXT DEFAULT 'max-age=31536000; includeSubDomains',
    secure_headers_referrer_policy TEXT DEFAULT 'strict-origin-when-cross-origin',
    secure_headers_csp TEXT DEFAULT '',
    secure_headers_permissions_policy TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Initialize security config singleton row
INSERT OR IGNORE INTO security_config (id) VALUES (1);