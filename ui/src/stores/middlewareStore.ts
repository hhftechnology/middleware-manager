import { create } from 'zustand'
import { middlewareApi } from '@/services/api'
import type { Middleware, CreateMiddlewareRequest, UpdateMiddlewareRequest } from '@/types'

interface MiddlewareState {
  // Data
  middlewares: Middleware[]
  selectedMiddleware: Middleware | null

  // Loading states
  loading: boolean
  loadingMiddleware: boolean
  saving: boolean

  // Error state
  error: string | null

  // Actions
  fetchMiddlewares: () => Promise<void>
  fetchMiddleware: (id: string) => Promise<void>
  createMiddleware: (data: CreateMiddlewareRequest) => Promise<Middleware | null>
  updateMiddleware: (id: string, data: UpdateMiddlewareRequest) => Promise<boolean>
  deleteMiddleware: (id: string) => Promise<boolean>
  clearError: () => void
  clearSelectedMiddleware: () => void
}

export const useMiddlewareStore = create<MiddlewareState>((set) => ({
  // Initial state
  middlewares: [],
  selectedMiddleware: null,
  loading: false,
  loadingMiddleware: false,
  saving: false,
  error: null,

  // Fetch all middlewares
  fetchMiddlewares: async () => {
    set({ loading: true, error: null })
    try {
      const middlewares = await middlewareApi.getAll()
      set({ middlewares, loading: false })
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load middlewares',
        loading: false,
      })
    }
  },

  // Fetch single middleware
  fetchMiddleware: async (id) => {
    set({ loadingMiddleware: true, error: null })
    try {
      const middleware = await middlewareApi.getById(id)
      set({ selectedMiddleware: middleware, loadingMiddleware: false })
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load middleware',
        loadingMiddleware: false,
      })
    }
  },

  // Create middleware
  createMiddleware: async (data) => {
    set({ saving: true, error: null })
    try {
      const middleware = await middlewareApi.create(data)
      set((state) => ({
        middlewares: [...state.middlewares, middleware],
        saving: false,
      }))
      return middleware
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to create middleware',
        saving: false,
      })
      return null
    }
  },

  // Update middleware
  updateMiddleware: async (id, data) => {
    set({ saving: true, error: null })
    try {
      const updated = await middlewareApi.update(id, data)
      set((state) => ({
        middlewares: state.middlewares.map((m) => (m.id === id ? updated : m)),
        selectedMiddleware: state.selectedMiddleware?.id === id ? updated : state.selectedMiddleware,
        saving: false,
      }))
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to update middleware',
        saving: false,
      })
      return false
    }
  },

  // Delete middleware
  deleteMiddleware: async (id) => {
    set({ loading: true, error: null })
    try {
      await middlewareApi.delete(id)
      set((state) => ({
        middlewares: state.middlewares.filter((m) => m.id !== id),
        selectedMiddleware: state.selectedMiddleware?.id === id ? null : state.selectedMiddleware,
        loading: false,
      }))
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to delete middleware',
        loading: false,
      })
      return false
    }
  },

  // Clear error
  clearError: () => set({ error: null }),

  // Clear selected middleware
  clearSelectedMiddleware: () => set({ selectedMiddleware: null }),
}))
