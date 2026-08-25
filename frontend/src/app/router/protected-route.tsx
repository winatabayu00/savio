import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { status } = useAuth();
  const location = useLocation();

  if (status === 'UNKNOWN') {
    return (
      <div className="d-flex align-items-center justify-content-center min-vh-100 text-muted">
        <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
        Loading…
      </div>
    );
  }
  if (status === 'UNAUTHENTICATED') {
    return <Navigate to="/login" state={{ from: location.pathname }} replace />;
  }
  return <>{children}</>;
}