import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ExternalLink, Download, Trash2, Loader2, Check } from 'lucide-react'
import type { Plugin } from '@/types'

interface PluginCardProps {
  plugin: Plugin
  onInstall: (moduleName: string, version: string) => Promise<void>
  onRemove: (moduleName: string) => Promise<void>
  installing: boolean
  removing: boolean
}

export function PluginCard({
  plugin,
  onInstall,
  onRemove,
  installing,
  removing,
}: PluginCardProps) {
  const isLoading = installing || removing

  return (
    <Card className="flex flex-col">
      <CardHeader>
        <div className="flex items-start justify-between">
          <div>
            <CardTitle className="text-lg">{plugin.name}</CardTitle>
            <CardDescription className="mt-1">
              {plugin.moduleName}
            </CardDescription>
          </div>
          {plugin.installed ? (
            <Badge variant="success" className="flex items-center gap-1">
              <Check className="h-3 w-3" />
              Installed
            </Badge>
          ) : (
            <Badge variant="outline">v{plugin.version}</Badge>
          )}
        </div>
      </CardHeader>
      <CardContent className="flex-1">
        <p className="text-sm text-muted-foreground">
          {plugin.description || 'No description available'}
        </p>
        {plugin.author && (
          <p className="text-xs text-muted-foreground mt-2">
            By {plugin.author}
          </p>
        )}
        {plugin.installedVersion && (
          <p className="text-xs text-muted-foreground mt-1">
            Installed version: v{plugin.installedVersion}
          </p>
        )}
      </CardContent>
      <CardFooter className="flex gap-2">
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
        {plugin.installed ? (
          <Button
            variant="destructive"
            size="sm"
            onClick={() => onRemove(plugin.moduleName)}
            disabled={isLoading}
            className="flex-1"
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
