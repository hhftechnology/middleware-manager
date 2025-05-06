// ui/src/components/resources/ResourceDetail.js
import React, { useEffect, useState, useCallback } from 'react';
import { useResources } from '../../contexts/ResourceContext';
import { useMiddlewares } from '../../contexts/MiddlewareContext';
import { useServices } from '../../contexts/ServiceContext';
import { LoadingSpinner, ErrorMessage, ConfirmationModal } from '../common';
import HTTPConfigModal from './config/HTTPConfigModal';
import TLSConfigModal from './config/TLSConfigModal';
import TCPConfigModal from './config/TCPConfigModal';
import HeadersConfigModal from './config/HeadersConfigModal';
import ServiceSelectModal from './config/ServiceSelectModal';
import { ResourceService, MiddlewareUtils } from '../../services/api';

// Reusable Modal Wrapper Component
const ModalWrapper = ({ title, children, onClose, show }) => {
  if (!show) return null;
  return (
    <div className="modal-overlay">
      <div className="modal-content max-w-lg"> {/* Standard size */}
        <div className="modal-header">
          <h3 className="modal-title">{title}</h3>
          <button onClick={onClose} className="modal-close-button" aria-label="Close">&times;</button>
        </div>
        {children} {/* Body and Footer passed as children */}
      </div>
    </div>
  );
};


