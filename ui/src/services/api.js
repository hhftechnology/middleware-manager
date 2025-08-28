// ui/src/services/api.js

/**
 * API service layer for the Middleware Manager
 * Centralizes all API calls and error handling
 */

const API_URL = '/api';

/**
 * Generic request handler with error handling
 * @param {string} url - API endpoint
 * @param {Object} options - Fetch options
 * @returns {Promise<any>} - Response data or throws an error
 */
const request = async (url, options = {}) => {
  try {
    const response = await fetch(url, options);

    // If the response is not ok, try to parse error JSON, otherwise throw status text
    if (!response.ok) {
      let errorData = { message: `Request failed with status: ${response.statusText} (${response.status})` };
      try {
        // Try to parse JSON error response from the API
        const parsedError = await response.json();
        // Use the message from the API if available
        if (parsedError && parsedError.message) {
          errorData.message = parsedError.message;
        }
        // Include details if provided by the API
        if (parsedError && parsedError.details) {
          errorData.details = parsedError.details;
        }
      } catch (e) {
        // Ignore JSON parsing error if the body is not JSON or empty
        console.debug("Could not parse error response body as JSON.");
      }

      const error = new Error(errorData.message);
      error.status = response.status;
      error.data = errorData; // Attach potentially parsed data
      throw error;
    }

    // Handle empty response body for methods like DELETE which might return 200 OK with no content
    if (response.status === 200 || response.status === 204) {
      const contentType = response.headers.get("content-type");
      if (contentType && contentType.indexOf("application/json") !== -1) {
        // Only parse JSON if the content type indicates it
         return await response.json();
      } else {
        // Return a success indicator for non-JSON 2xx responses
        return { success: true, status: response.status };
      }
    }

    // For other successful responses, parse JSON
    return await response.json();
  } catch (error) {
    console.error(`API Error (${options.method || 'GET'} ${url}): ${error.message}`, error);
    // Re-throw the error so it can be caught by the calling code (e.g., in contexts)
    throw error;
  }
};


/**
 * Resource-related API calls
 */
export const ResourceService = {
  getResources: () => request(`${API_URL}/resources`),
  getResource: (id) => request(`${API_URL}/resources/${id}`),
  deleteResource: (id) => request(`${API_URL}/resources/${id}`, { method: 'DELETE' }),
  assignMiddleware: (resourceId, data) => request(
    `${API_URL}/resources/${resourceId}/middlewares`,
    { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }
  ),
  assignMultipleMiddlewares: (resourceId, data) => request(
    `${API_URL}/resources/${resourceId}/middlewares/bulk`,
    { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }
  ),
  removeMiddleware: (resourceId, middlewareId) => request(
    `${API_URL}/resources/${resourceId}/middlewares/${middlewareId}`,
    { method: 'DELETE' }
  ),
  updateHTTPConfig: (resourceId, data) => request(
    `${API_URL}/resources/${resourceId}/config/http`,
    { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }
  ),
  updateTLSConfig: (resourceId, data) => request(
    `${API_URL}/resources/${resourceId}/config/tls`,
    { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }
  ),
  updateTCPConfig: (resourceId, data) => request(
    `${API_URL}/resources/${resourceId}/config/tcp`,
    { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }
  ),
  updateHeadersConfig: (resourceId, data) => request(
    `${API_URL}/resources/${resourceId}/config/headers`,
    { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }
  ),
  updateRouterPriority: (resourceId, data) => request(
    `${API_URL}/resources/${resourceId}/config/priority`,
    { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }
  ),
  // --- Resource-Service Assignment ---
  getResourceService: (resourceId) => request(`${API_URL}/resources/${resourceId}/service`),
  assignServiceToResource: (resourceId, data) => request(
    `${API_URL}/resources/${resourceId}/service`,
    { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }
  ),
  removeServiceFromResource: (resourceId) => request(
    `${API_URL}/resources/${resourceId}/service`,
    { method: 'DELETE' }
  ),
};

/**
 * Middleware-related API calls
 */
export const MiddlewareService = {
  getMiddlewares: () => request(`${API_URL}/middlewares`),
  getMiddleware: (id) => request(`${API_URL}/middlewares/${id}`),
  createMiddleware: (data) => request(
    `${API_URL}/middlewares`,
    { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }
  ),
  updateMiddleware: (id, data) => request(
    `${API_URL}/middlewares/${id}`,
    { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }
  ),
  deleteMiddleware: (id) => request(
    `${API_URL}/middlewares/${id}`,
    { method: 'DELETE' }
  )
};

// --- Add/Update ServiceService ---
/**
 * Service-related API calls
 */
