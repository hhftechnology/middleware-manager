import { useAppStore } from '@/stores/appStore'
import { Button } from '@/components/ui/button'
import { Moon, Sun } from 'lucide-react'

export function DarkModeToggle() {
  const { isDarkMode, toggleDarkMode } = useAppStore()

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={toggleDarkMode}
      aria-label={isDarkMode ? 'Switch to light mode' : 'Switch to dark mode'}
    >
      {isDarkMode ? (
        <Sun className="h-5 w-5" />
      ) : (
        <Moon className="h-5 w-5" />
      )}
    </Button>
  )
}