const ResourceDetail = ({ id, navigateTo }) => {
  // --- Context Hooks ---
  const {
    selectedResource,
    loading: resourceLoading,
    error: resourceError,
    fetchResource,
    assignMultipleMiddlewares,
    removeMiddleware,
    updateResourceConfig,
    deleteResource,
    setError: setResourceError,
  } = useResources();

  const {
    middlewares,
    loading: middlewaresLoading,
    error: middlewaresError,
    fetchMiddlewares,
    formatMiddlewareDisplay,
    setError: setMiddlewaresError,
  } = useMiddlewares();

  const {
    services,
    loading: servicesLoading,
    error: servicesError,
    loadServices,
    setError: setServicesError,
  } = useServices();

  // --- State Management ---
  const [modal, setModal] = useState({ isOpen: false, type: null });
  const [selectedMiddlewaresToAdd, setSelectedMiddlewaresToAdd] = useState([]);
  const [middlewarePriority, setMiddlewarePriority] = useState(100);
  const [routerPriority, setRouterPriority] = useState(100);
  const [resourceService, setResourceService] = useState(null);
  const [showServiceModal, setShowServiceModal] = useState(false);
  const [headerInput, setHeaderInput] = useState({ key: '', value: '' });
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [showRemoveMiddlewareModal, setShowRemoveMiddlewareModal] = useState(false);
  const [middlewareToRemove, setMiddlewareToRemove] = useState(null);
  const [showRemoveServiceModal, setShowRemoveServiceModal] = useState(false);

  // Configuration state
  const [config, setConfig] = useState({
    entrypoints: 'websecure',
    tlsDomains: '',
    tcpEnabled: false,
    tcpEntrypoints: 'tcp',
    tcpSNIRule: '',
    customHeaders: {},
  });

  // --- Data Fetching ---
  // Fetch resource details, middlewares, and services
  useEffect(() => {
    console.log("Fetching resource with ID:", id);
    if (id) {
      fetchResource(id);
      fetchMiddlewares();
      loadServices();
    } else {
      console.error("ResourceDetail: No ID provided.");
      setResourceError("No resource ID specified.");
    }
  }, [id, fetchResource, fetchMiddlewares, loadServices, setResourceError]);

  // Fetch the specific service assigned to this resource
  const fetchResourceService = useCallback(async () => {
    if (!id) return;
    try {
      setServicesError(null);
      const serviceData = await ResourceService.getResourceService(id);
      setResourceService(serviceData?.service || null);
    } catch (err) {
      if (err.status !== 404) {
        console.error("Error fetching resource service:", err);
        setServicesError(`Failed to fetch assigned service: ${err.message}`);
      } else {
        setResourceService(null);
      }
    }
  }, [id, setServicesError]);

  useEffect(() => {
    fetchResourceService();
  }, [fetchResourceService]);

  // Update local config state when the main resource data is loaded or changes
  useEffect(() => {
    if (selectedResource) {
      console.log("Updating local state from selectedResource:", selectedResource);
      try {
        let parsedHeaders = {};
        if (selectedResource.custom_headers) {
          if (typeof selectedResource.custom_headers === 'string' && selectedResource.custom_headers.trim()) {
            try {
              parsedHeaders = JSON.parse(selectedResource.custom_headers);
            } catch (e) {
              console.error("Error parsing custom_headers JSON:", e);
              setResourceError("Failed to parse custom headers configuration.");
              parsedHeaders = {};
            }
          } else if (typeof selectedResource.custom_headers === 'object') {
            parsedHeaders = selectedResource.custom_headers;
          }
        }

        setConfig({
          entrypoints: selectedResource.entrypoints || 'websecure',
          tlsDomains: selectedResource.tls_domains || '',
          tcpEnabled: selectedResource.tcp_enabled === true,
          tcpEntrypoints: selectedResource.tcp_entrypoints || 'tcp',
          tcpSNIRule: selectedResource.tcp_sni_rule || '',
          customHeaders: parsedHeaders,
        });
        setRouterPriority(selectedResource.router_priority || 100);
        // Don't clear errors here immediately, let specific actions clear them
      } catch (error) {
        console.error("Error updating local state from resource:", error);
        setResourceError(`Error processing resource data: ${error.message}`);
      }
    }
  }, [selectedResource, setResourceError]); // Only depend on selectedResource and setError


  // --- Loading & Error Handling ---
  const loading = resourceLoading || middlewaresLoading || servicesLoading;
  // Consolidate errors, prioritizing resource error
  const error = resourceError || middlewaresError || servicesError;

  const clearError = () => {
    setResourceError(null);
    setMiddlewaresError(null);
    setServicesError(null);
  };

  // Show loading spinner if still fetching essential data
  if (loading && !selectedResource) {
    return <LoadingSpinner message="Loading resource details..." />;
  }

  // Show error if resource couldn't be loaded
  if (!selectedResource && !loading) {
     return (
        <ErrorMessage
           message={error || "Resource not found."}
           details={!error ? `Could not load resource with ID: ${id}` : null}
           onRetry={() => fetchResource(id)}
           onDismiss={() => navigateTo('resources')}
        />
     );
  }
  // If resource loaded but other things failed, error is shown inline later

  // Ensure selectedResource is valid before proceeding
  if (!selectedResource) return null;


  // --- Helper Functions & Derived State ---
  const assignedMiddlewares = MiddlewareUtils.parseMiddlewares(selectedResource.middlewares);
  const assignedMiddlewareIds = new Set(assignedMiddlewares.map(m => m.id));
  const availableMiddlewares = middlewares.filter(m => !assignedMiddlewareIds.has(m.id));
  const isDisabled = selectedResource.status === 'disabled';

  const openConfigModal = (type) => {
      clearError(); // Clear errors when opening a modal
      setModal({ isOpen: true, type });
  }
  const closeModal = () => setModal({ isOpen: false, type: null });


  // --- Action Handlers ---

  // Middleware Assignment
  const handleMiddlewareSelectionChange = (e) => {
    const selectedOptions = Array.from(e.target.selectedOptions, option => option.value);
    setSelectedMiddlewaresToAdd(selectedOptions);
  };

  const handleAssignMiddlewareSubmit = async (e) => {
    e.preventDefault();
    if (isDisabled || !selectedMiddlewaresToAdd.length) return;
    clearError(); // Clear previous errors

    const middlewaresToAdd = selectedMiddlewaresToAdd.map(middlewareId => ({
      middleware_id: middlewareId,
      priority: parseInt(middlewarePriority, 10) || 100,
    }));

    const success = await assignMultipleMiddlewares(id, middlewaresToAdd);
    if (success) {
      closeModal();
      setSelectedMiddlewaresToAdd([]);
      setMiddlewarePriority(100);
    } else {
      alert(`Failed to assign middlewares. ${resourceError || 'Check console for details.'}`);
    }
  };

  // Confirm middleware removal
  const confirmRemoveMiddleware = (middlewareId) => {
    if (isDisabled) return;
    clearError();
    setMiddlewareToRemove(middlewareId);
    setShowRemoveMiddlewareModal(true);
  };

  // Handle middleware removal after confirmation
  const handleRemoveMiddlewareConfirmed = async () => {
    if (!middlewareToRemove) return;
    
    const success = await removeMiddleware(id, middlewareToRemove);
    setShowRemoveMiddlewareModal(false);
    setMiddlewareToRemove(null);
    
    if (!success) {
      alert(`Failed to remove middleware. ${resourceError || 'Check console for details.'}`);
    }
  };

  // Cancel middleware removal
  const cancelRemoveMiddleware = () => {
    setShowRemoveMiddlewareModal(false);
    setMiddlewareToRemove(null);
  };

  // Service Assignment
  const handleAssignService = async (serviceId) => {
    if (isDisabled) return;
    clearError();
    try {
      await ResourceService.assignServiceToResource(id, { service_id: serviceId });
      await fetchResourceService(); // Re-fetch the assigned service
      setShowServiceModal(false);
    } catch (err) {
      const errorMsg = `Failed to assign service: ${err.message || 'Unknown error'}`;
      setServicesError(errorMsg); // Use service context error
      alert(errorMsg);
      console.error('Error assigning service:', err);
    }
  };

  // Confirm service removal
  const confirmRemoveService = () => {
    if (isDisabled) return;
    clearError();
    setShowRemoveServiceModal(true);
  };

  // Handle service removal after confirmation
  const handleRemoveServiceConfirmed = async () => {
    try {
      await ResourceService.removeServiceFromResource(id);
      await fetchResourceService(); // Re-fetch (should be null now)
      setShowRemoveServiceModal(false);
    } catch (err) {
      const errorMsg = `Failed to remove service assignment: ${err.message || 'Unknown error'}`;
      setServicesError(errorMsg);
      alert(errorMsg);
      console.error('Error removing service assignment:', err);
    }
  };

  // Cancel service removal
  const cancelRemoveService = () => {
    setShowRemoveServiceModal(false);
  };

  // Render service summary
  const renderServiceSummary = (service) => {
    if (!service || !service.config) return 'Details unavailable';
    const config = typeof service.config === 'string' ? JSON.parse(service.config || '{}') : (service.config || {});

    switch (service.type) {
        case 'loadBalancer':
            const servers = config.servers || [];
            const serverInfo = servers.map(s => s.url || s.address).join(', ');
            return `Servers: ${serverInfo || 'None'}`;
        case 'weighted':
            const weightedServices = config.services || [];
            const weightedInfo = weightedServices.map(s => `${s.name}(${s.weight})`).join(', ');
            return `Weighted: ${weightedInfo || 'None'}`;
        case 'mirroring':
            const mirrors = config.mirrors || [];
            return `Primary: ${config.service || 'N/A'}, Mirrors: ${mirrors.length}`;
        case 'failover':
            return `Main: ${config.service || 'N/A'}, Fallback: ${config.fallback || 'N/A'}`;
        default: return `Type: ${service.type}`;
    }
  };

  // Config Updates
  const handleUpdateConfig = async (configType, data) => {
    if (isDisabled) return;
    clearError();
    const success = await updateResourceConfig(id, configType, data);
    if (success) {
      closeModal();
    } else {
      alert(`Failed to update ${configType} configuration. ${resourceError || 'Check console for details.'}`);
    }
  };

  const handleUpdateRouterPriority = async () => {
    if (isDisabled) return;
    await handleUpdateConfig('priority', { router_priority: routerPriority });
  };

  // Header Modal Helpers
  const addHeader = () => {
    if (!headerInput.key.trim()) {
      alert('Header name cannot be empty.');
      return;
    }
    setConfig(prev => ({
      ...prev,
      customHeaders: { ...prev.customHeaders, [headerInput.key.trim()]: headerInput.value },
    }));
    setHeaderInput({ key: '', value: '' });
  };

  const removeHeader = (keyToRemove) => {
    setConfig(prev => {
      const { [keyToRemove]: _, ...remainingHeaders } = prev.customHeaders;
      return { ...prev, customHeaders: remainingHeaders };
    });
  };

  // Deletion
  const confirmDeleteResource = () => {
    clearError();
    setShowDeleteModal(true);
  };

  const handleDeleteResourceConfirmed = async () => {
    const success = await deleteResource(id);
    if (success) navigateTo('resources');
    // Error handled by context otherwise
    setShowDeleteModal(false);
  };

  const cancelDeleteResource = () => {
    setShowDeleteModal(false);
  };

  // --- Modal Rendering ---
  // Define the renderModal function within the component scope
  const renderConfigModal = () => {
    if (!modal.isOpen) return null;

    switch (modal.type) {
      case 'http':
        return (
          <HTTPConfigModal
            initialEntrypoints={config.entrypoints} // Pass initial value
            onSave={(data) => handleUpdateConfig('http', data)}
            onClose={closeModal}
            isDisabled={isDisabled}
          />
        );
      case 'tls':
        return (
          <TLSConfigModal
            resource={selectedResource}
            initialTlsDomains={config.tlsDomains} // Pass initial value
            onSave={(data) => handleUpdateConfig('tls', data)}
            onClose={closeModal}
            isDisabled={isDisabled}
          />
        );
      case 'tcp':
        return (
          <TCPConfigModal
            resource={selectedResource}
            initialTcpEnabled={config.tcpEnabled}      // Pass initial values
            initialTcpEntrypoints={config.tcpEntrypoints}
            initialTcpSNIRule={config.tcpSNIRule}
            resourceHost={selectedResource.host}
            onSave={(data) => handleUpdateConfig('tcp', data)}
            onClose={closeModal}
            isDisabled={isDisabled}
          />
        );
      case 'headers':
        return (
          <HeadersConfigModal
            initialCustomHeaders={config.customHeaders} // Pass initial value
            headerKey={headerInput.key}                  // Manage input state locally or pass down
            setHeaderKey={(key) => setHeaderInput(prev => ({ ...prev, key }))}
            headerValue={headerInput.value}
            setHeaderValue={(value) => setHeaderInput(prev => ({ ...prev, value }))}
            addHeader={addHeader}                       // Pass add/remove handlers
            removeHeader={removeHeader}
            // onSave will be triggered by the form submit inside the modal
            onSave={(data) => handleUpdateConfig('headers', data)}
            onClose={closeModal}
            isDisabled={isDisabled}
          />
        );
      // Middleware modal is rendered separately below using ModalWrapper
      default:
        return null;
    }
  };


  // --- Main Render ---
  return (
    <div className="space-y-6">
      {/* Header and Disabled Warning */}
      <div className="mb-6 flex items-center flex-wrap gap-2">
           <button
             onClick={() => navigateTo('resources')}
             className="btn btn-secondary text-sm mr-4"
             aria-label="Back to resources list"
           >
             &larr; Back
           </button>
           <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mr-3">
             Resource: {selectedResource.host}
           </h1>
           {isDisabled && (
             <span className="badge badge-error">
               Disabled (Removed from Data Source)
             </span>
           )}
      </div>

      {isDisabled && (
          <div className="p-4 rounded-md bg-red-50 dark:bg-red-900 border border-red-300 dark:border-red-600">
              {/* ... disabled warning content ... */}
               <div className="flex">
                 <div className="flex-shrink-0">
                   {/* Error Icon */}
                   <svg className="h-5 w-5 text-red-500 dark:text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                     <path fillRule="evenodd" d="M8.485 3.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.875c.673 1.167-.17 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 3.495zM10 15.5a1 1 0 100-2 1 1 0 000 2zm-1.1-4.062l.25-4.5a.85.85 0 111.7 0l.25 4.5a.85.85 0 11-1.7 0z" clipRule="evenodd" />
                   </svg>
                 </div>
                 <div className="ml-3">
                   <p className="text-sm font-medium text-red-800 dark:text-red-200">
                       This resource is currently disabled.
                   </p>
                   <p className="mt-1 text-sm text-red-700 dark:text-red-300">
                       Configuration changes are saved but inactive. You can permanently delete the record.
                   </p>
                   <div className="mt-2">
                     <button onClick={confirmDeleteResource} className="text-sm text-red-700 dark:text-red-300 underline font-medium hover:text-red-600 dark:hover:text-red-200">
                       Permanently Delete Record
                     </button>
                   </div>
                 </div>
               </div>
          </div>
      )}

      {/* Inline Error Display */}
       {error && !modal.isOpen && (
            <ErrorMessage
                message={error}
                onDismiss={clearError}
            />
        )}

      {/* Resource Details Card */}
      <div className="card p-6">
        <h2 className="text-xl font-semibold mb-4 text-gray-800 dark:text-gray-200">Resource Details</h2>
         <div className="grid grid-cols-1 md:grid-cols-2 gap-x-6 gap-y-4">
              <div>
                  <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Host</p>
                  <div className="flex items-center mt-1">
                      <p className="font-medium text-gray-900 dark:text-gray-100">{selectedResource.host}</p>
                      <a href={`https://${selectedResource.host}`} target="_blank" rel="noopener noreferrer" className="ml-3 btn-link text-xs">
                          Visit &rarr;
                      </a>
                  </div>
              </div>
              <div>
                  <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Default Service ID</p>
                  <p className="mt-1 font-medium text-gray-900 dark:text-gray-100">{selectedResource.service_id}<span className="text-gray-400 dark:text-gray-500">@http</span></p>
              </div>
              <div>
                  <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Status</p>
                  <p className="mt-1">
                      <span className={`badge ${isDisabled ? 'badge-error' : assignedMiddlewares.length > 0 ? 'badge-success' : 'badge-warning'}`}>
                          {isDisabled ? 'Disabled' : assignedMiddlewares.length > 0 ? 'Protected' : 'Not Protected'}
                      </span>
                  </p>
              </div>
               <div>
                  <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Resource ID</p>
                  <p className="mt-1 font-medium text-gray-900 dark:text-gray-100 font-mono text-xs break-all">{selectedResource.id}</p>
              </div>
               <div>
                   <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Data Source</p>
                   <p className="mt-1 font-medium text-gray-900 dark:text-gray-100 capitalize">{selectedResource.source_type || 'Unknown'}</p>
               </div>
         </div>
      </div>

      {/* Router Configuration Card */}
       <div className="card p-6">
         <h2 className="text-xl font-semibold mb-4 text-gray-800 dark:text-gray-200">Router Configuration</h2>
         <div className="flex flex-wrap gap-3 mb-6">
           <button onClick={() => openConfigModal('http')} disabled={isDisabled} className="btn btn-primary text-sm bg-blue-600 hover:bg-blue-700">HTTP Entrypoints</button>
           <button onClick={() => openConfigModal('tls')} disabled={isDisabled} className="btn btn-primary text-sm bg-green-600 hover:bg-green-700">TLS Domains</button>
           <button onClick={() => openConfigModal('tcp')} disabled={isDisabled} className="btn btn-primary text-sm bg-purple-600 hover:bg-purple-700">TCP Routing</button>
           <button onClick={() => openConfigModal('headers')} disabled={isDisabled} className="btn btn-primary text-sm bg-yellow-500 hover:bg-yellow-600">Custom Headers</button>
         </div>
         <div className="p-4 bg-gray-50 dark:bg-gray-700 rounded border border-gray-200 dark:border-gray-600">
             <h3 className="font-medium mb-3 text-gray-700 dark:text-gray-300">Current Settings</h3>
             <div className="grid grid-cols-1 md:grid-cols-2 gap-x-6 gap-y-3 text-sm">
                 {/* Details */}
                 <div><strong className="text-gray-500 dark:text-gray-400">HTTP Entrypoints:</strong> <span className="font-medium text-gray-900 dark:text-gray-100">{config.entrypoints || 'websecure'}</span></div>
                 <div><strong className="text-gray-500 dark:text-gray-400">TLS SANs:</strong> <span className="font-medium text-gray-900 dark:text-gray-100">{config.tlsDomains || 'None'}</span></div>
                 <div><strong className="text-gray-500 dark:text-gray-400">TCP SNI Routing:</strong> <span className={`font-medium ${config.tcpEnabled ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>{config.tcpEnabled ? 'Enabled' : 'Disabled'}</span></div>
                 {config.tcpEnabled && <div><strong className="text-gray-500 dark:text-gray-400">TCP Entrypoints:</strong> <span className="font-medium text-gray-900 dark:text-gray-100">{config.tcpEntrypoints || 'tcp'}</span></div>}
                 {config.tcpEnabled && config.tcpSNIRule && <div className="md:col-span-2"><strong className="text-gray-500 dark:text-gray-400">TCP SNI Rule:</strong> <code className="text-xs font-mono bg-gray-200 dark:bg-gray-600 px-1 py-0.5 rounded">{config.tcpSNIRule}</code></div>}
                 {Object.keys(config.customHeaders || {}).length > 0 && (
                    <div className="md:col-span-2"><strong className="text-gray-500 dark:text-gray-400">Custom Headers:</strong>
                         <ul className="list-disc pl-5 mt-1">
                             {Object.entries(config.customHeaders).map(([key, value]) => (
                                 <li key={key}><code className="text-xs font-mono bg-gray-200 dark:bg-gray-600 px-1 py-0.5 rounded">{key}: {value || '(empty)'}</code></li>
                             ))}
                         </ul>
                     </div>
                 )}
             </div>
         </div>
       </div>


      {/* Service Configuration Card */}
       <div className="card p-6">
         {/* ... Service config content ... */}
          <h2 className="text-xl font-semibold mb-4 text-gray-800 dark:text-gray-200">Service Configuration</h2>
             <div className="mb-4">
                 {servicesLoading && !resourceService && !servicesError ? ( // Show loading only if no service AND no error
                     <div className="text-center py-4 text-gray-500 dark:text-gray-400"><LoadingSpinner size="sm" message="Loading service info..." /></div>
                 ) : servicesError ? ( // Show service-specific error here
                      <ErrorMessage message={servicesError} onDismiss={() => setServicesError(null)} />
                 ) : resourceService ? (
                     <div className="border dark:border-gray-600 rounded p-4 bg-gray-50 dark:bg-gray-700">
                         {/* ... details of assigned service ... */}
                          <div className="flex flex-col sm:flex-row justify-between items-start gap-2">
                             <div className="flex-1">
                                 <h3 className="font-semibold text-gray-900 dark:text-gray-100">{resourceService.name}</h3>
                                 <div className="flex items-center gap-2 mt-1">
                                     <span className="badge badge-info bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200">{resourceService.type}</span>
                                     <span className="text-xs font-mono text-gray-500 dark:text-gray-400">({resourceService.id}@file)</span>
                                 </div>
                                 <div className="mt-2 text-sm text-gray-600 dark:text-gray-300 break-words">
                                     {renderServiceSummary(resourceService)}
                                 </div>
                             </div>
                             <div className="flex space-x-2 flex-shrink-0 mt-2 sm:mt-0 self-start sm:self-center">
                                 <button onClick={() => navigateTo('service-form', resourceService.id)} className="btn-link text-xs" disabled={isDisabled} title="Edit the base service definition">Edit Base Service</button>
                                 <button onClick={confirmRemoveService} className="btn-link text-xs text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-300" disabled={isDisabled} title="Remove custom service assignment">Remove Assignment</button>
                             </div>
                         </div>
                         <p className="text-xs text-gray-500 dark:text-gray-400 mt-3 italic">
                             This resource uses the custom service configured above.
                         </p>
                     </div>
                 ) : (
                     <div className="text-center py-4 text-gray-500 dark:text-gray-400 border dark:border-gray-600 rounded bg-gray-50 dark:bg-gray-700">
                         <p>Using default service: <code className="text-xs font-mono bg-gray-200 dark:bg-gray-600 px-1 rounded">{selectedResource.service_id}@http</code></p>
                         <p className="text-xs mt-1">Assign a custom service to override routing behavior.</p>
                     </div>
                 )}
             </div>
             <button
                 onClick={() => setShowServiceModal(true)}
                 disabled={isDisabled}
                 className="btn btn-primary bg-purple-600 hover:bg-purple-700 text-sm"
             >
                 {resourceService ? 'Change Assigned Service' : 'Assign Custom Service'}
             </button>
       </div>

      {/* Router Priority Card */}
      <div className="card p-6">
        <h2 className="text-xl font-semibold mb-4 text-gray-800 dark:text-gray-200">Router Priority</h2>
        {/* ... Priority content ... */}
          <div className="mb-4">
            <p className="text-sm text-gray-700 dark:text-gray-300">
              Control router evaluation order. Higher numbers are checked first (default 100).
            </p>
             <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
               Useful for overlapping rules or specific overrides.
             </p>
          </div>
          <div className="flex items-center gap-3">
            <label htmlFor="router-priority-input" className="form-label mb-0 text-sm">Priority:</label>
            <input
              id="router-priority-input"
              type="number"
              value={routerPriority}
              onChange={(e) => setRouterPriority(parseInt(e.target.value) || 100)}
              className="form-input w-24 text-sm"
              min="1"
              max="1000" // Adjust max as needed
              disabled={isDisabled}
            />
            <button
              onClick={handleUpdateRouterPriority}
              className="btn btn-secondary text-sm"
              disabled={isDisabled || (selectedResource && routerPriority === selectedResource.router_priority)}
            >
              Update Priority
            </button>
          </div>
      </div>

      {/* Attached Middlewares Card */}
      <div className="card p-6">
        {/* ... Middlewares content ... */}
         <div className="flex justify-between items-center mb-4">
           <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200">Attached Middlewares</h2>
           <button
             onClick={() => openConfigModal('middlewares')} // Use openConfigModal
             disabled={isDisabled || availableMiddlewares.length === 0}
             className="btn btn-primary text-sm"
             title={availableMiddlewares.length === 0 ? "All available middlewares are assigned" : "Assign middlewares"}
           >
             Add Middleware
           </button>
         </div>
         {assignedMiddlewares.length === 0 ? (
           <div className="text-center py-6 text-gray-500 dark:text-gray-400 border dark:border-gray-600 rounded bg-gray-50 dark:bg-gray-700">
             <p>No middlewares attached.</p>
           </div>
         ) : (
           <div className="overflow-x-auto">
             <table className="table min-w-full">
               <thead>
                 <tr>
                   <th>Middleware</th>
                   <th className="text-center">Priority</th>
                   <th className="text-right">Actions</th>
                 </tr>
               </thead>
               <tbody>
                 {assignedMiddlewares.map(middleware => {
                   const middlewareDetails = middlewares.find(m => m.id === middleware.id) || {
                     id: middleware.id, name: middleware.name || middleware.id, type: 'unknown',
                   };
                   return (
                     <tr key={middleware.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                       <td className="py-2 px-6">
                         {formatMiddlewareDisplay(middlewareDetails)}
                       </td>
                       <td className="py-2 px-6 text-center text-sm text-gray-500 dark:text-gray-400">{middleware.priority}</td>
                       <td className="py-2 px-6 text-right">
                         <button
                           onClick={() => confirmRemoveMiddleware(middleware.id)}
                           className="btn-link text-xs text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-300"
                           disabled={isDisabled}
                         >
                           Remove
                         </button>
                       </td>
                     </tr>
                   );
                 })}
               </tbody>
             </table>
           </div>
         )}
      </div>

      {/* --- Modals --- */}

      {/* Render configuration modals */}
      {renderConfigModal()} {/* Call the defined render function */}

      {/* Middleware Assignment Modal */}
      <ModalWrapper
          title={`Assign Middlewares to ${selectedResource.host}`}
          show={modal.isOpen && modal.type === 'middlewares'}
          onClose={closeModal}
      >
          <form onSubmit={handleAssignMiddlewareSubmit}>
              <div className="modal-body">
                  {isDisabled && (
                     <div className="mb-4 p-3 text-sm text-red-700 bg-red-100 dark:bg-red-900 dark:text-red-200 border border-red-300 dark:border-red-600 rounded-md">
                         Cannot assign middlewares while the resource is disabled.
                     </div>
                   )}
                  {availableMiddlewares.length === 0 && !isDisabled ? (
                      <div className="text-center py-4 text-gray-500 dark:text-gray-400">
                          <p>All available middlewares are assigned.</p>
                          <button type="button" onClick={() => { navigateTo('middleware-form'); closeModal(); }} className="mt-2 btn-link text-sm">Create New</button>
                      </div>
                  ) : (
                      <>
                          <div className="mb-4">
                              <label htmlFor="middleware-select-add" className="form-label">Select Middlewares</label>
                              <select
                                  id="middleware-select-add"
                                  multiple
                                  value={selectedMiddlewaresToAdd}
                                  onChange={handleMiddlewareSelectionChange} // Corrected handler
                                  className="form-input"
                                  size={Math.min(8, availableMiddlewares.length || 1)}
                                  disabled={isDisabled}
                              >
                                  {availableMiddlewares.map(middleware => (
                                      <option key={middleware.id} value={middleware.id}>
                                          {middleware.name} ({middleware.type})
                                      </option>
                                  ))}
                              </select>
                              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Hold Ctrl/Cmd to select multiple.</p>
                          </div>
                          <div className="mb-4">
                              <label htmlFor="middleware-priority-add" className="form-label">Priority</label>
                              <input
                                  id="middleware-priority-add"
                                  type="number"
                                  value={middlewarePriority}
                                  onChange={(e) => setMiddlewarePriority(e.target.value)} // Corrected handler
                                  className="form-input w-full"
                                  min="1"
                                  max="1000"
                                  required
                                  disabled={isDisabled}
                              />
                              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Higher priority runs first (1-1000). Default: 100.</p>
                          </div>
                      </>
                  )}
              </div>
              <div className="modal-footer">
                  <button type="button" onClick={closeModal} className="btn btn-secondary">Cancel</button>
                  <button
                      type="submit"
                      className="btn btn-primary"
                      disabled={isDisabled || availableMiddlewares.length === 0 || selectedMiddlewaresToAdd.length === 0}
                  >
                      Assign Selected
                  </button>
              </div>
          </form>
      </ModalWrapper>


      {/* Service Selection Modal */}
      {showServiceModal && (
        <ServiceSelectModal
          services={services}
          currentServiceId={resourceService?.id}
          onSelect={handleAssignService}
          onClose={() => setShowServiceModal(false)}
          isDisabled={isDisabled}
        />
      )}

      {/* Delete Resource Confirmation Modal */}
      <ConfirmationModal
        show={showDeleteModal}
        title="Confirm Resource Deletion"
        message={`Are you sure you want to delete the resource "${selectedResource?.host}"?`}
        details={isDisabled 
          ? "This action cannot be undone." 
          : "Resource must be disabled (removed from data source) to be deleted. This action cannot be undone."}
        confirmText="Delete Resource"
        cancelText="Cancel"
        onConfirm={handleDeleteResourceConfirmed}
        onCancel={cancelDeleteResource}
      />

      {/* Remove Middleware Confirmation Modal */}
      <ConfirmationModal
        show={showRemoveMiddlewareModal}
        title="Confirm Middleware Removal"
        message="Are you sure you want to remove this middleware?"
        details="This will remove the middleware from this resource but won't delete the middleware itself."
        confirmText="Remove"
        cancelText="Cancel"
        onConfirm={handleRemoveMiddlewareConfirmed}
        onCancel={cancelRemoveMiddleware}
      />

      {/* Remove Service Confirmation Modal */}
      <ConfirmationModal
        show={showRemoveServiceModal}
        title="Confirm Service Removal"
        message="Remove custom service assignment?"
        details="The resource will use its default service."
        confirmText="Remove Assignment"
        cancelText="Cancel"
        onConfirm={handleRemoveServiceConfirmed}
        onCancel={cancelRemoveService}
      />
    </div>
  );
};

export default ResourceDetail;
