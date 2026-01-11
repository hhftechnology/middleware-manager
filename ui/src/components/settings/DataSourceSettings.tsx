import { useEffect, useState } from 'react'
import { useDataSourceStore } from '@/stores/dataSourceStore'
import { useAppStore } from '@/stores/appStore'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'
import {
  Check,
  X,
  RefreshCw,
  Database,
  Loader2,
  CheckCircle,
  XCircle,
} from 'lucide-react'
import { DATA_SOURCE_TYPE_LABELS } from '@/types'
import type { DataSourceType } from '@/types'

export function DataSourceSettings() {
  const { showSettings, setShowSettings } = useAppStore()
  const {
    dataSources,
    loading,
    testing,
    testResult,
    error,
    fetchDataSources,
    setActiveDataSource,
    updateDataSource,
    testConnection,
    clearTestResult,
  } = useDataSourceStore()

  const [editingSource, setEditingSource] = useState<string | null>(null)
  const [editUrl, setEditUrl] = useState('')
  const [editUsername, setEditUsername] = useState('')
  const [editPassword, setEditPassword] = useState('')

  useEffect(() => {
    if (showSettings) {
      fetchDataSources()
    }
  }, [showSettings, fetchDataSources])

  const handleEdit = (name: string) => {
    const source = dataSources.find((ds) => ds.name === name)
    if (source) {
      setEditingSource(name)
      setEditUrl(source.url)
      setEditUsername('')
      setEditPassword('')
    }
  }

  const handleSave = async () => {
    if (!editingSource) return

    await updateDataSource(editingSource, {
      url: editUrl,
      basicAuth:
        editUsername && editPassword
          ? { username: editUsername, password: editPassword }
          : undefined,
    })
    setEditingSource(null)
  }

  const handleCancel = () => {
    setEditingSource(null)
    setEditUrl('')
    setEditUsername('')
    setEditPassword('')
  }

  const handleTest = async (name: string) => {
    clearTestResult()
    await testConnection(name)
  }

  const handleSetActive = async (name: string) => {
    await setActiveDataSource(name)
  }

  return (
    <Dialog open={showSettings} onOpenChange={setShowSettings}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Database className="h-5 w-5" />
            Data Source Settings
          </DialogTitle>
          <DialogDescription>
            Configure your data source connections for Pangolin or Traefik API
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {loading ? (
            <div className="flex justify-center py-8">
              <Spinner size="lg" />
            </div>
          ) : (
            <>
              {error && (
                <Alert variant="destructive">
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}

              {testResult && (
                <Alert variant={testResult.success ? 'default' : 'destructive'}>
                  <AlertDescription className="flex items-center gap-2">
                    {testResult.success ? (
                      <CheckCircle className="h-4 w-4 text-success" />
                    ) : (
                      <XCircle className="h-4 w-4" />
                    )}
                    {testResult.message || (testResult.success ? 'Connection successful' : 'Connection failed')}
                  </AlertDescription>
                </Alert>
              )}

              {dataSources.map((source) => (
                <div
                  key={source.name}
                  className="border rounded-lg p-4 space-y-3"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <h3 className="font-medium">
                        {DATA_SOURCE_TYPE_LABELS[source.type as DataSourceType] || source.name}
                      </h3>
                      {source.isActive && (
                        <Badge variant="success">Active</Badge>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      {!source.isActive && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleSetActive(source.name)}
                        >
                          <Check className="h-4 w-4 mr-1" />
                          Set Active
                        </Button>
                      )}
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleTest(source.name)}
                        disabled={testing}
                      >
                        {testing ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <RefreshCw className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </div>

                  {editingSource === source.name ? (
                    <div className="space-y-3 pt-2 border-t">
                      <div className="space-y-2">
                        <Label>URL</Label>
                        <Input
                          value={editUrl}
                          onChange={(e) => setEditUrl(e.target.value)}
                          placeholder="http://localhost:8080"
                        />
                      </div>
                      <div className="grid grid-cols-2 gap-2">
                        <div className="space-y-2">
                          <Label>Username (optional)</Label>
                          <Input
                            value={editUsername}
                            onChange={(e) => setEditUsername(e.target.value)}
                            placeholder="admin"
                          />
                        </div>
                        <div className="space-y-2">
                          <Label>Password (optional)</Label>
                          <Input
                            type="password"
                            value={editPassword}
                            onChange={(e) => setEditPassword(e.target.value)}
                            placeholder="••••••••"
                          />
                        </div>
                      </div>
                      <div className="flex justify-end gap-2">
                        <Button variant="outline" size="sm" onClick={handleCancel}>
                          <X className="h-4 w-4 mr-1" />
                          Cancel
                        </Button>
                        <Button size="sm" onClick={handleSave}>
                          <Check className="h-4 w-4 mr-1" />
                          Save
                        </Button>
                      </div>
                    </div>
                  ) : (
                    <div className="flex items-center justify-between text-sm text-muted-foreground">
                      <span className="font-mono truncate">{source.url}</span>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleEdit(source.name)}
                      >
                        Edit
                      </Button>
                    </div>
                  )}
                </div>
              ))}
            </>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => setShowSettings(false)}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
