import { create } from 'zustand'
import { pluginApi } from '@/services/api'
import type { Plugin, PluginUsage, CataloguePlugin } from '@/types'

function derivePluginKey(input?: string): string {
  if (!input) return ''
  const parts = input.split('/')
  let key = parts[parts.length - 1] || input
  key = key.split('@')[0]
  key = key.replace(/\.git$/, '')
  key = key.replace(/-plugin$/, '')
  return key.toLowerCase()
}

function enrichPluginsWithCatalogue(plugins: Plugin[], cataloguePlugins: CataloguePlugin[]): Plugin[] {
  const catalogueMap = new Map<string, CataloguePlugin>()
  cataloguePlugins.forEach((cp) => {
    const key = derivePluginKey(cp.import) || derivePluginKey(cp.name) || cp.id
    if (key) {
      catalogueMap.set(key, cp)
    }
  })

  return plugins.map((plugin) => {
    const keyCandidates = [
      derivePluginKey(plugin.moduleName),
      derivePluginKey(plugin.name),
      derivePluginKey(plugin.displayName),
    ].filter(Boolean)

    const matchedCatalogue = keyCandidates
      .map((key) => (key ? catalogueMap.get(key) : undefined))
      .find(Boolean)

    const version =
      plugin.version ||
      matchedCatalogue?.latestVersion ||
      plugin.installedVersion ||
      ''

    const enriched: Plugin = {
      ...plugin,
      displayName: plugin.displayName || matchedCatalogue?.displayName || matchedCatalogue?.name,
      description: plugin.description || matchedCatalogue?.summary,
      summary: plugin.summary || matchedCatalogue?.summary,
      author: plugin.author || matchedCatalogue?.author,
      version,
      iconUrl: plugin.iconUrl || matchedCatalogue?.iconUrl,
      installSource: plugin.installSource ?? 'config',
    }
    return enriched
  })
}

interface PluginState {
  // Data
  plugins: Plugin[]
  cataloguePlugins: CataloguePlugin[]
  configPath: string
  selectedPlugin: Plugin | null
  selectedCataloguePlugin: CataloguePlugin | null

  // Loading states
  loading: boolean
  loadingCatalogue: boolean
  installing: boolean
  removing: boolean

  // Error state
  error: string | null

  // Restart warning state
  showRestartWarning: boolean
  lastInstalledPlugin: string | null

  // Actions
  fetchPlugins: () => Promise<void>
  fetchCatalogue: () => Promise<void>
  fetchConfigPath: () => Promise<void>
  fetchPluginUsage: (name: string) => Promise<PluginUsage | null>
  installPlugin: (moduleName: string, version?: string) => Promise<boolean>
  removePlugin: (moduleName: string) => Promise<boolean>
  updateConfigPath: (path: string) => Promise<boolean>
  selectPlugin: (plugin: Plugin | null) => void
  selectCataloguePlugin: (plugin: CataloguePlugin | null) => void
  clearError: () => void
  dismissRestartWarning: () => void
}

export const usePluginStore = create<PluginState>((set, get) => ({
  // Initial state
  plugins: [],
  cataloguePlugins: [],
  configPath: '/etc/traefik/traefik.yml',
  selectedPlugin: null,
  selectedCataloguePlugin: null,
  loading: false,
  loadingCatalogue: false,
  installing: false,
  removing: false,
  error: null,
  showRestartWarning: false,
  lastInstalledPlugin: null,

  // Fetch all plugins from Traefik API
  fetchPlugins: async () => {
    set({ loading: true, error: null })
    try {
      const plugins = await pluginApi.getAll()
      const cataloguePlugins = get().cataloguePlugins
      const enrichedPlugins = enrichPluginsWithCatalogue(plugins, cataloguePlugins)
      set({ plugins: enrichedPlugins, loading: false })
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load plugins from Traefik API',
        loading: false,
      })
    }
  },

  // Fetch plugin catalogue from plugins.traefik.io
  fetchCatalogue: async () => {
    set({ loadingCatalogue: true, error: null })
    try {
      const cataloguePlugins = await pluginApi.getCatalogue()
      set((state) => {
        const safeCatalogue = Array.isArray(cataloguePlugins) ? cataloguePlugins : []
        return {
          cataloguePlugins: safeCatalogue,
          plugins: enrichPluginsWithCatalogue(state.plugins, safeCatalogue),
          loadingCatalogue: false,
        }
      })
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load plugin catalogue from plugins.traefik.io',
        loadingCatalogue: false,
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
        plugins: enrichPluginsWithCatalogue(
          state.plugins.some((p) => p.moduleName === moduleName || p.name === response.pluginKey)
            ? state.plugins.map((p) =>
                p.moduleName === moduleName || p.name === response.pluginKey
                  ? {
                      ...p,
                      isInstalled: true,
                      installedVersion: version,
                      status: 'configured' as const,
                      installSource: 'catalogue',
                    }
                  : p
              )
            : [
                ...state.plugins,
                {
                  name: response.pluginKey || moduleName,
                  moduleName: response.moduleName || moduleName,
                  version: response.version || version || '',
                  type: 'middleware',
                  description: '',
                  summary: '',
                  author: '',
                  homepage: '',
                  status: 'configured',
                  isInstalled: true,
                  installedVersion: version,
                  usageCount: 0,
                  installSource: 'catalogue',
                },
              ],
          state.cataloguePlugins
        ),
        installing: false,
        // Show restart warning after successful installation
        showRestartWarning: true,
        lastInstalledPlugin: response.pluginKey || moduleName,
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

  // Select a catalogue plugin for viewing details
  selectCataloguePlugin: (plugin) => {
    set({ selectedCataloguePlugin: plugin })
  },

  // Clear error
  clearError: () => set({ error: null }),

  // Dismiss restart warning
  dismissRestartWarning: () => set({ showRestartWarning: false, lastInstalledPlugin: null }),
}))
