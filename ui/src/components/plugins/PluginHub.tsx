import { useEffect, useState } from 'react'
import { usePluginStore } from '@/stores/pluginStore'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { PluginCard } from './PluginCard'
import { PageLoader } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { EmptyState } from '@/components/common/EmptyState'
import { Search, Puzzle, RefreshCw, Settings, Save, Loader2 } from 'lucide-react'

export function PluginHub() {
  const {
    plugins,
    configPath,
    loading,
    installing,
    removing,
    error,
    fetchPlugins,
    fetchConfigPath,
    installPlugin,
    removePlugin,
    updateConfigPath,
    clearError,
  } = usePluginStore()

  const [searchTerm, setSearchTerm] = useState('')
  const [showInstalled, setShowInstalled] = useState(false)
  const [newConfigPath, setNewConfigPath] = useState(configPath)
  const [savingPath, setSavingPath] = useState(false)

  useEffect(() => {
    fetchPlugins()
    fetchConfigPath()
  }, [fetchPlugins, fetchConfigPath])

  useEffect(() => {
    setNewConfigPath(configPath)
  }, [configPath])

  const filteredPlugins = plugins.filter((plugin) => {
    const matchesSearch =
      plugin.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      plugin.moduleName.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesInstalled = showInstalled ? plugin.installed : true
    return matchesSearch && matchesInstalled
  })

  const installedCount = plugins.filter((p) => p.installed).length

  const handleInstall = async (moduleName: string, version: string) => {
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

  if (loading && plugins.length === 0) {
    return <PageLoader message="Loading plugins..." />
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Plugin Hub</h1>
          <p className="text-muted-foreground">
            Discover and install Traefik plugins
          </p>
        </div>
        <Button variant="outline" onClick={() => fetchPlugins()}>
          <RefreshCw className="h-4 w-4 mr-2" />
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

      {/* Config Path Setting */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Settings className="h-5 w-5" />
            Traefik Configuration
          </CardTitle>
          <CardDescription>
            Path to your Traefik static configuration file
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
              <CardTitle>Available Plugins</CardTitle>
              <CardDescription>
                {filteredPlugins.length} plugins ({installedCount} installed)
              </CardDescription>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="showInstalled"
                  checked={showInstalled}
                  onChange={(e) => setShowInstalled(e.target.checked)}
                  className="rounded border-input"
                />
                <Label htmlFor="showInstalled" className="text-sm">
                  Installed only
                </Label>
              </div>
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
                  : 'No plugins available'
              }
            />
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {filteredPlugins.map((plugin) => (
                <PluginCard
                  key={plugin.moduleName}
                  plugin={plugin}
                  onInstall={handleInstall}
                  onRemove={handleRemove}
                  installing={installing}
                  removing={removing}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
