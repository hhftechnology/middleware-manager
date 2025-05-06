// ui/src/components/services/ServicesList.js
import React, { useEffect, useState, useContext } from 'react';
import { ServiceContext } from '../../contexts/ServiceContext';
import { LoadingSpinner, ErrorMessage, ConfirmationModal } from '../common';

const ServicesList = ({ navigateTo }) => {
  const {
    services,
    loading,
    error,
    loadServices,
    removeService,
    setError
  } = useContext(ServiceContext);

  const [searchTerm, setSearchTerm] = useState('');
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [serviceToDelete, setServiceToDelete] = useState(null);

  useEffect(() => {
    loadServices();
  }, [loadServices]);

  const confirmDelete = (service) => {
    setServiceToDelete(service);
    setShowDeleteModal(true);
    setError(null); // Clear previous errors
  };

  const handleDeleteService = async () => {
    if (!serviceToDelete) return;

    const success = await removeService(serviceToDelete.id);
    if (success) {
        setShowDeleteModal(false);
        setServiceToDelete(null);
    } else {
        // Error handled by context
        alert(`Failed to delete service: ${error || 'Unknown error'}`);
    }
  };

  const cancelDelete = () => {
    setShowDeleteModal(false);
    setServiceToDelete(null);
    setError(null); // Clear error on cancel
  };

  const filteredServices = services.filter(
    (service) =>
      (service.name && service.name.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (service.type && service.type.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (service.id && service.id.toLowerCase().includes(searchTerm.toLowerCase()))
  );

  // Initial loading state
  if (loading && services.length === 0) {
    return <LoadingSpinner message="Loading services..." />;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Services</h1>

      {error && !showDeleteModal && ( // Only show list error if delete modal isn't open
        <ErrorMessage
          message="Failed to manage services"
          details={error}
          onRetry={loadServices}
          onDismiss={() => setError(null)}
        />
      )}

      {/* Search and Actions */}
       <div className="flex flex-col sm:flex-row justify-between items-center gap-4">
         <div className="relative w-full sm:w-64">
           <label htmlFor="service-search-list" className="sr-only">Search Services</label>
           <input
             id="service-search-list"
             type="text"
             placeholder="Search by name, type, ID..."
             className="form-input w-full"
             value={searchTerm}
             onChange={(e) => setSearchTerm(e.target.value)}
           />
         </div>
         <div className="flex space-x-3 w-full sm:w-auto">
           <button
             onClick={loadServices}
             className="btn btn-secondary flex-1 sm:flex-none"
             disabled={loading}
           >
             {loading ? 'Refreshing...' : 'Refresh'}
           </button>
           <button
             onClick={() => navigateTo('service-form')}
             className="btn btn-primary flex-1 sm:flex-none"
           >
             Create Service
           </button>
         </div>
       </div>


      {/* Services Table */}
      <div className="card overflow-hidden"> {/* Use card class */}
        <div className="overflow-x-auto">
          <table className="table min-w-full">
            <thead>
              <tr>
                <th>Name / ID</th>
                <th>Type</th>
                <th className="text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredServices.map((service) => (
                <tr key={service.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                  <td className="py-3 px-6">
                     <div className="font-medium text-gray-900 dark:text-gray-100">{service.name}</div>
                     <div className="text-xs font-mono text-gray-500 dark:text-gray-400">{service.id}</div>
                  </td>
                  <td className="py-3 px-6">
                     <span className="badge badge-info bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200">
                       {service.type}
                     </span>
                  </td>
                  <td className="py-3 px-6 whitespace-nowrap text-right space-x-3">
                    <button
                      onClick={() => navigateTo('service-form', service.id)}
                       className="btn-link text-sm"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => confirmDelete(service)}
                       className="btn-link text-sm text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-300"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
              {filteredServices.length === 0 && (
                <tr>
                  <td colSpan="3" className="py-4 text-center text-gray-500 dark:text-gray-400">
                    {searchTerm ? 'No services match your search.' : 'No services found.'}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      <ConfirmationModal
        show={showDeleteModal && !!serviceToDelete}
        title="Confirm Service Deletion"
        message={`Are you sure you want to delete the service "${serviceToDelete?.name}"?`}
        details="This action cannot be undone. Ensure no resources are using this service before deleting."
        confirmText="Delete Service"
        cancelText="Cancel"
        onConfirm={handleDeleteService}
        onCancel={cancelDelete}
      />
    </div>
  );
};

export default ServicesList;