export const ServiceService = {
  /**
   * Get all services
   * @returns {Promise<Array>} - List of services
   */
  getServices: () => request(`${API_URL}/services`), // Matches draft: fetchServices

  /**
   * Get a specific service by ID
   * @param {string} id - Service ID
   * @returns {Promise<Object>} - Service data
   */
  getService: (id) => request(`${API_URL}/services/${id}`), // Matches draft: fetchService

  /**
   * Create a new service
   * @param {Object} data - Service data
   * @returns {Promise<Object>} - Created service
   */
  createService: (data) => request( // Matches draft: createService
    `${API_URL}/services`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    }
  ),

  /**
   * Update a service
   * @param {string} id - Service ID
   * @param {Object} data - Updated service data
   * @returns {Promise<Object>} - Updated service
   */
  updateService: (id, data) => request( // Matches draft: updateService
    `${API_URL}/services/${id}`,
    {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    }
  ),

  /**
   * Delete a service
   * @param {string} id - Service ID
   * @returns {Promise<Object>} - Response message (or {success: true})
   */
  deleteService: (id) => request( // Matches draft: deleteService
    `${API_URL}/services/${id}`,
    { method: 'DELETE' }
  ),
};

/**
 * Utility functions for middleware management
 */
export const MiddlewareUtils = {
  parseMiddlewares: (middlewaresStr) => {
    if (!middlewaresStr || typeof middlewaresStr !== 'string') return [];
    return middlewaresStr
      .split(',')
      .filter(Boolean)
      .map((item) => {
        const [id, name, priority] = item.split(':');
        return {
          id,
          name,
          priority: parseInt(priority, 10) || 200,
        };
      });
  },
  getConfigTemplate: (type) => {
      const templates = {
          basicAuth: '{\n  "users": [\n    "admin:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"\n  ]\n}',
          digestAuth: '{\n  "users": [\n    "test:traefik:a2688e031edb4be6a3797f3882655c05"\n  ]\n}',
          forwardAuth: '{\n  "address": "http://auth-service:9090/auth",\n  "trustForwardHeader": true,\n  "authResponseHeaders": [\n    "X-Auth-User",\n    "X-Auth-Roles"\n  ]\n}',
          ipWhiteList: '{\n  "sourceRange": [\n    "127.0.0.1/32",\n    "192.168.1.0/24"\n  ]\n}',
          ipAllowList: '{\n  "sourceRange": [\n    "127.0.0.1/32",\n    "192.168.1.0/24"\n  ]\n}',
          rateLimit: '{\n  "average": 100,\n  "burst": 50\n}',
          headers: '{\n  "browserXssFilter": true,\n  "contentTypeNosniff": true,\n  "customFrameOptionsValue": "SAMEORIGIN",\n  "forceSTSHeader": true,\n  "stsIncludeSubdomains": true,\n  "stsSeconds": 63072000,\n  "customResponseHeaders": {\n    "X-Custom-Header": "value",\n    "Server": ""\n  }\n}',
          stripPrefix: '{\n  "prefixes": [\n    "/api"\n  ],\n  "forceSlash": true\n}',
          addPrefix: '{\n  "prefix": "/api"\n}',
          redirectRegex: '{\n  "regex": "^http://(.*)$",\n  "replacement": "https://${1}",\n  "permanent": true\n}',
          redirectScheme: '{\n  "scheme": "https",\n  "permanent": true,\n  "port": "443"\n}',
          chain: '{\n  "middlewares": []\n}',
          replacePath: '{\n  "path": "/newpath"\n}',
          replacePathRegex: '{\n  "regex": "^/api/(.*)",\n  "replacement": "/bar/$1"\n}',
          stripPrefixRegex: '{\n  "regex": [\n    "^/api/v\\\\d+/"\n  ]\n}',
          plugin: '{\n  "plugin-name": {\n    "option1": "value1",\n    "option2": "value2"\n  }\n}',
          buffering: '{\n  "maxRequestBodyBytes": 10485760,\n  "memRequestBodyBytes": 2097152,\n  "maxResponseBodyBytes": 10485760,\n  "memResponseBodyBytes": 2097152,\n  "retryExpression": "IsNetworkError() && Attempts() < 2"\n}',
          circuitBreaker: '{\n  "expression": "NetworkErrorRatio() > 0.5 && Requests > 10"\n}',
          compress: '{\n  "excludedContentTypes": [\n    "text/event-stream"\n  ]\n}',
          contentType: '{}', // Empty config
          errors: '{\n  "status": ["500-599"],\n  "service": "error-service@file",\n  "query": "/{status}.html"\n}',
          grpcWeb: '{}', // Empty config
          inFlightReq: '{\n  "amount": 10\n}',
          passTLSClientCert: '{\n  "pem": true\n}',
          retry: '{\n  "attempts": 4\n}'
      };
      const templateJson = templates[type] || '{}';
      // Ensure it's nicely formatted
      try {
          return JSON.stringify(JSON.parse(templateJson), null, 2);
      } catch (e) {
          return '{}';
      }
  }
};

