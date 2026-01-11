import { useEffect, useState, useMemo } from 'react'
import { usePluginStore } from '@/stores/pluginStore'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { PluginCard } from './PluginCard'
import { PageLoader } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { EmptyState } from '@/components/common/EmptyState'
import {
  Search,
  Puzzle,
  RefreshCw,
  Settings,
  Save,
  Loader2,
  Power,
  AlertCircle,
  Activity,
} from 'lucide-react'
import type { Plugin } from '@/types'

type StatusFilter = 'all' | 'enabled' | 'disabled' | 'error' | 'installed'

export function PluginHub() {
  const {
    plugins,
    configPath,
    selectedPlugin,
    loading,
    installing,
    removing,
    error,
    fetchPlugins,
    fetchConfigPath,
    installPlugin,
    removePlugin,
    updateConfigPath,
    selectPlugin,
    clearError,
  } = usePluginStore()

  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [newConfigPath, setNewConfigPath] = useState(configPath)
  const [savingPath, setSavingPath] = useState(false)

  useEffect(() => {
    fetchPlugins()
    fetchConfigPath()
  }, [fetchPlugins, fetchConfigPath])

  useEffect(() => {
    setNewConfigPath(configPath)
  }, [configPath])

  // Compute plugin statistics
  const stats = useMemo(() => {
    return {
      total: plugins.length,
      enabled: plugins.filter((p) => p.status === 'enabled').length,
      disabled: plugins.filter((p) => p.status === 'disabled').length,
      error: plugins.filter((p) => p.status === 'error').length,
      installed: plugins.filter((p) => p.isInstalled).length,
    }
  }, [plugins])

  // Filter plugins
  const filteredPlugins = useMemo(() => {
    return plugins.filter((plugin) => {
      // Search filter
      const matchesSearch =
        searchTerm === '' ||
        plugin.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        plugin.moduleName.toLowerCase().includes(searchTerm.toLowerCase())

      // Status filter
      let matchesStatus = true
      switch (statusFilter) {
        case 'enabled':
          matchesStatus = plugin.status === 'enabled'
          break
        case 'disabled':
          matchesStatus = plugin.status === 'disabled'
          break
        case 'error':
          matchesStatus = plugin.status === 'error'
          break
        case 'installed':
          matchesStatus = plugin.isInstalled
          break
      }

      return matchesSearch && matchesStatus
    })
  }, [plugins, searchTerm, statusFilter])

  const handleInstall = async (moduleName: string, version?: string) => {
    await installPlugin(moduleName, version)
  }

  const handleRemove = async (moduleName: string) => {
    await removePlugin(moduleName)
  }

  const handleSaveConfigPath = async () => {
    setSavingPath(true)
    await updateConfigPath(newConfigPath)
    setSavingPath(false)
  }

  const handleSelectPlugin = (plugin: Plugin) => {
    selectPlugin(plugin)
  }

  if (loading && plugins.length === 0) {
    return <PageLoader message="Loading plugins from Traefik API..." />
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Plugin Hub</h1>
          <p className="text-muted-foreground">
            Manage Traefik plugins from the API
          </p>
        </div>
        <Button variant="outline" onClick={() => fetchPlugins()} disabled={loading}>
          <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {error && (
        <ErrorMessage
          message={error}
          onRetry={fetchPlugins}
          onDismiss={clearError}
        />
      )}

      {/* Statistics */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card
          className={`cursor-pointer transition-colors ${statusFilter === 'all' ? 'border-primary' : 'hover:border-primary/50'}`}
          onClick={() => setStatusFilter('all')}
        >
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Plugins</CardTitle>
            <Puzzle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total}</div>
          </CardContent>
        </Card>
        <Card
          className={`cursor-pointer transition-colors ${statusFilter === 'enabled' ? 'border-primary' : 'hover:border-primary/50'}`}
          onClick={() => setStatusFilter('enabled')}
        >
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Enabled</CardTitle>
            <Power className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-500">{stats.enabled}</div>
          </CardContent>
        </Card>
        <Card
          className={`cursor-pointer transition-colors ${statusFilter === 'error' ? 'border-primary' : 'hover:border-primary/50'}`}
          onClick={() => setStatusFilter('error')}
        >
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">With Errors</CardTitle>
            <AlertCircle className="h-4 w-4 text-destructive" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-destructive">{stats.error}</div>
          </CardContent>
        </Card>
        <Card
          className={`cursor-pointer transition-colors ${statusFilter === 'installed' ? 'border-primary' : 'hover:border-primary/50'}`}
          onClick={() => setStatusFilter('installed')}
        >
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Configured</CardTitle>
            <Activity className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-500">{stats.installed}</div>
          </CardContent>
        </Card>
      </div>

      {/* Config Path Setting */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Settings className="h-5 w-5" />
            Traefik Configuration
          </CardTitle>
          <CardDescription>
            Path to your Traefik static configuration file for plugin installation
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2">
            <div className="flex-1">
              <Label htmlFor="configPath" className="sr-only">
                Config Path
              </Label>
              <Input
                id="configPath"
                value={newConfigPath}
                onChange={(e) => setNewConfigPath(e.target.value)}
                placeholder="/etc/traefik/traefik.yml"
              />
            </div>
            <Button
              onClick={handleSaveConfigPath}
              disabled={savingPath || newConfigPath === configPath}
            >
              {savingPath ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Save className="h-4 w-4" />
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Search and Filter */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Plugins</CardTitle>
              <CardDescription>
                {filteredPlugins.length} plugin{filteredPlugins.length !== 1 ? 's' : ''} found
                {statusFilter !== 'all' && (
                  <Badge variant="secondary" className="ml-2">
                    {statusFilter}
                  </Badge>
                )}
              </CardDescription>
            </div>
            <div className="flex items-center gap-4">
              <div className="relative w-64">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search plugins..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-8"
                />
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {filteredPlugins.length === 0 ? (
            <EmptyState
              icon={Puzzle}
              title="No plugins found"
              description={
                searchTerm
                  ? 'Try adjusting your search terms'
                  : statusFilter !== 'all'
                    ? `No ${statusFilter} plugins found`
                    : 'No plugins detected from Traefik API. Make sure Traefik is running with plugins enabled.'
              }
              action={
                statusFilter !== 'all'
                  ? { label: 'Show All Plugins', onClick: () => setStatusFilter('all') }
                  : undefined
              }
            />
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {filteredPlugins.map((plugin) => (
                <PluginCard
                  key={plugin.name}
                  plugin={plugin}
                  onInstall={handleInstall}
                  onRemove={handleRemove}
                  onSelect={handleSelectPlugin}
                  installing={installing}
                  removing={removing}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Selected Plugin Details (could be expanded into a modal) */}
      {selectedPlugin && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>Plugin Details: {selectedPlugin.name}</CardTitle>
              <Button variant="ghost" size="sm" onClick={() => selectPlugin(null)}>
                Close
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <dl className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <dt className="font-medium text-muted-foreground">Module Name</dt>
                <dd className="mt-1">{selectedPlugin.moduleName}</dd>
              </div>
              <div>
                <dt className="font-medium text-muted-foreground">Version</dt>
                <dd className="mt-1">{selectedPlugin.version || 'N/A'}</dd>
              </div>
              <div>
                <dt className="font-medium text-muted-foreground">Status</dt>
                <dd className="mt-1 capitalize">{selectedPlugin.status}</dd>
              </div>
              <div>
                <dt className="font-medium text-muted-foreground">Provider</dt>
                <dd className="mt-1">{selectedPlugin.provider || 'N/A'}</dd>
              </div>
              <div>
                <dt className="font-medium text-muted-foreground">Usage Count</dt>
                <dd className="mt-1">{selectedPlugin.usageCount}</dd>
              </div>
              {selectedPlugin.usedBy && selectedPlugin.usedBy.length > 0 && (
                <div className="col-span-2">
                  <dt className="font-medium text-muted-foreground">Used By</dt>
                  <dd className="mt-1 flex flex-wrap gap-1">
                    {selectedPlugin.usedBy.map((name) => (
                      <Badge key={name} variant="outline">
                        {name}
                      </Badge>
                    ))}
                  </dd>
                </div>
              )}
              {selectedPlugin.error && (
                <div className="col-span-2">
                  <dt className="font-medium text-destructive">Error</dt>
                  <dd className="mt-1 text-destructive">{selectedPlugin.error}</dd>
                </div>
              )}
            </dl>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
