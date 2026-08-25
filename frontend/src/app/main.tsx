import React from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';
import '@/styles/savio.scss';
import { AppProviders } from '@/app/providers/app-providers';
import { router } from '@/app/router';
import { ApplicationErrorBoundary } from '@/shared/components/ui/application-error-boundary';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <AppProviders>
      <ApplicationErrorBoundary>
        <RouterProvider router={router} />
      </ApplicationErrorBoundary>
    </AppProviders>
  </React.StrictMode>,
);
