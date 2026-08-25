import { useEffect, useState } from 'react';
import { ApiError, errorMessage } from '@/shared/api/client';

export const UI_ERROR_EVENT = 'savio:ui-error';

export function ErrorNotice() {
  const [error, setError] = useState<unknown>(null);

  useEffect(() => {
    const show = (event: Event) => {
      setError((event as CustomEvent<unknown>).detail);
    };
    const showRuntimeError = (event: ErrorEvent) => setError(event.error ?? new Error('Frontend runtime error'));
    const showRejectedPromise = (event: PromiseRejectionEvent) => setError(event.reason ?? new Error('Unhandled promise rejection'));
    window.addEventListener(UI_ERROR_EVENT, show);
    window.addEventListener('error', showRuntimeError);
    window.addEventListener('unhandledrejection', showRejectedPromise);
    return () => {
      window.removeEventListener(UI_ERROR_EVENT, show);
      window.removeEventListener('error', showRuntimeError);
      window.removeEventListener('unhandledrejection', showRejectedPromise);
    };
  }, []);

  if (!error) return null;
  const message = errorMessage(error);
  const requestId = error instanceof ApiError ? error.requestId : undefined;
  return (
    <div role="alert" className="alert alert-danger position-fixed bottom-0 end-0 m-4 shadow-lg" style={{ maxWidth: '24rem', zIndex: 2000 }}>
      <p className="mb-1">{message}</p>
      {requestId && <p className="fs-12 mb-2">Reference: <code>{requestId}</code></p>}
      <button type="button" className="btn btn-sm btn-outline-danger" onClick={() => setError(null)}>Dismiss</button>
    </div>
  );
}
