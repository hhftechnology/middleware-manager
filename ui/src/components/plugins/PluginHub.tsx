import { useEffect, useState, useMemo } from 'react'
import { usePluginStore } from '@/stores/pluginStore'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { PluginCard } from './PluginCard'
import { CataloguePluginCard } from './CataloguePluginCard'
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
  Store,
  Download,
} from 'lucide-react'
import type { Plugin, CataloguePlugin } from '@/types'

type StatusFilter = 'all' | 'enabled' | 'disabled' | 'error' | 'installed'
type CatalogueFilter = 'all' | 'installed' | 'not_installed'

export function PluginHub() {
  const {
    plugins,
    cataloguePlugins,
    configPath,
    selectedPlugin,
    selectedCataloguePlugin,
    loading,
    loadingCatalogue,
    installing,
    removing,
    error,
    fetchPlugins,
    fetchCatalogue,
    fetchConfigPath,
    installPlugin,
    removePlugin,
    updateConfigPath,
    selectPlugin,
    selectCataloguePlugin,
    clearError,
  } = usePluginStore()

  const [activeTab, setActiveTab] = useState<'installed' | 'catalogue'>('installed')
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [catalogueFilter, setCatalogueFilter] = useState<CatalogueFilter>('all')
  const [newConfigPath, setNewConfigPath] = useState(configPath)
  const [savingPath, setSavingPath] = useState(false)

  useEffect(() => {
    fetchPlugins()
    fetchConfigPath()
  }, [fetchPlugins, fetchConfigPath])

  useEffect(() => {
    setNewConfigPath(configPath)
  }, [configPath])

  // Fetch catalogue when switching to that tab
  useEffect(() => {
    if (activeTab === 'catalogue' && cataloguePlugins.length === 0 && !loadingCatalogue) {
      fetchCatalogue()
    }
  }, [activeTab, cataloguePlugins.length, loadingCatalogue, fetchCatalogue])

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

  // Filter installed plugins
  const filteredPlugins = useMemo(() => {
    return plugins.filter((plugin) => {
      const matchesSearch =
        searchTerm === '' ||
        plugin.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        plugin.moduleName.toLowerCase().includes(searchTerm.toLowerCase())

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

  // Filter catalogue plugins
  const filteredCataloguePlugins = useMemo(() => {
    return cataloguePlugins.filter((plugin) => {
      const matchesSearch =
        searchTerm === '' ||
        plugin.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        plugin.displayName.toLowerCase().includes(searchTerm.toLowerCase()) ||
        plugin.summary.toLowerCase().includes(searchTerm.toLowerCase())

      let matchesFilter = true
      switch (catalogueFilter) {
        case 'installed':
          matchesFilter = plugin.isInstalled
          break
        case 'not_installed':
          matchesFilter = !plugin.isInstalled
          break
      }

      return matchesSearch && matchesFilter
    })
  }, [cataloguePlugins, searchTerm, catalogueFilter])

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

  const handleSelectCataloguePlugin = (plugin: CataloguePlugin) => {
    selectCataloguePlugin(plugin)
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
            Manage installed plugins and browse the Traefik plugin catalogue
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => activeTab === 'installed' ? fetchPlugins() : fetchCatalogue()}
            disabled={loading || loadingCatalogue}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${(loading || loadingCatalogue) ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {error && (
        <ErrorMessage
          message={error}
          onRetry={activeTab === 'installed' ? fetchPlugins : fetchCatalogue}
          onDismiss={clearError}
        />
      )}

      {/* Statistics (for installed plugins) */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card
          className={`cursor-pointer transition-colors ${statusFilter === 'all' && activeTab === 'installed' ? 'border-primary' : 'hover:border-primary/50'}`}
          onClick={() => { setActiveTab('installed'); setStatusFilter('all') }}
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
          className={`cursor-pointer transition-colors ${statusFilter === 'enabled' && activeTab === 'installed' ? 'border-primary' : 'hover:border-primary/50'}`}
          onClick={() => { setActiveTab('installed'); setStatusFilter('enabled') }}
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
          className={`cursor-pointer transition-colors ${statusFilter === 'error' && activeTab === 'installed' ? 'border-primary' : 'hover:border-primary/50'}`}
          onClick={() => { setActiveTab('installed'); setStatusFilter('error') }}
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
          className={`cursor-pointer transition-colors ${activeTab === 'catalogue' ? 'border-primary' : 'hover:border-primary/50'}`}
          onClick={() => setActiveTab('catalogue')}
        >
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Browse Catalogue</CardTitle>
            <Store className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-500">{cataloguePlugins.length || '...'}</div>
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

      {/* Tabs for Installed vs Catalogue */}
      <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as 'installed' | 'catalogue')}>
        <TabsList className="grid w-full grid-cols-2 max-w-md">
          <TabsTrigger value="installed" className="flex items-center gap-2">
            <Activity className="h-4 w-4" />
            Installed ({stats.total})
          </TabsTrigger>
          <TabsTrigger value="catalogue" className="flex items-center gap-2">
            <Store className="h-4 w-4" />
            Catalogue ({cataloguePlugins.length})
          </TabsTrigger>
        </TabsList>

        {/* Installed Plugins Tab */}
        <TabsContent value="installed">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Installed Plugins</CardTitle>
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
                        : 'No plugins detected from Traefik API. Browse the catalogue to install plugins.'
                  }
                  action={
                    statusFilter !== 'all'
                      ? { label: 'Show All Plugins', onClick: () => setStatusFilter('all') }
                      : { label: 'Browse Catalogue', onClick: () => setActiveTab('catalogue') }
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
        </TabsContent>

        {/* Catalogue Tab */}
        <TabsContent value="catalogue">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Plugin Catalogue</CardTitle>
                  <CardDescription>
                    {filteredCataloguePlugins.length} plugin{filteredCataloguePlugins.length !== 1 ? 's' : ''} available from plugins.traefik.io
                    {catalogueFilter !== 'all' && (
                      <Badge variant="secondary" className="ml-2">
                        {catalogueFilter === 'installed' ? 'Installed' : 'Not Installed'}
                      </Badge>
                    )}
                  </CardDescription>
                </div>
                <div className="flex items-center gap-4">
                  <select
                    value={catalogueFilter}
                    onChange={(e) => setCatalogueFilter(e.target.value as CatalogueFilter)}
                    className="h-10 rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                  >
                    <option value="all">All Plugins</option>
                    <option value="not_installed">Not Installed</option>
                    <option value="installed">Already Installed</option>
                  </select>
                  <div className="relative w-64">
                    <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                      placeholder="Search catalogue..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className="pl-8"
                    />
                  </div>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {loadingCatalogue ? (
                <div className="flex justify-center py-12">
                  <div className="flex flex-col items-center gap-4">
                    <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                    <p className="text-sm text-muted-foreground">Loading plugin catalogue from plugins.traefik.io...</p>
                  </div>
                </div>
              ) : filteredCataloguePlugins.length === 0 ? (
                <EmptyState
                  icon={Store}
                  title="No plugins found"
                  description={
                    searchTerm
                      ? 'Try adjusting your search terms'
                      : catalogueFilter !== 'all'
                        ? `No ${catalogueFilter === 'installed' ? 'installed' : 'uninstalled'} plugins found`
                        : 'Failed to load plugin catalogue. Try refreshing.'
                  }
                  action={
                    catalogueFilter !== 'all'
                      ? { label: 'Show All Plugins', onClick: () => setCatalogueFilter('all') }
                      : { label: 'Refresh Catalogue', onClick: () => fetchCatalogue() }
                  }
                />
              ) : (
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                  {filteredCataloguePlugins.map((plugin) => (
                    <CataloguePluginCard
                      key={plugin.id}
                      plugin={plugin}
                      onInstall={handleInstall}
                      onSelect={handleSelectCataloguePlugin}
                      installing={installing}
                    />
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Selected Plugin Details */}
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

      {/* Selected Catalogue Plugin Details */}
      {selectedCataloguePlugin && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center gap-2">
                {selectedCataloguePlugin.iconUrl && (
                  <img
                    src={selectedCataloguePlugin.iconUrl}
                    alt=""
                    className="h-8 w-8 rounded"
                  />
                )}
                {selectedCataloguePlugin.displayName}
              </CardTitle>
              <div className="flex items-center gap-2">
                {!selectedCataloguePlugin.isInstalled && (
                  <Button
                    size="sm"
                    onClick={() => handleInstall(selectedCataloguePlugin.import, selectedCataloguePlugin.latestVersion)}
                    disabled={installing}
                  >
                    {installing ? (
                      <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    ) : (
                      <Download className="h-4 w-4 mr-2" />
                    )}
                    Install v{selectedCataloguePlugin.latestVersion}
                  </Button>
                )}
                <Button variant="ghost" size="sm" onClick={() => selectCataloguePlugin(null)}>
                  Close
                </Button>
              </div>
            </div>
            <CardDescription>{selectedCataloguePlugin.summary}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <dl className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <dt className="font-medium text-muted-foreground">Author</dt>
                <dd className="mt-1">{selectedCataloguePlugin.author}</dd>
              </div>
              <div>
                <dt className="font-medium text-muted-foreground">Type</dt>
                <dd className="mt-1 capitalize">{selectedCataloguePlugin.type}</dd>
              </div>
              <div>
                <dt className="font-medium text-muted-foreground">Latest Version</dt>
                <dd className="mt-1">{selectedCataloguePlugin.latestVersion}</dd>
              </div>
              <div>
                <dt className="font-medium text-muted-foreground">Stars</dt>
                <dd className="mt-1">{selectedCataloguePlugin.stars}</dd>
              </div>
              <div className="col-span-2">
                <dt className="font-medium text-muted-foreground">Import Path</dt>
                <dd className="mt-1 font-mono text-xs bg-muted p-2 rounded">
                  {selectedCataloguePlugin.import}
                </dd>
              </div>
              {selectedCataloguePlugin.snippet?.yaml && (
                <div className="col-span-2">
                  <dt className="font-medium text-muted-foreground">YAML Configuration Example</dt>
                  <dd className="mt-1">
                    <pre className="text-xs bg-muted p-3 rounded overflow-x-auto max-h-48">
                      {selectedCataloguePlugin.snippet.yaml}
                    </pre>
                  </dd>
                </div>
              )}
            </dl>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
