import { useAppStore } from '@/stores/appStore'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ExternalLink } from 'lucide-react'
import type { Resource } from '@/types'
import { truncate } from '@/lib/utils'

interface ResourceSummaryProps {
  resources: Resource[]
  limit?: number
}

export function ResourceSummary({ resources, limit = 5 }: ResourceSummaryProps) {
  const { navigateTo } = useAppStore()

  const displayResources = resources.slice(0, limit)

  if (resources.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No resources found
      </div>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Host</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Middlewares</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {displayResources.map((resource) => {
          const middlewareCount = resource.middlewares
            ? resource.middlewares.split(',').filter(Boolean).length
            : 0

          return (
            <TableRow key={resource.id}>
              <TableCell className="font-medium">
                <div className="flex items-center gap-2">
                  {truncate(resource.host, 30)}
                  {resource.tcp_enabled && (
                    <Badge variant="secondary" className="text-xs">
                      TCP
                    </Badge>
                  )}
                </div>
              </TableCell>
              <TableCell>
                <Badge
                  variant={resource.status === 'active' ? 'success' : 'secondary'}
                >
                  {resource.status}
                </Badge>
              </TableCell>
              <TableCell>
                <Badge variant="outline">{middlewareCount}</Badge>
              </TableCell>
              <TableCell className="text-right">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => navigateTo('resource-detail', resource.id)}
                >
                  <ExternalLink className="h-4 w-4" />
                </Button>
              </TableCell>
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}
