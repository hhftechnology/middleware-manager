// ui/src/components/plugins/PluginCard.js
import React, { useState, useEffect, useCallback } from 'react';
import { usePlugins } from '../../contexts/PluginContext';

// Default placeholder icon (SVG)
const DefaultPluginIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12 text-gray-300 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M20.25 7.5l-.625 10.632a2.25 2.25 0 01-2.247 2.118H6.622a2.25 2.25 0 01-2.247-2.118L3.75 7.5M10 11.25h4M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z" />
  </svg>
);

// Star icon (SVG)
const StarIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1" viewBox="0 0 20 20" fill="currentColor">
    <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
  </svg>
);

const PluginCard = ({ plugin }) => {
  const {
    installPlugin,
    removePlugin,
    // error: contextError, // Context error can be shown globally by PluginHub
    setError: setContextError,
  } = usePlugins();

  const [actionInProgress, setActionInProgress] = useState(false);
  // Initialize versionToInstall with the plugin's suggested version or its currently installed version if available
  const [versionToInstall, setVersionToInstall] = useState(plugin.installedVersion || plugin.version || '');

  // Update versionToInstall if the plugin prop changes (e.g., after an install/remove action refreshes the list)
  useEffect(() => {
    setVersionToInstall(plugin.installedVersion || plugin.version || '');
  }, [plugin.installedVersion, plugin.version]);

  const handleInstallOrUpdate = useCallback(async () => {
    setActionInProgress(true);
    setContextError(null); // Clear previous specific errors
    const success = await installPlugin({
      moduleName: plugin.import,
      version: versionToInstall.trim() || undefined, // Send user-specified version or undefined for latest
    });
    if (!success) {
      // Error is set in context and typically displayed by a global error component or PluginHub
      // For immediate feedback, an alert can be used, but context should be the source of truth.
      // alert(`Error installing/updating plugin. Check console or error messages.`);
    }
    setActionInProgress(false);
  }, [installPlugin, plugin.import, versionToInstall, setContextError]);

  const handleRemove = useCallback(async () => {
    setActionInProgress(true);
    setContextError(null); // Clear previous specific errors
    const success = await removePlugin(plugin.import);
    if (!success) {
      // Error is set in context.
      // alert(`Error removing plugin. Check console or error messages.`);
    }
    setActionInProgress(false);
  }, [removePlugin, plugin.import, setContextError]);

  const isActuallyInstalled = plugin.isInstalled;
  const canUpdate = isActuallyInstalled && versionToInstall && versionToInstall.trim() !== (plugin.installedVersion || '');

  return (
    <div
      className={`card flex flex-col justify-between p-4 md:p-5 transform transition-all duration-300 hover:shadow-xl hover:-translate-y-1 
                  ${isActuallyInstalled ? 'border-l-4 border-green-500 dark:border-green-400 bg-green-50 dark:bg-gray-700' : 'dark:bg-gray-800'}`}
    >
      {/* Top section: Icon, Title, Author, Badge, Stars */}
      <div className="flex items-start space-x-4 mb-3">
        <div className="flex-shrink-0">
          {plugin.iconPath ? (
            <img
              src={plugin.iconPath.startsWith('http') ? plugin.iconPath : `https://plugins.traefik.io/assets/${plugin.import.replace('github.com/', '')}/${plugin.iconPath.replace('.assets/', '')}`}
              alt="" // Decorative, alt provided by displayName
              className="w-12 h-12 object-contain rounded"
              onError={(e) => {
                // Fallback to default icon if image fails to load
                const parent = e.target.parentNode;
                if (parent) {
                  const defaultIconContainer = parent.querySelector('.default-icon-container');
                  if (defaultIconContainer) defaultIconContainer.style.display = 'block';
                }
                e.target.style.display = 'none';
              }}
            />
          ) : null}
          <div className="default-icon-container" style={{ display: plugin.iconPath ? 'none' : 'block' }}>
            <DefaultPluginIcon />
          </div>
        </div>

        <div className="flex-1 min-w-0"> {/* Allows text to wrap/truncate */}
          <div className="flex justify-between items-start">
            <div className="flex-1 min-w-0 mr-2"> {/* Title and Author container */}
              <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 break-words">
                {plugin.displayName}
              </h3>
              {plugin.author && (
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">by {plugin.author}</p>
              )}
            </div>
            <div className="flex-shrink-0 flex flex-col items-end space-y-1"> {/* Badge and Stars container */}
              {isActuallyInstalled && (
                <span className="badge badge-success text-xs whitespace-nowrap py-0.5 px-1.5">
                  Installed {plugin.installedVersion && `(${plugin.installedVersion})`}
                </span>
              )}
              {plugin.stars !== undefined && ( // Show stars always if available, or conditionally if space is an issue
                <div className="flex items-center text-xs text-yellow-500 dark:text-yellow-400">
                  <StarIcon />
                  {plugin.stars}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Middle section: Summary, TestedWith, Import Path */}
      <div className="mb-4">
        <p className="text-sm text-gray-600 dark:text-gray-300 leading-relaxed" style={{ minHeight: '3.5em' }}>
          {plugin.summary || 'No summary available.'}
        </p>
        {plugin.testedWith && (
          <p className="text-xs text-gray-500 dark:text-gray-400 mt-2 mb-1">Tested with: {plugin.testedWith}</p>
        )}
        {plugin.import && (
          <p className="text-xs text-gray-500 dark:text-gray-400 mb-1 font-mono break-all">
            {plugin.import}
          </p>
        )}
      </div>

      {/* Bottom section: Version Input and Action Buttons */}
      <div className="mt-auto space-y-3">
        <div className="flex items-center space-x-2">
          <label htmlFor={`version-${plugin.import}`} className="form-label text-xs !mb-0 whitespace-nowrap">Version:</label>
          <input
            id={`version-${plugin.import}`}
            type="text"
            value={versionToInstall}
            onChange={(e) => setVersionToInstall(e.target.value)}
            placeholder={isActuallyInstalled && plugin.installedVersion ? `current: ${plugin.installedVersion}` : "latest (e.g., v1.2.0)"}
            className="form-input text-xs py-1 flex-grow"
            disabled={actionInProgress}
            aria-label={`Version for ${plugin.displayName}`}
          />
        </div>
        <div className="flex flex-col sm:flex-row sm:space-x-2 space-y-2 sm:space-y-0">
          {plugin.homepage && (
            <a
              href={plugin.homepage}
              target="_blank"
              rel="noopener noreferrer"
              className="btn btn-secondary text-xs py-2 flex-1 text-center" // Ensure button styles are applied
              role="button" // For accessibility if it acts like a button
            >
              View Details
            </a>
          )}
          {isActuallyInstalled ? (
            <button
              onClick={handleRemove}
              disabled={actionInProgress}
              className="btn btn-danger text-xs py-2 flex-1"
            >
              {actionInProgress ? 'Removing...' : 'Remove Plugin'}
            </button>
          ) : (
            <button
              onClick={handleInstallOrUpdate}
              disabled={actionInProgress || !plugin.import}
              className="btn btn-primary text-xs py-2 flex-1"
            >
              {actionInProgress ? 'Installing...' : 'Install Plugin'}
            </button>
          )}
        </div>
        {canUpdate && ( // Show update button only if installed and version input differs
          <button
            onClick={handleInstallOrUpdate}
            disabled={actionInProgress || !plugin.import}
            className="btn btn-primary w-full text-xs py-2 mt-2 bg-orange-500 hover:bg-orange-600 dark:bg-orange-400 dark:hover:bg-orange-500 border-orange-500 hover:border-orange-600 dark:border-orange-400 dark:hover:border-orange-500"
          >
            {actionInProgress ? 'Updating...' : `Update to ${versionToInstall.trim() || 'latest'}`}
          </button>
        )}
      </div>
    </div>
  );
};

export default React.memo(PluginCard); // Memoize PluginCard if plugins list can be large and re-renders frequently