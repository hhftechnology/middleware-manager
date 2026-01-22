import { useEffect } from 'react'
import { Info } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { useTraefikStore } from '@/stores/traefikStore'

export function VersionInfo() {
  const { version, overview, fetchVersion, isLoading } = useTraefikStore()

  useEffect(() => {
    fetchVersion()
  }, [fetchVersion])

  if (isLoading && !version) {
    return (
      <Card>
        <CardHeader className="pb-2">
          <Skeleton className="h-4 w-24" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-6 w-32" />
          <Skeleton className="mt-2 h-4 w-20" />
        </CardContent>
      </Card>
    )
  }

  const providers = overview?.providers || []

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">Traefik Version</CardTitle>
        <Info className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">
          {version?.version || 'Unknown'}
        </div>
        {version?.codename && (
          <p className="text-sm text-muted-foreground mt-1">
            {version.codename}
          </p>
        )}
        {providers.length > 0 && (
          <div className="mt-3 flex flex-wrap gap-1">
            {providers.map((provider) => (
              <Badge key={provider} variant="secondary" className="text-xs">
                {provider}
              </Badge>
            ))}
          </div>
        )}
        {overview?.features && (
          <div className="mt-3 flex flex-wrap gap-2">
            {overview.features.accessLog && (
              <Badge variant="outline" className="text-xs">
                Access Log
              </Badge>
            )}
            {overview.features.tracing && (
              <Badge variant="outline" className="text-xs">
                Tracing: {overview.features.tracing}
              </Badge>
            )}
            {overview.features.metrics && (
              <Badge variant="outline" className="text-xs">
                Metrics: {overview.features.metrics}
              </Badge>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export function VersionBadge() {
  const { version } = useTraefikStore()

  if (!version?.version) {
    return null
  }

  return (
    <Badge variant="outline" className="text-xs">
      Traefik {version.version}
    </Badge>
  )
}
