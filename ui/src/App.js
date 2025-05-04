import React from 'react';
import { AppProvider, useApp } from './contexts/AppContext';
import { ResourceProvider } from './contexts/ResourceContext';
import { MiddlewareProvider } from './contexts/MiddlewareContext';
import { DataSourceProvider } from './contexts/DataSourceContext';
import { Header } from './components/common';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

// Import page components
import Dashboard from './components/dashboard/Dashboard';
import ResourcesList from './components/resources/ResourcesList';
import ResourceDetail from './components/resources/ResourceDetail';
import MiddlewaresList from './components/middlewares/MiddlewaresList';
import MiddlewareForm from './components/middlewares/MiddlewareForm';
import DataSourceSettings from './components/settings/DataSourceSettings';

/**
 * Main application component that renders the current page
 * based on the navigation state
 */
const MainContent = () => {
  const { 
    page, 
    resourceId, 
    middlewareId, 
    isEditing, 
    navigateTo, 
    isDarkMode, 
    setIsDarkMode,
    showSettings,
    setShowSettings
  } = useApp();

  // Render the active page based on state
  const renderPage = () => {
    // If settings panel should be displayed, show it as an overlay
    if (showSettings) {
      return (
      <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center z-50 p-4 overflow-y-auto">
        <div className="w-full max-w-3xl max-h-screen">
          <DataSourceSettings onClose={() => setShowSettings(false)} />
        </div>
      </div>
      );
    }
    
    // Otherwise, render the current page
    switch (page) {
      case 'dashboard':
        return <Dashboard navigateTo={navigateTo} />;
      case 'resources':
        return <ResourcesList navigateTo={navigateTo} />;
      case 'resource-detail':
        return <ResourceDetail id={resourceId} navigateTo={navigateTo} />;
      case 'middlewares':
        return <MiddlewaresList navigateTo={navigateTo} />;
      case 'middleware-form':
        return (
          <MiddlewareForm
            id={middlewareId}
            isEditing={isEditing}
            navigateTo={navigateTo}
          />
        );
      default:
        return <Dashboard navigateTo={navigateTo} />;
    }
  };

  return (
    <div className={`min-h-screen ${isDarkMode ? 'dark-mode' : ''} bg-gray-100`}>
      <Header 
        currentPage={page} 
        navigateTo={navigateTo}
        isDarkMode={isDarkMode}
        setIsDarkMode={setIsDarkMode}
        openSettings={() => setShowSettings(true)}
      />
      <main className="container mx-auto px-6 py-6">
        {renderPage()}
      </main>
      <ToastContainer 
        position="top-center"
        autoClose={5000}
        hideProgressBar={false}
        newestOnTop={true}
        closeOnClick
        rtl={false}
        pauseOnFocusLoss
        draggable
        pauseOnHover
        style={{ width: "500px" }}
      />
    </div>
  );
};

/**
 * Application root component with all providers
 */
const App = () => {
  return (
    <AppProvider>
      <DataSourceProvider>
        <ResourceProvider>
          <MiddlewareProvider>
            <MainContent />
          </MiddlewareProvider>
        </ResourceProvider>
      </DataSourceProvider>
    </AppProvider>
  );
};

export default App;
