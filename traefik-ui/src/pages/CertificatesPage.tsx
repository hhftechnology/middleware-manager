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

export default function CertificatesPage() {
  const certsQuery = useQuery({ queryKey: ['certs'], queryFn: api.traefik.certs })

  return (
    <div className="space-y-6">
      <PageHeader
        title="Certificates"
        description="Read ACME and file-based certificate metadata from the Traefik Manager runtime."
      />
      <Card>
        <CardContent className="pt-6">
          {certsQuery.data?.certs.length ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Main</TableHead>
                  <TableHead>Resolver</TableHead>
                  <TableHead>SANs</TableHead>
                  <TableHead>Not After</TableHead>
                  <TableHead>File</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {certsQuery.data.certs.map((cert) => (
                  <TableRow key={cert.main + cert.resolver}>
                    <TableCell className="font-medium">{cert.main}</TableCell>
                    <TableCell>{cert.resolver}</TableCell>
                    <TableCell className="text-xs">{cert.sans.join(', ') || 'n/a'}</TableCell>
                    <TableCell>{cert.not_after || 'unknown'}</TableCell>
                    <TableCell className="text-xs">{cert.certFile || 'n/a'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : certsQuery.isLoading ? (
            <p className="text-sm text-muted-foreground">Loading certificates...</p>
          ) : (
            <p className="text-sm text-muted-foreground">
              The configured ACME store and TLS certificate files did not return any certificate metadata.
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
