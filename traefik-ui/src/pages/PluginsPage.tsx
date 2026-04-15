import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import { PageHeader } from '@/components/common/PageHeader'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

export default function PluginsPage() {
  const pluginsQuery = useQuery({ queryKey: ['plugins'], queryFn: api.traefik.plugins })

  return (
    <div className="space-y-6">
      <PageHeader
        title="Plugins"
        description="Read the plugin definitions from the configured Traefik static config file."
      />
      <Card>
        <CardContent className="pt-6">
          {pluginsQuery.data?.plugins.length ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Module</TableHead>
                  <TableHead>Version</TableHead>
                  <TableHead>Settings</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {pluginsQuery.data.plugins.map((plugin) => (
                  <TableRow key={plugin.name}>
                    <TableCell className="font-medium">{plugin.name}</TableCell>
                    <TableCell className="font-mono text-xs">{plugin.moduleName}</TableCell>
                    <TableCell>{plugin.version}</TableCell>
                    <TableCell className="font-mono text-xs">
                      {JSON.stringify(plugin.settings ?? {}, null, 2)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : pluginsQuery.isLoading ? (
            <p className="text-sm text-muted-foreground">Loading plugins...</p>
          ) : (
            <p className="text-sm text-muted-foreground">
              {pluginsQuery.data?.error || 'No experimental plugins were found in the static config file.'}
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
