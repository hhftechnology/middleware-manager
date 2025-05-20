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
 * @param {string} props.mode - Whether this is a confirmation or an alert
 * @returns {JSX.Element}
 */
const ConfirmationModal = ({ 
  title, 
  message, 
  details, 
  confirmText, 
  cancelText, 
  onConfirm, 
  onCancel, 
  show,
  mode = "confirm" // new prop, default to "confirm"
}) => {
  if (!show) return null;

  // Use different button texts and visibility based on mode
  const isAlert = mode === "alert";
  const confirmButtonText = confirmText || (isAlert ? "OK" : "Confirm");
  const cancelButtonText = cancelText || "Cancel";

  return (
    <div className="modal-overlay">
      <div className="modal-content max-w-md">
        <div className="modal-header">
          <h3 className={`modal-title ${isAlert ? 'text-blue-600 dark:text-blue-400' : 'text-red-600 dark:text-red-400'}`}>
            {title}
          </h3>
          {/* For alerts, closing the modal can just call onConfirm or onCancel or be hidden */}
          <button onClick={isAlert ? onConfirm : onCancel} className="modal-close-button">&times;</button>
        </div>
        <div className="modal-body">
          <p className="text-sm text-gray-700 dark:text-gray-300 mb-2">{message}</p>
          {details && <p className="text-xs text-gray-500 dark:text-gray-400 mb-4">{details}</p>}
        </div>
        <div className="modal-footer">
          {!isAlert && (
            <button onClick={onCancel} className="btn btn-secondary">
              {cancelButtonText}
            </button>
          )}
          <button onClick={onConfirm} className={isAlert ? "btn btn-primary" : "btn btn-danger"}>
            {confirmButtonText}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ConfirmationModal;