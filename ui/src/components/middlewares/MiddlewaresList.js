// ui/src/components/middlewares/MiddlewaresList.js
import React, { useEffect, useState } from 'react';
import { useMiddlewares } from '../../contexts/MiddlewareContext';
import { LoadingSpinner, ErrorMessage, ConfirmationModal } from '../common';

const MiddlewaresList = ({ navigateTo }) => {
  const {
    middlewares,
    loading,
    error,
    fetchMiddlewares,
    deleteMiddleware,
    setError // Get setError to clear errors
  } = useMiddlewares();

  const [searchTerm, setSearchTerm] = useState('');
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [middlewareToDelete, setMiddlewareToDelete] = useState(null);

  useEffect(() => {
    fetchMiddlewares();
  }, [fetchMiddlewares]);

  const confirmDelete = (middleware) => {
    setMiddlewareToDelete(middleware);
    setShowDeleteModal(true);
    setError(null); // Clear previous errors when opening modal
  };

  const handleDeleteConfirmed = async () => {
    if (!middlewareToDelete) return;

    const success = await deleteMiddleware(middlewareToDelete.id);
    if (success) {
        setShowDeleteModal(false);
        setMiddlewareToDelete(null);
        // Optionally show a success notification here
    } else {
        // Error is handled by context and will be shown
        // Keep the modal open if deletion fails
        alert(`Failed to delete middleware: ${error || 'Unknown error'}`);
    }
  };

  const cancelDelete = () => {
    setShowDeleteModal(false);
    setMiddlewareToDelete(null);
    setError(null); // Clear error on cancel
  };

  const filteredMiddlewares = middlewares.filter(
    (middleware) =>
      (middleware.name && middleware.name.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (middleware.type && middleware.type.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (middleware.id && middleware.id.toLowerCase().includes(searchTerm.toLowerCase()))
  );

  if (loading && middlewares.length === 0) {
    return <LoadingSpinner message="Loading middlewares..." />;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Middlewares</h1>

       {error && !showDeleteModal && ( // Only show list error if delete modal isn't open
          <ErrorMessage
              message="Failed to manage middlewares"
              details={error}
              onRetry={fetchMiddlewares}
              onDismiss={() => setError(null)}
          />
       )}

      {/* Search and Actions */}
      <div className="flex flex-col sm:flex-row justify-between items-center gap-4">
        <div className="relative w-full sm:w-64">
          <label htmlFor="middleware-search" className="sr-only">Search Middlewares</label>
          <input
            id="middleware-search"
            type="text"
            placeholder="Search by name, type, ID..."
            className="form-input w-full"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
           {/* Optional: Add search icon */}
        </div>
        <div className="flex space-x-3 w-full sm:w-auto">
          <button
            onClick={fetchMiddlewares}
            className="btn btn-secondary flex-1 sm:flex-none"
            disabled={loading}
          >
            {loading ? 'Refreshing...' : 'Refresh'}
          </button>
          <button
            onClick={() => navigateTo('middleware-form')}
            className="btn btn-primary flex-1 sm:flex-none"
          >
            Create Middleware
          </button>
        </div>
      </div>

      {/* Middlewares Table */}
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
              {filteredMiddlewares.map((middleware) => (
                <tr key={middleware.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                  <td className="py-3 px-6"> {/* Adjusted padding */}
                    <div className="font-medium text-gray-900 dark:text-gray-100">{middleware.name}</div>
                    <div className="text-xs font-mono text-gray-500 dark:text-gray-400">{middleware.id}</div>
                  </td>
                  <td className="py-3 px-6">
                    <span className="badge badge-info bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                      {middleware.type}
                    </span>
                  </td>
                  <td className="py-3 px-6 whitespace-nowrap text-right space-x-3">
                    <button
                      onClick={() => navigateTo('middleware-form', middleware.id)}
                      className="btn-link text-sm"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => confirmDelete(middleware)}
                      className="btn-link text-sm text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-300"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
              {filteredMiddlewares.length === 0 && (
                <tr>
                  <td colSpan="3" className="py-4 text-center text-gray-500 dark:text-gray-400">
                    No middlewares found.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      <ConfirmationModal
          show={showDeleteModal && !!middlewareToDelete}
          title="Confirm Deletion"
          message={`Are you sure you want to delete the middleware "${middlewareToDelete?.name}"?`}
          details="This action cannot be undone and may affect resources using it."
          confirmText="Delete"
          cancelText="Cancel"
          onConfirm={handleDeleteConfirmed}
          onCancel={cancelDelete}
      />
    </div>
  );
};

export default MiddlewaresList;
