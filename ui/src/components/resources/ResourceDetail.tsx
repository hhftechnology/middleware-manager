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
  Save,
  X,
  Loader2,
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
    updateHTTPConfig,
    updateTLSConfig,
    updateTCPConfig,
    updateRouterPriority,
    clearError,
  } = useResourceStore()

  const { middlewares, fetchMiddlewares } = useMiddlewareStore()
  const { services, fetchServices } = useServiceStore()

  const [selectedMiddlewareId, setSelectedMiddlewareId] = useState('')
  const [middlewarePriority, setMiddlewarePriority] = useState('100')
  const [selectedServiceId, setSelectedServiceId] = useState('')
  const [removeMiddlewareModal, setRemoveMiddlewareModal] = useState<string | null>(null)

  // Configuration editing state
  const [isEditingConfig, setIsEditingConfig] = useState(false)
  const [savingConfig, setSavingConfig] = useState(false)
  const [editEntrypoints, setEditEntrypoints] = useState('')
  const [editPriority, setEditPriority] = useState('')
  const [editTlsDomains, setEditTlsDomains] = useState('')
  const [editTcpEntrypoints, setEditTcpEntrypoints] = useState('')
  const [editTcpSniRule, setEditTcpSniRule] = useState('')

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

  // Parse middleware assignments from the format "id:name:priority,id:name:priority,..."
  const assignedMiddlewares = selectedResource.middlewares
    ? selectedResource.middlewares.split(',').filter(Boolean).map(mwStr => {
        const parts = mwStr.split(':')
        return {
          id: parts[0] || '',
          name: parts[1] || 'Unknown',
          priority: parseInt(parts[2], 10) || 100,
        }
      }).sort((a, b) => b.priority - a.priority) // Sort by priority descending (higher priority first)
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

  const assignedMiddlewareIds = assignedMiddlewares.map(m => m.id)
  const availableMiddlewares = middlewares.filter(
    (m) => !assignedMiddlewareIds.includes(m.id)
  )

  // Start editing configuration
  const handleStartEditConfig = () => {
    if (selectedResource) {
      setEditEntrypoints(selectedResource.entrypoints || 'websecure')
      setEditPriority(String(selectedResource.router_priority || 200))
      setEditTlsDomains(selectedResource.tls_domains || '')
      setEditTcpEntrypoints(selectedResource.tcp_entrypoints || '')
      setEditTcpSniRule(selectedResource.tcp_sni_rule || '')
      setIsEditingConfig(true)
    }
  }

  // Cancel editing
  const handleCancelEditConfig = () => {
    setIsEditingConfig(false)
  }

  // Save configuration changes
  const handleSaveConfig = async () => {
    if (!resourceId) return

    setSavingConfig(true)
    try {
      // Update HTTP config (entrypoints)
      await updateHTTPConfig(resourceId, { entrypoints: editEntrypoints })

      // Update router priority
      const newPriority = parseInt(editPriority, 10) || 200
      await updateRouterPriority(resourceId, newPriority)

      // Update TLS config
      if (editTlsDomains !== selectedResource?.tls_domains) {
        await updateTLSConfig(resourceId, { tls_domains: editTlsDomains })
      }

      // Update TCP config if TCP is enabled
      if (selectedResource?.tcp_enabled) {
        await updateTCPConfig(resourceId, {
          tcp_entrypoints: editTcpEntrypoints,
          tcp_sni_rule: editTcpSniRule,
        })
      }

      setIsEditingConfig(false)
    } finally {
      setSavingConfig(false)
    }
  }

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
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <Settings className="h-5 w-5" />
                  Configuration
                </CardTitle>
                <CardDescription>
                  Router settings and entrypoints
                </CardDescription>
              </div>
              {!isEditingConfig ? (
                <Button variant="outline" size="sm" onClick={handleStartEditConfig}>
                  <Pencil className="h-4 w-4 mr-2" />
                  Edit
                </Button>
              ) : (
                <div className="flex gap-2">
                  <Button variant="outline" size="sm" onClick={handleCancelEditConfig} disabled={savingConfig}>
                    <X className="h-4 w-4 mr-2" />
                    Cancel
                  </Button>
                  <Button size="sm" onClick={handleSaveConfig} disabled={savingConfig}>
                    {savingConfig ? (
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    ) : (
                      <Save className="h-4 w-4 mr-2" />
                    )}
                    Save
                  </Button>
                </div>
              )}
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            {isEditingConfig ? (
              // Edit Mode
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="entrypoints">Entrypoints</Label>
                    <Input
                      id="entrypoints"
                      value={editEntrypoints}
                      onChange={(e) => setEditEntrypoints(e.target.value)}
                      placeholder="websecure"
                    />
                    <p className="text-xs text-muted-foreground">
                      Comma-separated list of entrypoints
                    </p>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="priority">Router Priority</Label>
                    <Input
                      id="priority"
                      type="number"
                      value={editPriority}
                      onChange={(e) => setEditPriority(e.target.value)}
                      placeholder="200"
                    />
                    <p className="text-xs text-muted-foreground">
                      Higher priority routers are evaluated first
                    </p>
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="tlsDomains">TLS Domains</Label>
                  <Input
                    id="tlsDomains"
                    value={editTlsDomains}
                    onChange={(e) => setEditTlsDomains(e.target.value)}
                    placeholder="example.com, *.example.com"
                  />
                  <p className="text-xs text-muted-foreground">
                    Comma-separated list of TLS certificate SANs
                  </p>
                </div>

                {selectedResource.tcp_enabled && (
                  <div className="border-t pt-4 mt-4">
                    <h4 className="font-medium mb-3">TCP Configuration</h4>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="tcpEntrypoints">TCP Entrypoints</Label>
                        <Input
                          id="tcpEntrypoints"
                          value={editTcpEntrypoints}
                          onChange={(e) => setEditTcpEntrypoints(e.target.value)}
                          placeholder="tcpep"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="tcpSniRule">SNI Rule</Label>
                        <Input
                          id="tcpSniRule"
                          value={editTcpSniRule}
                          onChange={(e) => setEditTcpSniRule(e.target.value)}
                          placeholder="HostSNI(`example.com`)"
                        />
                      </div>
                    </div>
                  </div>
                )}
              </div>
            ) : (
              // View Mode
              <>
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
              </>
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
                  <TableHead className="w-12">#</TableHead>
                  <TableHead>Middleware ID</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Priority</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {assignedMiddlewares.map((mw, index) => {
                  // Look up middleware in the full middlewares list for type info
                  const middlewareInfo = middlewares.find((m) => m.id === mw.id)
                  return (
                    <TableRow key={mw.id}>
                      <TableCell className="text-muted-foreground">{index + 1}</TableCell>
                      <TableCell className="font-mono text-sm">{mw.id}</TableCell>
                      <TableCell>{mw.name}</TableCell>
                      <TableCell>
                        <Badge variant="outline">
                          {middlewareInfo?.type || 'plugin'}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant="secondary">{mw.priority}</Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => {
                              // Re-assign with edit - prompt for new priority
                              const newPriority = prompt('Enter new priority (higher = runs first):', String(mw.priority))
                              if (newPriority && resourceId) {
                                const priority = parseInt(newPriority, 10)
                                if (!isNaN(priority)) {
                                  assignMiddleware(resourceId, {
                                    middleware_id: mw.id,
                                    priority: priority,
                                  })
                                }
                              }
                            }}
                            title="Edit priority"
                          >
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setRemoveMiddlewareModal(mw.id)}
                            title="Remove middleware"
                          >
                            <Trash2 className="h-4 w-4 text-destructive" />
                          </Button>
                        </div>
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
