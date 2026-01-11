import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  ExternalLink,
  Download,
  Trash2,
  Loader2,
  Check,
  AlertCircle,
  Power,
  PowerOff,
  Activity,
} from 'lucide-react'
import type { Plugin } from '@/types'

interface PluginCardProps {
  plugin: Plugin
  onInstall: (moduleName: string, version?: string) => Promise<void>
  onRemove: (moduleName: string) => Promise<void>
  onSelect?: (plugin: Plugin) => void
  installing: boolean
  removing: boolean
}

// Get status badge variant and icon
function getStatusBadge(status: Plugin['status'], isInstalled?: boolean) {
  switch (status) {
    case 'enabled':
      return {
        variant: 'success' as const,
        icon: <Power className="h-3 w-3" />,
        label: 'Loaded',
      }
    case 'disabled':
      return {
        variant: 'secondary' as const,
        icon: <PowerOff className="h-3 w-3" />,
        label: 'Disabled',
      }
    case 'error':
      return {
        variant: 'destructive' as const,
        icon: <AlertCircle className="h-3 w-3" />,
        label: 'Error',
      }
    case 'not_loaded':
      // If installed but not loaded, it's a user-configured plugin pending restart
      if (isInstalled) {
        return {
          variant: 'warning' as const,
          icon: <Activity className="h-3 w-3" />,
          label: 'User (Restart Required)',
        }
      }
      return {
        variant: 'outline' as const,
        icon: <Activity className="h-3 w-3" />,
        label: 'Not Loaded',
      }
    case 'configured':
      return {
        variant: 'default' as const,
        icon: <Check className="h-3 w-3" />,
        label: 'User',
      }
    default:
      return {
        variant: 'outline' as const,
        icon: null,
        label: status,
      }
  }
}

export function PluginCard({
  plugin,
  onInstall,
  onRemove,
  onSelect,
  installing,
  removing,
}: PluginCardProps) {
  const isLoading = installing || removing
  const statusBadge = getStatusBadge(plugin.status, plugin.isInstalled)

  return (
    <Card
      className="flex flex-col cursor-pointer hover:border-primary/50 transition-colors"
      onClick={() => onSelect?.(plugin)}
    >
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="flex-1 min-w-0">
            <CardTitle className="text-lg truncate">{plugin.name}</CardTitle>
            <CardDescription className="mt-1 truncate">
              {plugin.moduleName}
            </CardDescription>
          </div>
          <div className="flex flex-col gap-1 items-end ml-2">
            <Badge variant={statusBadge.variant} className="flex items-center gap-1">
              {statusBadge.icon}
              {statusBadge.label}
            </Badge>
            {plugin.version && (
              <Badge variant="outline" className="text-xs">
                v{plugin.version}
              </Badge>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex-1">
        <p className="text-sm text-muted-foreground line-clamp-2">
          {plugin.description || 'No description available'}
        </p>
        {plugin.author && (
          <p className="text-xs text-muted-foreground mt-2">
            By {plugin.author}
          </p>
        )}
        {plugin.error && (
          <p className="text-xs text-destructive mt-2">
            Error: {plugin.error}
          </p>
        )}
        {plugin.usageCount > 0 && (
          <p className="text-xs text-muted-foreground mt-2">
            Used by {plugin.usageCount} middleware{plugin.usageCount > 1 ? 's' : ''}
          </p>
        )}
        {plugin.installedVersion && plugin.installedVersion !== plugin.version && (
          <p className="text-xs text-amber-500 mt-1">
            Installed: v{plugin.installedVersion}
          </p>
        )}
      </CardContent>
      <CardFooter className="flex gap-2" onClick={(e) => e.stopPropagation()}>
        {plugin.homepage && (
          <Button
            variant="outline"
            size="sm"
            asChild
            className="flex-1"
          >
            <a
              href={plugin.homepage}
              target="_blank"
              rel="noopener noreferrer"
            >
              <ExternalLink className="h-4 w-4 mr-2" />
              Docs
            </a>
          </Button>
        )}
        {plugin.isInstalled ? (
          <Button
            variant="destructive"
            size="sm"
            onClick={() => onRemove(plugin.moduleName)}
            disabled={isLoading || plugin.usageCount > 0}
            className="flex-1"
            title={plugin.usageCount > 0 ? 'Cannot remove: plugin is in use' : undefined}
          >
            {removing ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Removing...
              </>
            ) : (
              <>
                <Trash2 className="h-4 w-4 mr-2" />
                Remove
              </>
            )}
          </Button>
        ) : (
          <Button
            size="sm"
            onClick={() => onInstall(plugin.moduleName, plugin.version)}
            disabled={isLoading}
            className="flex-1"
          >
            {installing ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Installing...
              </>
            ) : (
              <>
                <Download className="h-4 w-4 mr-2" />
                Install
              </>
            )}
          </Button>
        )}
      </CardFooter>
    </Card>
  )
}
