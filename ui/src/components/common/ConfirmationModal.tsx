import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'

interface ConfirmationModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: string
  confirmLabel?: string
  cancelLabel?: string
  variant?: 'default' | 'destructive'
  onConfirm: () => void
  onCancel?: () => void
}

export function ConfirmationModal({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  variant = 'default',
  onConfirm,
  onCancel,
}: ConfirmationModalProps) {
  const handleConfirm = () => {
    onConfirm()
    onOpenChange(false)
  }

  const handleCancel = () => {
    onCancel?.()
    onOpenChange(false)
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription>{description}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={handleCancel}>
            {cancelLabel}
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={handleConfirm}
            className={
              variant === 'destructive'
                ? 'bg-destructive text-destructive-foreground hover:bg-destructive/90'
                : ''
            }
          >
            {confirmLabel}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

// Hook for easy confirmation dialog usage
import { useState, useCallback } from 'react'

interface UseConfirmationOptions {
  title: string
  description: string
  confirmLabel?: string
  cancelLabel?: string
  variant?: 'default' | 'destructive'
}

export function useConfirmation(options: UseConfirmationOptions) {
  const [isOpen, setIsOpen] = useState(false)
  const [resolveRef, setResolveRef] = useState<((value: boolean) => void) | null>(null)

  const confirm = useCallback(() => {
    return new Promise<boolean>((resolve) => {
      setResolveRef(() => resolve)
      setIsOpen(true)
    })
  }, [])

  const handleConfirm = useCallback(() => {
    resolveRef?.(true)
    setIsOpen(false)
  }, [resolveRef])

  const handleCancel = useCallback(() => {
    resolveRef?.(false)
    setIsOpen(false)
  }, [resolveRef])

  const ConfirmationDialog = useCallback(
    () => (
      <ConfirmationModal
        open={isOpen}
        onOpenChange={setIsOpen}
        title={options.title}
        description={options.description}
        confirmLabel={options.confirmLabel}
        cancelLabel={options.cancelLabel}
        variant={options.variant}
        onConfirm={handleConfirm}
        onCancel={handleCancel}
      />
    ),
    [isOpen, options, handleConfirm, handleCancel]
  )

  return { confirm, ConfirmationDialog }
}
