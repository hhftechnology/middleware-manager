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
  { id: 'plugin-hub', label: 'Plugin Hub', icon: <Puzzle className="h-4 w-4" /> },
]

export function Header() {
  const { page, navigateTo, setShowSettings } = useAppStore()

  return (
    <header className="sticky top-0 z-40 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center">
        {/* Logo */}
        <div className="mr-4 flex">
          <button
            onClick={() => navigateTo('dashboard')}
            className="mr-6 flex items-center space-x-2"
          >
            <Layers className="h-6 w-6" />
            <span className="hidden font-bold sm:inline-block">
              Middleware Manager
            </span>
          </button>
        </div>

        {/* Desktop Navigation */}
        <nav className="hidden md:flex flex-1 items-center space-x-1">
          {navItems.map((item) => (
            <Button
              key={item.id}
              variant={page === item.id ? 'secondary' : 'ghost'}
              size="sm"
              onClick={() => navigateTo(item.id)}
              className={cn(
                'gap-2',
                page === item.id && 'bg-secondary'
              )}
            >
              {item.icon}
              {item.label}
            </Button>
          ))}
        </nav>

        {/* Right side actions */}
        <div className="flex items-center space-x-2">
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
                      page === item.id && 'bg-secondary'
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
