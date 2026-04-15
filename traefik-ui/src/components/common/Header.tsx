import { NavLink } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import {
  LayoutDashboard,
  Globe,
  Layers,
  Settings as SettingsIcon,
  Shield,
  FileText,
  Puzzle,
  Archive,
  Menu,
} from 'lucide-react'
import { api } from '@/api/client'
import { Button } from '@/components/ui/button'
import { ThemeToggle } from '@/components/theme-toggle'
import { cn } from '@/lib/utils'
import { useUIStore } from '@/store/ui'

interface NavItem {
  to: string
  label: string
  icon: React.ReactNode
  key?: string
}

const baseItems: NavItem[] = [
  { to: '/', label: 'Dashboard', icon: <LayoutDashboard className="h-4 w-4" /> },
  { to: '/routes', label: 'Routes', icon: <Globe className="h-4 w-4" /> },
  { to: '/middlewares', label: 'Middlewares', icon: <Layers className="h-4 w-4" /> },
  { to: '/backups', label: 'Backups', icon: <Archive className="h-4 w-4" /> },
  { to: '/settings', label: 'Settings', icon: <SettingsIcon className="h-4 w-4" /> },
]

const optionalItems: NavItem[] = [
  { to: '/certificates', label: 'Certificates', icon: <Shield className="h-4 w-4" />, key: 'certs' },
  { to: '/plugins', label: 'Plugins', icon: <Puzzle className="h-4 w-4" />, key: 'plugins' },
  { to: '/logs', label: 'Logs', icon: <FileText className="h-4 w-4" />, key: 'logs' },
]

export function Header() {
  const { toggleSidebar } = useUIStore()
  const settingsQuery = useQuery({ queryKey: ['settings'], queryFn: api.settings.get })
  const visibleTabs = settingsQuery.data?.visible_tabs ?? {}
  const visibleOptional = optionalItems.filter((item) => (item.key ? visibleTabs[item.key] : true))
  const navItems = [...baseItems, ...visibleOptional]

  return (
    <header className="sticky top-0 z-40 w-full border-b border-border/80 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center">
        <div className="mr-6 flex">
          <NavLink to="/" className="flex items-center space-x-2.5 hover:opacity-80 transition-opacity">
            <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary">
              <Layers className="h-4 w-4 text-primary-foreground" />
            </div>
            <span className="hidden font-semibold sm:inline-block tracking-tight">Traefik Manager</span>
          </NavLink>
        </div>

        <nav className="hidden md:flex flex-1 items-center gap-0.5">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/'}
              className={({ isActive }) =>
                cn(
                  'relative inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
                  isActive
                    ? 'text-primary'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent',
                )
              }
            >
              {({ isActive }) => (
                <>
                  {item.icon}
                  {item.label}
                  {isActive && (
                    <span className="absolute -bottom-[13px] left-2 right-2 h-0.5 rounded-full bg-primary" />
                  )}
                </>
              )}
            </NavLink>
          ))}
        </nav>

        <div className="ml-auto flex items-center gap-2">
          <ThemeToggle />
          <Button
            variant="ghost"
            size="icon"
            className="md:hidden"
            onClick={toggleSidebar}
            aria-label="Toggle menu"
          >
            <Menu className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </header>
  )
}
