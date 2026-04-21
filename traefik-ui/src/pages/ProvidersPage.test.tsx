import { fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'
import ProvidersPage from './ProvidersPage'

const toastMock = vi.fn()

vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: toastMock,
  }),
}))

function renderPage(path = '/providers/docker') {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/providers/:provider" element={<ProvidersPage />} />
        <Route path="/providers" element={<ProvidersPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

const providers = [
  { key: 'docker', title: 'Docker' },
  { key: 'file-external', title: 'File (External)' },
  { key: 'http-provider', title: 'HTTP Provider' },
  { key: 'swarm', title: 'Swarm' },
  { key: 'ecs', title: 'ECS' },
  { key: 'kubernetes', title: 'Kubernetes' },
  { key: 'consul', title: 'Consul' },
  { key: 'consulcatalog', title: 'Consul Catalog' },
  { key: 'etcd', title: 'Etcd' },
  { key: 'redis', title: 'Redis' },
  { key: 'zookeeper', title: 'ZooKeeper' },
  { key: 'nomad', title: 'Nomad' },
]

describe('ProvidersPage', () => {
  it('renders provider tabs and validates required fields for each provider', async () => {
    for (const provider of providers) {
      const view = renderPage(`/providers/${provider.key}`)
      fireEvent.click(await screen.findByRole('button', { name: `Save ${provider.title}` }))
      expect(toastMock).toHaveBeenCalledWith(
        expect.objectContaining({
          title: `${provider.title} validation failed`,
          variant: 'destructive',
        }),
      )
      view.unmount()
    }
  })

  it('accepts valid provider draft and shows staged save message', async () => {
    renderPage('/providers/docker')

    fireEvent.change(await screen.findByLabelText('Endpoint *'), {
      target: { value: 'unix:///var/run/docker.sock' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'Save Docker' }))

    expect(toastMock).toHaveBeenCalledWith(
      expect.objectContaining({
        title: 'Docker configuration prepared',
      }),
    )
  })
})
