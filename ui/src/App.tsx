import { useEffect } from 'react'
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

function App() {
  const { page, isDarkMode, showSettings } = useAppStore()

  // Apply dark mode class to document
  useEffect(() => {
    if (isDarkMode) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }, [isDarkMode])

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
        {renderPage()}
      </main>
      {showSettings && <DataSourceSettings />}
    </div>
  )
}

export default App
