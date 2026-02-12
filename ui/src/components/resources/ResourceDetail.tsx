import { useEffect, useState } from 'react'
import { useResourceStore } from '@/stores/resourceStore'
import { useMiddlewareStore } from '@/stores/middlewareStore'
import { useServiceStore } from '@/stores/serviceStore'
import { useMTLSStore } from '@/stores/mtlsStore'
import { useSecurityStore } from '@/stores/securityStore'
import { useTraefikStore } from '@/stores/traefikStore'
import { useAppStore } from '@/stores/appStore'
import { resourceApi } from '@/services/api'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
  Shield,
  Pencil,
  Save,
  X,
  Loader2,
  Globe,
} from 'lucide-react'
import { parseJSON } from '@/lib/utils'
import type { MTLSWhitelistExternalData } from '@/types'

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
    assignExternalMiddleware,
    removeExternalMiddleware,
  } = useResourceStore()

  const { middlewares, fetchMiddlewares } = useMiddlewareStore()
  const { services, fetchServices } = useServiceStore()
  const { config: mtlsConfig, fetchConfig: fetchMTLSConfig } = useMTLSStore()
  const { config: securityConfig, fetchConfig: fetchSecurityConfig } = useSecurityStore()
  const { httpMiddlewares, fetchMiddlewares: fetchTraefikMiddlewares } = useTraefikStore()
  const { updateMTLSConfig, updateMTLSWhitelistConfig } = useResourceStore()

  const [selectedMiddlewareId, setSelectedMiddlewareId] = useState('')
  const [middlewarePriority, setMiddlewarePriority] = useState('100')
  const [selectedServiceId, setSelectedServiceId] = useState('')
  const [removeMiddlewareModal, setRemoveMiddlewareModal] = useState<string | null>(null)

  // External middleware state
  const [selectedExternalMw, setSelectedExternalMw] = useState('')
  const [customExternalMwName, setCustomExternalMwName] = useState('')
  const [externalMwPriority, setExternalMwPriority] = useState('100')
  const [removeExternalMwModal, setRemoveExternalMwModal] = useState<string | null>(null)
  const [useCustomExternalName, setUseCustomExternalName] = useState(false)

  // Edit priority dialog state
  const [editPriorityDialog, setEditPriorityDialog] = useState<{ middlewareId: string; currentPriority: number } | null>(null)
  const [newPriorityValue, setNewPriorityValue] = useState('')

  // Configuration editing state
  const [isEditingConfig, setIsEditingConfig] = useState(false)
  const [savingConfig, setSavingConfig] = useState(false)
  const [editEntrypoints, setEditEntrypoints] = useState('')
  const [editPriority, setEditPriority] = useState('')
  const [editTlsDomains, setEditTlsDomains] = useState('')
  const [editTcpEntrypoints, setEditTcpEntrypoints] = useState('')
  const [editTcpSniRule, setEditTcpSniRule] = useState('')
  const [mtlsToggleLoading, setMtlsToggleLoading] = useState(false)
  const [mtlsRulesText, setMtlsRulesText] = useState('[]')
  const [requestHeaderEntries, setRequestHeaderEntries] = useState<Array<{ key: string; value: string }>>([])
  const [mtlsRejectMessage, setMtlsRejectMessage] = useState('')
  const [mtlsRejectCode, setMtlsRejectCode] = useState<number>(403)
  const [mtlsRefreshInterval, setMtlsRefreshInterval] = useState('')
  const [externalUrl, setExternalUrl] = useState('')
  const [externalDataKey, setExternalDataKey] = useState('')
  const [externalSkipTls, setExternalSkipTls] = useState(false)
  const [externalHeaderEntries, setExternalHeaderEntries] = useState<Array<{ key: string; value: string }>>([])
  const [savingWhitelist, setSavingWhitelist] = useState(false)
  const [whitelistError, setWhitelistError] = useState<string | null>(null)
  const [tlsHardeningLoading, setTLSHardeningLoading] = useState(false)
  const [secureHeadersLoading, setSecureHeadersLoading] = useState(false)

  useEffect(() => {
    if (resourceId) {
      fetchResource(resourceId)
      fetchMiddlewares()
      fetchServices()
      fetchMTLSConfig()
      fetchSecurityConfig()
      fetchTraefikMiddlewares('http')
    }
  }, [resourceId, fetchResource, fetchMiddlewares, fetchServices, fetchMTLSConfig, fetchSecurityConfig, fetchTraefikMiddlewares])

  useEffect(() => {
    if (!selectedResource) return

    const rules = parseJSON<unknown[]>(selectedResource.mtls_rules || '[]', [])
    setMtlsRulesText(JSON.stringify(rules, null, 2))

    const parsedHeaders = parseJSON<Record<string, string>>(selectedResource.mtls_request_headers || '{}', {})
    setRequestHeaderEntries(Object.entries(parsedHeaders).map(([key, value]) => ({ key, value })))

    setMtlsRejectMessage(selectedResource.mtls_reject_message || '')
    setMtlsRejectCode(selectedResource.mtls_reject_code ?? 403)
    setMtlsRefreshInterval(selectedResource.mtls_refresh_interval || '')

    const external = parseJSON<MTLSWhitelistExternalData>(selectedResource.mtls_external_data || '{}', {})
    setExternalUrl(external.url || '')
    setExternalDataKey(external.dataKey || '')
    setExternalSkipTls(Boolean(external.skipTlsVerify))
    const extHeaders = external.headers || {}
    setExternalHeaderEntries(Object.entries(extHeaders).map(([key, value]) => ({ key, value })))
  }, [selectedResource])

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
          priority: parseInt(parts[2] || '100', 10) || 100,
        }
      }).sort((a, b) => b.priority - a.priority) // Sort by priority descending (higher priority first)
    : []

  // Parse external middleware assignments from the format "name:priority:provider,..."
  const assignedExternalMiddlewares = selectedResource.external_middlewares
    ? selectedResource.external_middlewares.split(',').filter(Boolean).map(mwStr => {
        const parts = mwStr.split(':')
        return {
          name: parts[0] || '',
          priority: parseInt(parts[1] || '100', 10) || 100,
          provider: parts[2] || '',
        }
      }).sort((a, b) => b.priority - a.priority)
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

  const handleAssignExternalMiddleware = async () => {
    const name = useCustomExternalName ? customExternalMwName.trim() : selectedExternalMw
    if (!name || !resourceId) return

    // Try to find provider info from Traefik data
    const traefikMw = httpMiddlewares.find(m => m.name === name)
    await assignExternalMiddleware(resourceId, {
      middleware_name: name,
      priority: parseInt(externalMwPriority, 10) || 100,
      provider: traefikMw?.provider || '',
    })
    setSelectedExternalMw('')
    setCustomExternalMwName('')
    setExternalMwPriority('100')
  }

  const handleRemoveExternalMiddleware = async (name: string) => {
    if (resourceId) {
      await removeExternalMiddleware(resourceId, name)
    }
    setRemoveExternalMwModal(null)
  }

  const assignedMiddlewareIds = assignedMiddlewares.map(m => m.id)
  const availableMiddlewares = middlewares.filter(
    (m) => !assignedMiddlewareIds.includes(m.id)
  )

  // Filter Traefik middlewares: exclude already-assigned external ones and MW-manager's own
  const assignedExternalNames = assignedExternalMiddlewares.map(m => m.name)
  const availableExternalMiddlewares = httpMiddlewares.filter(
    (m) => !assignedExternalNames.includes(m.name)
  )

  // Start editing configuration
  const handleStartEditConfig = () => {
    if (selectedResource) {
      setEditEntrypoints(selectedResource.entrypoints || 'websecure')
      setEditPriority(String(selectedResource.router_priority || 100))
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
      const newPriority = parseInt(editPriority, 10) || 100
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

  // Handle mTLS toggle
  const handleMTLSToggle = async () => {
    if (!resourceId) return
    setMtlsToggleLoading(true)
    try {
      await updateMTLSConfig(resourceId, !selectedResource?.mtls_enabled)
    } finally {
      setMtlsToggleLoading(false)
    }
  }

  const handleAddRequestHeaderRow = () => {
    setRequestHeaderEntries((prev) => [...prev, { key: '', value: '' }])
  }

  const handleRemoveRequestHeaderRow = (index: number) => {
    setRequestHeaderEntries((prev) => prev.filter((_, i) => i !== index))
  }

  const handleAddExternalHeaderRow = () => {
    setExternalHeaderEntries((prev) => [...prev, { key: '', value: '' }])
  }

  const handleRemoveExternalHeaderRow = (index: number) => {
    setExternalHeaderEntries((prev) => prev.filter((_, i) => i !== index))
  }

  const handleSaveWhitelistConfig = async () => {
    if (!resourceId) return
    setWhitelistError(null)
    setSavingWhitelist(true)
    try {
      let parsedRules: unknown[] = []
      const trimmedRules = mtlsRulesText.trim()
      if (trimmedRules) {
        try {
          const candidate = JSON.parse(trimmedRules)
          if (Array.isArray(candidate)) {
            parsedRules = candidate
          } else {
            throw new Error('Rules must be a JSON array')
          }
        } catch (err) {
          setWhitelistError(err instanceof Error ? err.message : 'Invalid rules JSON')
          return
        }
      }

      const requestHeaders = requestHeaderEntries.reduce<Record<string, string>>((acc, entry) => {
        if (entry.key.trim()) {
          acc[entry.key.trim()] = entry.value
        }
        return acc
      }, {})

      const externalHeaders = externalHeaderEntries.reduce<Record<string, string>>((acc, entry) => {
        if (entry.key.trim()) {
          acc[entry.key.trim()] = entry.value
        }
        return acc
      }, {})

      const externalData: MTLSWhitelistExternalData = {}
      if (externalUrl.trim()) externalData.url = externalUrl.trim()
      if (externalDataKey.trim()) externalData.dataKey = externalDataKey.trim()
      if (externalSkipTls) externalData.skipTlsVerify = true
      if (Object.keys(externalHeaders).length > 0) externalData.headers = externalHeaders

      const payload = {
        rules: parsedRules,
        request_headers: requestHeaders,
        reject_message: mtlsRejectMessage || undefined,
        reject_code: mtlsRejectCode || undefined,
        refresh_interval: mtlsRefreshInterval || undefined,
        external_data: Object.keys(externalData).length > 0 ? externalData : undefined,
      }

      await updateMTLSWhitelistConfig(resourceId, payload)
    } finally {
      setSavingWhitelist(false)
    }
  }

  // Handle TLS hardening toggle
  const handleTLSHardeningToggle = async () => {
    if (!resourceId) return
    setTLSHardeningLoading(true)
    try {
      await resourceApi.updateTLSHardeningConfig(resourceId, !selectedResource?.tls_hardening_enabled)
      await fetchResource(resourceId)
    } catch (err) {
      console.error('Failed to update TLS hardening:', err)
    } finally {
      setTLSHardeningLoading(false)
    }
  }

  // Handle secure headers toggle
  const handleSecureHeadersToggle = async () => {
    if (!resourceId) return
    setSecureHeadersLoading(true)
    try {
      await resourceApi.updateSecureHeadersConfig(resourceId, !selectedResource?.secure_headers_enabled)
      await fetchResource(resourceId)
    } catch (err) {
      console.error('Failed to update secure headers:', err)
    } finally {
      setSecureHeadersLoading(false)
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
                      placeholder="100"
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
                    <p className="font-medium">{selectedResource.router_priority || 100}</p>
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
              <>
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
                {/* Display server targets from the assigned service */}
                {(() => {
                  const assignedService = services.find(s => s.id === selectedResource.service_id)
                  if (!assignedService) return null

                  const config = assignedService.config as Record<string, unknown>
                  const servers = config?.servers as Array<{ url?: string; address?: string; weight?: number }> | undefined

                  if (!servers || servers.length === 0) return null

                  return (
                    <div className="space-y-2">
                      <Label className="text-muted-foreground">Server Targets ({servers.length})</Label>
                      <div className="space-y-1">
                        {servers.map((server, index) => (
                          <div
                            key={index}
                            className="flex items-center justify-between p-2 bg-muted rounded text-sm font-mono"
                          >
                            <span>{server.url || server.address || 'Unknown'}</span>
                            {server.weight !== undefined && (
                              <Badge variant="outline" className="ml-2">
                                weight: {server.weight}
                              </Badge>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>
                  )
                })()}
              </>
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

      {/* mTLS Configuration Card */}
      {mtlsConfig?.enabled && mtlsConfig?.has_ca && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              mTLS Authentication
            </CardTitle>
            <CardDescription>
              Require client certificates for this resource
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="mtls-toggle" className="text-base">Enable mTLS</Label>
                <p className="text-sm text-muted-foreground">
                  When enabled, clients must present a valid certificate to access this resource
                </p>
              </div>
              <Switch
                id="mtls-toggle"
                checked={selectedResource?.mtls_enabled ?? false}
                onCheckedChange={handleMTLSToggle}
                disabled={mtlsToggleLoading}
              />
            </div>
            {selectedResource?.mtls_enabled && (
              <div className="mt-4 p-3 bg-muted rounded-lg text-sm">
                <p className="font-medium">mTLS is active for this resource</p>
                <p className="text-muted-foreground mt-1">
                  Only clients with valid certificates signed by your CA can access this resource.
                  Issue certificates in the Security tab.
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {mtlsConfig?.enabled && mtlsConfig?.has_ca && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              mTLS Whitelist (mtlswhitelist)
            </CardTitle>
            <CardDescription>
              Configure per-resource whitelist rules, request headers, refresh interval, external data, and reject message.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {whitelistError && (
              <div className="rounded-md bg-destructive/10 text-destructive p-3 text-sm">
                {whitelistError}
              </div>
            )}

            <div className="space-y-2">
              <Label>Rules JSON</Label>
              <Textarea
                value={mtlsRulesText}
                onChange={(e) => setMtlsRulesText(e.target.value)}
                className="font-mono text-sm"
                rows={8}
                placeholder={`[{"type":"ipRange","ranges":["192.168.0.0/24"],"addInterface":true}]`}
              />
              <p className="text-xs text-muted-foreground">
                Use the mtlswhitelist rule schema (ipRange/header/anyOf/allOf/noneOf). Rules must be a JSON array.
              </p>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <Label>Request Headers (templates allowed)</Label>
                <Button variant="outline" size="sm" onClick={handleAddRequestHeaderRow}>
                  <Plus className="h-4 w-4 mr-1" />
                  Add header
                </Button>
              </div>
              <div className="space-y-2">
                {requestHeaderEntries.length === 0 && (
                  <p className="text-sm text-muted-foreground">No request headers configured.</p>
                )}
                {requestHeaderEntries.map((entry, idx) => (
                  <div key={`${entry.key}-${idx}`} className="grid grid-cols-2 gap-2 items-center">
                    <Input
                      placeholder="Header name"
                      value={entry.key}
                      onChange={(e) => {
                        const newEntries = [...requestHeaderEntries]
                        newEntries[idx] = { ...entry, key: e.target.value }
                        setRequestHeaderEntries(newEntries)
                      }}
                    />
                    <div className="flex gap-2">
                      <Input
                        placeholder="Header value template"
                        value={entry.value}
                        onChange={(e) => {
                          const newEntries = [...requestHeaderEntries]
                          newEntries[idx] = { ...entry, value: e.target.value }
                          setRequestHeaderEntries(newEntries)
                        }}
                      />
                      <Button variant="ghost" size="icon" onClick={() => handleRemoveRequestHeaderRow(idx)}>
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="grid md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Reject Message</Label>
                <Input
                  placeholder="Forbidden"
                  value={mtlsRejectMessage}
                  onChange={(e) => setMtlsRejectMessage(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label>Reject Code</Label>
                <Input
                  type="number"
                  value={mtlsRejectCode}
                  onChange={(e) => setMtlsRejectCode(parseInt(e.target.value || '0', 10))}
                />
              </div>
              <div className="space-y-2">
                <Label>Refresh Interval</Label>
                <Input
                  placeholder="30m, 300s"
                  value={mtlsRefreshInterval}
                  onChange={(e) => setMtlsRefreshInterval(e.target.value)}
                />
                <p className="text-xs text-muted-foreground">Duration string understood by the plugin.</p>
              </div>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <div>
                  <Label>External Data</Label>
                  <p className="text-xs text-muted-foreground">Optional fetch for IP ranges or headers.</p>
                </div>
              </div>
              <div className="grid md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>URL</Label>
                  <Input
                    placeholder="https://api.example/config"
                    value={externalUrl}
                    onChange={(e) => setExternalUrl(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Data Key</Label>
                  <Input
                    placeholder="data"
                    value={externalDataKey}
                    onChange={(e) => setExternalDataKey(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label className="flex items-center gap-2">
                    <Switch checked={externalSkipTls} onCheckedChange={setExternalSkipTls} />
                    Skip TLS Verify
                  </Label>
                </div>
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label>External Headers</Label>
                  <Button variant="outline" size="sm" onClick={handleAddExternalHeaderRow}>
                    <Plus className="h-4 w-4 mr-1" />
                    Add header
                  </Button>
                </div>
                {externalHeaderEntries.length === 0 && (
                  <p className="text-sm text-muted-foreground">No external headers configured.</p>
                )}
                {externalHeaderEntries.map((entry, idx) => (
                  <div key={`ext-${entry.key}-${idx}`} className="grid grid-cols-2 gap-2 items-center">
                    <Input
                      placeholder="Header name"
                      value={entry.key}
                      onChange={(e) => {
                        const next = [...externalHeaderEntries]
                        next[idx] = { ...entry, key: e.target.value }
                        setExternalHeaderEntries(next)
                      }}
                    />
                    <div className="flex gap-2">
                      <Input
                        placeholder="Header value"
                        value={entry.value}
                        onChange={(e) => {
                          const next = [...externalHeaderEntries]
                          next[idx] = { ...entry, value: e.target.value }
                          setExternalHeaderEntries(next)
                        }}
                      />
                      <Button variant="ghost" size="icon" onClick={() => handleRemoveExternalHeaderRow(idx)}>
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="flex justify-end">
              <Button onClick={handleSaveWhitelistConfig} disabled={savingWhitelist || !selectedResource?.mtls_enabled}>
                {savingWhitelist ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : <Save className="h-4 w-4 mr-2" />}
                Save whitelist
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Security Settings Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Security Settings
          </CardTitle>
          <CardDescription>
            TLS hardening and secure headers configuration
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* TLS Hardening Toggle - hidden when mTLS is enabled */}
          {!selectedResource?.mtls_enabled && securityConfig?.tls_hardening_enabled && (
            <div className="flex items-center justify-between p-3 border rounded-lg">
              <div className="space-y-0.5">
                <Label htmlFor="tls-hardening-toggle" className="text-base">TLS Hardening</Label>
                <p className="text-sm text-muted-foreground">
                  Apply hardened TLS settings (TLS 1.2+, secure ciphers)
                </p>
              </div>
              <Switch
                id="tls-hardening-toggle"
                checked={selectedResource?.tls_hardening_enabled ?? false}
                onCheckedChange={handleTLSHardeningToggle}
                disabled={tlsHardeningLoading}
              />
            </div>
          )}

          {selectedResource?.mtls_enabled && (
            <div className="p-3 bg-muted rounded-lg text-sm">
              <p className="font-medium">TLS Hardening via mTLS</p>
              <p className="text-muted-foreground mt-1">
                mTLS is enabled, which already includes TLS hardening via the mtls-verify options.
              </p>
            </div>
          )}

          {!securityConfig?.tls_hardening_enabled && !selectedResource?.mtls_enabled && (
            <div className="p-3 bg-muted rounded-lg text-sm">
              <p className="text-muted-foreground">
                TLS Hardening is disabled globally. Enable it in the Security tab first.
              </p>
            </div>
          )}

          {/* Secure Headers Toggle */}
          {securityConfig?.secure_headers_enabled && (
            <div className="flex items-center justify-between p-3 border rounded-lg">
              <div className="space-y-0.5">
                <Label htmlFor="secure-headers-toggle" className="text-base">Secure Headers</Label>
                <p className="text-sm text-muted-foreground">
                  Add security response headers to this resource
                </p>
              </div>
              <Switch
                id="secure-headers-toggle"
                checked={selectedResource?.secure_headers_enabled ?? false}
                onCheckedChange={handleSecureHeadersToggle}
                disabled={secureHeadersLoading}
              />
            </div>
          )}

          {!securityConfig?.secure_headers_enabled && (
            <div className="p-3 bg-muted rounded-lg text-sm">
              <p className="text-muted-foreground">
                Secure Headers are disabled globally. Enable and configure them in the Security tab first.
              </p>
            </div>
          )}
        </CardContent>
      </Card>

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
                              setEditPriorityDialog({ middlewareId: mw.id, currentPriority: mw.priority })
                              setNewPriorityValue(String(mw.priority))
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

      {/* External / Traefik-Native Middlewares Card */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Globe className="h-5 w-5" />
                External Middlewares
              </CardTitle>
              <CardDescription>
                {assignedExternalMiddlewares.length} Traefik-native middleware(s) assigned
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Add External Middleware */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Label className="text-sm font-medium whitespace-nowrap">
                {useCustomExternalName ? 'Custom name:' : 'From Traefik:'}
              </Label>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  setUseCustomExternalName(!useCustomExternalName)
                  setSelectedExternalMw('')
                  setCustomExternalMwName('')
                }}
                className="text-xs"
              >
                {useCustomExternalName ? 'Switch to dropdown' : 'Enter custom name'}
              </Button>
            </div>
            <div className="flex gap-2">
              {useCustomExternalName ? (
                <Input
                  placeholder="e.g., my-auth@file or plugin-ratelimit@docker"
                  value={customExternalMwName}
                  onChange={(e) => setCustomExternalMwName(e.target.value)}
                  className="flex-1"
                />
              ) : (
                <Select value={selectedExternalMw} onValueChange={setSelectedExternalMw}>
                  <SelectTrigger className="flex-1">
                    <SelectValue placeholder="Select Traefik middleware" />
                  </SelectTrigger>
                  <SelectContent>
                    {availableExternalMiddlewares.map((mw) => (
                      <SelectItem key={mw.name} value={mw.name}>
                        {mw.name} {mw.provider ? `(${mw.provider})` : ''} {mw.type ? `[${mw.type}]` : ''}
                      </SelectItem>
                    ))}
                    {availableExternalMiddlewares.length === 0 && (
                      <div className="px-2 py-1.5 text-sm text-muted-foreground">
                        No external middlewares available
                      </div>
                    )}
                  </SelectContent>
                </Select>
              )}
              <Input
                type="number"
                placeholder="Priority"
                value={externalMwPriority}
                onChange={(e) => setExternalMwPriority(e.target.value)}
                className="w-24"
              />
              <Button
                onClick={handleAssignExternalMiddleware}
                disabled={useCustomExternalName ? !customExternalMwName.trim() : !selectedExternalMw}
              >
                <Plus className="h-4 w-4 mr-2" />
                Add
              </Button>
            </div>
          </div>

          {/* Assigned External Middlewares Table */}
          {assignedExternalMiddlewares.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">#</TableHead>
                  <TableHead>Middleware Name</TableHead>
                  <TableHead>Provider</TableHead>
                  <TableHead>Priority</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {assignedExternalMiddlewares.map((mw, index) => (
                  <TableRow key={mw.name}>
                    <TableCell className="text-muted-foreground">{index + 1}</TableCell>
                    <TableCell className="font-mono text-sm">{mw.name}</TableCell>
                    <TableCell>
                      <Badge variant="outline" className="bg-blue-50 dark:bg-blue-950 text-blue-700 dark:text-blue-300">
                        {mw.provider || 'external'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant="secondary">{mw.priority}</Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setRemoveExternalMwModal(mw.name)}
                        title="Remove external middleware"
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="text-center py-8 text-muted-foreground border rounded-lg">
              No external middlewares assigned. Use this section to reference middlewares defined in Traefik&apos;s dynamic config or plugins.
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

      {/* Remove External Middleware Confirmation */}
      <ConfirmationModal
        open={!!removeExternalMwModal}
        onOpenChange={(open) => !open && setRemoveExternalMwModal(null)}
        title="Remove External Middleware"
        description={`Are you sure you want to remove the external middleware "${removeExternalMwModal}" from this resource?`}
        confirmLabel="Remove"
        variant="destructive"
        onConfirm={() => removeExternalMwModal && handleRemoveExternalMiddleware(removeExternalMwModal)}
      />

      {/* Edit Priority Dialog */}
      <Dialog
        open={!!editPriorityDialog}
        onOpenChange={(open) => !open && setEditPriorityDialog(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Middleware Priority</DialogTitle>
            <DialogDescription>
              Higher priority middlewares run first. Enter a new priority value.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Label htmlFor="priority-input">Priority</Label>
            <Input
              id="priority-input"
              type="number"
              value={newPriorityValue}
              onChange={(e) => setNewPriorityValue(e.target.value)}
              placeholder="Enter priority (e.g., 100)"
              className="mt-2"
              autoFocus
              onKeyDown={(e) => {
                if (e.key === 'Enter' && editPriorityDialog && resourceId) {
                  const priority = parseInt(newPriorityValue, 10)
                  if (!isNaN(priority)) {
                    assignMiddleware(resourceId, {
                      middleware_id: editPriorityDialog.middlewareId,
                      priority: priority,
                    })
                    setEditPriorityDialog(null)
                  }
                }
              }}
            />
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setEditPriorityDialog(null)}
            >
              Cancel
            </Button>
            <Button
              onClick={() => {
                if (editPriorityDialog && resourceId) {
                  const priority = parseInt(newPriorityValue, 10)
                  if (!isNaN(priority)) {
                    assignMiddleware(resourceId, {
                      middleware_id: editPriorityDialog.middlewareId,
                      priority: priority,
                    })
                    setEditPriorityDialog(null)
                  }
                }
              }}
            >
              Save
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
