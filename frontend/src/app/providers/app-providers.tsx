import type { ReactNode } from 'react';
import { QueryProvider } from './query-provider';
import { AuthProvider } from './auth-provider';
import { ErrorNotice } from '@/shared/components/ui/error-notice';

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <QueryProvider>
      <AuthProvider>
        {children}
        <ErrorNotice />
      </AuthProvider>
    </QueryProvider>
  );
}
