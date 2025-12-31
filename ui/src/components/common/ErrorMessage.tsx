import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { AlertCircle, RefreshCw, X } from 'lucide-react'

interface ErrorMessageProps {
  title?: string
  message: string
  onRetry?: () => void
  onDismiss?: () => void
}

export function ErrorMessage({
  title = 'Error',
  message,
  onRetry,
  onDismiss,
}: ErrorMessageProps) {
  return (
    <Alert variant="destructive" className="relative">
      <AlertCircle className="h-4 w-4" />
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription className="flex items-center justify-between">
        <span>{message}</span>
        <div className="flex items-center gap-2 ml-4">
          {onRetry && (
            <Button
              variant="outline"
              size="sm"
              onClick={onRetry}
              className="gap-1"
            >
              <RefreshCw className="h-3 w-3" />
              Retry
            </Button>
          )}
          {onDismiss && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onDismiss}
              className="h-6 w-6 p-0"
            >
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
      </AlertDescription>
    </Alert>
  )
}

interface InlineErrorProps {
  message: string
}

export function InlineError({ message }: InlineErrorProps) {
  return (
    <p className="text-sm text-destructive flex items-center gap-1">
      <AlertCircle className="h-3 w-3" />
      {message}
    </p>
  )
}
