import { useAppStore } from '@/stores/appStore'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ThemeToggle } from '@/components/theme-toggle'
import {
  LayoutDashboard,
  Globe,
  Layers,
  Server,
  Puzzle,
  Shield,
  Settings,
  Menu,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Page } from '@/types'

interface NavItem {
  id: Page
  label: string
  icon: React.ReactNode
}

const navItems: NavItem[] = [
  { id: 'dashboard', label: 'Dashboard', icon: <LayoutDashboard className="h-4 w-4" /> },
  { id: 'resources', label: 'Resources', icon: <Globe className="h-4 w-4" /> },
  { id: 'middlewares', label: 'Middlewares', icon: <Layers className="h-4 w-4" /> },
  { id: 'services', label: 'Services', icon: <Server className="h-4 w-4" /> },
  { id: 'plugin-hub', label: 'Plugins', icon: <Puzzle className="h-4 w-4" /> },
  { id: 'security', label: 'Security', icon: <Shield className="h-4 w-4" /> },
]

export function Header() {
  const { page, navigateTo, setShowSettings } = useAppStore()

  return (
    <header className="sticky top-0 z-40 w-full border-b border-border/80 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center">
        {/* Logo */}
        <div className="mr-6 flex">
          <button
            onClick={() => navigateTo('dashboard')}
            className="flex items-center space-x-2.5 hover:opacity-80 transition-opacity"
          >
            <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary">
              <Layers className="h-4 w-4 text-primary-foreground" />
            </div>
            <span className="hidden font-semibold sm:inline-block tracking-tight">
              Middleware Manager
            </span>
          </button>
        </div>

        {/* Desktop Navigation */}
        <nav className="hidden md:flex flex-1 items-center gap-0.5">
          {navItems.map((item) => (
            <button
              key={item.id}
              onClick={() => navigateTo(item.id)}
              className={cn(
                'relative inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
                page === item.id
                  ? 'text-primary'
                  : 'text-muted-foreground hover:text-foreground hover:bg-accent'
              )}
            >
              {item.icon}
              {item.label}
              {page === item.id && (
                <span className="absolute -bottom-[13px] left-2 right-2 h-0.5 rounded-full bg-primary" />
              )}
            </button>
          ))}
        </nav>

        {/* Right side actions */}
        <div className="flex items-center gap-1">
          <ThemeToggle />

          <Button
            variant="ghost"
            size="icon"
            onClick={() => setShowSettings(true)}
            aria-label="Settings"
          >
            <Settings className="h-5 w-5" />
          </Button>

          {/* Mobile Menu */}
          <div className="md:hidden">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon">
                  <Menu className="h-5 w-5" />
                  <span className="sr-only">Toggle menu</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-48">
                <DropdownMenuLabel>Navigation</DropdownMenuLabel>
                <DropdownMenuSeparator />
                {navItems.map((item) => (
                  <DropdownMenuItem
                    key={item.id}
                    onClick={() => navigateTo(item.id)}
                    className={cn(
                      'gap-2 cursor-pointer',
                      page === item.id && 'bg-primary/10 text-primary'
                    )}
                  >
                    {item.icon}
                    {item.label}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>
    </header>
  )
}
