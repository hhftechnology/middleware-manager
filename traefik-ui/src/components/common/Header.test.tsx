import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useUIStore } from '@/store/ui'
import { Header } from './Header'

const settingsGet = vi.fn()

vi.mock('@/api/client', () => ({
  api: {
    settings: {
      get: () => settingsGet(),
    },
  },
}))

vi.mock('@/components/theme-toggle', () => ({
  ThemeToggle: () => <button type="button">Theme</button>,
}))

function createMatchMedia() {
  return vi.fn().mockImplementation(() => ({
    matches: false,
    media: '',
    onchange: null,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    addListener: vi.fn(),
    removeListener: vi.fn(),
    dispatchEvent: vi.fn(),
  }))
}

function renderHeader(path = '/') {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })

  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[path]}>
        <Routes>
          <Route path="*" element={<Header />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('Header', () => {
  beforeEach(() => {
    useUIStore.setState({ sidebarOpen: false })
    vi.stubGlobal('matchMedia', createMatchMedia())
    vi.stubGlobal('localStorage', {
      getItem: vi.fn(() => 'light'),
      setItem: vi.fn(),
      removeItem: vi.fn(),
      clear: vi.fn(),
      key: vi.fn(),
      length: 0,
    })
    settingsGet.mockResolvedValue({
      visible_tabs: { certs: true, plugins: true, logs: false },
    })
  })

  it('renders grouped top-level navigation', async () => {
    renderHeader('/routes')
    expect(await screen.findByRole('button', { name: /Overview/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Traffic/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Operations/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Configuration/i })).toBeInTheDocument()
  })

  it('filters optional operations by visible tabs', async () => {
    renderHeader('/')
    await waitFor(() => expect(settingsGet).toHaveBeenCalled())
    const operationButtons = await screen.findAllByRole('button', { name: /Operations/i })
    expect(operationButtons.length).toBeGreaterThan(0)
    fireEvent.click(operationButtons[0]!)
    expect(screen.getByText('Certificates')).toBeInTheDocument()
    expect(screen.getByText('Plugins')).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.queryAllByText('Logs').length).toBe(0)
    })
  })

  it('opens mobile menu hierarchy', async () => {
    renderHeader('/')
    fireEvent.click(await screen.findByLabelText('Open navigation menu'))
    expect(screen.getByText('Navigation')).toBeInTheDocument()
    expect(screen.getAllByText('Overview').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Traffic').length).toBeGreaterThan(0)
  })
})
