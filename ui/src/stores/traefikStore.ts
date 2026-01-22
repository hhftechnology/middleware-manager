import { create } from 'zustand'
import { traefikApi } from '@/services/api'
import type {
  TraefikOverview,
  TraefikVersion,
  TraefikEntrypoint,
  HTTPRouter,
  TCPRouter,
  UDPRouter,
  HTTPService,
  TCPService,
  UDPService,
  HTTPMiddleware,
  TCPMiddleware,
} from '@/types'

interface TraefikState {
  // Data
  overview: TraefikOverview | null
  version: TraefikVersion | null
  entrypoints: TraefikEntrypoint[]
  httpRouters: HTTPRouter[]
  tcpRouters: TCPRouter[]
  udpRouters: UDPRouter[]
  httpServices: HTTPService[]
  tcpServices: TCPService[]
  udpServices: UDPService[]
  httpMiddlewares: HTTPMiddleware[]
  tcpMiddlewares: TCPMiddleware[]

  // Loading states
  isLoading: boolean
  isLoadingOverview: boolean
  isLoadingRouters: boolean
  isLoadingServices: boolean
  isLoadingMiddlewares: boolean
  error: string | null

  // Actions
  fetchOverview: () => Promise<void>
  fetchVersion: () => Promise<void>
  fetchEntrypoints: () => Promise<void>
  fetchRouters: (type?: 'http' | 'tcp' | 'udp' | 'all') => Promise<void>
  fetchServices: (type?: 'http' | 'tcp' | 'udp' | 'all') => Promise<void>
  fetchMiddlewares: (type?: 'http' | 'tcp' | 'all') => Promise<void>
  fetchFullData: () => Promise<void>
  clearError: () => void
  reset: () => void
}

const initialState = {
  overview: null,
  version: null,
  entrypoints: [],
  httpRouters: [],
  tcpRouters: [],
  udpRouters: [],
  httpServices: [],
  tcpServices: [],
  udpServices: [],
  httpMiddlewares: [],
  tcpMiddlewares: [],
  isLoading: false,
  isLoadingOverview: false,
  isLoadingRouters: false,
  isLoadingServices: false,
  isLoadingMiddlewares: false,
  error: null,
}

export const useTraefikStore = create<TraefikState>((set) => ({
  ...initialState,

  fetchOverview: async () => {
    set({ isLoadingOverview: true, error: null })
    try {
      const overview = await traefikApi.getOverview()
      set({ overview, isLoadingOverview: false })
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to fetch overview',
        isLoadingOverview: false,
      })
    }
  },

  fetchVersion: async () => {
    try {
      const version = await traefikApi.getVersion()
      set({ version })
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to fetch version',
      })
    }
  },

  fetchEntrypoints: async () => {
    try {
      const entrypoints = await traefikApi.getEntrypoints()
      set({ entrypoints })
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to fetch entrypoints',
      })
    }
  },

  fetchRouters: async (type) => {
    set({ isLoadingRouters: true, error: null })
    try {
      if (type === 'all' || !type) {
        const data = await traefikApi.getAllRouters()
        set({
          httpRouters: data.http || [],
          tcpRouters: data.tcp || [],
          udpRouters: data.udp || [],
          isLoadingRouters: false,
        })
      } else if (type === 'http') {
        const routers = await traefikApi.getHTTPRouters()
        set({ httpRouters: routers || [], isLoadingRouters: false })
      } else if (type === 'tcp') {
        const routers = await traefikApi.getTCPRouters()
        set({ tcpRouters: routers || [], isLoadingRouters: false })
      } else if (type === 'udp') {
        const routers = await traefikApi.getUDPRouters()
        set({ udpRouters: routers || [], isLoadingRouters: false })
      }
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to fetch routers',
        isLoadingRouters: false,
      })
    }
  },

  fetchServices: async (type) => {
    set({ isLoadingServices: true, error: null })
    try {
      if (type === 'all' || !type) {
        const data = await traefikApi.getAllServices()
        set({
          httpServices: data.http || [],
          tcpServices: data.tcp || [],
          udpServices: data.udp || [],
          isLoadingServices: false,
        })
      } else if (type === 'http') {
        const services = await traefikApi.getHTTPServices()
        set({ httpServices: services || [], isLoadingServices: false })
      } else if (type === 'tcp') {
        const services = await traefikApi.getTCPServices()
        set({ tcpServices: services || [], isLoadingServices: false })
      } else if (type === 'udp') {
        const services = await traefikApi.getUDPServices()
        set({ udpServices: services || [], isLoadingServices: false })
      }
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to fetch services',
        isLoadingServices: false,
      })
    }
  },

  fetchMiddlewares: async (type) => {
    set({ isLoadingMiddlewares: true, error: null })
    try {
      if (type === 'all' || !type) {
        const data = await traefikApi.getAllMiddlewares()
        set({
          httpMiddlewares: data.http || [],
          tcpMiddlewares: data.tcp || [],
          isLoadingMiddlewares: false,
        })
      } else if (type === 'http') {
        const middlewares = await traefikApi.getHTTPMiddlewares()
        set({ httpMiddlewares: middlewares || [], isLoadingMiddlewares: false })
      } else if (type === 'tcp') {
        const middlewares = await traefikApi.getTCPMiddlewares()
        set({ tcpMiddlewares: middlewares || [], isLoadingMiddlewares: false })
      }
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to fetch middlewares',
        isLoadingMiddlewares: false,
      })
    }
  },

  fetchFullData: async () => {
    set({ isLoading: true, error: null })
    try {
      const data = await traefikApi.getFullData()
      set({
        httpRouters: data.httpRouters || [],
        tcpRouters: data.tcpRouters || [],
        udpRouters: data.udpRouters || [],
        httpServices: data.httpServices || [],
        tcpServices: data.tcpServices || [],
        udpServices: data.udpServices || [],
        httpMiddlewares: data.httpMiddlewares || [],
        tcpMiddlewares: data.tcpMiddlewares || [],
        overview: data.overview || null,
        version: data.version || null,
        entrypoints: data.entrypoints || [],
        isLoading: false,
      })
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to fetch Traefik data',
        isLoading: false,
      })
    }
  },

  clearError: () => set({ error: null }),

  reset: () => set(initialState),
}))

// Selectors
export const selectTotalRouters = (state: TraefikState) =>
  state.httpRouters.length + state.tcpRouters.length + state.udpRouters.length

export const selectTotalServices = (state: TraefikState) =>
  state.httpServices.length + state.tcpServices.length + state.udpServices.length

export const selectTotalMiddlewares = (state: TraefikState) =>
  state.httpMiddlewares.length + state.tcpMiddlewares.length
