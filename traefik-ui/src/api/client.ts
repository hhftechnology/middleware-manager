import type {
  ApiErrorShape,
  BackupEntry,
  CertificatesResponse,
  ConfigListResponse,
  DashboardConfig,
  LogsResponse,
  ManagerVersion,
  MiddlewareEntry,
  MiddlewareRequest,
  PluginsResponse,
  RouteRequest,
  RouteResponse,
  SelfRoute,
  Settings,
  TraefikPing,
} from '@/types'

const API_BASE = '/api'

export class ApiError extends Error {
  status: number
  details?: string

  constructor(status: number, data: ApiErrorShape) {
    super(data.message || data.error || `Request failed with status ${status}`)
    this.name = 'ApiError'
    this.status = status
    this.details = data.details
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
  })

  if (!response.ok) {
    let data: ApiErrorShape = {}
    try {
      data = await response.json()
    } catch {
      data = {}
    }
    throw new ApiError(response.status, data)
  }

  if (response.status === 204) {
    return {} as T
  }

  return response.json() as Promise<T>
}

export const api = {
  routes: {
    list: () => request<RouteResponse>('/routes'),
    create: (payload: RouteRequest) => request<{ ok: boolean }>('/routes', { method: 'POST', body: JSON.stringify(payload) }),
    update: (id: string, payload: RouteRequest) => request<{ ok: boolean }>(`/routes/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(payload) }),
    delete: (id: string) => request<{ ok: boolean }>(`/routes/${encodeURIComponent(id)}`, { method: 'DELETE' }),
    toggle: (id: string, enable: boolean) => request<{ ok: boolean }>(`/routes/${encodeURIComponent(id)}/toggle`, { method: 'POST', body: JSON.stringify({ enable }) }),
    routerNames: () => request<string[]>('/manager/router-names'),
  },
  configs: {
    list: () => request<ConfigListResponse>('/configs'),
  },
  middlewares: {
    list: () => request<MiddlewareEntry[]>('/middlewares'),
    create: (payload: MiddlewareRequest) => request<{ ok: boolean }>('/middlewares', { method: 'POST', body: JSON.stringify(payload) }),
    update: (name: string, payload: MiddlewareRequest) => request<{ ok: boolean }>(`/middlewares/${encodeURIComponent(name)}`, { method: 'PUT', body: JSON.stringify(payload) }),
    delete: (name: string, configFile?: string) => request<{ ok: boolean }>(`/middlewares/${encodeURIComponent(name)}${configFile ? `?configFile=${encodeURIComponent(configFile)}` : ''}`, { method: 'DELETE' }),
  },
  settings: {
    get: () => request<Settings>('/settings'),
    save: (payload: Partial<Settings>) => request<{ success: boolean; settings: Settings }>('/settings', { method: 'POST', body: JSON.stringify(payload) }),
    saveTabs: (payload: Record<string, boolean>) => request<{ success: boolean }>('/settings/tabs', { method: 'POST', body: JSON.stringify(payload) }),
    getSelfRoute: () => request<SelfRoute>('/settings/self-route'),
    saveSelfRoute: (payload: SelfRoute) => request<{ ok: boolean }>('/settings/self-route', { method: 'POST', body: JSON.stringify(payload) }),
    testConnection: (url: string) => request<{ ok: boolean; error?: string }>('/settings/test-connection', { method: 'POST', body: JSON.stringify({ url }) }),
  },
  backups: {
    list: () => request<BackupEntry[]>('/backups'),
    create: () => request<{ success: boolean; name: string }>('/backup/create', { method: 'POST' }),
    restore: (name: string) => request<{ success: boolean }>(`/restore/${encodeURIComponent(name)}`, { method: 'POST' }),
    remove: (name: string) => request<{ success: boolean }>(`/backup/delete/${encodeURIComponent(name)}`, { method: 'POST' }),
  },
  traefik: {
    overview: () => request<Record<string, unknown>>('/traefik/overview'),
    ping: () => request<TraefikPing>('/traefik/ping'),
    certs: () => request<CertificatesResponse>('/traefik/certs'),
    logs: () => request<LogsResponse>('/traefik/logs'),
    plugins: () => request<PluginsResponse>('/traefik/plugins'),
  },
  dashboard: {
    get: () => request<DashboardConfig>('/dashboard/config'),
    save: (payload: DashboardConfig) => request<{ ok: boolean }>('/dashboard/config', { method: 'POST', body: JSON.stringify(payload) }),
  },
  manager: {
    version: () => request<ManagerVersion>('/manager/version'),
  },
  health: async () => {
    const response = await fetch('/health')
    return response.json() as Promise<{ status: string; mode: string }>
  },
}
