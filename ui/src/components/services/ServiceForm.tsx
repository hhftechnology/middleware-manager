import { useEffect, useState } from 'react'
import { useServiceStore } from '@/stores/serviceStore'
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
import { SERVICE_TYPE_LABELS } from '@/types'
import type { ServiceType } from '@/types'

const SERVICE_TYPES = Object.keys(SERVICE_TYPE_LABELS) as ServiceType[]

const DEFAULT_CONFIGS: Record<ServiceType, object> = {
  loadBalancer: {
    servers: [{ url: 'http://localhost:8080' }],
    passHostHeader: true,
  },
  weighted: {
    services: [{ name: 'service1', weight: 1 }],
  },
  mirroring: {
    service: 'main-service',
    mirrors: [{ name: 'mirror-service', percent: 10 }],
  },
  failover: {
    service: 'main-service',
    fallback: 'fallback-service',
  },
}

export function ServiceForm() {
  const { serviceId, isEditing, navigateTo } = useAppStore()
  const {
    selectedService,
    loadingService,
    saving,
    error,
    fetchService,
    createService,
    updateService,
    clearError,
    clearSelectedService,
  } = useServiceStore()

  const [name, setName] = useState('')
  const [type, setType] = useState<ServiceType>('loadBalancer')
  const [configJson, setConfigJson] = useState(
    JSON.stringify(DEFAULT_CONFIGS.loadBalancer, null, 2)
  )
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    if (isEditing && serviceId) {
      fetchService(serviceId)
    } else {
      clearSelectedService()
    }
  }, [isEditing, serviceId, fetchService, clearSelectedService])

  useEffect(() => {
    if (selectedService && isEditing) {
      setName(selectedService.name)
      setType(selectedService.type as ServiceType)
      setConfigJson(JSON.stringify(selectedService.config, null, 2))
    }
  }, [selectedService, isEditing])

  const handleTypeChange = (newType: string) => {
    setType(newType as ServiceType)
    if (!isEditing) {
      setConfigJson(JSON.stringify(DEFAULT_CONFIGS[newType as ServiceType], null, 2))
    }
  }

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

    if (isEditing && serviceId) {
      const success = await updateService(serviceId, {
        name,
        type,
        config,
      })
      if (success) {
        navigateTo('services')
      }
    } else {
      const service = await createService({
        name,
        type,
        config,
      })
      if (service) {
        navigateTo('services')
      }
    }
  }

  if (loadingService) {
    return <PageLoader message="Loading service..." />
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4 pb-2 border-b border-border/60">
        <Button variant="ghost" onClick={() => navigateTo('services')}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back
        </Button>
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            {isEditing ? 'Edit Service' : 'Create Service'}
          </h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            {isEditing
              ? 'Update service configuration'
              : 'Create a new Traefik service'}
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
            <CardTitle>Service Configuration</CardTitle>
            <CardDescription>
              Configure your load balancer or service settings
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Name */}
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                placeholder="my-service"
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
              <Select value={type} onValueChange={handleTypeChange}>
                <SelectTrigger>
                  <SelectValue placeholder="Select service type" />
                </SelectTrigger>
                <SelectContent>
                  {SERVICE_TYPES.map((t) => (
                    <SelectItem key={t} value={t}>
                      {SERVICE_TYPE_LABELS[t]}
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
                className="font-mono min-h-[250px]"
              />
              {validationErrors.config && (
                <InlineError message={validationErrors.config} />
              )}
              <p className="text-sm text-muted-foreground">
                Enter the service-specific configuration as JSON. Refer to the{' '}
                <a
                  href="https://doc.traefik.io/traefik/routing/services/"
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
                onClick={() => navigateTo('services')}
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
