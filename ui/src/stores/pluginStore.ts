import { create } from 'zustand'
import { pluginApi } from '@/services/api'
import type { Plugin, PluginUsage } from '@/types'

interface PluginState {
  // Data
  plugins: Plugin[]
  configPath: string
  selectedPlugin: Plugin | null

  // Loading states
  loading: boolean
  installing: boolean
  removing: boolean

  // Error state
  error: string | null

  // Actions
  fetchPlugins: () => Promise<void>
  fetchConfigPath: () => Promise<void>
  fetchPluginUsage: (name: string) => Promise<PluginUsage | null>
  installPlugin: (moduleName: string, version?: string) => Promise<boolean>
  removePlugin: (moduleName: string) => Promise<boolean>
  updateConfigPath: (path: string) => Promise<boolean>
  selectPlugin: (plugin: Plugin | null) => void
  clearError: () => void
}

export const usePluginStore = create<PluginState>((set, get) => ({
  // Initial state
  plugins: [],
  configPath: '/etc/traefik/traefik.yml',
  selectedPlugin: null,
  loading: false,
  installing: false,
  removing: false,
  error: null,

  // Fetch all plugins from Traefik API
  fetchPlugins: async () => {
    set({ loading: true, error: null })
    try {
      const plugins = await pluginApi.getAll()
      set({ plugins, loading: false })
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load plugins from Traefik API',
        loading: false,
      })
    }
  },

  // Fetch config path
  fetchConfigPath: async () => {
    try {
      const result = await pluginApi.getConfigPath()
      set({ configPath: result.path || '/etc/traefik/traefik.yml' })
    } catch {
      // Use default path silently
    }
  },

  // Fetch plugin usage details
  fetchPluginUsage: async (name: string) => {
    try {
      const usage = await pluginApi.getUsage(name)
      return usage
    } catch (err) {
      console.error('Failed to fetch plugin usage:', err)
      return null
    }
  },

  // Install plugin
  installPlugin: async (moduleName, version) => {
    set({ installing: true, error: null })
    try {
      const response = await pluginApi.install({ moduleName, version })

      // Update local state to mark plugin as installed
      set((state) => ({
        plugins: state.plugins.map((p) =>
          p.moduleName === moduleName || p.name === response.pluginKey
            ? { ...p, isInstalled: true, installedVersion: version, status: 'configured' as const }
            : p
        ),
        installing: false,
      }))

      // Refresh plugins list to get latest state
      setTimeout(() => get().fetchPlugins(), 1000)

      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to install plugin',
        installing: false,
      })
      return false
    }
  },

  // Remove plugin
  removePlugin: async (moduleName) => {
    set({ removing: true, error: null })
    try {
      await pluginApi.remove({ moduleName })

      // Update local state to mark plugin as not installed
      set((state) => ({
        plugins: state.plugins.map((p) =>
          p.moduleName === moduleName
            ? { ...p, isInstalled: false, installedVersion: undefined, status: 'not_loaded' as const }
            : p
        ),
        removing: false,
      }))

      // Refresh plugins list to get latest state
      setTimeout(() => get().fetchPlugins(), 1000)

      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to remove plugin',
        removing: false,
      })
      return false
    }
  },

  // Update config path
  updateConfigPath: async (path) => {
    set({ error: null })
    try {
      await pluginApi.updateConfigPath(path)
      set({ configPath: path })
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to update config path',
      })
      return false
    }
  },

  // Select a plugin for viewing details
  selectPlugin: (plugin) => {
    set({ selectedPlugin: plugin })
  },

  // Clear error
  clearError: () => set({ error: null }),
}))
