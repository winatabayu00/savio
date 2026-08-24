import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { status } = useAuth();
  const location = useLocation();

  if (status === 'UNKNOWN') {
    return (
      <div className="flex min-h-screen items-center justify-center text-gray-500">
        Loading…
      </div>
    );
  }
  if (status === 'UNAUTHENTICATED') {
    return <Navigate to="/login" state={{ from: location.pathname }} replace />;
  }
  return <>{children}</>;
}