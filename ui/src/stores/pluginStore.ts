import { create } from 'zustand'
import { pluginApi } from '@/services/api'
import type { Plugin } from '@/types'

interface PluginState {
  // Data
  plugins: Plugin[]
  configPath: string

  // Loading states
  loading: boolean
  installing: boolean
  removing: boolean

  // Error state
  error: string | null

  // Actions
  fetchPlugins: () => Promise<void>
  fetchConfigPath: () => Promise<void>
  installPlugin: (moduleName: string, version: string) => Promise<boolean>
  removePlugin: (moduleName: string) => Promise<boolean>
  updateConfigPath: (path: string) => Promise<boolean>
  clearError: () => void
}

export const usePluginStore = create<PluginState>((set) => ({
  // Initial state
  plugins: [],
  configPath: '/etc/traefik/traefik.yml',
  loading: false,
  installing: false,
  removing: false,
  error: null,

  // Fetch all plugins
  fetchPlugins: async () => {
    set({ loading: true, error: null })
    try {
      const plugins = await pluginApi.getAll()
      set({ plugins, loading: false })
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load plugins',
        loading: false,
      })
    }
  },

  // Fetch config path
  fetchConfigPath: async () => {
    try {
      const result = await pluginApi.getConfigPath()
      set({ configPath: result.path })
    } catch (err) {
      // Silently fail - use default
    }
  },

  // Install plugin
  installPlugin: async (moduleName, version) => {
    set({ installing: true, error: null })
    try {
      await pluginApi.install({ moduleName, version })
      // Update local state to mark plugin as installed
      set((state) => ({
        plugins: state.plugins.map((p) =>
          p.moduleName === moduleName
            ? { ...p, installed: true, installedVersion: version }
            : p
        ),
        installing: false,
      }))
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
            ? { ...p, installed: false, installedVersion: undefined }
            : p
        ),
        removing: false,
      }))
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

  // Clear error
  clearError: () => set({ error: null }),
}))
