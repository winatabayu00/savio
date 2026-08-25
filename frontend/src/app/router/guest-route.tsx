import { Navigate } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';

export function GuestRoute({ children }: { children: React.ReactNode }) {
  const { status } = useAuth();
  if (status === 'UNKNOWN') {
    return (
      <div className="d-flex align-items-center justify-content-center min-vh-100 text-muted">
        <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
        Loading…
      </div>
    );
  }
  if (status === 'AUTHENTICATED') {
    return <Navigate to="/dashboard" replace />;
  }
  return <>{children}</>;
}