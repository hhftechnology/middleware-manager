import { useAppStore } from '@/stores/appStore'
import { Header } from '@/components/common/Header'
import { Dashboard } from '@/components/dashboard/Dashboard'
import { ResourcesList } from '@/components/resources/ResourcesList'
import { ResourceDetail } from '@/components/resources/ResourceDetail'
import { MiddlewaresList } from '@/components/middlewares/MiddlewaresList'
import { MiddlewareForm } from '@/components/middlewares/MiddlewareForm'
import { ServicesList } from '@/components/services/ServicesList'
import { ServiceForm } from '@/components/services/ServiceForm'
import { PluginHub } from '@/components/plugins/PluginHub'
import { DataSourceSettings } from '@/components/settings/DataSourceSettings'
import { ThemeProvider } from '@/components/theme-provider'
import { ErrorBoundary } from '@/components/error-boundary'
import { Toaster, TooltipProvider } from '@/components/ui'

function AppContent() {
  const { page, showSettings } = useAppStore()

  const renderPage = () => {
    switch (page) {
      case 'dashboard':
        return <Dashboard />
      case 'resources':
        return <ResourcesList />
      case 'resource-detail':
        return <ResourceDetail />
      case 'middlewares':
        return <MiddlewaresList />
      case 'middleware-form':
        return <MiddlewareForm />
      case 'services':
        return <ServicesList />
      case 'service-form':
        return <ServiceForm />
      case 'plugin-hub':
        return <PluginHub />
      default:
        return <Dashboard />
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="container mx-auto px-4 py-6">
        <ErrorBoundary>
          {renderPage()}
        </ErrorBoundary>
      </main>
      {showSettings && <DataSourceSettings />}
      <Toaster />
    </div>
  )
}

function App() {
  return (
    <ThemeProvider defaultTheme="system" storageKey="middleware-manager-theme">
      <TooltipProvider>
        <AppContent />
      </TooltipProvider>
    </ThemeProvider>
  )
}

export default App
