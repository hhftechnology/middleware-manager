import React, { useEffect, useState } from 'react';
import { useMiddlewares } from '../../contexts/MiddlewareContext';
import { LoadingSpinner, ErrorMessage } from '../common';
import ConfirmationModal from '../common/ConfirmationModal';
import { showErrorToast, showSuccessToast } from '../common/Toast';

const MiddlewaresList = ({ navigateTo }) => {
  const {
    middlewares,
    loading,
    error,
    fetchMiddlewares,
    deleteMiddleware
  } = useMiddlewares();
  
  const [searchTerm, setSearchTerm] = useState('');
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [middlewareToDelete, setMiddlewareToDelete] = useState(null);
  const [isDeleting, setIsDeleting] = useState(false);

  useEffect(() => {
    fetchMiddlewares();
  }, [fetchMiddlewares]);

  const confirmDelete = (middleware) => {
    setMiddlewareToDelete(middleware);
    setShowDeleteModal(true);
  };

  const handleDeleteMiddleware = async () => {
    if (!middlewareToDelete) return;
    
    try {
      setIsDeleting(true);
      await deleteMiddleware(middlewareToDelete.id);
      showSuccessToast(`Middleware "${middlewareToDelete.name}" was successfully deleted.`);
      setShowDeleteModal(false);
      setMiddlewareToDelete(null);
    } catch (err) {
      showErrorToast('Failed to delete middleware', err.message);
      console.error('Delete middleware error:', err);
    } finally {
      setIsDeleting(false);
    }
  };

  const cancelDelete = () => {
    setShowDeleteModal(false);
    setMiddlewareToDelete(null);
  };

  const filteredMiddlewares = middlewares.filter(
    (middleware) =>
      middleware.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      middleware.type.toLowerCase().includes(searchTerm.toLowerCase())
  );

  if (loading && !middlewares.length) {
    return <LoadingSpinner message="Loading middlewares..." />;
  }

  if (error) {
    return (
      <ErrorMessage 
        message="Failed to load middlewares" 
        details={error}
        onRetry={fetchMiddlewares}
      />
    );
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Middlewares</h1>
      <div className="mb-6 flex justify-between">
        <div className="relative w-64">
          <input
            type="text"
            placeholder="Search middlewares..."
            className="w-full px-4 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <div className="space-x-3">
          <button
            onClick={fetchMiddlewares}
            className="px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
            disabled={loading}
          >
            Refresh
          </button>
          <button
            onClick={() => navigateTo('middleware-form')}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Create Middleware
          </button>
        </div>
      </div>
      
      <div className="bg-white shadow rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Name
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Type
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {filteredMiddlewares.map((middleware) => (
              <tr key={middleware.id}>
                <td className="px-6 py-4 whitespace-nowrap">
                  {middleware.name}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800">
                    {middleware.type}
                    {middleware.type === 'chain' && " (Middleware Chain)"}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex justify-end space-x-2">
                    <button
                      onClick={() => navigateTo('middleware-form', middleware.id)}
                      className="text-blue-600 hover:text-blue-900 mr-3"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => confirmDelete(middleware)}
                      className="text-red-600 hover:text-red-900"
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            ))}
            {filteredMiddlewares.length === 0 && (
              <tr>
                <td
                  colSpan="3"
                  className="px-6 py-4 text-center text-gray-500"
                >
                  No middlewares found
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Professional Delete Confirmation Modal */}
      <ConfirmationModal
        isOpen={showDeleteModal}
        title="Confirm Deletion"
        message={middlewareToDelete ? `Are you sure you want to delete the middleware "${middlewareToDelete.name}"?` : ''}
        details="This action cannot be undone and may affect any resources currently using this middleware."
        confirmText="Delete"
        cancelText="Cancel"
        confirmButtonType="danger"
        onConfirm={handleDeleteMiddleware}
        onCancel={cancelDelete}
        isProcessing={isDeleting}
      />
    </div>
  );
};

export default MiddlewaresList;
