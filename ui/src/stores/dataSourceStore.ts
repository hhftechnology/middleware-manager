import { create } from 'zustand'
import { dataSourceApi } from '@/services/api'
import type { DataSourceInfo, DataSourceConfig } from '@/types'

interface DataSourceState {
  // Data
  dataSources: DataSourceInfo[]
  activeDataSource: DataSourceConfig | null

  // Loading states
  loading: boolean
  testing: boolean

  // Error state
  error: string | null

  // Test result
  testResult: { success: boolean; message?: string } | null

  // Actions
  fetchDataSources: () => Promise<void>
  fetchActiveDataSource: () => Promise<void>
  setActiveDataSource: (name: string) => Promise<boolean>
  updateDataSource: (name: string, config: Partial<DataSourceConfig>) => Promise<boolean>
  testConnection: (name: string) => Promise<boolean>
  clearError: () => void
  clearTestResult: () => void
}

export const useDataSourceStore = create<DataSourceState>((set) => ({
  // Initial state
  dataSources: [],
  activeDataSource: null,
  loading: false,
  testing: false,
  error: null,
  testResult: null,

  // Fetch all data sources
  fetchDataSources: async () => {
    set({ loading: true, error: null })
    try {
      const result = await dataSourceApi.getAll()
      // Ensure we always have an array
      const dataSources = Array.isArray(result) ? result : []
      set({ dataSources, loading: false })
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load data sources',
        loading: false,
        dataSources: [], // Reset to empty array on error
      })
    }
  },

  // Fetch active data source
  fetchActiveDataSource: async () => {
    set({ loading: true, error: null })
    try {
      const activeDataSource = await dataSourceApi.getActive()
      set({ activeDataSource, loading: false })
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load active data source',
        loading: false,
      })
    }
  },

  // Set active data source
  setActiveDataSource: async (name) => {
    set({ loading: true, error: null })
    try {
      await dataSourceApi.setActive(name)
      // Update local state
      set((state) => ({
        dataSources: state.dataSources.map((ds) => ({
          ...ds,
          isActive: ds.name === name,
        })),
        loading: false,
      }))
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to set active data source',
        loading: false,
      })
      return false
    }
  },

  // Update data source configuration
  updateDataSource: async (name, config) => {
    set({ loading: true, error: null })
    try {
      await dataSourceApi.update(name, config)
      set((state) => ({
        dataSources: state.dataSources.map((ds) =>
          ds.name === name ? { ...ds, ...config } : ds
        ),
        loading: false,
      }))
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to update data source',
        loading: false,
      })
      return false
    }
  },

  // Test connection
  testConnection: async (name) => {
    set({ testing: true, error: null, testResult: null })
    try {
      const result = await dataSourceApi.testConnection(name)
      set({
        testResult: { success: result.success, message: result.message || result.error },
        testing: false,
      })
      return result.success
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Connection test failed'
      set({
        testResult: { success: false, message },
        testing: false,
      })
      return false
    }
  },

  // Clear error
  clearError: () => set({ error: null }),

  // Clear test result
  clearTestResult: () => set({ testResult: null }),
}))
