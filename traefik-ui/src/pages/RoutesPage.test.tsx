import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import RoutesPage from './RoutesPage'

const routesList = vi.fn()
const configsList = vi.fn()
const routesToggle = vi.fn()
const routesCreate = vi.fn()
const routesUpdate = vi.fn()
const routesDelete = vi.fn()

vi.mock('@/api/client', () => ({
  api: {
    routes: {
      list: () => routesList(),
      create: (payload: unknown) => routesCreate(payload),
      update: (id: string, payload: unknown) => routesUpdate(id, payload),
      delete: (id: string) => routesDelete(id),
      toggle: (id: string, enable: boolean) => routesToggle(id, enable),
    },
    configs: {
      list: () => configsList(),
    },
  },
}))

function renderPage() {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return render(
    <QueryClientProvider client={client}>
      <RoutesPage />
    </QueryClientProvider>,
  )
}

describe('RoutesPage', () => {
  beforeEach(() => {
    routesList.mockResolvedValue({
      apps: [
        {
          id: 'dynamic.yml::whoami',
          name: 'whoami',
          rule: 'Host(`whoami.example.com`)',
          service_name: 'whoami-service',
          target: 'http://whoami:80',
          middlewares: ['security-headers'],
          entryPoints: ['websecure'],
          protocol: 'http',
          tls: true,
          enabled: true,
          configFile: 'dynamic.yml',
          provider: 'file',
          certResolver: 'cloudflare',
          passHostHeader: true,
          insecureSkipVerify: false,
        },
      ],
      middlewares: [],
    })
    configsList.mockResolvedValue({
      files: [{ label: 'dynamic.yml', path: '/tmp/dynamic.yml' }],
      configDirSet: false,
    })
    routesToggle.mockResolvedValue({ ok: true })
    routesCreate.mockResolvedValue({ ok: true })
    routesUpdate.mockResolvedValue({ ok: true })
    routesDelete.mockResolvedValue({ ok: true })
    vi.stubGlobal('alert', vi.fn())
  })

  it('renders routes and toggles one off', async () => {
    renderPage()

    await screen.findByText('whoami')
    fireEvent.click(screen.getByRole('button', { name: 'Disable' }))

    await waitFor(() => {
      expect(routesToggle).toHaveBeenCalledWith('dynamic.yml::whoami', false)
    })
  })
})
