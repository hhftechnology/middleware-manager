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
  AssignServiceRequest,
  HTTPConfig,
  TLSConfig,
  TCPConfig,
  HeadersConfig,
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

// Health check
export const healthApi = {
  check: () => request<{ status: string }>('/health'),
}
