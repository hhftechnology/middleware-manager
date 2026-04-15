import { createContext, useContext, useState } from 'react'
import { useMountEffect } from '@/hooks/useMountEffect'

type Theme = 'dark' | 'light' | 'system'

interface ThemeProviderProps {
  children: React.ReactNode
  defaultTheme?: Theme
  storageKey?: string
}

interface ThemeProviderState {
  theme: Theme
  setTheme: (theme: Theme) => void
  resolvedTheme: 'dark' | 'light'
}

const ThemeProviderContext = createContext<ThemeProviderState | undefined>(undefined)

function resolve(theme: Theme): 'dark' | 'light' {
  if (theme !== 'system') return theme
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function apply(next: 'dark' | 'light') {
  const root = window.document.documentElement
  root.classList.remove('light', 'dark')
  root.classList.add(next)
}

export function ThemeProvider({
  children,
  defaultTheme = 'system',
  storageKey = 'traefik-manager-theme',
}: ThemeProviderProps) {
  const [theme, setThemeState] = useState<Theme>(() => {
    if (typeof window === 'undefined') return defaultTheme
    return (localStorage.getItem(storageKey) as Theme) || defaultTheme
  })
  const [resolvedTheme, setResolvedTheme] = useState<'dark' | 'light'>(() =>
    typeof window === 'undefined' ? 'light' : resolve(theme),
  )

  useMountEffect(() => {
    apply(resolve(theme))
    setResolvedTheme(resolve(theme))

    const media = window.matchMedia('(prefers-color-scheme: dark)')
    const handler = () => {
      const current = (localStorage.getItem(storageKey) as Theme) || defaultTheme
      if (current !== 'system') return
      const next = resolve('system')
      apply(next)
      setResolvedTheme(next)
    }
    media.addEventListener('change', handler)
    return () => media.removeEventListener('change', handler)
  })

  const setTheme = (next: Theme) => {
    localStorage.setItem(storageKey, next)
    setThemeState(next)
    const r = resolve(next)
    apply(r)
    setResolvedTheme(r)
  }

  return (
    <ThemeProviderContext.Provider value={{ theme, setTheme, resolvedTheme }}>
      {children}
    </ThemeProviderContext.Provider>
  )
}

export function useTheme() {
  const context = useContext(ThemeProviderContext)
  if (!context) throw new Error('useTheme must be used within ThemeProvider')
  return context
}
