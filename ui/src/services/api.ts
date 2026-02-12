import type {
  Resource,
  Middleware,
  Service,
  DataSourceConfig,
  DataSourceInfo,
  Plugin,
  CataloguePlugin,
  CreateMiddlewareRequest,
  UpdateMiddlewareRequest,
  CreateServiceRequest,
  UpdateServiceRequest,
  AssignMiddlewareRequest,
  AssignExternalMiddlewareRequest,
  ExternalMiddleware,
  AssignServiceRequest,
  HTTPConfig,
  TLSConfig,
  TCPConfig,
  HeadersConfig,
  MTLSWhitelistConfigRequest,
  TestConnectionResponse,
  PluginInstallRequest,
  PluginRemoveRequest,
  TraefikOverview,
  TraefikVersion,
  TraefikEntrypoint,
  HTTPRouter,
  TCPRouter,
  UDPRouter,
  HTTPService,
  TCPService,
  UDPService,
  HTTPMiddleware,
  TCPMiddleware,
  FullTraefikData,
  AllRoutersResponse,
  AllServicesResponse,
  AllMiddlewaresResponse,
  ProtocolType,
  MTLSConfig,
  MTLSClient,
  CreateCARequest,
  CreateClientRequest,
  PluginCheckResponse,
  MTLSMiddlewareConfig,
  SecurityConfig,
  SecureHeadersConfig,
  DuplicateCheckResult,
  DuplicateCheckRequest,
  UpdateResourceSecurityRequest,
} from '@/types'

const API_BASE = '/api'

// Custom API Error class
export class ApiError extends Error {
  status: number
  details?: unknown

  constructor(message: string, status: number, details?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.details = details
  }
}

// Generic request function with error handling
async function request<T>(
  url: string,
  options?: RequestInit
): Promise<T> {
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!response.ok) {
    let errorData: { message?: string; error?: string; details?: unknown } = {}
    try {
      errorData = await response.json()
    } catch {
      // Response is not JSON
    }
    throw new ApiError(
      errorData.message || errorData.error || `Request failed: ${response.statusText}`,
      response.status,
      errorData.details
    )
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return {} as T
  }

  // Check if response has content
  const contentType = response.headers.get('content-type')
  if (contentType && contentType.includes('application/json')) {
    return response.json()
  }

  return {} as T
}

