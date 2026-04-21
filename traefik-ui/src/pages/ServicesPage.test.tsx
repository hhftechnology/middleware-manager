import { fireEvent, render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, expect, it, vi } from 'vitest'
import ServicesPage from './ServicesPage'

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
      <ServicesPage />
    </QueryClientProvider>,
  )
}

describe('ServicesPage', () => {
  it('aggregates route data into services and opens detail modal', async () => {
    routesList.mockResolvedValue({
      apps: [
        {
          id: 'r1',
          name: 'whoami',
          service_name: 'whoami-svc',
          rule: 'Host(`a.example.com`)',
          target: 'http://whoami:80',
          middlewares: ['mw-a'],
          entryPoints: ['websecure'],
          protocol: 'http',
          tls: true,
          enabled: true,
          configFile: 'dynamic.yml',
          provider: 'file',
        },
        {
          id: 'r2',
          name: 'whoami-alt',
          service_name: 'whoami-svc',
          rule: 'Host(`b.example.com`)',
          target: 'http://whoami:80',
          middlewares: ['mw-b'],
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
    expect(await screen.findByText('whoami-svc')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'View' }))
    expect(screen.getByRole('heading', { name: 'whoami-svc' })).toBeInTheDocument()
  })
})
