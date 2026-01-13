import { useEffect, useState } from 'react'
import { useMTLSStore } from '@/stores/mtlsStore'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Loader2, Shield, ShieldCheck, ShieldAlert, AlertTriangle, Info } from 'lucide-react'
import { CAManager } from './CAManager'
import { ClientCertList } from './ClientCertList'
import { CertImportGuide } from './CertImportGuide'

export function SecurityHub() {
  const {
    config,
    loading,
    error,
    fetchConfig,
    enableMTLS,
    disableMTLS,
    clearError,
  } = useMTLSStore()

  const [showSetupWarning, setShowSetupWarning] = useState(false)
  const [switchLoading, setSwitchLoading] = useState(false)

  useEffect(() => {
    fetchConfig()
  }, [fetchConfig])

  const handleToggleMTLS = async () => {
    setSwitchLoading(true)
    if (config?.enabled) {
      await disableMTLS()
    } else {
      const success = await enableMTLS()
      if (success) {
        setShowSetupWarning(true)
      }
    }
    setSwitchLoading(false)
  }

  if (loading && !config) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Shield className="h-8 w-8" />
            Security
          </h1>
          <p className="text-muted-foreground">
            Manage mTLS (mutual TLS) for secure client authentication
          </p>
        </div>
        {config?.enabled ? (
          <Badge variant="default" className="flex items-center gap-1">
            <ShieldCheck className="h-4 w-4" />
            mTLS Enabled
          </Badge>
        ) : (
          <Badge variant="secondary" className="flex items-center gap-1">
            <ShieldAlert className="h-4 w-4" />
            mTLS Disabled
          </Badge>
        )}
      </div>

      {/* Error Alert */}
      {error && (
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription className="flex justify-between items-center">
            {error}
            <Button variant="outline" size="sm" onClick={clearError}>
              Dismiss
            </Button>
          </AlertDescription>
        </Alert>
      )}

      {/* Setup Warning */}
      {showSetupWarning && (
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Traefik Configuration Required</AlertTitle>
          <AlertDescription>
            <p className="mb-2">
              mTLS has been enabled. For it to take effect, you need to:
            </p>
            <ol className="list-decimal list-inside space-y-1 text-sm">
              <li>Add the certificates volume mount to your docker-compose.yml traefik service:</li>
              <pre className="bg-muted p-2 rounded text-xs my-2 overflow-x-auto">
{`volumes:
  - ./config/traefik/certs:/etc/traefik/certs:ro`}
              </pre>
              <li>Restart Traefik:</li>
              <pre className="bg-muted p-2 rounded text-xs my-2">
                docker compose down && docker compose up -d
              </pre>
            </ol>
            <Button variant="outline" size="sm" className="mt-2" onClick={() => setShowSetupWarning(false)}>
              Got it
            </Button>
          </AlertDescription>
        </Alert>
      )}

      <Tabs defaultValue="overview" className="space-y-4">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="ca">Certificate Authority</TabsTrigger>
          <TabsTrigger value="clients">Client Certificates</TabsTrigger>
          <TabsTrigger value="guide">Import Guide</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            {/* mTLS Status Card */}
            <Card>
              <CardHeader>
                <CardTitle>mTLS Status</CardTitle>
                <CardDescription>
                  Enable or disable mutual TLS authentication globally
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="mtls-toggle">Enable mTLS</Label>
                    <p className="text-sm text-muted-foreground">
                      Require client certificates for all mTLS-enabled resources
                    </p>
                  </div>
                  <Switch
                    id="mtls-toggle"
                    checked={config?.enabled ?? false}
                    onCheckedChange={handleToggleMTLS}
                    disabled={switchLoading || !config?.has_ca}
                  />
                </div>
                {!config?.has_ca && (
                  <Alert>
                    <Info className="h-4 w-4" />
                    <AlertDescription>
                      Create a Certificate Authority first to enable mTLS
                    </AlertDescription>
                  </Alert>
                )}
              </CardContent>
            </Card>

            {/* Statistics Card */}
            <Card>
              <CardHeader>
                <CardTitle>Certificate Statistics</CardTitle>
                <CardDescription>
                  Overview of your mTLS certificates
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">CA Status</p>
                    <p className="text-2xl font-bold">
                      {config?.has_ca ? 'Configured' : 'Not Set'}
                    </p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Client Certs</p>
                    <p className="text-2xl font-bold">{config?.client_count ?? 0}</p>
                  </div>
                  {config?.ca_expiry && (
                    <div className="col-span-2 space-y-1">
                      <p className="text-sm text-muted-foreground">CA Expiry</p>
                      <p className="text-sm font-medium">
                        {new Date(config.ca_expiry).toLocaleDateString()}
                      </p>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* How it Works */}
          <Card>
            <CardHeader>
              <CardTitle>How mTLS Works</CardTitle>
              <CardDescription>
                Mutual TLS provides strong client authentication
              </CardDescription>
            </CardHeader>
            <CardContent className="prose prose-sm dark:prose-invert max-w-none">
              <p>
                With mTLS enabled, clients must present a valid certificate signed by your CA
                to access protected resources. This provides:
              </p>
              <ul>
                <li><strong>Strong Authentication:</strong> Only clients with valid certificates can connect</li>
                <li><strong>No Login Page:</strong> Unauthorized users are rejected at the TLS level</li>
                <li><strong>Per-Resource Control:</strong> Enable mTLS selectively on specific routers</li>
              </ul>
              <p className="text-muted-foreground">
                After enabling mTLS globally, go to Resources and enable it on individual routers.
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        {/* CA Management Tab */}
        <TabsContent value="ca">
          <CAManager />
        </TabsContent>

        {/* Client Certificates Tab */}
        <TabsContent value="clients">
          <ClientCertList />
        </TabsContent>

        {/* Import Guide Tab */}
        <TabsContent value="guide">
          <CertImportGuide />
        </TabsContent>
      </Tabs>
    </div>
  )
}
