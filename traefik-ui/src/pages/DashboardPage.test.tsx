import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import DashboardPage from './DashboardPage'

const overview = vi.fn()
const ping = vi.fn()
const version = vi.fn()
const routesList = vi.fn()

vi.mock('@/api/client', () => ({
  api: {
    traefik: {
      overview: () => overview(),
      ping: () => ping(),
    },
    manager: {
      version: () => version(),
    },
    routes: {
      list: () => routesList(),
    },
  },
}))

function renderPage() {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })

  return render(
    <QueryClientProvider client={client}>
      <DashboardPage />
    </QueryClientProvider>,
  )
}

describe('DashboardPage', () => {
  beforeEach(() => {
    overview.mockResolvedValue({ http: {}, tcp: {}, udp: {} })
    ping.mockResolvedValue({ ok: true, latency_ms: 12 })
    version.mockResolvedValue({ version: 'v1.2.3', repo: 'org/repo' })
    routesList.mockResolvedValue({
      apps: [
        {
          id: 'r1',
          name: 'whoami',
          service_name: 'whoami-svc',
          rule: 'Host(`a.example.com`)',
          target: 'http://whoami:80',
          middlewares: ['auth', 'headers'],
          entryPoints: ['websecure'],
          protocol: 'http',
          tls: true,
          enabled: true,
          configFile: 'dynamic.yml',
          provider: 'file',
        },
      ],
      middlewares: [{ name: 'auth' }, { name: 'headers' }],
    })
  })

  it('renders structured dashboard widgets without raw payload block', async () => {
    renderPage()
    expect(await screen.findByText('Traffic Snapshot')).toBeInTheDocument()
    expect(screen.getByText('Runtime Health')).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.getByText('Reachable')).toBeInTheDocument()
    })
    expect(screen.queryByText('Raw payload returned by the configured Traefik API.')).not.toBeInTheDocument()
  })

  it('shows unreachable status when ping fails', async () => {
    ping.mockResolvedValue({ ok: false, latency_ms: null })
    renderPage()
    expect(await screen.findByText('Unreachable')).toBeInTheDocument()
  })
})
