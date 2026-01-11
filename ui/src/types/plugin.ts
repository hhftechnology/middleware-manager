// Plugin represents a Traefik plugin fetched from the API
export interface Plugin {
  // Basic info
  name: string
  moduleName: string
  version: string
  type: string
  description?: string
  author?: string
  homepage?: string

  // Status from Traefik
  status: 'enabled' | 'disabled' | 'error' | 'not_loaded' | 'configured'
  error?: string
  provider?: string

  // Installation info
  isInstalled: boolean
  installedVersion?: string

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