// Resource API
export const resourceApi = {
  getAll: () => request<Resource[]>(`${API_BASE}/resources`),

  getById: (id: string) => request<Resource>(`${API_BASE}/resources/${encodeURIComponent(id)}`),

  delete: (id: string) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    }),

  bulkDeleteDisabled: (ids: string[]) =>
    request<{ deleted: number; ids: string[] }>(`${API_BASE}/resources/bulk-delete-disabled`, {
      method: 'POST',
      body: JSON.stringify({ ids }),
    }),

  // Middleware assignment
  assignMiddleware: (resourceId: string, data: AssignMiddlewareRequest) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/middlewares`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  assignMultipleMiddlewares: (resourceId: string, middlewares: AssignMiddlewareRequest[]) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/middlewares/bulk`, {
      method: 'POST',
      body: JSON.stringify({ middlewares }),
    }),

  removeMiddleware: (resourceId: string, middlewareId: string) =>
    request<void>(
      `${API_BASE}/resources/${encodeURIComponent(resourceId)}/middlewares/${encodeURIComponent(middlewareId)}`,
      { method: 'DELETE' }
    ),

  // External (Traefik-native) middleware assignment
  getExternalMiddlewares: (resourceId: string) =>
    request<ExternalMiddleware[]>(
      `${API_BASE}/resources/${encodeURIComponent(resourceId)}/external-middlewares`
    ),

  assignExternalMiddleware: (resourceId: string, data: AssignExternalMiddlewareRequest) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/external-middlewares`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  removeExternalMiddleware: (resourceId: string, name: string) =>
    request<void>(
      `${API_BASE}/resources/${encodeURIComponent(resourceId)}/external-middlewares/${encodeURIComponent(name)}`,
      { method: 'DELETE' }
    ),

  // Service assignment
  getService: (resourceId: string) =>
    request<Service>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/service`),

  assignService: (resourceId: string, data: AssignServiceRequest) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/service`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  removeService: (resourceId: string) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/service`, {
      method: 'DELETE',
    }),

  // Configuration updates
  updateHTTPConfig: (resourceId: string, config: HTTPConfig) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/config/http`, {
      method: 'PUT',
      body: JSON.stringify(config),
    }),

  updateTLSConfig: (resourceId: string, config: TLSConfig) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/config/tls`, {
      method: 'PUT',
      body: JSON.stringify(config),
    }),

  updateTCPConfig: (resourceId: string, config: TCPConfig) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/config/tcp`, {
      method: 'PUT',
      body: JSON.stringify(config),
    }),

  updateHeadersConfig: (resourceId: string, config: HeadersConfig) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/config/headers`, {
      method: 'PUT',
      body: JSON.stringify(config),
    }),

  updateRouterPriority: (resourceId: string, priority: number) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/config/priority`, {
      method: 'PUT',
      body: JSON.stringify({ router_priority: priority }),
    }),

  updateMTLSConfig: (resourceId: string, mtlsEnabled: boolean) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/config/mtls`, {
      method: 'PUT',
      body: JSON.stringify({ mtls_enabled: mtlsEnabled }),
    }),

  updateMTLSWhitelistConfig: (resourceId: string, config: MTLSWhitelistConfigRequest) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/config/mtlswhitelist`, {
      method: 'PUT',
      body: JSON.stringify(config),
    }),

  updateTLSHardeningConfig: (resourceId: string, enabled: boolean) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/config/tls-hardening`, {
      method: 'PUT',
      body: JSON.stringify({ enabled } as UpdateResourceSecurityRequest),
    }),

  updateSecureHeadersConfig: (resourceId: string, enabled: boolean) =>
    request<void>(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/config/secure-headers`, {
      method: 'PUT',
      body: JSON.stringify({ enabled } as UpdateResourceSecurityRequest),
    }),
}

// Middleware API
export const middlewareApi = {
  getAll: () => request<Middleware[]>(`${API_BASE}/middlewares`),

  getById: (id: string) => request<Middleware>(`${API_BASE}/middlewares/${encodeURIComponent(id)}`),

  create: (data: CreateMiddlewareRequest) =>
    request<Middleware>(`${API_BASE}/middlewares`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: string, data: UpdateMiddlewareRequest) =>
    request<Middleware>(`${API_BASE}/middlewares/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    request<void>(`${API_BASE}/middlewares/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    }),
}

// Service API
export const serviceApi = {
  getAll: () => request<Service[]>(`${API_BASE}/services`),

  getById: (id: string) => request<Service>(`${API_BASE}/services/${encodeURIComponent(id)}`),

  create: (data: CreateServiceRequest) =>
    request<Service>(`${API_BASE}/services`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: string, data: UpdateServiceRequest) =>
    request<Service>(`${API_BASE}/services/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    request<void>(`${API_BASE}/services/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    }),
}

// Data Source API response types (backend format)
interface DataSourcesResponse {
  active_source: string
  sources: Record<string, { type: string; url: string; basicAuth?: { username: string; password: string } }>
}

// Data Source API
export const dataSourceApi = {
  getAll: async (): Promise<DataSourceInfo[]> => {
    const response = await request<DataSourcesResponse>(`${API_BASE}/datasource`)
    // Transform backend response to array format
    if (!response.sources) {
      return []
    }
    return Object.entries(response.sources).map(([name, config]) => ({
      name,
      type: config.type as DataSourceInfo['type'],
      url: config.url,
      isActive: name === response.active_source,
    }))
  },

  getActive: () => request<DataSourceConfig>(`${API_BASE}/datasource/active`),

  setActive: (name: string) =>
    request<void>(`${API_BASE}/datasource/active`, {
      method: 'PUT',
      body: JSON.stringify({ name }),
    }),

  update: (name: string, config: Partial<DataSourceConfig>) =>
    request<void>(`${API_BASE}/datasource/${encodeURIComponent(name)}`, {
      method: 'PUT',
      body: JSON.stringify(config),
    }),

  testConnection: (name: string) =>
    request<TestConnectionResponse>(`${API_BASE}/datasource/${encodeURIComponent(name)}/test`, {
      method: 'POST',
    }),
}

// Plugin API - fetches plugins from Traefik API
export const pluginApi = {
  getAll: () => request<Plugin[]>(`${API_BASE}/plugins`),

  // Fetch the full plugin catalogue from plugins.traefik.io
  getCatalogue: () => request<CataloguePlugin[]>(`${API_BASE}/plugins/catalogue`),

  getUsage: (name: string) =>
    request<{ name: string; usageCount: number; usedBy: string[]; status: string }>(
      `${API_BASE}/plugins/${encodeURIComponent(name)}/usage`
    ),

  install: (data: PluginInstallRequest) =>
    request<{ message: string; pluginKey: string; moduleName: string; version?: string }>(
      `${API_BASE}/plugins/install`,
      {
        method: 'POST',
        body: JSON.stringify(data),
      }
    ),

  remove: (data: PluginRemoveRequest) =>
    request<{ message: string; pluginKey: string; moduleName: string }>(
      `${API_BASE}/plugins/remove`,
      {
        method: 'DELETE',
        body: JSON.stringify(data),
      }
    ),

  getConfigPath: () => request<{ path: string; message?: string }>(`${API_BASE}/plugins/configpath`),

  updateConfigPath: (path: string) =>
    request<{ message: string; path: string }>(`${API_BASE}/plugins/configpath`, {
      method: 'PUT',
      body: JSON.stringify({ path }),
    }),
}

// Traefik API - Direct access to Traefik data following Mantrae patterns
export const traefikApi = {
  // Get Traefik overview statistics
  getOverview: () => request<TraefikOverview>(`${API_BASE}/traefik/overview`),

  // Get Traefik version
  getVersion: () => request<TraefikVersion>(`${API_BASE}/traefik/version`),

  // Get Traefik entrypoints
  getEntrypoints: () => request<TraefikEntrypoint[]>(`${API_BASE}/traefik/entrypoints`),

  // Get routers with optional protocol filter
  getRouters: (type?: ProtocolType) => {
    const params = type ? `?type=${type}` : ''
    return request<HTTPRouter[] | TCPRouter[] | UDPRouter[] | AllRoutersResponse>(
      `${API_BASE}/traefik/routers${params}`
    )
  },

  // Get HTTP routers only
  getHTTPRouters: () => request<HTTPRouter[]>(`${API_BASE}/traefik/routers?type=http`),

  // Get TCP routers only
  getTCPRouters: () => request<TCPRouter[]>(`${API_BASE}/traefik/routers?type=tcp`),

  // Get UDP routers only
  getUDPRouters: () => request<UDPRouter[]>(`${API_BASE}/traefik/routers?type=udp`),

  // Get all routers (aggregated)
  getAllRouters: () => request<AllRoutersResponse>(`${API_BASE}/traefik/routers?type=all`),

  // Get services with optional protocol filter
  getServices: (type?: ProtocolType) => {
    const params = type ? `?type=${type}` : ''
    return request<HTTPService[] | TCPService[] | UDPService[] | AllServicesResponse>(
      `${API_BASE}/traefik/services${params}`
    )
  },

  // Get HTTP services only
  getHTTPServices: () => request<HTTPService[]>(`${API_BASE}/traefik/services?type=http`),

  // Get TCP services only
  getTCPServices: () => request<TCPService[]>(`${API_BASE}/traefik/services?type=tcp`),

  // Get UDP services only
  getUDPServices: () => request<UDPService[]>(`${API_BASE}/traefik/services?type=udp`),

  // Get all services (aggregated)
  getAllServices: () => request<AllServicesResponse>(`${API_BASE}/traefik/services?type=all`),

  // Get middlewares with optional protocol filter
  getMiddlewares: (type?: 'http' | 'tcp' | 'all') => {
    const params = type ? `?type=${type}` : ''
    return request<HTTPMiddleware[] | TCPMiddleware[] | AllMiddlewaresResponse>(
      `${API_BASE}/traefik/middlewares${params}`
    )
  },

  // Get HTTP middlewares only
  getHTTPMiddlewares: () => request<HTTPMiddleware[]>(`${API_BASE}/traefik/middlewares?type=http`),

  // Get TCP middlewares only
  getTCPMiddlewares: () => request<TCPMiddleware[]>(`${API_BASE}/traefik/middlewares?type=tcp`),

  // Get all middlewares (aggregated)
  getAllMiddlewares: () => request<AllMiddlewaresResponse>(`${API_BASE}/traefik/middlewares?type=all`),

  // Get full Traefik data in one request
  getFullData: () => request<FullTraefikData>(`${API_BASE}/traefik/data`),
}

// mTLS API - Certificate Authority and client certificate management
export const mtlsApi = {
  // Get mTLS configuration
  getConfig: () => request<MTLSConfig>(`${API_BASE}/mtls/config`),

  // Enable mTLS globally
  enableMTLS: () =>
    request<{ message: string; enabled: boolean }>(`${API_BASE}/mtls/enable`, {
      method: 'PUT',
    }),

  // Disable mTLS globally
  disableMTLS: () =>
    request<{ message: string; enabled: boolean }>(`${API_BASE}/mtls/disable`, {
      method: 'PUT',
    }),

  // Create Certificate Authority
  createCA: (data: CreateCARequest) =>
    request<MTLSConfig>(`${API_BASE}/mtls/ca`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Delete Certificate Authority (and all client certs)
  deleteCA: () =>
    request<{ message: string }>(`${API_BASE}/mtls/ca`, {
      method: 'DELETE',
    }),

  // Update certificates base path
  updateCertsBasePath: (certsBasePath: string) =>
    request<{ message: string; certs_base_path: string }>(`${API_BASE}/mtls/config/path`, {
      method: 'PUT',
      body: JSON.stringify({ certs_base_path: certsBasePath }),
    }),

  // Get all client certificates
  getClients: () => request<MTLSClient[]>(`${API_BASE}/mtls/clients`),

  // Create a new client certificate
  createClient: (data: CreateClientRequest) =>
    request<MTLSClient>(`${API_BASE}/mtls/clients`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Get a specific client certificate
  getClient: (id: string) =>
    request<MTLSClient>(`${API_BASE}/mtls/clients/${encodeURIComponent(id)}`),

  // Get download URL for client P12 file (use with window.open or anchor download)
  getClientP12Url: (id: string) =>
    `${API_BASE}/mtls/clients/${encodeURIComponent(id)}/download`,

  // Revoke a client certificate
  revokeClient: (id: string) =>
    request<{ message: string; id: string }>(`${API_BASE}/mtls/clients/${encodeURIComponent(id)}/revoke`, {
      method: 'PUT',
    }),

  // Delete a client certificate
  deleteClient: (id: string) =>
    request<{ message: string; id: string }>(`${API_BASE}/mtls/clients/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    }),

  // Check if mtlswhitelist plugin is installed
  checkPlugin: () => request<PluginCheckResponse>(`${API_BASE}/mtls/plugin/check`),

  // Get middleware plugin configuration
  getMiddlewareConfig: () => request<MTLSMiddlewareConfig>(`${API_BASE}/mtls/middleware/config`),

  // Update middleware plugin configuration
  updateMiddlewareConfig: (config: MTLSMiddlewareConfig) =>
    request<{ message: string }>(`${API_BASE}/mtls/middleware/config`, {
      method: 'PUT',
      body: JSON.stringify(config),
    }),
}

