import { useEffect, useState } from 'react';
import { ApiError, errorMessage } from '@/shared/api/client';

export const UI_ERROR_EVENT = 'savio:ui-error';

export function ErrorNotice() {
  const [error, setError] = useState<unknown>(null);

  useEffect(() => {
    const show = (event: Event) => {
      setError((event as CustomEvent<unknown>).detail);
    };
    window.addEventListener(UI_ERROR_EVENT, show);
    return () => window.removeEventListener(UI_ERROR_EVENT, show);
  }, []);

  if (!error) return null;
  const message = errorMessage(error);
  const requestId = error instanceof ApiError ? error.requestId : undefined;
  return (
    <div role="alert" className="fixed bottom-4 right-4 z-50 max-w-sm rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-800 shadow-lg">
      <p>{message}</p>
      {requestId && <p className="mt-2 text-xs">Reference: <code>{requestId}</code></p>}
      <button type="button" className="mt-2 font-medium underline" onClick={() => setError(null)}>Dismiss</button>
    </div>
  );
}
