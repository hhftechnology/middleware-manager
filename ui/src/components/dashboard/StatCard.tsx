import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import { LucideIcon } from 'lucide-react'

type StatColor = 'blue' | 'violet' | 'emerald' | 'cyan' | 'neutral'

const iconColorMap: Record<StatColor, string> = {
  blue: 'bg-rose-800/10 text-rose-800 dark:bg-rose-400/15 dark:text-rose-300',
  violet: 'bg-amber-700/10 text-amber-700 dark:bg-amber-400/15 dark:text-amber-300',
  emerald: 'bg-emerald-700/10 text-emerald-700 dark:bg-emerald-400/15 dark:text-emerald-300',
  cyan: 'bg-stone-600/10 text-stone-600 dark:bg-stone-400/15 dark:text-stone-300',
  neutral: 'bg-muted text-foreground/70',
}

interface StatCardProps {
  title: string
  value: number | string
  description?: string
  icon?: LucideIcon
  color?: StatColor
  className?: string
  onClick?: () => void
}

export function StatCard({
  title,
  value,
  description,
  icon: Icon,
  color = 'neutral',
  className,
  onClick,
}: StatCardProps) {
  return (
    <Card
      className={cn(
        'transition-all duration-150',
        onClick && 'cursor-pointer hover:border-primary/40 hover:shadow-md',
        className
      )}
      onClick={onClick}
    >
      <CardContent className="p-5">
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">{title}</p>
            <p className="text-3xl font-bold tracking-tight">{value}</p>
            {description && (
              <p className="text-xs text-muted-foreground pt-0.5">{description}</p>
            )}
          </div>
          {Icon && (
            <div className={cn('rounded-lg p-2.5', iconColorMap[color])}>
              <Icon className="h-5 w-5" />
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
