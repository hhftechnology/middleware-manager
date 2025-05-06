import React, { useEffect, useState } from 'react';
import { useResources } from '../../contexts/ResourceContext';
import { LoadingSpinner, ErrorMessage, ConfirmationModal } from '../common';
import { MiddlewareUtils } from '../../services/api';

/**
 * ResourcesList component displays all resources with filtering and management options
 */
const ResourcesList = ({ navigateTo }) => {
  const { 
    resources, 
    loading, 
    error, 
    fetchResources,
    deleteResource,
    setError
  } = useResources();
  
  const [searchTerm, setSearchTerm] = useState('');
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [resourceToDelete, setResourceToDelete] = useState(null);

  // Load resources when component mounts
  useEffect(() => {
    fetchResources();
  }, [fetchResources]);

  // Open confirmation modal for resource deletion
  const confirmDelete = (resource) => {
    setResourceToDelete(resource);
    setShowDeleteModal(true);
    if (setError) setError(null); // Clear previous errors if setError exists
  };

  // Handle resource deletion after confirmation
  const handleDeleteConfirmed = async () => {
    if (!resourceToDelete) return;
    
    const success = await deleteResource(resourceToDelete.id);
    if (success) {
      setShowDeleteModal(false);
      setResourceToDelete(null);
      // Success already handled in the context
    }
    // Error is handled by context
  };

  // Cancel deletion
  const cancelDelete = () => {
    setShowDeleteModal(false);
    setResourceToDelete(null);
    if (setError) setError(null); // Clear error on cancel if setError exists
  };

  // Filter resources based on search term
  const filteredResources = resources.filter((resource) =>
    resource.host.toLowerCase().includes(searchTerm.toLowerCase())
  );

  if (loading && !resources.length) {
    return <LoadingSpinner message="Loading resources..." />;
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Resources</h1>
      
      {error && !showDeleteModal && (
        <ErrorMessage 
          message="Failed to load resources" 
          details={error}
          onRetry={fetchResources}
          onDismiss={() => setError(null)}
        />
      )}
      
      <div className="mb-6 flex justify-between">
        <div className="relative w-64">
          <input
            type="text"
            placeholder="Search resources..."
            className="w-full px-4 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <button
          onClick={fetchResources}
          className="px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
          disabled={loading}
        >
          Refresh
        </button>
      </div>
      
      <div className="bg-white shadow rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Host
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Middlewares
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {filteredResources.map((resource) => {
              const middlewaresList = MiddlewareUtils.parseMiddlewares(resource.middlewares);
              const isProtected = middlewaresList.length > 0;
              const isDisabled = resource.status === 'disabled';

              return (
                <tr
                  key={resource.id}
                  className={isDisabled ? 'bg-gray-100' : ''}
                >
                  <td className="px-6 py-4 whitespace-nowrap">
                    {resource.host}
                    {isDisabled && (
                      <span className="ml-2 px-2 py-1 text-xs rounded-full bg-red-100 text-red-800">
                        Removed from Pangolin
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${
                        isDisabled
                          ? 'bg-gray-100 text-gray-800'
                          : isProtected
                          ? 'bg-green-100 text-green-800'
                          : 'bg-yellow-100 text-yellow-800'
                      }`}
                    >
                      {isDisabled
                        ? 'Disabled'
                        : isProtected
                        ? 'Protected'
                        : 'Not Protected'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {middlewaresList.length > 0
                      ? middlewaresList.length
                      : 'None'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <button
                      onClick={() => navigateTo('resource-detail', resource.id)}
                      className="text-blue-600 hover:text-blue-900 mr-3"
                    >
                      {isDisabled ? 'View' : 'Manage'}
                    </button>
                    {isDisabled && (
                      <button
                        onClick={() => confirmDelete(resource)}
                        className="text-red-600 hover:text-red-900"
                      >
                        Delete
                      </button>
                    )}
                  </td>
                </tr>
              );
            })}
            {filteredResources.length === 0 && (
              <tr>
                <td
                  colSpan="4"
                  className="px-6 py-4 text-center text-gray-500"
                >
                  No resources found
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Delete Confirmation Modal */}
      <ConfirmationModal
        show={showDeleteModal && !!resourceToDelete}
        title="Confirm Deletion"
        message={`Are you sure you want to delete the resource "${resourceToDelete?.host}"?`}
        details="This action cannot be undone."
        confirmText="Delete"
        cancelText="Cancel"
        onConfirm={handleDeleteConfirmed}
        onCancel={cancelDelete}
      />
    </div>
  );
};

export default ResourcesList;
