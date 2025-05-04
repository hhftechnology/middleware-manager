import { toast } from 'react-toastify';

// Toast for errors
export const showErrorToast = (message, details = null) => {
  const content = (
    <div>
      <strong>{message}</strong>
      {details && <div className="mt-1 text-sm">{details}</div>}
    </div>
  );
  
  return toast.error(content, {
    position: "top-center",
    autoClose: 5000,
    hideProgressBar: false,
    closeOnClick: true,
    pauseOnHover: true,
    draggable: true,
    style: { width: "100%" }
  });
};

// Toast for success
export const showSuccessToast = (message) => {
  return toast.success(message, {
    position: "top-center",
    autoClose: 3000,
    hideProgressBar: false,
    closeOnClick: true,
    style: { width: "100%" }
  });
};

// Toast for warnings
export const showWarningToast = (message) => {
  return toast.warning(message, {
    position: "top-right",
    autoClose: 4000,
    hideProgressBar: false,
    closeOnClick: true
  });
};

// Toast for info
export const showInfoToast = (message) => {
  return toast.info(message, {
    position: "top-right",
    autoClose: 3000,
    hideProgressBar: false,
    closeOnClick: true
  });
};
