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
    
    -- Router priority configuration
    router_priority INTEGER DEFAULT 200,
    
    -- Source type for tracking data origin
    source_type TEXT DEFAULT '',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Resource_middlewares table stores the relationship between resources and middlewares
CREATE TABLE IF NOT EXISTS resource_middlewares (
    resource_id TEXT NOT NULL,
    middleware_id TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 200,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (resource_id, middleware_id),
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE,
    FOREIGN KEY (middleware_id) REFERENCES middlewares(id) ON DELETE CASCADE
);

-- Insert default middlewares
INSERT OR IGNORE INTO middlewares (id, name, type, config) VALUES 
('authelia', 'Authelia', 'forwardAuth', '{"address":"http://authelia:9091/api/authz/forward-auth","trustForwardHeader":true,"authResponseHeaders":["Remote-User","Remote-Groups","Remote-Name","Remote-Email"]}'),
('authentik', 'Authentik', 'forwardAuth', '{"address":"http://authentik:9000/outpost.goauthentik.io/auth/traefik","trustForwardHeader":true,"authResponseHeaders":["X-authentik-username","X-authentik-groups","X-authentik-email","X-authentik-name","X-authentik-uid"]}'),
('basic-auth', 'Basic Auth', 'basicAuth', '{"users":["admin:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"]}');

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

-- Insert default services
INSERT OR IGNORE INTO services (id, name, type, config) VALUES
('simple-lb', 'Simple LoadBalancer', 'loadBalancer', '{"servers":[{"url":"http://localhost:8080"}]}'),
('weighted-demo', 'Weighted Service Demo', 'weighted', '{"services":[{"name":"service1","weight":3},{"name":"service2","weight":1}]}'),
('failover-demo', 'Failover Demo', 'failover', '{"service":"main","fallback":"backup"}');

-- Deleted templates table tracks template IDs that users have explicitly deleted
-- This prevents templates from being re-created on application restart
CREATE TABLE IF NOT EXISTS deleted_templates (
    id TEXT NOT NULL,
    type TEXT NOT NULL, -- 'middleware' or 'service'
    deleted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, type)
);