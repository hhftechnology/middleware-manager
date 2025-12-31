import type {
  Resource,
  Middleware,
  Service,
  DataSourceConfig,
  DataSourceInfo,
  Plugin,
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
      body: JSON.stringify({ priority }),
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

// Data Source API
export const dataSourceApi = {
  getAll: () => request<DataSourceInfo[]>(`${API_BASE}/datasource`),

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

// Plugin API
export const pluginApi = {
  getAll: () => request<Plugin[]>(`${API_BASE}/plugins`),

  install: (data: PluginInstallRequest) =>
    request<void>(`${API_BASE}/plugins/install`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  remove: (data: PluginRemoveRequest) =>
    request<void>(`${API_BASE}/plugins/remove`, {
      method: 'DELETE',
      body: JSON.stringify(data),
    }),

  getConfigPath: () => request<{ path: string }>(`${API_BASE}/plugins/configpath`),

  updateConfigPath: (path: string) =>
    request<void>(`${API_BASE}/plugins/configpath`, {
      method: 'PUT',
      body: JSON.stringify({ path }),
    }),
}

// Health check
export const healthApi = {
  check: () => request<{ status: string }>('/health'),
}