// Security API - TLS hardening, secure headers, duplicate detection
export const securityApi = {
  // Get security configuration
  getConfig: () => request<SecurityConfig>(`${API_BASE}/security/config`),

  // TLS Hardening
  enableTLSHardening: () =>
    request<{ message: string; enabled: boolean }>(`${API_BASE}/security/tls-hardening/enable`, {
      method: 'PUT',
    }),

  disableTLSHardening: () =>
    request<{ message: string; enabled: boolean }>(`${API_BASE}/security/tls-hardening/disable`, {
      method: 'PUT',
    }),

  // Secure Headers
  enableSecureHeaders: () =>
    request<{ message: string; enabled: boolean }>(`${API_BASE}/security/secure-headers/enable`, {
      method: 'PUT',
    }),

  disableSecureHeaders: () =>
    request<{ message: string; enabled: boolean }>(`${API_BASE}/security/secure-headers/disable`, {
      method: 'PUT',
    }),

  updateSecureHeadersConfig: (config: SecureHeadersConfig) =>
    request<{ message: string }>(`${API_BASE}/security/secure-headers/config`, {
      method: 'PUT',
      body: JSON.stringify(config),
    }),

  // Duplicate Detection
  checkDuplicates: (req: DuplicateCheckRequest) =>
    request<DuplicateCheckResult>(`${API_BASE}/security/check-duplicates`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),
}

// Health check
export const healthApi = {
  check: () => request<{ status: string }>('/health'),
}
