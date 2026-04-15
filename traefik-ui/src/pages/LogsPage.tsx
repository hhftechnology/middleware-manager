import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import { PageHeader } from '@/components/common/PageHeader'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

export default function LogsPage() {
  const logsQuery = useQuery({
    queryKey: ['logs'],
    queryFn: api.traefik.logs,
    refetchInterval: 10_000,
  })

  return (
    <div className="space-y-6">
      <PageHeader
        title="Logs"
        description="Tail the configured Traefik access log file from the manager container."
        actions={
          <Button variant="outline" onClick={() => logsQuery.refetch()}>
            Refresh
          </Button>
        }
      />
      <Card>
        <CardContent className="pt-6">
          <pre className="min-h-[420px] overflow-x-auto rounded-md bg-muted p-4 text-xs">
            {(logsQuery.data?.lines ?? []).join('\n') || logsQuery.data?.error || 'No log lines available.'}
          </pre>
        </CardContent>
      </Card>
    </div>
  )
}
