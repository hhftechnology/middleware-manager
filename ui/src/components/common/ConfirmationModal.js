import React from 'react';

/**
 * Reusable confirmation modal component
 * 
 * @param {Object} props
 * @param {string} props.title - Modal title
 * @param {string} props.message - Main confirmation message
 * @param {string} props.details - Optional additional details
 * @param {string} props.confirmText - Text for confirm button
 * @param {string} props.cancelText - Text for cancel button
 * @param {Function} props.onConfirm - Function to call when confirmed
 * @param {Function} props.onCancel - Function to call when cancelled
 * @param {boolean} props.show - Whether to show the modal
 * @returns {JSX.Element}
 */
const ConfirmationModal = ({ 
  title, 
  message, 
  details, 
  confirmText = "Confirm", 
  cancelText = "Cancel", 
  onConfirm, 
  onCancel, 
  show 
}) => {
  if (!show) return null;
  
  return (
    <div className="modal-overlay">
      <div className="modal-content max-w-md"> {/* Standard width */}
        <div className="modal-header">
          <h3 className="modal-title text-red-600 dark:text-red-400">{title}</h3>
          <button onClick={onCancel} className="modal-close-button">&times;</button>
        </div>
        <div className="modal-body">
          <p className="text-sm text-gray-700 dark:text-gray-300 mb-2">{message}</p>
          {details && <p className="text-xs text-gray-500 dark:text-gray-400 mb-4">{details}</p>}
        </div>
        <div className="modal-footer">
          <button onClick={onCancel} className="btn btn-secondary">{cancelText}</button>
          <button onClick={onConfirm} className="btn btn-danger">{confirmText}</button>
        </div>
      </div>
    </div>
  );
};

export default ConfirmationModal;