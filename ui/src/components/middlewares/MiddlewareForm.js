import React, { useState, useEffect } from 'react';
import { useMiddlewares } from '../../contexts/MiddlewareContext';
import { LoadingSpinner } from '../common';
import { showErrorToast } from '../common/Toast';

const MiddlewareForm = ({ id, isEditing, navigateTo }) => {
  const {
    middlewares,
    selectedMiddleware,
    loading,
    fetchMiddleware,
    createMiddleware,
    updateMiddleware,
    getConfigTemplate
  } = useMiddlewares();

  // Form state
  const [middleware, setMiddleware] = useState({
    name: '',
    type: 'basicAuth',
    config: {}
  });
  const [configText, setConfigText] = useState('{\n  "users": [\n    "admin:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"\n  ]\n}');
  const [orderedMiddlewares, setOrderedMiddlewares] = useState([]);
  
  // Available middleware types
  const middlewareTypes = [
    { value: 'basicAuth', label: 'Basic Authentication' },
    { value: 'digestAuth', label: 'Digest Authentication' },
    { value: 'forwardAuth', label: 'Forward Authentication' },
    { value: 'ipWhiteList', label: 'IP Whitelist' },
    { value: 'ipAllowList', label: 'IP Allow List' },
    { value: 'rateLimit', label: 'Rate Limiting' },
    { value: 'headers', label: 'HTTP Headers' },
    { value: 'stripPrefix', label: 'Strip Prefix' },
    { value: 'stripPrefixRegex', label: 'Strip Prefix Regex' },
    { value: 'addPrefix', label: 'Add Prefix' },
    { value: 'redirectRegex', label: 'Redirect Regex' },
    { value: 'redirectScheme', label: 'Redirect Scheme' },
    { value: 'replacePath', label: 'Replace Path' },
    { value: 'replacePathRegex', label: 'Replace Path Regex' },
    { value: 'chain', label: 'Middleware Chain' },
    { value: 'plugin', label: 'Traefik Plugin' },
    { value: 'buffering', label: 'Buffering' },
    { value: 'circuitBreaker', label: 'Circuit Breaker' },
    { value: 'compress', label: 'Compression' },
    { value: 'contentType', label: 'Content Type' },
    { value: 'errors', label: 'Error Pages' },
    { value: 'grpcWeb', label: 'gRPC Web' },
    { value: 'inFlightReq', label: 'In-Flight Request Limiter' },
    { value: 'passTLSClientCert', label: 'Pass TLS Client Certificate' },
    { value: 'retry', label: 'Retry' }
  ];

  // Fetch middleware details if editing
  useEffect(() => {
    if (isEditing && id) {
      fetchMiddleware(id);
    }
  }, [id, isEditing, fetchMiddleware]);

  // Update form state when middleware data is loaded
  useEffect(() => {
    if (isEditing && selectedMiddleware) {
      // Format config as pretty JSON string
      const configJson = typeof selectedMiddleware.config === 'string' 
        ? selectedMiddleware.config 
        : JSON.stringify(selectedMiddleware.config, null, 2);
      
      setMiddleware({
        name: selectedMiddleware.name,
        type: selectedMiddleware.type,
        config: selectedMiddleware.config
      });
      
      setConfigText(configJson);
      
      // Extract and set ordered middlewares for chain type
      if (selectedMiddleware.type === 'chain' && 
          selectedMiddleware.config && 
          selectedMiddleware.config.middlewares) {
        setOrderedMiddlewares(selectedMiddleware.config.middlewares);
      }
    }
  }, [isEditing, selectedMiddleware]);

  // Update config template when type changes
  const handleTypeChange = (e) => {
    const newType = e.target.value;
    setMiddleware({ ...middleware, type: newType });
    
    // Get template for this middleware type
    const template = getConfigTemplate(newType);
    setConfigText(template);
    
    // Reset ordered middlewares when switching to/from chain type
    if (newType === 'chain') {
      setOrderedMiddlewares([]);
    }
  };

  // Handle middleware selection for chain type
  const handleMiddlewareSelection = (e) => {
    const options = e.target.options;
    const selectedValues = Array.from(options)
      .filter(option => option.selected)
      .map(option => option.value);
    
    // Keep track of previously selected middlewares
    const previouslySelected = orderedMiddlewares;
    
    // Create a new ordered array that preserves existing order
    // and adds new selections at the end (or removes unselected ones)
    const newOrderedMiddlewares = [];
    
    // First add all previously selected items that are still selected
    for (const id of previouslySelected) {
      if (selectedValues.includes(id)) {
        newOrderedMiddlewares.push(id);
      }
    }
    
    // Then add any newly selected items in their original DOM order
    for (const id of selectedValues) {
      if (!newOrderedMiddlewares.includes(id)) {
        newOrderedMiddlewares.push(id);
      }
    }
    
    // Update the ordered state
    setOrderedMiddlewares(newOrderedMiddlewares);
    
    // Update the config text with the ordered array
    const configObj = { middlewares: newOrderedMiddlewares };
    setConfigText(JSON.stringify(configObj, null, 2));
  };

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();
    
    try {
      // Parse config JSON
      let configObj;
      
      if (middleware.type === 'chain') {
        // For chain type, extract from configText
        try {
          configObj = JSON.parse(configText);
        } catch (err) {
          const errorMsg = 'Invalid JSON configuration';
          showErrorToast(errorMsg, 'Please check your JSON syntax and try again.');
          return;
        }
      } else {
        try {
          configObj = JSON.parse(configText);
        } catch (err) {
          const errorMsg = 'Invalid JSON configuration';
          showErrorToast(errorMsg, 'Please check your JSON syntax and try again.');
          return;
        }
      }
      
      const middlewareData = {
        name: middleware.name,
        type: middleware.type,
        config: configObj
      };
      
      if (isEditing) {
        await updateMiddleware(id, middlewareData);
      } else {
        await createMiddleware(middlewareData);
      }
      
      navigateTo('middlewares');
    } catch (err) {
      const errorMsg = `Failed to ${isEditing ? 'update' : 'create'} middleware`;
      showErrorToast(errorMsg, err.message);
    }
  };

  if (loading && isEditing) {
    return <LoadingSpinner message="Loading middleware..." />;
  }

  return (
    <div>
      <div className="mb-6 flex items-center">
        <button
          onClick={() => navigateTo('middlewares')}
          className="mr-4 px-3 py-1 bg-gray-200 rounded hover:bg-gray-300"
        >
          Back
        </button>
        <h1 className="text-2xl font-bold">
          {isEditing ? 'Edit Middleware' : 'Create Middleware'}
        </h1>
      </div>


      <div className="bg-white p-6 rounded-lg shadow">
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label className="block text-gray-700 text-sm font-bold mb-2">
              Middleware Name
            </label>
            <input
              type="text"
              value={middleware.name}
              onChange={(e) => setMiddleware({ ...middleware, name: e.target.value })}
              className="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="e.g., production-authentication"
              required
            />
          </div>

          <div className="mb-4">
            <label className="block text-gray-700 text-sm font-bold mb-2">
              Middleware Type
            </label>
            <select
              value={middleware.type}
              onChange={handleTypeChange}
              className="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
              disabled={isEditing}
            >
              {middlewareTypes.map((type) => (
                <option key={type.value} value={type.value}>
                  {type.label}
                </option>
              ))}
            </select>
            {isEditing && (
              <p className="text-xs text-gray-500 mt-1">
                Middleware type cannot be changed after creation
              </p>
            )}
          </div>

          {middleware.type === 'chain' ? (
            <div className="mb-4">
              <label className="block text-gray-700 text-sm font-bold mb-2">
                Select Middlewares for Chain
              </label>
              {middlewares.length > 0 ? (
                <>
                  <select
                    multiple
                    onChange={handleMiddlewareSelection}
                    className="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                    size={Math.min(8, middlewares.length)}
                    value={orderedMiddlewares}
                  >
                    {middlewares
                      .filter(m => m.id !== id) // Filter out current middleware if editing
                      .map((mw) => (
                        <option key={mw.id} value={mw.id}>
                          {mw.name} ({mw.type})
                        </option>
                      ))}
                  </select>
                  <p className="text-xs text-gray-500 mt-1">
                    Hold Ctrl (or Cmd) to select multiple middlewares. Middlewares will be applied in the order selected.
                  </p>
                  
                  {orderedMiddlewares.length > 0 && (
                    <div className="mt-4 p-3 bg-gray-50 border rounded">
                      <h4 className="text-sm font-bold mb-2">Current Middleware Order:</h4>
                      <ol className="list-decimal pl-5">
                        {orderedMiddlewares.map((mwId, index) => {
                          const mw = middlewares.find(m => m.id === mwId);
                          return (
                            <li key={mwId} className="text-sm py-1">
                              {mw ? `${mw.name} (${mw.type})` : mwId}
                            </li>
                          );
                        })}
                      </ol>
                      <p className="text-xs text-gray-500 mt-2">Middlewares will be applied in this order.</p>
                    </div>
                  )}
                </>
              ) : (
                <div className="p-3 bg-blue-50 border border-blue-200 rounded text-blue-700">
                  <p className="mb-2">You need to create other middlewares first before creating a chain.</p>
                  <button
                    type="button"
                    onClick={() => navigateTo('middleware-form')}
                    className="text-blue-600 hover:underline"
                  >
                    Create a new middleware
                  </button>
                </div>
              )}
            </div>
          ) : (
            <div className="mb-4">
              <label className="block text-gray-700 text-sm font-bold mb-2">
                Configuration (JSON)
              </label>
              <textarea
                value={configText}
                onChange={(e) => setConfigText(e.target.value)}
                className="w-full px-3 py-2 border font-mono h-64 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Enter JSON configuration"
                required
              />
              <div className="text-xs text-gray-500 mt-1">
                <p>Configuration must be valid JSON for the selected middleware type</p>
                {middleware.type === 'headers' && (
                  <p className="mt-1 text-amber-600 font-medium">
                    Special note for Headers middleware: Use empty strings ("") to remove headers. 
                    Example: <code className="bg-gray-100 px-1 rounded">{'{"Server": ""}'}</code>
                  </p>
                )}
              </div>
            </div>
          )}

          <div className="flex justify-end space-x-3">
            <button
              type="button"
              onClick={() => navigateTo('middlewares')}
              className="px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
              disabled={loading || (middleware.type === 'chain' && !configText.includes('"middlewares"'))}
            >
              {loading ? 'Saving...' : isEditing ? 'Update Middleware' : 'Create Middleware'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default MiddlewareForm;
