import { Link, useLocation } from 'react-router-dom'
import type { NavGroup } from '@/components/common/nav-config'
import { cn } from '@/lib/utils'
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface MobileMenuProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  groups: NavGroup[]
}

export function MobileMenu({ open, onOpenChange, groups }: MobileMenuProps) {
  const location = useLocation()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="left-0 top-0 h-screen max-w-none translate-x-0 translate-y-0 rounded-none border-0 p-0 sm:max-w-none">
        <DialogHeader className="border-b px-5 py-4">
          <DialogTitle>Navigation</DialogTitle>
          <DialogDescription>Browse all manager sections by category.</DialogDescription>
        </DialogHeader>
        <div className="h-[calc(100vh-80px)] overflow-y-auto px-3 py-3">
          <Accordion type="multiple" className="w-full">
            {groups.map((group) => (
              <AccordionItem key={group.key} value={group.key}>
                <AccordionTrigger className="px-2 text-sm font-semibold text-foreground">
                  <span>{group.label}</span>
                </AccordionTrigger>
                <AccordionContent className="pt-1">
                  <div className="grid gap-1">
                    {group.items.map((item) => {
                      const isActive = location.pathname === item.to
                      return (
                        <Link
                          key={item.to}
                          to={item.to}
                          className={cn(
                            'rounded-md px-3 py-2 text-sm transition-colors',
                            isActive
                              ? 'bg-accent text-accent-foreground'
                              : 'text-muted-foreground hover:bg-accent/70 hover:text-foreground',
                          )}
                          onClick={() => onOpenChange(false)}
                        >
                          {item.label}
                        </Link>
                      )
                    })}
                  </div>
                </AccordionContent>
              </AccordionItem>
            ))}
          </Accordion>
        </div>
      </DialogContent>
    </Dialog>
  )
}
