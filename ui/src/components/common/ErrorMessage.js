import React, { useEffect } from 'react';
import { showErrorToast } from './Toast';

/**
 * Error message component for displaying errors
 * 
 * @param {Object} props
 * @param {string} props.message - Error message to display
 * @param {string} props.details - Optional error details
 * @param {function} props.onRetry - Optional retry function
 * @param {function} props.onDismiss - Optional dismiss function
 * @param {boolean} props.showToast - Whether to also show a toast notification
 * @returns {JSX.Element}
 */
const ErrorMessage = ({ 
  message, 
  details = null, 
  onRetry = null, 
  onDismiss = null,
  showToast = false
}) => {
  useEffect(() => {
    if (showToast) {
      showErrorToast(message, details);
    }
  }, [message, details, showToast]);

  return (
    <div className="bg-red-100 text-red-700 p-6 rounded-lg border border-red-300 mb-4">
      <div className="flex items-start">
        <div className="flex-shrink-0 pt-0.5">
          <svg 
            className="h-5 w-5 text-red-500"
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
              clipRule="evenodd"
            />
          </svg>
        </div>
        <div className="ml-3 flex-1">
          <h3 className="text-md font-medium text-red-800">{message}</h3>
          
          {details && (
            <div className="mt-2 text-sm text-red-700">
              <p>{details}</p>
            </div>
          )}
          
          {(onRetry || onDismiss) && (
            <div className="mt-4 flex">
              {onRetry && (
                <button
                  type="button"
                  onClick={onRetry}
                  className="mr-3 inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                >
                  Retry
                </button>
              )}
              
              {onDismiss && (
                <button
                  type="button"
                  onClick={onDismiss}
                  className="inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  Dismiss
                </button>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ErrorMessage;
