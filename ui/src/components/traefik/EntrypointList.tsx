import { useEffect } from 'react'
import { Globe, Lock, Unlock } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useTraefikStore } from '@/stores/traefikStore'
import { EntrypointsSkeleton } from '@/components/loading-skeleton'

export function EntrypointList() {
  const { entrypoints, fetchEntrypoints, isLoading } = useTraefikStore()

  useEffect(() => {
    fetchEntrypoints()
  }, [fetchEntrypoints])

  if (isLoading && entrypoints.length === 0) {
    return <EntrypointsSkeleton />
  }

  if (entrypoints.length === 0) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-8 text-muted-foreground">
          No entrypoints configured
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Entrypoints</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {entrypoints.map((ep) => {
            const hasTLS = ep.http?.tls !== undefined
            const hasRedirect = ep.http?.redirections?.entryPoint !== undefined
            const redirectTo = ep.http?.redirections?.entryPoint?.to

            return (
              <div
                key={ep.name}
                className="flex items-center justify-between rounded-lg border p-3"
              >
                <div className="flex items-center gap-3">
                  <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10">
                    <Globe className="h-4 w-4 text-primary" />
                  </div>
                  <div>
                    <div className="font-medium">{ep.name}</div>
                    <div className="text-sm text-muted-foreground">
                      {ep.address}
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {hasTLS && (
                    <Badge variant="outline" className="text-green-600 border-green-200">
                      <Lock className="mr-1 h-3 w-3" />
                      TLS
                    </Badge>
                  )}
                  {!hasTLS && ep.name.toLowerCase().includes('web') && (
                    <Badge variant="outline" className="text-muted-foreground">
                      <Unlock className="mr-1 h-3 w-3" />
                      HTTP
                    </Badge>
                  )}
                  {hasRedirect && redirectTo && (
                    <Badge variant="secondary">
                      Redirects to {redirectTo}
                    </Badge>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      </CardContent>
    </Card>
  )
}
