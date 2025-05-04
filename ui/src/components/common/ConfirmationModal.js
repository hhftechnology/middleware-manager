import React from 'react';

/**
 * Ant Design styled confirmation modal component
 */
const ConfirmationModal = ({
  isOpen,
  title = 'Confirm',
  message,
  details,
  confirmText = 'OK',
  cancelText = 'Cancel',
  confirmButtonType = 'primary',
  onConfirm,
  onCancel,
  isProcessing = false
}) => {
  if (!isOpen) return null;
    
  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center z-50 ant-modal-root">
      <div className="ant-modal-wrap flex items-center justify-center">
        <div className="ant-modal" style={{ width: '416px', maxWidth: '95vw', margin: '0 auto', top: '0' }}>
          <div className="ant-modal-content">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-semibold">{title}</h3>
              <button
                onClick={onCancel}
                className="text-gray-500 hover:text-gray-700"
                disabled={isProcessing}
              >
                ×
              </button>
            </div>
            
            <div className="mb-4">
              <p className="text-gray-700">{message}</p>
              {details && (
                <p className="text-sm text-gray-500 mt-2">{details}</p>
              )}
            </div>
            
            <div className="flex justify-end space-x-3">
              <button
                type="button"
                className="px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
                onClick={onCancel}
                disabled={isProcessing}
              >
                <span>{cancelText}</span>
              </button>
              <button
                type="button"
                className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                onClick={onConfirm}
                disabled={isProcessing}
              >
                {isProcessing ? (
                  <span className="flex items-center">
                    <span role="img" aria-label="loading" className="anticon anticon-loading anticon-spin" style={{ marginRight: '8px' }}>
                      <svg viewBox="0 0 1024 1024" focusable="false" data-icon="loading" width="1em" height="1em" fill="currentColor" aria-hidden="true">
                        <path d="M988 548c-19.9 0-36-16.1-36-36 0-59.4-11.6-117-34.6-171.3a440.45 440.45 0 00-94.3-139.9 437.71 437.71 0 00-139.9-94.3C629 83.6 571.4 72 512 72c-19.9 0-36-16.1-36-36s16.1-36 36-36c69.1 0 136.2 13.5 199.3 40.3C772.3 66 827 103 874 150c47 47 83.9 101.8 109.7 162.7 26.7 63.1 40.2 130.2 40.2 199.3.1 19.9-16 36-35.9 36z"></path>
                      </svg>
                    </span>
                    <span>{confirmText}</span>
                  </span>
                ) : (
                  <span>{confirmText}</span>
                )}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
  
};

export default ConfirmationModal;
