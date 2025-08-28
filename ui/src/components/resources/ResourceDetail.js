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
  const [middlewarePriority, setMiddlewarePriority] = useState(200);
  const [routerPriority, setRouterPriority] = useState(200);
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
  useEffect(() => {
    if (id) {
      fetchResource(id);
      fetchMiddlewares();
      loadServices();
    } else {
      setResourceError("No resource ID specified.");
    }
  }, [id, fetchResource, fetchMiddlewares, loadServices, setResourceError]);

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
        setResourceService(null); // No custom service assigned
      }
    }
  }, [id, setServicesError]);

  useEffect(() => {
    fetchResourceService();
  }, [fetchResourceService]);

  useEffect(() => {
    if (selectedResource) {
      try {
        let parsedHeaders = {};
        if (selectedResource.custom_headers) {
          if (typeof selectedResource.custom_headers === 'string' && selectedResource.custom_headers.trim()) {
            parsedHeaders = JSON.parse(selectedResource.custom_headers);
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
          customHeaders: parsedHeaders || {}, // Ensure it's an object
        });
        setRouterPriority(selectedResource.router_priority || 200);
      } catch (error) {
        console.error("Error updating local state from resource:", error);
        setResourceError(`Error processing resource data: ${error.message}`);
      }
    }
  }, [selectedResource, setResourceError]);


  // --- Loading & Error Handling ---
  const loading = resourceLoading || middlewaresLoading || servicesLoading;
  const error = resourceError || middlewaresError || servicesError;

  const clearError = () => {
    if (setResourceError) setResourceError(null);
    if (setMiddlewaresError) setMiddlewaresError(null);
    if (setServicesError) setServicesError(null);
  };

  if (loading && !selectedResource) {
    return <LoadingSpinner message="Loading resource details..." />;
  }

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
  if (!selectedResource) return null;


  // --- Helper Functions & Derived State ---
  const assignedMiddlewares = MiddlewareUtils.parseMiddlewares(selectedResource.middlewares);
  const assignedMiddlewareIds = new Set(assignedMiddlewares.map(m => m.id));
  const availableMiddlewares = (middlewares || []).filter(m => !assignedMiddlewareIds.has(m.id));
  const isDisabled = selectedResource.status === 'disabled';

  const openConfigModal = (type) => {
      clearError();
      setModal({ isOpen: true, type });
  }
  const closeModal = () => setModal({ isOpen: false, type: null });


  // --- Action Handlers ---
  const handleMiddlewareSelectionChange = (e) => {
    const selectedOptions = Array.from(e.target.selectedOptions, option => option.value);
    setSelectedMiddlewaresToAdd(selectedOptions);
  };

  const handleAssignMiddlewareSubmit = async (e) => {
    e.preventDefault();
    if (isDisabled || !selectedMiddlewaresToAdd.length) return;
    clearError();

    const middlewaresToAdd = selectedMiddlewaresToAdd.map(middlewareId => ({
      middleware_id: middlewareId,
      priority: parseInt(middlewarePriority, 10) || 200,
    }));

    const success = await assignMultipleMiddlewares(id, middlewaresToAdd);
    if (success) {
      closeModal();
      setSelectedMiddlewaresToAdd([]);
      setMiddlewarePriority(200);
    } else {
      alert(`Failed to assign middlewares. ${resourceError || 'Check console for details.'}`);
    }
  };

  const confirmRemoveMiddleware = (middlewareId) => {
    if (isDisabled) return;
    clearError();
    setMiddlewareToRemove(middlewareId);
    setShowRemoveMiddlewareModal(true);
  };

  const handleRemoveMiddlewareConfirmed = async () => {
    if (!middlewareToRemove) return;
    
    const success = await removeMiddleware(id, middlewareToRemove);
    setShowRemoveMiddlewareModal(false);
    setMiddlewareToRemove(null);
    
    if (!success) {
      alert(`Failed to remove middleware. ${resourceError || 'Check console for details.'}`);
    }
  };

  const cancelRemoveMiddleware = () => {
    setShowRemoveMiddlewareModal(false);
    setMiddlewareToRemove(null);
  };

  const handleAssignService = async (serviceIdToAssign) => {
    if (isDisabled) return;
    clearError();
    try {
      // If serviceIdToAssign is empty, it means we are removing the custom assignment.
      if (!serviceIdToAssign) {
        await ResourceService.removeServiceFromResource(id);
      } else {
        await ResourceService.assignServiceToResource(id, { service_id: serviceIdToAssign });
      }
      await fetchResourceService(); 
      setShowServiceModal(false);
    } catch (err) {
      const errorMsg = `Failed to assign/remove service: ${err.message || 'Unknown error'}`;
      setServicesError(errorMsg); 
      alert(errorMsg);
      console.error('Error assigning/removing service:', err);
    }
  };
  
  const confirmRemoveService = () => {
    if (isDisabled) return;
    clearError();
    setShowRemoveServiceModal(true);
  };

  const handleRemoveServiceConfirmed = async () => {
    // This now calls handleAssignService with an empty ID to signify removal
    await handleAssignService(''); 
    setShowRemoveServiceModal(false);
  };

  const cancelRemoveService = () => {
    setShowRemoveServiceModal(false);
  };

  const renderServiceSummary = (service) => {
    if (!service || !service.config) return 'Details unavailable';
    const serviceConfigObj = typeof service.config === 'string' ? JSON.parse(service.config || '{}') : (service.config || {});

    switch (service.type) {
        case 'loadBalancer':
            const servers = serviceConfigObj.servers || [];
            const serverInfo = servers.map(s => s.url || s.address).join(', ');
            return `Servers: ${serverInfo || 'None'}`;
        case 'weighted':
            const weightedServices = serviceConfigObj.services || [];
            const weightedInfo = weightedServices.map(s => `${s.name}(${s.weight})`).join(', ');
            return `Weighted: ${weightedInfo || 'None'}`;
        case 'mirroring':
            const mirrors = serviceConfigObj.mirrors || [];
            return `Primary: ${serviceConfigObj.service || 'N/A'}, Mirrors: ${mirrors.length}`;
        case 'failover':
            return `Main: ${serviceConfigObj.service || 'N/A'}, Fallback: ${serviceConfigObj.fallback || 'N/A'}`;
        default: return `Type: ${service.type}`;
    }
  };

  const handleUpdateConfig = async (configType, data) => {
    if (isDisabled) return;
    clearError();
    // For clearing, ensure data contains empty/default values
    // e.g., for tls: data = { tls_domains: "" }
    // e.g., for tcp (disabled): data = { tcp_enabled: false, tcp_entrypoints: "", tcp_sni_rule: "" }
    // e.g., for headers: data = { custom_headers: {} }
    const success = await updateResourceConfig(id, configType, data);
    if (success) {
      closeModal(); // Close modal on successful save
    } else {
      // Error should be set in context, alert is for immediate user feedback
      alert(`Failed to update ${configType} configuration. ${resourceError || 'Check console for details.'}`);
    }
  };

  const handleUpdateRouterPriority = async () => {
    if (isDisabled) return;
    await handleUpdateConfig('priority', { router_priority: routerPriority });
  };

  const addHeader = () => { // Used by HeadersConfigModal internally
    if (!headerInput.key.trim()) {
      alert('Header name cannot be empty.');
      return;
    }
    const newHeaders = { ...config.customHeaders, [headerInput.key.trim()]: headerInput.value };
    setConfig(prev => ({ ...prev, customHeaders: newHeaders }));
    setHeaderInput({ key: '', value: '' }); // Reset for next input
  };

  const removeHeader = (keyToRemove) => { // Used by HeadersConfigModal internally
    const { [keyToRemove]: _, ...remainingHeaders } = config.customHeaders;
    setConfig(prev => ({ ...prev, customHeaders: remainingHeaders }));
  };

  const confirmDeleteResource = () => {
    clearError();
    setShowDeleteModal(true);
  };

  const handleDeleteResourceConfirmed = async () => {
    const success = await deleteResource(id);
    if (success) navigateTo('resources');
    setShowDeleteModal(false);
  };

  const cancelDeleteResource = () => {
    setShowDeleteModal(false);
  };
  
  const renderConfigModal = () => {
    if (!modal.isOpen) return null;

    switch (modal.type) {
      case 'http':
        return (
          <HTTPConfigModal
            entrypoints={config.entrypoints} // Pass current value from local state
            onSave={(data) => handleUpdateConfig('http', data)}
            onClose={closeModal}
            isDisabled={isDisabled}
          />
        );
      case 'tls':
        return (
          <TLSConfigModal
            resource={selectedResource}
            tlsDomains={config.tlsDomains} // Pass current value
            onSave={(data) => handleUpdateConfig('tls', data)} // data will be { tls_domains: "..." } or { tls_domains: "" }
            onClose={closeModal}
            isDisabled={isDisabled}
          />
        );
      case 'tcp':
        return (
          <TCPConfigModal
            resource={selectedResource}
            tcpEnabled={config.tcpEnabled}      
            tcpEntrypoints={config.tcpEntrypoints}
            tcpSNIRule={config.tcpSNIRule}
            resourceHost={selectedResource.host} // Pass host for default SNI rule
            onSave={(data) => handleUpdateConfig('tcp', data)} // data will include tcp_enabled, and empty strings for others if disabled
            onClose={closeModal}
            isDisabled={isDisabled}
          />
        );
      case 'headers':
        return (
          <HeadersConfigModal
            customHeaders={config.customHeaders} // Pass current value
            // setParentCustomHeaders is optional, modal can manage its state for adding/removing
            // and then call onSave with the final state.
            onSave={(data) => handleUpdateConfig('headers', data)} // data will be { custom_headers: {...} } or { custom_headers: {} }
            onClose={closeModal}
            isDisabled={isDisabled}
          />
        );
      default:
        return null;
    }
  };

  return (
    <div className="space-y-6">
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
               <div className="flex">
                 <div className="flex-shrink-0">
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

       {error && !modal.isOpen && (
            <ErrorMessage
                message={error}
                onDismiss={clearError}
            />
        )}

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
                 {Object.keys(config.customHeaders || {}).length === 0 && (
                     <div className="md:col-span-2"><strong className="text-gray-500 dark:text-gray-400">Custom Headers:</strong> <span className="font-medium text-gray-900 dark:text-gray-100">None</span></div>
                 )}
             </div>
         </div>
       </div>

       <div className="card p-6">
          <h2 className="text-xl font-semibold mb-4 text-gray-800 dark:text-gray-200">Service Configuration</h2>
             <div className="mb-4">
                 {servicesLoading && !resourceService && !servicesError ? ( 
                     <div className="text-center py-4 text-gray-500 dark:text-gray-400"><LoadingSpinner size="sm" message="Loading service info..." /></div>
                 ) : servicesError ? ( 
                      <ErrorMessage message={servicesError} onDismiss={() => setServicesError(null)} />
                 ) : resourceService ? (
                     <div className="border dark:border-gray-600 rounded p-4 bg-gray-50 dark:bg-gray-700">
                          <div className="flex flex-col sm:flex-row justify-between items-start gap-2">
                             <div className="flex-1">
                                 <h3 className="font-semibold text-gray-900 dark:text-gray-100">{resourceService.name}</h3>
                                 <div className="flex items-center gap-2 mt-1">
                                     <span className="badge badge-info bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200">{resourceService.type}</span>
                                     <span className="text-xs font-mono text-gray-500 dark:text-gray-400">({resourceService.id})</span>
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
                         <p>Using default service: <code className="text-xs font-mono bg-gray-200 dark:bg-gray-600 px-1 rounded">{selectedResource.service_id}</code></p>
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

      <div className="card p-6">
        <h2 className="text-xl font-semibold mb-4 text-gray-800 dark:text-gray-200">Router Priority</h2>
          <div className="mb-4">
            <p className="text-sm text-gray-700 dark:text-gray-300">
              Control router evaluation order. Higher numbers are checked first (default 200).
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
              onChange={(e) => setRouterPriority(parseInt(e.target.value) || 200)}
              className="form-input w-24 text-sm"
              min="1"
              max="1000" 
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

      <div className="card p-6">
         <div className="flex justify-between items-center mb-4">
           <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200">Attached Middlewares</h2>
           <button
             onClick={() => openConfigModal('middlewares')} 
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
                   const middlewareDetails = (middlewares || []).find(m => m.id === middleware.id) || {
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

      {renderConfigModal()}

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
                                  onChange={handleMiddlewareSelectionChange}
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
                                  onChange={(e) => setMiddlewarePriority(e.target.value)}
                                  className="form-input w-full"
                                  min="1"
                                  max="1000"
                                  required
                                  disabled={isDisabled}
                              />
                              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Higher priority runs first (1-1000). Default: 200.</p>
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

      {showServiceModal && (
        <ServiceSelectModal
          services={services || []} // Ensure services is an array
          currentServiceId={resourceService?.id}
          onSelect={handleAssignService}
          onClose={() => setShowServiceModal(false)}
          isDisabled={isDisabled}
        />
      )}

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

      <ConfirmationModal
        show={showRemoveServiceModal}
        title="Confirm Service Assignment Removal"
        message="Remove custom service assignment?"
        details="The resource will revert to its default service routing."
        confirmText="Remove Assignment"
        cancelText="Cancel"
        onConfirm={handleRemoveServiceConfirmed}
        onCancel={cancelRemoveService}
      />
    </div>
  );
};

export default ResourceDetail;