// ui/src/contexts/PluginContext.js
import React, { createContext, useState, useContext, useCallback, useEffect } from 'react';
// Ensure GlobalErrorMessage is imported correctly if you use it elsewhere in this file.
// For now, local error display via alert/setError should suffice for these new functions.
import { ConfirmationModal } from '../components/common';

const API_URL = '/api/plugins';

const PluginContext = createContext();

export const PluginProvider = ({ children }) => {
  const [plugins, setPlugins] = useState([]); // Will now store plugins with status
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [traefikConfigPath, setTraefikConfigPath] = useState('');
  const [fetchingPath, setFetchingPath] = useState(true);
  const [alert, setAlert] = useState({ show: false, title: '', message: '' });

  const fetchPlugins = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch(API_URL);
      if (!response.ok) {
        const errData = await response.json().catch(() => ({ message: `HTTP error ${response.status}` }));
        throw new Error(errData.message);
      }
      const data = await response.json();
      setPlugins(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Failed to fetch plugins:', err);
      setError(`Failed to load plugins: ${err.message}`);
      setPlugins([]);
    } finally {
      setLoading(false);
    }
  }, []);

  const installPlugin = useCallback(async (pluginData) => {
    // setLoading(true); // Individual button will handle its loading state
    setError(null);
    try {
      const response = await fetch(`${API_URL}/install`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(pluginData),
      });
      if (!response.ok) {
        const errData = await response.json().catch(() => ({ message: `HTTP error ${response.status}` }));
        throw new Error(errData.message);
      }
      const result = await response.json();
      //alert(result.message || 'Plugin installation initiated successfully! Restart Traefik to apply.'); // Alert is now in the Confirmation Dialog component
      setAlert({
        show: true,
        title: 'Success',
        message: result.message || 'Plugin installation initiated successfully! Restart Traefik to apply.',
      });
      fetchPlugins(); // Refresh plugin list to show installed status
      return true;
    } catch (err) {
      console.error('Failed to install plugin:', err);
      setError(`Failed to install plugin: ${err.message}`);
      // alert(`Failed to install plugin: ${err.message}`); // Alert is now in the component
      return false;
    } finally {
      // setLoading(false);
    }
  }, [fetchPlugins]);

  const removePlugin = useCallback(async (moduleName) => {
    // setLoading(true); // Individual button will handle its loading state
    setError(null);
    try {
      const response = await fetch(`${API_URL}/remove`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ moduleName }),
      });
      if (!response.ok) {
        const errData = await response.json().catch(() => ({ message: `HTTP error ${response.status}` }));
        throw new Error(errData.message);
      }
      const result = await response.json();
      //alert(result.message || 'Plugin removal initiated successfully! Restart Traefik to apply.'); // Alert is now in the Confirmation Dialog component
      setAlert({
        show: true,
        title: 'Success',
        message: result.message || 'Plugin removal initiated successfully! Restart Traefik to apply.',
      });
      fetchPlugins(); // Refresh plugin list
      return true;
    } catch (err) {
      console.error('Failed to remove plugin:', err);
      setError(`Failed to remove plugin: ${err.message}`);
      // alert(`Failed to remove plugin: ${err.message}`); // Alert is now in the component
      return false;
    } finally {
      // setLoading(false);
    }
  }, [fetchPlugins]);


  const fetchTraefikConfigPath = useCallback(async () => {
    setFetchingPath(true);
    setError(null);
    try {
      const response = await fetch(`${API_URL}/configpath`);
      if (!response.ok) {
        const errData = await response.json().catch(() => ({ message: `HTTP error ${response.status}` }));
        throw new Error(errData.message);
      }
      const data = await response.json();
      setTraefikConfigPath(data.path || '');
    } catch (err) {
      console.error('Failed to fetch Traefik config path:', err);
      setError(`Failed to load Traefik config path: ${err.message}`);
      setTraefikConfigPath('');
    } finally {
      setFetchingPath(false);
    }
  }, []);

  const updateTraefikConfigPath = useCallback(async (newPath) => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch(`${API_URL}/configpath`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: newPath }),
      });
      if (!response.ok) {
        const errData = await response.json().catch(() => ({ message: `HTTP error ${response.status}` }));
        throw new Error(errData.message);
      }
      const data = await response.json();
      setTraefikConfigPath(data.path || '');
      //alert(data.message || 'Traefik config path updated successfully!'); // Alert is now in the Confirmation Dialog component
      setAlert({
        show: true,
        title: 'Success',
        message: data.message || 'Traefik config path updated successfully!',
      });
      return true;
    } catch (err) {
      console.error('Failed to update Traefik config path:', err);
      setError(`Failed to update Traefik config path: ${err.message}`);
      return false;
    } finally {
      setLoading(false);
    }
  }, []);


  useEffect(() => {
    fetchPlugins();
    fetchTraefikConfigPath();
  }, [fetchPlugins, fetchTraefikConfigPath]);

  const value = {
    plugins, // This will now be []PluginWithStatus
    loading,
    error,
    fetchPlugins,
    installPlugin,
    removePlugin, // Add removePlugin to context
    traefikConfigPath,
    fetchingPath,
    fetchTraefikConfigPath,
    updateTraefikConfigPath,
    setError,
  };

  return (
    <PluginContext.Provider value={value}>
      {children}

      {/* Alert modal rendered here */}
      <ConfirmationModal
        show={alert.show}
        mode="alert"
        title={alert.title}
        message={alert.message}
        onConfirm={() => setAlert(prev => ({ ...prev, show: false }))}
      />

    </PluginContext.Provider>
  );
};

export const usePlugins = () => {
  const context = useContext(PluginContext);
  if (context === undefined) {
    throw new Error('usePlugins must be used within a PluginProvider');
  }
  return context;
};