import { useEffect, useState, useMemo } from 'react'
import { useMiddlewareStore } from '@/stores/middlewareStore'
import { useAppStore } from '@/stores/appStore'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
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
import { ArrowLeft, Save, Loader2, GripVertical, X } from 'lucide-react'
import { MIDDLEWARE_TYPE_LABELS } from '@/types'
import type { MiddlewareType, Middleware } from '@/types'

const MIDDLEWARE_TYPES = Object.keys(MIDDLEWARE_TYPE_LABELS) as MiddlewareType[]

// Chain middleware builder component
interface ChainBuilderProps {
  selectedMiddlewares: string[]
  availableMiddlewares: Middleware[]
  onChange: (middlewares: string[]) => void
}

function ChainBuilder({ selectedMiddlewares, availableMiddlewares, onChange }: ChainBuilderProps) {
  // Get middleware names for the selected ones (handles both IDs and names)
  const resolveMiddlewareName = (idOrName: string) => {
    // First try to find by ID
    const byId = availableMiddlewares.find((m) => m.id === idOrName)
    if (byId) return byId.name
    // Then try by name
    const byName = availableMiddlewares.find((m) => m.name === idOrName)
    if (byName) return byName.name
    // Return as-is if not found (external middleware reference)
    return idOrName
  }

  const handleToggle = (middlewareName: string) => {
    if (selectedMiddlewares.includes(middlewareName)) {
      onChange(selectedMiddlewares.filter((m) => m !== middlewareName))
    } else {
      onChange([...selectedMiddlewares, middlewareName])
    }
  }

  const handleRemove = (index: number) => {
    const newList = [...selectedMiddlewares]
    newList.splice(index, 1)
    onChange(newList)
  }

  const handleMoveUp = (index: number) => {
    if (index === 0) return
    const newList = [...selectedMiddlewares]
    const temp = newList[index - 1] as string
    newList[index - 1] = newList[index] as string
    newList[index] = temp
    onChange(newList)
  }

  const handleMoveDown = (index: number) => {
    if (index === selectedMiddlewares.length - 1) return
    const newList = [...selectedMiddlewares]
    const temp = newList[index] as string
    newList[index] = newList[index + 1] as string
    newList[index + 1] = temp
    onChange(newList)
  }

  // Convert any IDs in selectedMiddlewares to names
  const normalizedSelected = selectedMiddlewares.map(resolveMiddlewareName)

  // Filter out chain middlewares to prevent circular references
  const selectableMiddlewares = availableMiddlewares.filter((m) => m.type !== 'chain')

  return (
    <div className="space-y-4">
      {/* Selected middlewares with order controls */}
      <div className="space-y-2">
        <Label>Chain Order (top to bottom = execution order)</Label>
        {normalizedSelected.length === 0 ? (
          <p className="text-sm text-muted-foreground py-2">
            No middlewares selected. Select middlewares below to add to the chain.
          </p>
        ) : (
          <div className="space-y-1 border rounded-md p-2">
            {normalizedSelected.map((name, index) => {
              const mw = availableMiddlewares.find((m) => m.name === name)
              return (
                <div
                  key={`${name}-${index}`}
                  className="flex items-center gap-2 p-2 bg-muted/50 rounded hover:bg-muted"
                >
                  <GripVertical className="h-4 w-4 text-muted-foreground cursor-grab" />
                  <span className="flex-1 font-medium">{name}</span>
                  {mw && (
                    <Badge variant="outline" className="text-xs">
                      {MIDDLEWARE_TYPE_LABELS[mw.type as MiddlewareType] || mw.type}
                    </Badge>
                  )}
                  <div className="flex gap-1">
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => handleMoveUp(index)}
                      disabled={index === 0}
                      className="h-6 w-6 p-0"
                    >
                      ↑
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => handleMoveDown(index)}
                      disabled={index === normalizedSelected.length - 1}
                      className="h-6 w-6 p-0"
                    >
                      ↓
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => handleRemove(index)}
                      className="h-6 w-6 p-0 text-destructive hover:text-destructive"
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>

      {/* Available middlewares to select */}
      <div className="space-y-2">
        <Label>Available Middlewares</Label>
        <div className="border rounded-md p-2 max-h-48 overflow-y-auto space-y-1">
          {selectableMiddlewares.length === 0 ? (
            <p className="text-sm text-muted-foreground py-2">
              No other middlewares available. Create middlewares first to add them to a chain.
            </p>
          ) : (
            selectableMiddlewares.map((mw) => (
              <label
                key={mw.id}
                className="flex items-center gap-3 p-2 rounded hover:bg-muted cursor-pointer"
              >
                <Checkbox
                  checked={normalizedSelected.includes(mw.name)}
                  onCheckedChange={() => handleToggle(mw.name)}
                />
                <span className="flex-1">{mw.name}</span>
                <Badge variant="outline" className="text-xs">
                  {MIDDLEWARE_TYPE_LABELS[mw.type as MiddlewareType] || mw.type}
                </Badge>
              </label>
            ))
          )}
        </div>
      </div>
    </div>
  )
}

export function MiddlewareForm() {
  const { middlewareId, isEditing, navigateTo } = useAppStore()
  const {
    middlewares,
    selectedMiddleware,
    loadingMiddleware,
    saving,
    error,
    fetchMiddlewares,
    fetchMiddleware,
    createMiddleware,
    updateMiddleware,
    clearError,
    clearSelectedMiddleware,
  } = useMiddlewareStore()

  const [name, setName] = useState('')
  const [type, setType] = useState<MiddlewareType>('headers')
  const [configJson, setConfigJson] = useState('{\n  \n}')
  const [chainMiddlewares, setChainMiddlewares] = useState<string[]>([])
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({})

  // Fetch all middlewares for chain builder
  useEffect(() => {
    fetchMiddlewares()
  }, [fetchMiddlewares])

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
      
      // If it's a chain middleware, extract the middlewares array
      if (selectedMiddleware.type === 'chain' && selectedMiddleware.config) {
        const config = selectedMiddleware.config as { middlewares?: string[] }
        setChainMiddlewares(config.middlewares || [])
      }
    }
  }, [selectedMiddleware, isEditing])

  // Sync chain middlewares back to config JSON when they change
  useEffect(() => {
    if (type === 'chain') {
      setConfigJson(JSON.stringify({ middlewares: chainMiddlewares }, null, 2))
    }
  }, [chainMiddlewares, type])

  // Filter out the current middleware from the list (for chain builder)
  const availableMiddlewares = useMemo(() => {
    return middlewares.filter((m) => m.id !== middlewareId)
  }, [middlewares, middlewareId])

  const validate = (): boolean => {
    const errors: Record<string, string> = {}

    if (!name.trim()) {
      errors.name = 'Name is required'
    } else if (!/^[a-z0-9-]+$/.test(name)) {
      errors.name = 'Name must contain only lowercase letters, numbers, and hyphens'
    }

    if (!type) {
      errors.type = 'Type is required'
    }

    // For chain type, validate the chain middlewares
    if (type === 'chain') {
      if (chainMiddlewares.length === 0) {
        errors.config = 'A chain must contain at least one middleware'
      }
    } else {
      try {
        JSON.parse(configJson)
      } catch {
        errors.config = 'Invalid JSON configuration'
      }
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
      <div className="flex items-center gap-4 pb-2 border-b border-border/60">
        <Button variant="ghost" onClick={() => navigateTo('middlewares')}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back
        </Button>
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            {isEditing ? 'Edit Middleware' : 'Create Middleware'}
          </h1>
          <p className="text-sm text-muted-foreground mt-0.5">
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
                onChange={(e) => setName(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))}
              />
              {validationErrors.name && (
                <InlineError message={validationErrors.name} />
              )}
            </div>

            {/* Type */}
            <div className="space-y-2">
              <Label htmlFor="type">Type</Label>
              <Select value={type} onValueChange={(v) => {
                const newType = v as MiddlewareType
                setType(newType)
                // Reset config when type changes
                if (newType === 'chain') {
                  setChainMiddlewares([])
                  setConfigJson('{\n  "middlewares": []\n}')
                } else {
                  setConfigJson('{\n  \n}')
                }
              }}>
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

            {/* Chain Builder (shown only for chain type) */}
            {type === 'chain' ? (
              <div className="space-y-2">
                <ChainBuilder
                  selectedMiddlewares={chainMiddlewares}
                  availableMiddlewares={availableMiddlewares}
                  onChange={setChainMiddlewares}
                />
                {validationErrors.config && (
                  <InlineError message={validationErrors.config} />
                )}
                <p className="text-sm text-muted-foreground">
                  Select and order middlewares to create a chain. Middlewares will be executed in the order shown (top to bottom).
                </p>
              </div>
            ) : (
              /* Configuration JSON (for non-chain types) */
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
            )}

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
