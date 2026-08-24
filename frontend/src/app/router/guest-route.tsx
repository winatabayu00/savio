import { Navigate } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';

export function GuestRoute({ children }: { children: React.ReactNode }) {
  const { status } = useAuth();
  if (status === 'UNKNOWN') {
    return (
      <div className="flex min-h-screen items-center justify-center text-gray-500">
        Loading…
      </div>
    );
  }
  if (status === 'AUTHENTICATED') {
    return <Navigate to="/dashboard" replace />;
  }
  return <>{children}</>;
}