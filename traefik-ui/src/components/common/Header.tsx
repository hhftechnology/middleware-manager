import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ChevronDown, Layers, Menu } from 'lucide-react'
import { api } from '@/api/client'
import { Button } from '@/components/ui/button'
import { ThemeToggle } from '@/components/theme-toggle'
import { cn } from '@/lib/utils'
import { useUIStore } from '@/store/ui'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { MobileMenu } from '@/components/common/MobileMenu'
import { buildNavigation } from '@/components/common/nav-config'

export function Header() {
  const { sidebarOpen, setSidebarOpen } = useUIStore()
  const navigate = useNavigate()
  const location = useLocation()
  const settingsQuery = useQuery({ queryKey: ['settings'], queryFn: api.settings.get })
  const visibleTabs = settingsQuery.data?.visible_tabs ?? {}
  const navGroups = buildNavigation(visibleTabs)

  function groupIsActive(paths: string[]): boolean {
    return paths.some((path) => {
      if (path === '/') return location.pathname === '/'
      return location.pathname === path || location.pathname.startsWith(`${path}/`)
    })
  }

  return (
    <header className="sticky top-0 z-40 w-full border-b border-border/80 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center">
        <div className="mr-4 flex">
          <Link to="/" className="flex items-center space-x-2.5 transition-opacity hover:opacity-80">
            <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary">
              <Layers className="h-4 w-4 text-primary-foreground" />
            </div>
            <span className="hidden font-semibold tracking-tight sm:inline-block">Traefik Manager</span>
          </Link>
        </div>

        <nav className="hidden flex-1 items-center gap-1 md:flex">
          {navGroups.map((group) => (
            <DropdownMenu key={group.key}>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  className={cn(
                    'h-9 gap-1 px-3 text-sm font-medium',
                    groupIsActive(group.items.map((item) => item.to)) ? 'text-primary' : 'text-muted-foreground',
                  )}
                >
                  {group.label}
                  <ChevronDown className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="w-64">
                <DropdownMenuLabel>{group.label}</DropdownMenuLabel>
                <DropdownMenuSeparator />
                {group.items.map((item) => (
                  <DropdownMenuItem
                    key={item.to}
                    className={cn(
                      'cursor-pointer rounded-md',
                      location.pathname === item.to ? 'bg-accent text-accent-foreground' : '',
                    )}
                    onClick={() => navigate(item.to)}
                  >
                    {item.label}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          ))}
        </nav>

        <div className="ml-auto flex items-center gap-2">
          <ThemeToggle />
          <Button
            variant="ghost"
            size="icon"
            className="md:hidden"
            onClick={() => setSidebarOpen(true)}
            aria-label="Open navigation menu"
          >
            <Menu className="h-4 w-4" />
          </Button>
        </div>

        <MobileMenu open={sidebarOpen} onOpenChange={setSidebarOpen} groups={navGroups} />
      </div>
    </header>
  )
}
