import { useEffect, useState, useMemo } from 'react'
import { usePluginStore } from '@/stores/pluginStore'
import { cn } from '@/lib/utils'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
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
  AlertTriangle,
  Activity,
  Store,
  Download,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'
import type { Plugin, CataloguePlugin } from '@/types'

type StatusFilter = 'all' | 'enabled' | 'disabled' | 'error' | 'installed'
type CatalogueFilter = 'all' | 'installed' | 'not_installed'

const CATALOGUE_PAGE_SIZE = 12

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
    showRestartWarning,
    lastInstalledPlugin,
    fetchPlugins,
    fetchCatalogue,
    fetchConfigPath,
    installPlugin,
    removePlugin,
    updateConfigPath,
    selectPlugin,
    selectCataloguePlugin,
    clearError,
    dismissRestartWarning,
  } = usePluginStore()

  const [activeTab, setActiveTab] = useState<'installed' | 'catalogue'>('installed')
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [catalogueFilter, setCatalogueFilter] = useState<CatalogueFilter>('all')
  const [newConfigPath, setNewConfigPath] = useState(configPath)
  const [savingPath, setSavingPath] = useState(false)
  const [cataloguePage, setCataloguePage] = useState(1)

  useEffect(() => {
    fetchPlugins()
    fetchConfigPath()
  }, [fetchPlugins, fetchConfigPath])

  // Fetch catalogue in the background so installed cards have metadata
  useEffect(() => {
    if (cataloguePlugins.length === 0 && !loadingCatalogue) {
      fetchCatalogue()
    }
  }, [cataloguePlugins.length, loadingCatalogue, fetchCatalogue])

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
  // A plugin is considered enabled if status is 'enabled' OR if it has active usage
  const stats = useMemo(() => {
    const isPluginEnabled = (p: Plugin) => 
      p.status === 'enabled' || (p.usageCount && p.usageCount > 0)
    
    return {
      total: plugins.length,
      enabled: plugins.filter(isPluginEnabled).length,
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

  // Filter catalogue plugins (full list for pagination)
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

  // Reset page when filters change
  useEffect(() => {
    setCataloguePage(1)
  }, [searchTerm, catalogueFilter])

  // Paginated catalogue plugins
  const paginatedCataloguePlugins = useMemo(() => {
    const startIndex = (cataloguePage - 1) * CATALOGUE_PAGE_SIZE
    return filteredCataloguePlugins.slice(startIndex, startIndex + CATALOGUE_PAGE_SIZE)
  }, [filteredCataloguePlugins, cataloguePage])

  const totalCataloguePages = Math.ceil(filteredCataloguePlugins.length / CATALOGUE_PAGE_SIZE)

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
      <div className="flex items-center justify-between pb-2 border-b border-border/60">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Plugin Hub</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
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

      {/* Statistics */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card
          className={cn(
            'cursor-pointer transition-all hover:shadow-md',
            statusFilter === 'all' && activeTab === 'installed' ? 'border-primary/40 shadow-md' : 'hover:border-primary/30'
          )}
          onClick={() => { setActiveTab('installed'); setStatusFilter('all') }}
        >
          <CardContent className="p-5">
            <div className="flex items-start justify-between">
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">Total Plugins</p>
                <p className="text-3xl font-bold tracking-tight">{stats.total}</p>
              </div>
              <div className="rounded-lg p-2.5 bg-rose-800/10 text-rose-800 dark:bg-rose-400/15 dark:text-rose-300">
                <Puzzle className="h-5 w-5" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card
          className={cn(
            'cursor-pointer transition-all hover:shadow-md',
            statusFilter === 'enabled' && activeTab === 'installed' ? 'border-primary/40 shadow-md' : 'hover:border-primary/30'
          )}
          onClick={() => { setActiveTab('installed'); setStatusFilter('enabled') }}
        >
          <CardContent className="p-5">
            <div className="flex items-start justify-between">
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">Enabled</p>
                <p className="text-3xl font-bold tracking-tight text-emerald-700 dark:text-emerald-400">{stats.enabled}</p>
              </div>
              <div className="rounded-lg p-2.5 bg-emerald-700/10 text-emerald-700 dark:bg-emerald-400/15 dark:text-emerald-300">
                <Power className="h-5 w-5" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card
          className={cn(
            'cursor-pointer transition-all hover:shadow-md',
            statusFilter === 'error' && activeTab === 'installed' ? 'border-primary/40 shadow-md' : 'hover:border-primary/30'
          )}
          onClick={() => { setActiveTab('installed'); setStatusFilter('error') }}
        >
          <CardContent className="p-5">
            <div className="flex items-start justify-between">
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">With Errors</p>
                <p className="text-3xl font-bold tracking-tight text-destructive">{stats.error}</p>
              </div>
              <div className="rounded-lg p-2.5 bg-destructive/10 text-destructive">
                <AlertCircle className="h-5 w-5" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card
          className={cn(
            'cursor-pointer transition-all hover:shadow-md',
            activeTab === 'catalogue' ? 'border-primary/40 shadow-md' : 'hover:border-primary/30'
          )}
          onClick={() => setActiveTab('catalogue')}
        >
          <CardContent className="p-5">
            <div className="flex items-start justify-between">
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">Browse Catalogue</p>
                <p className="text-3xl font-bold tracking-tight text-amber-700 dark:text-amber-400">{cataloguePlugins.length || '...'}</p>
              </div>
              <div className="rounded-lg p-2.5 bg-amber-700/10 text-amber-700 dark:bg-amber-400/15 dark:text-amber-300">
                <Store className="h-5 w-5" />
              </div>
            </div>
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
        <TabsList>
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
                    className="h-10 rounded-md border border-input bg-muted/40 px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-1 transition-colors"
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
                <div className="space-y-4">
                  <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                    {paginatedCataloguePlugins.map((plugin) => (
                      <CataloguePluginCard
                        key={plugin.id}
                        plugin={plugin}
                        onInstall={handleInstall}
                        onSelect={handleSelectCataloguePlugin}
                        installing={installing}
                      />
                    ))}
                  </div>
                  {/* Pagination Controls */}
                  {totalCataloguePages > 1 && (
                    <div className="flex items-center justify-between border-t pt-4">
                      <p className="text-sm text-muted-foreground">
                        Showing {((cataloguePage - 1) * CATALOGUE_PAGE_SIZE) + 1} to {Math.min(cataloguePage * CATALOGUE_PAGE_SIZE, filteredCataloguePlugins.length)} of {filteredCataloguePlugins.length} plugins
                      </p>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setCataloguePage(p => Math.max(1, p - 1))}
                          disabled={cataloguePage === 1}
                        >
                          <ChevronLeft className="h-4 w-4 mr-1" />
                          Previous
                        </Button>
                        <span className="text-sm text-muted-foreground px-2">
                          Page {cataloguePage} of {totalCataloguePages}
                        </span>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setCataloguePage(p => Math.min(totalCataloguePages, p + 1))}
                          disabled={cataloguePage === totalCataloguePages}
                        >
                          Next
                          <ChevronRight className="h-4 w-4 ml-1" />
                        </Button>
                      </div>
                    </div>
                  )}
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
                    Install v{selectedCataloguePlugin.latestVersion.replace(/^v/, '')}
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

      {/* Restart Traefik Warning Dialog */}
      <AlertDialog open={showRestartWarning} onOpenChange={(open) => !open && dismissRestartWarning()}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-amber-100 dark:bg-amber-900">
                <AlertTriangle className="h-5 w-5 text-amber-600 dark:text-amber-400" />
              </div>
              <AlertDialogTitle>Traefik Restart Required</AlertDialogTitle>
            </div>
            <AlertDialogDescription className="pt-2">
              <span className="font-medium text-foreground">{lastInstalledPlugin}</span> has been successfully configured in your Traefik static configuration.
              <br /><br />
              <span className="text-amber-600 dark:text-amber-400 font-medium">
                The plugin will NOT be activated until you restart Traefik.
              </span>
              <br /><br />
              After restarting Traefik, you can use this plugin by creating a middleware of type "plugin" with the plugin key "{lastInstalledPlugin}".
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogAction onClick={dismissRestartWarning}>
              I Understand
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
