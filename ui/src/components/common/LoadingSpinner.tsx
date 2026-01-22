import { Spinner, LoadingOverlay, LoadingCard } from '@/components/ui/spinner'

export { Spinner, LoadingOverlay, LoadingCard }

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg'
  className?: string
  message?: string
}

export function LoadingSpinner({ size = 'md', className, message }: LoadingSpinnerProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-2">
      <Spinner size={size} className={className} />
      {message && (
        <p className="text-sm text-muted-foreground">{message}</p>
      )}
    </div>
  )
}

export function PageLoader({ message = 'Loading...' }: { message?: string }) {
  return (
    <div className="flex h-[50vh] items-center justify-center">
      <LoadingSpinner size="lg" message={message} />
    </div>
  )
}
