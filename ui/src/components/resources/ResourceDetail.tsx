import { useEffect, useState } from 'react'
import { useResourceStore } from '@/stores/resourceStore'
import { useMiddlewareStore } from '@/stores/middlewareStore'
import { useServiceStore } from '@/stores/serviceStore'
import { useAppStore } from '@/stores/appStore'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { PageLoader } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { ConfirmationModal } from '@/components/common/ConfirmationModal'
import {
  ArrowLeft,
  Plus,
  Trash2,
  Server,
  Layers,
  Settings,
  Pencil,
} from 'lucide-react'
import { parseJSON } from '@/lib/utils'

export function ResourceDetail() {
  const { resourceId, navigateTo } = useAppStore()
  const {
    selectedResource,
    loadingResource,
    error,
    fetchResource,
    assignMiddleware,
    removeMiddleware,
    assignService,
    removeService,
    clearError,
  } = useResourceStore()

  const { middlewares, fetchMiddlewares } = useMiddlewareStore()
  const { services, fetchServices } = useServiceStore()

  const [selectedMiddlewareId, setSelectedMiddlewareId] = useState('')
  const [middlewarePriority, setMiddlewarePriority] = useState('100')
  const [selectedServiceId, setSelectedServiceId] = useState('')
  const [removeMiddlewareModal, setRemoveMiddlewareModal] = useState<string | null>(null)

  useEffect(() => {
    if (resourceId) {
      fetchResource(resourceId)
      fetchMiddlewares()
      fetchServices()
    }
  }, [resourceId, fetchResource, fetchMiddlewares, fetchServices])

  if (loadingResource) {
    return <PageLoader message="Loading resource..." />
  }

  if (!selectedResource) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" onClick={() => navigateTo('resources')}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Resources
        </Button>
        <ErrorMessage
          message="Resource not found"
          onRetry={() => resourceId && fetchResource(resourceId)}
        />
      </div>
    )
  }

  const assignedMiddlewares = selectedResource.middlewares
    ? selectedResource.middlewares.split(',').filter(Boolean)
    : []

  const customHeaders = parseJSON<Record<string, string>>(
    selectedResource.custom_headers,
    {}
  )

  const handleAssignMiddleware = async () => {
    if (selectedMiddlewareId && resourceId) {
      await assignMiddleware(resourceId, {
        middleware_id: selectedMiddlewareId,
        priority: parseInt(middlewarePriority, 10) || 100,
      })
      setSelectedMiddlewareId('')
    }
  }

  const handleRemoveMiddleware = async (middlewareId: string) => {
    if (resourceId) {
      await removeMiddleware(resourceId, middlewareId)
    }
    setRemoveMiddlewareModal(null)
  }

  const handleAssignService = async () => {
    if (selectedServiceId && resourceId) {
      await assignService(resourceId, selectedServiceId)
      setSelectedServiceId('')
    }
  }

  const handleRemoveService = async () => {
    if (resourceId) {
      await removeService(resourceId)
    }
  }

  const availableMiddlewares = middlewares.filter(
    (m) => !assignedMiddlewares.includes(m.id)
  )

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button variant="ghost" onClick={() => navigateTo('resources')}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-2xl font-bold">{selectedResource.host}</h1>
            <div className="flex items-center gap-2 mt-1">
              <Badge
                variant={selectedResource.status === 'active' ? 'success' : 'secondary'}
              >
                {selectedResource.status}
              </Badge>
              <Badge variant="outline">
                {selectedResource.source_type || 'unknown'}
              </Badge>
              {selectedResource.tcp_enabled && (
                <Badge variant="secondary">TCP Enabled</Badge>
              )}
            </div>
          </div>
        </div>
      </div>

      {error && (
        <ErrorMessage
          message={error}
          onDismiss={clearError}
        />
      )}

      <div className="grid gap-6 md:grid-cols-2">
        {/* Configuration Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Settings className="h-5 w-5" />
              Configuration
            </CardTitle>
            <CardDescription>
              Router settings and entrypoints
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label className="text-muted-foreground">Entrypoints</Label>
                <p className="font-medium">{selectedResource.entrypoints || 'websecure'}</p>
              </div>
              <div>
                <Label className="text-muted-foreground">Priority</Label>
                <p className="font-medium">{selectedResource.router_priority || 200}</p>
              </div>
              <div>
                <Label className="text-muted-foreground">Service ID</Label>
                <p className="font-medium">{selectedResource.service_id || 'Not assigned'}</p>
              </div>
              <div>
                <Label className="text-muted-foreground">TLS Domains</Label>
                <p className="font-medium text-sm">
                  {selectedResource.tls_domains || 'None'}
                </p>
              </div>
            </div>

            {selectedResource.tcp_enabled && (
              <div className="border-t pt-4 mt-4">
                <h4 className="font-medium mb-2">TCP Configuration</h4>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <Label className="text-muted-foreground">TCP Entrypoints</Label>
                    <p>{selectedResource.tcp_entrypoints || 'None'}</p>
                  </div>
                  <div>
                    <Label className="text-muted-foreground">SNI Rule</Label>
                    <p>{selectedResource.tcp_sni_rule || 'None'}</p>
                  </div>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Service Assignment Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="h-5 w-5" />
              Service
            </CardTitle>
            <CardDescription>
              Assigned load balancer configuration
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {selectedResource.service_id ? (
              <div className="flex items-center justify-between p-3 border rounded-lg">
                <div>
                  <p className="font-medium">{selectedResource.service_id}</p>
                  <p className="text-sm text-muted-foreground">Currently assigned</p>
                </div>
                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => navigateTo('service-form', selectedResource.service_id)}
                    title="Edit service"
                  >
                    <Pencil className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleRemoveService}
                    title="Remove service"
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              </div>
            ) : (
              <div className="flex gap-2">
                <Select value={selectedServiceId} onValueChange={setSelectedServiceId}>
                  <SelectTrigger className="flex-1">
                    <SelectValue placeholder="Select a service" />
                  </SelectTrigger>
                  <SelectContent>
                    {services.map((service) => (
                      <SelectItem key={service.id} value={service.id}>
                        {service.name} ({service.type})
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Button onClick={handleAssignService} disabled={!selectedServiceId}>
                  <Plus className="h-4 w-4" />
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Middlewares Card */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Layers className="h-5 w-5" />
                Middlewares
              </CardTitle>
              <CardDescription>
                {assignedMiddlewares.length} middleware(s) assigned
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Add Middleware */}
          <div className="flex gap-2">
            <Select value={selectedMiddlewareId} onValueChange={setSelectedMiddlewareId}>
              <SelectTrigger className="flex-1">
                <SelectValue placeholder="Select middleware to add" />
              </SelectTrigger>
              <SelectContent>
                {availableMiddlewares.map((mw) => (
                  <SelectItem key={mw.id} value={mw.id}>
                    {mw.name} ({mw.type})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Input
              type="number"
              placeholder="Priority"
              value={middlewarePriority}
              onChange={(e) => setMiddlewarePriority(e.target.value)}
              className="w-24"
            />
            <Button onClick={handleAssignMiddleware} disabled={!selectedMiddlewareId}>
              <Plus className="h-4 w-4 mr-2" />
              Add
            </Button>
          </div>

          {/* Assigned Middlewares Table */}
          {assignedMiddlewares.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Middleware ID</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {assignedMiddlewares.map((mwId) => {
                  const middleware = middlewares.find((m) => m.id === mwId)
                  return (
                    <TableRow key={mwId}>
                      <TableCell className="font-mono text-sm">{mwId}</TableCell>
                      <TableCell>{middleware?.name || 'Unknown'}</TableCell>
                      <TableCell>
                        <Badge variant="outline">
                          {middleware?.type || 'unknown'}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setRemoveMiddlewareModal(mwId)}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          ) : (
            <div className="text-center py-8 text-muted-foreground border rounded-lg">
              No middlewares assigned
            </div>
          )}
        </CardContent>
      </Card>

      {/* Custom Headers Card */}
      {Object.keys(customHeaders).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Custom Headers</CardTitle>
            <CardDescription>
              Additional headers configured for this resource
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Header Name</TableHead>
                  <TableHead>Value</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {Object.entries(customHeaders).map(([key, value]) => (
                  <TableRow key={key}>
                    <TableCell className="font-mono">{key}</TableCell>
                    <TableCell className="font-mono text-sm">{value}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {/* Remove Middleware Confirmation */}
      <ConfirmationModal
        open={!!removeMiddlewareModal}
        onOpenChange={(open) => !open && setRemoveMiddlewareModal(null)}
        title="Remove Middleware"
        description={`Are you sure you want to remove this middleware from the resource?`}
        confirmLabel="Remove"
        variant="destructive"
        onConfirm={() => removeMiddlewareModal && handleRemoveMiddleware(removeMiddlewareModal)}
      />
    </div>
  )
}
