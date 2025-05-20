// ui/src/components/plugins/PluginHub.js
import React, { useEffect, useState } from 'react';
import { usePlugins } from '../../contexts/PluginContext';
import { LoadingSpinner, ErrorMessage as GlobalErrorMessage, ConfirmationModal } from '../common'; // Renamed to avoid conflict
import PluginCard from './PluginCard';

const PluginHub = () => {
  const {
    plugins,
    loading,
    error,
    fetchPlugins,
    traefikConfigPath,
    fetchingPath,
    updateTraefikConfigPath,
    setError,
  } = usePlugins();

  const [searchTerm, setSearchTerm] = useState('');
  const [currentPath, setCurrentPath] = useState('');
  const [pathError, setPathError] = useState('');
  const [pathSaving, setPathSaving] = useState(false);
  const [showWarningModal, setShowWarningModal] = useState(false);

  const cancelWarning = () => {
    setShowWarningModal(false);
  };

  useEffect(() => {
    if (traefikConfigPath) {
      setCurrentPath(traefikConfigPath);
    }
  }, [traefikConfigPath]);

  const handlePathUpdate = async (e) => {
    e.preventDefault();
    setPathError('');
    if (!currentPath.trim()) {
      setPathError('Traefik static configuration path cannot be empty.');
      return;
    }
    setPathSaving(true);
    const success = await updateTraefikConfigPath(currentPath.trim());
    if (!success) {
      // Error is set in context, this is for immediate feedback
      setPathError(error || 'Failed to update path. Check console.');
    } else {
        {/* Warning Confirmation Modal */}
        <ConfirmationModal
        show={showWarningModal}
        title="Warning"
        message={`Path updated. `}
        details="This change is in-memory and will be lost on application restart unless persisted in the backend configuration (e.g., environment variable or config file)."
        cancelText="Cancel"
        onCancel={cancelWarning}
        />
//        alert('Path updated. This change is in-memory and will be lost on application restart unless persisted in the backend configuration (e.g., environment variable or config file).');

      }
    setPathSaving(false);
  };

  const filteredPlugins = plugins.filter(plugin =>
    (plugin.displayName && plugin.displayName.toLowerCase().includes(searchTerm.toLowerCase())) ||
    (plugin.summary && plugin.summary.toLowerCase().includes(searchTerm.toLowerCase())) ||
    (plugin.import && plugin.import.toLowerCase().includes(searchTerm.toLowerCase())) ||
    (plugin.author && plugin.author.toLowerCase().includes(searchTerm.toLowerCase()))
  );

  if (loading && plugins.length === 0 && fetchingPath) {
    return <LoadingSpinner message="Loading Traefik plugins..." />;
  }

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
            <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Traefik Plugin Hub</h1>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                Discover and install Traefik plugins. Restart Traefik after installation.
            </p>
        </div>
        <button
            onClick={() => { fetchPlugins(); setError(null); }} // Clear error on refresh
            className="btn btn-secondary text-sm"
            disabled={loading}
        >
            {loading ? 'Refreshing...' : 'Refresh Plugins'}
        </button>
      </div>

      {error && (
        <GlobalErrorMessage
          message={error}
          onDismiss={() => setError(null)}
        />
      )}

      {/* Traefik Static Config Path Setting */}
      <div className="card p-5">
        <h2 className="text-lg font-semibold mb-3 text-gray-800 dark:text-gray-200">Traefik Static Configuration Path</h2>
        <form onSubmit={handlePathUpdate} className="space-y-3">
          <p className="text-xs text-gray-600 dark:text-gray-400">
            Specify the absolute path to your main Traefik static configuration file (e.g., <code className="bg-gray-200 dark:bg-gray-600 px-1 rounded">/etc/traefik/traefik.yml</code> or <code className="bg-gray-200 dark:bg-gray-600 px-1 rounded">/data/traefik.yml</code>).
            This is where plugin configurations will be added.
          </p>
          <div className="flex items-center gap-3">
            <label htmlFor="traefik-config-path-input" className="sr-only">Traefik Config Path</label>
            <input
              id="traefik-config-path-input"
              type="text"
              value={currentPath}
              onChange={(e) => setCurrentPath(e.target.value)}
              className="form-input flex-grow text-sm"
              placeholder="/etc/traefik/traefik.yml"
              disabled={pathSaving || fetchingPath}
            />
            <button
              type="submit"
              className="btn btn-primary text-sm"
              disabled={pathSaving || fetchingPath || currentPath === traefikConfigPath}
            >
              {pathSaving ? 'Saving...' : 'Set Path'}
            </button>
          </div>
          {fetchingPath && <p className="text-xs text-gray-500 dark:text-gray-400">Loading current path...</p>}
          {pathError && <p className="text-xs text-red-600 dark:text-red-400 mt-1">{pathError}</p>}
           <p className="text-xs text-amber-600 dark:text-amber-400 mt-1">
             <strong>Note:</strong> Setting the path here updates it temporarily for the current session. For persistent changes, set the <code className="bg-gray-200 dark:bg-gray-600 px-1 rounded">TRAEFIK_STATIC_CONFIG_PATH</code> environment variable for the Middleware Manager container.
           </p>
        </form>
      </div>


      <div className="relative mb-6">
        <label htmlFor="plugin-search-hub" className="sr-only">Search Plugins</label>
        <input
          id="plugin-search-hub"
          type="text"
          placeholder="Search plugins by name, author, keyword..."
          className="form-input w-full pl-10" // Add padding for icon
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
        />
        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          <svg className="h-5 w-5 text-gray-400 dark:text-gray-500" xmlns="[http://www.w3.org/2000/svg](http://www.w3.org/2000/svg)" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clipRule="evenodd" />
          </svg>
        </div>
      </div>

      {loading && plugins.length === 0 ? (
        <LoadingSpinner message="Fetching plugin list..." />
      ) : !loading && plugins.length === 0 && !error ? (
         <div className="text-center py-10 card">
            <svg xmlns="[http://www.w3.org/2000/svg](http://www.w3.org/2000/svg)" className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M20.25 7.5l-.625 10.632a2.25 2.25 0 01-2.247 2.118H6.622a2.25 2.25 0 01-2.247-2.118L3.75 7.5M10 11.25h4M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z" />
            </svg>
           <h3 className="mt-2 text-lg font-medium text-gray-900 dark:text-gray-100">No Plugins Found</h3>
           <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
             Could not load plugins. Check the <code className="text-xs">PLUGINS_JSON_URL</code> configuration or network.
           </p>
         </div>
      ) : filteredPlugins.length === 0 && searchTerm ? (
        <div className="text-center py-10 card">
            <svg xmlns="[http://www.w3.org/2000/svg](http://www.w3.org/2000/svg)" className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
           <h3 className="mt-2 text-lg font-medium text-gray-900 dark:text-gray-100">No Plugins Match Your Search</h3>
           <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
             Try a different search term.
           </p>
         </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {filteredPlugins.map(plugin => (
            <PluginCard key={plugin.import || plugin.displayName} plugin={plugin} />
          ))}
        </div>
      )}
    </div>
  );
};

export default PluginHub;