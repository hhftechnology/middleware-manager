import { fireEvent, render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, expect, it, vi } from 'vitest'
import RouteMapPage from './RouteMapPage'

const routesList = vi.fn()

vi.mock('@/api/client', () => ({
  api: {
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
      <RouteMapPage />
    </QueryClientProvider>,
  )
}

describe('RouteMapPage', () => {
  it('builds node relationships and opens details', async () => {
    routesList.mockResolvedValue({
      apps: [
        {
          id: 'dynamic.yml::whoami',
          name: 'whoami',
          service_name: 'whoami-service',
          rule: 'Host(`whoami.example.com`)',
          target: 'http://whoami:80',
          middlewares: ['security-headers'],
          entryPoints: ['websecure'],
          protocol: 'http',
          tls: true,
          enabled: true,
          configFile: 'dynamic.yml',
          provider: 'file',
        },
      ],
    })

    renderPage()

    expect(await screen.findByText('whoami')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'whoami' }))
    expect(screen.getByRole('heading', { name: 'whoami' })).toBeInTheDocument()
    expect(screen.getByText('route')).toBeInTheDocument()
  })
})