/**
 * Utility functions for service management
 */
export const ServiceUtils = {
  /**
   * Returns a template configuration for the specified service type
   * @param {string} type - Service type (loadBalancer, weighted, mirroring, failover)
   * @returns {string} JSON template as a formatted string
   */
  getConfigTemplate: (type) => {
    const templates = {
      // HTTP Load Balancer with multiple configuration options
      loadBalancer: `{
  "servers": [
    {
      "url": "http://backend1:80"
    },
    {
      "url": "http://backend2:80",
      "weight": 2
    }
  ],
  "healthCheck": {
    "path": "/health",
    "interval": "10s",
    "timeout": "3s"
  },
  "sticky": {
    "cookie": {
      "name": "sticky_session",
      "secure": true,
      "httpOnly": true
    }
  },
  "passHostHeader": true
}`,

      // Weighted service with multiple service references and sticky session
      weighted: `{
  "services": [
    {
      "name": "primary-service@file",
      "weight": 3
    },
    {
      "name": "backup-service@file",
      "weight": 1
    }
  ],
  "healthCheck": {},
  "sticky": {
    "cookie": {
      "name": "weighted_sticky"
    }
  }
}`,

      // Mirroring service with full configuration options
      mirroring: `{
  "service": "primary-service@file",
  "mirrors": [
    {
      "name": "analytics-service@file",
      "percent": 10
    },
    {
      "name": "testing-service@file",
      "percent": 5
    }
  ],
  "mirrorBody": true,
  "maxBodySize": 10485760,
  "healthCheck": {}
}`,

      // Failover service with health checks
      failover: `{
  "service": "main-service@file",
  "fallback": "backup-service@file",
  "healthCheck": {}
}`
    };

    // Get the template for the requested type or return empty object
    const templateJson = templates[type] || '{}';

    // Ensure it's nicely formatted
    try {
      return JSON.stringify(JSON.parse(templateJson), null, 2);
    } catch (e) {
      console.error(`Error parsing template for ${type}:`, e);
      return '{}';
    }
  },

  /**
   * Returns descriptive information about service types
   * @param {string} type - Service type
   * @returns {Object} Description and examples
   */
  getServiceTypeInfo: (type) => {
    const info = {
      loadBalancer: {
        description: "Routes requests to multiple backend servers with load balancing",
        serverFormat: "url: \"http://server-address:port\" (HTTP) or address: \"server:port\" (TCP/UDP)",
        commonOptions: [
          "sticky: For session persistence",
          "healthCheck: For server health monitoring",
          "passHostHeader: To control Host header forwarding",
          "weight: To distribute load proportionally"
        ]
      },
      weighted: {
        description: "Distributes requests across multiple named services with weighted load balancing",
        format: "services: Array of {name, weight} objects",
        serviceNames: "Must reference existing services (e.g., service-name@file)"
      },
      mirroring: {
        description: "Forwards requests to a primary service while mirroring a percentage of traffic to other services",
        format: "service: Primary service, mirrors: Array of {name, percent} objects",
        notes: "Used for testing, analytics, or gradual deployments"
      },
      failover: {
        description: "Routes to a primary service until it fails, then uses a fallback service",
        format: "service: Main service, fallback: Backup service",
        notes: "Typically used with healthCheck configuration"
      }
    };

    return info[type] || {
      description: "Unknown service type"
    };
  },

  /**
   * Returns special configurations for TCP/UDP services (used with loadBalancer type)
   * @param {string} protocol - "tcp" or "udp"
   * @returns {string} JSON template as a formatted string
   */
  getProtocolTemplate: (protocol) => {
    const templates = {
      tcp: `{
  "servers": [
    {
      "address": "backend1:8080"
    },
    {
      "address": "backend2:8080",
      "tls": true
    }
  ],
  "terminationDelay": 100
}`,
      udp: `{
  "servers": [
    {
      "address": "backend1:53"
    },
    {
      "address": "backend2:53"
    }
  ]
}`
    };

    const templateJson = templates[protocol] || '{}';

    try {
      return JSON.stringify(JSON.parse(templateJson), null, 2);
    } catch (e) {
      return '{}';
    }
  }
}