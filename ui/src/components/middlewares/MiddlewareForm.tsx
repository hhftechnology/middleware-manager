import { useEffect, useState } from 'react'
import { useMiddlewareStore } from '@/stores/middlewareStore'
import { useAppStore } from '@/stores/appStore'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { PageLoader } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { InlineError } from '@/components/common/ErrorMessage'
import { ArrowLeft, Save, Loader2 } from 'lucide-react'
import { MIDDLEWARE_TYPE_LABELS } from '@/types'
import type { MiddlewareType } from '@/types'

const MIDDLEWARE_TYPES = Object.keys(MIDDLEWARE_TYPE_LABELS) as MiddlewareType[]

export function MiddlewareForm() {
  const { middlewareId, isEditing, navigateTo } = useAppStore()
  const {
    selectedMiddleware,
    loadingMiddleware,
    saving,
    error,
    fetchMiddleware,
    createMiddleware,
    updateMiddleware,
    clearError,
    clearSelectedMiddleware,
  } = useMiddlewareStore()

  const [name, setName] = useState('')
  const [type, setType] = useState<MiddlewareType>('headers')
  const [configJson, setConfigJson] = useState('{\n  \n}')
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    if (isEditing && middlewareId) {
      fetchMiddleware(middlewareId)
    } else {
      clearSelectedMiddleware()
    }
  }, [isEditing, middlewareId, fetchMiddleware, clearSelectedMiddleware])

  useEffect(() => {
    if (selectedMiddleware && isEditing) {
      setName(selectedMiddleware.name)
      setType(selectedMiddleware.type as MiddlewareType)
      setConfigJson(JSON.stringify(selectedMiddleware.config, null, 2))
    }
  }, [selectedMiddleware, isEditing])

  const validate = (): boolean => {
    const errors: Record<string, string> = {}

    if (!name.trim()) {
      errors.name = 'Name is required'
    }

    if (!type) {
      errors.type = 'Type is required'
    }

    try {
      JSON.parse(configJson)
    } catch {
      errors.config = 'Invalid JSON configuration'
    }

    setValidationErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validate()) {
      return
    }

    const config = JSON.parse(configJson)

    if (isEditing && middlewareId) {
      const success = await updateMiddleware(middlewareId, {
        name,
        type,
        config,
      })
      if (success) {
        navigateTo('middlewares')
      }
    } else {
      const middleware = await createMiddleware({
        name,
        type,
        config,
      })
      if (middleware) {
        navigateTo('middlewares')
      }
    }
  }

  if (loadingMiddleware) {
    return <PageLoader message="Loading middleware..." />
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button variant="ghost" onClick={() => navigateTo('middlewares')}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back
        </Button>
        <div>
          <h1 className="text-2xl font-bold">
            {isEditing ? 'Edit Middleware' : 'Create Middleware'}
          </h1>
          <p className="text-muted-foreground">
            {isEditing
              ? 'Update middleware configuration'
              : 'Create a new Traefik middleware'}
          </p>
        </div>
      </div>

      {error && (
        <ErrorMessage
          message={error}
          onDismiss={clearError}
        />
      )}

      <form onSubmit={handleSubmit}>
        <Card>
          <CardHeader>
            <CardTitle>Middleware Configuration</CardTitle>
            <CardDescription>
              Configure your middleware settings
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Name */}
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                placeholder="my-middleware"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
              {validationErrors.name && (
                <InlineError message={validationErrors.name} />
              )}
            </div>

            {/* Type */}
            <div className="space-y-2">
              <Label htmlFor="type">Type</Label>
              <Select value={type} onValueChange={(v) => setType(v as MiddlewareType)}>
                <SelectTrigger>
                  <SelectValue placeholder="Select middleware type" />
                </SelectTrigger>
                <SelectContent>
                  {MIDDLEWARE_TYPES.map((t) => (
                    <SelectItem key={t} value={t}>
                      {MIDDLEWARE_TYPE_LABELS[t]}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {validationErrors.type && (
                <InlineError message={validationErrors.type} />
              )}
            </div>

            {/* Configuration JSON */}
            <div className="space-y-2">
              <Label htmlFor="config">Configuration (JSON)</Label>
              <Textarea
                id="config"
                placeholder="{}"
                value={configJson}
                onChange={(e) => setConfigJson(e.target.value)}
                className="font-mono min-h-[200px]"
              />
              {validationErrors.config && (
                <InlineError message={validationErrors.config} />
              )}
              <p className="text-sm text-muted-foreground">
                Enter the middleware-specific configuration as JSON. Refer to the{' '}
                <a
                  href="https://doc.traefik.io/traefik/middlewares/overview/"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="underline"
                >
                  Traefik documentation
                </a>{' '}
                for configuration options.
              </p>
            </div>

            {/* Submit */}
            <div className="flex justify-end gap-2 pt-4 border-t">
              <Button
                type="button"
                variant="outline"
                onClick={() => navigateTo('middlewares')}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={saving}>
                {saving ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Saving...
                  </>
                ) : (
                  <>
                    <Save className="h-4 w-4 mr-2" />
                    {isEditing ? 'Update' : 'Create'}
                  </>
                )}
              </Button>
            </div>
          </CardContent>
        </Card>
      </form>
    </div>
  )
}
