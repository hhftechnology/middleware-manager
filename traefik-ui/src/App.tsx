import { Suspense, lazy } from 'react'
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { Header } from '@/components/common/Header'
import { Footer } from '@/components/common/Footer'
import { ErrorBoundary } from '@/components/error-boundary'
import { Card, CardContent } from '@/components/ui/card'

const DashboardPage = lazy(() => import('@/pages/DashboardPage'))
const RoutesPage = lazy(() => import('@/pages/RoutesPage'))
const MiddlewaresPage = lazy(() => import('@/pages/MiddlewaresPage'))
const SettingsPage = lazy(() => import('@/pages/SettingsPage'))
const CertificatesPage = lazy(() => import('@/pages/CertificatesPage'))
const LogsPage = lazy(() => import('@/pages/LogsPage'))
const PluginsPage = lazy(() => import('@/pages/PluginsPage'))
const BackupsPage = lazy(() => import('@/pages/BackupsPage'))

function PageFallback() {
  return (
    <Card>
      <CardContent className="flex items-center justify-center py-10 text-sm text-muted-foreground">
        Loading...
      </CardContent>
    </Card>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <div className="min-h-screen bg-background flex flex-col">
        <Header />
        <main className="container mx-auto px-4 py-8 flex-1">
          <ErrorBoundary>
            <Suspense fallback={<PageFallback />}>
              <Routes>
                <Route path="/" element={<DashboardPage />} />
                <Route path="/routes" element={<RoutesPage />} />
                <Route path="/middlewares" element={<MiddlewaresPage />} />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="/certificates" element={<CertificatesPage />} />
                <Route path="/logs" element={<LogsPage />} />
                <Route path="/plugins" element={<PluginsPage />} />
                <Route path="/backups" element={<BackupsPage />} />
              </Routes>
            </Suspense>
          </ErrorBoundary>
        </main>
        <Footer />
      </div>
    </BrowserRouter>
  )
}
