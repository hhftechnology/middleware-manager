// Plugin represents a Traefik plugin fetched from the API
export interface Plugin {
  // Basic info
  name: string
  displayName?: string
  moduleName: string
  version: string
  type: string
  description?: string
  summary?: string
  author?: string
  homepage?: string
  iconUrl?: string

  // Status from Traefik
  status: 'enabled' | 'disabled' | 'error' | 'not_loaded' | 'configured'
  error?: string
  provider?: string

  // Installation info
  isInstalled: boolean
  installedVersion?: string
  installSource?: 'catalogue' | 'config'

  // Usage info
  usageCount: number
  usedBy?: string[]

  // Config
  config?: Record<string, unknown>
}

export interface PluginUsage {
  name: string
  usageCount: number
  usedBy: string[]
  status: string
}

export interface PluginInstallRequest {
  moduleName: string
  version?: string
}

export interface PluginInstallResponse {
  message: string
  pluginKey: string
  moduleName: string
  version?: string
}

export interface PluginRemoveRequest {
  moduleName: string
}

export interface PluginRemoveResponse {
  message: string
  pluginKey: string
  moduleName: string
}

export interface PluginConfigPathResponse {
  path: string
  message?: string
}

export interface PluginConfigPathRequest {
  path: string
}

// CataloguePlugin represents a plugin from the Traefik plugin catalogue (plugins.traefik.io)
export interface CataloguePlugin {
  id: string
  name: string
  displayName: string
  author: string
  type: string
  import: string
  summary: string
  iconUrl?: string
  bannerUrl?: string
  latestVersion: string
  versions?: string[]
  stars: number
  snippet?: {
    yaml?: string
    toml?: string
    kubernetes?: string
  }
  isInstalled: boolean
}
