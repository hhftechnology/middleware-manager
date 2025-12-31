export interface Plugin {
  name: string
  moduleName: string
  version: string
  description?: string
  author?: string
  homepage?: string
  installed: boolean
  installedVersion?: string
  configPath?: string
}

export interface PluginInstallRequest {
  moduleName: string
  version: string
}

export interface PluginRemoveRequest {
  moduleName: string
}

export interface PluginConfigPathResponse {
  path: string
}

export interface PluginConfigPathRequest {
  path: string
}
