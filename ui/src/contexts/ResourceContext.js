import React, { createContext, useState, useContext, useEffect, useCallback } from 'react';
import { ResourceService } from '../services/api';


// Create the context
const ResourceContext = createContext();

/**
 * Resource provider component
 * Manages resource state and provides data to child components
 * 
 * @param {Object} props
 * @param {ReactNode} props.children - Child components
 * @returns {JSX.Element}
 */
export const ResourceProvider = ({ children }) => {
  const [resources, setResources] = useState([]);
  const [selectedResource, setSelectedResource] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  /**
   * Fetch all resources
   */
  const fetchResources = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await ResourceService.getResources();
      setResources(data);
    } catch (err) {
      setError('Failed to load resources');
      console.error('Error fetching resources:', err);
    } finally {
      setLoading(false);
    }
  }, []);
  
  /**
   * Fetch a specific resource by ID
   * 
   * @param {string} id - Resource ID
   */
  const fetchResource = useCallback(async (id) => {
    if (!id) return;
    
    try {
      setLoading(true);
      setError(null);
      const data = await ResourceService.getResource(id);
      setSelectedResource(data);
    } catch (err) {
      setError(`Failed to load resource: ${err.message}`);
      console.error(`Error fetching resource ${id}:`, err);
    } finally {
      setLoading(false);
    }
  }, []);
  
  /**
   * Delete a resource
   * 
   * @param {string} id - Resource ID
   * @returns {Promise<boolean>} - Success status
   */
  const deleteResource = useCallback(async (id) => {
    try {
      setLoading(true);
      setError(null);
      await ResourceService.deleteResource(id);
      
      // Update resources list
      setResources(prevResources => 
        prevResources.filter(resource => resource.id !== id)
      );
      
      return true;
    } catch (err) {
      setError(`Failed to delete resource: ${err.message}`);
      console.error(`Error deleting resource ${id}:`, err);
      return false;
    } finally {
      setLoading(false);
    }
  }, []);
  
  /**
   * Assign a middleware to a resource
   * 
   * @param {string} resourceId - Resource ID
   * @param {string} middlewareId - Middleware ID
   * @param {number} priority - Priority level
   * @returns {Promise<boolean>} - Success status
   */
  const assignMiddleware = useCallback(async (resourceId, middlewareId, priority = 200) => {
    try {
      setLoading(true);
      setError(null);
      
      await ResourceService.assignMiddleware(resourceId, {
        middleware_id: middlewareId,
        priority
      });
      
      // Refresh the resource
      await fetchResource(resourceId);
      return true;
    } catch (err) {
      setError(`Failed to assign middleware: ${err.message}`);
      console.error('Error assigning middleware:', err);
      return false;
    } finally {
      setLoading(false);
    }
  }, [fetchResource]);
  
  /**
   * Assign multiple middlewares to a resource
   * 
   * @param {string} resourceId - Resource ID
   * @param {Array} middlewares - Array of middleware objects
   * @returns {Promise<boolean>} - Success status
   */
  const assignMultipleMiddlewares = useCallback(async (resourceId, middlewares) => {
    try {
      setLoading(true);
      setError(null);
      
      await ResourceService.assignMultipleMiddlewares(resourceId, {
        middlewares: middlewares
      });
      
      // Refresh the resource
      await fetchResource(resourceId);
      return true;
    } catch (err) {
      setError(`Failed to assign middlewares: ${err.message}`);
      console.error('Error assigning middlewares:', err);
      return false;
    } finally {
      setLoading(false);
    }
  }, [fetchResource]);
  
  /**
   * Remove a middleware from a resource
   * 
   * @param {string} resourceId - Resource ID
   * @param {string} middlewareId - Middleware ID to remove
   * @returns {Promise<boolean>} - Success status
   */
  const removeMiddleware = useCallback(async (resourceId, middlewareId) => {
    try {
      setLoading(true);
      setError(null);
      
      await ResourceService.removeMiddleware(resourceId, middlewareId);
      
      // Refresh the resource
      await fetchResource(resourceId);
      return true;
    } catch (err) {
      setError(`Failed to remove middleware: ${err.message}`);
      console.error('Error removing middleware:', err);
      return false;
    } finally {
      setLoading(false);
    }
  }, [fetchResource]);
  
  /**
   * Update resource configuration
   * 
   * @param {string} resourceId - Resource ID
   * @param {string} configType - Configuration type (http, tls, tcp, headers, priority)
   * @param {Object} data - Configuration data
   * @returns {Promise<boolean>} - Success status
   */
  const updateResourceConfig = useCallback(async (resourceId, configType, data) => {
    try {
      setLoading(true);
      setError(null);
      
      // Call the appropriate API method based on config type
      switch (configType) {
        case 'http':
          await ResourceService.updateHTTPConfig(resourceId, data);
          break;
        case 'tls':
          await ResourceService.updateTLSConfig(resourceId, data);
          break;
        case 'tcp':
          await ResourceService.updateTCPConfig(resourceId, data);
          break;
        case 'headers':
          await ResourceService.updateHeadersConfig(resourceId, data);
          break;
        case 'priority':
          await ResourceService.updateRouterPriority(resourceId, data);
          break;
        default:
          throw new Error(`Unknown config type: ${configType}`);
      }
      
      // Refresh the resource
      await fetchResource(resourceId);
      return true;
    } catch (err) {
      setError(`Failed to update ${configType} configuration: ${err.message}`);
      console.error(`Error updating ${configType} config:`, err);
      return false;
    } finally {
      setLoading(false);
    }
  }, [fetchResource]);
  
  // Fetch resources on initial mount
  useEffect(() => {
    fetchResources();
  }, [fetchResources]);
  
  // Create context value object
  const value = {
    resources,
    selectedResource,
    loading,
    error,
    fetchResources,
    fetchResource,
    deleteResource,
    assignMiddleware,
    assignMultipleMiddlewares,
    removeMiddleware,
    updateResourceConfig,
    setError, // Expose setError for components to clear errors
  };
  
  return (
    <ResourceContext.Provider value={value}>
      {children}
    </ResourceContext.Provider>
  );
};

/**
 * Custom hook to use the resource context
 * 
 * @returns {Object} Resource context value
 */
export const useResources = () => {
  const context = useContext(ResourceContext);
  if (context === undefined) {
    throw new Error('useResources must be used within a ResourceProvider');
  }
  return context;
};