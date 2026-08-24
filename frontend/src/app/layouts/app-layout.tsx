import { Link, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';

const NAV = [
  { to: '/dashboard', label: 'Dashboard' },
];

export function AppLayout() {
  const { auth, logout } = useAuth();
  const navigate = useNavigate();

  const onLogout = async () => {
    await logout();
    navigate('/login', { replace: true });
  };

  return (
    <div className="flex min-h-screen">
      <aside className="hidden w-60 shrink-0 flex-col border-r border-gray-200 bg-white md:flex">
        <Link to="/dashboard" className="px-6 py-5 text-xl font-semibold text-brand">
          Savio
        </Link>
        <nav className="flex-1 space-y-1 px-3">
          {NAV.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              className="block rounded-lg px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
            >
              {item.label}
            </Link>
          ))}
        </nav>
        <div className="border-t border-gray-200 p-4">
          <div className="text-sm font-medium">{auth?.user.name}</div>
          <div className="text-xs text-gray-500">{auth?.user.email}</div>
          <button
            type="button"
            onClick={onLogout}
            className="mt-3 text-sm font-medium text-gray-600 hover:text-brand"
          >
            Sign out
          </button>
        </div>
      </aside>
      <main className="min-w-0 flex-1 overflow-x-hidden">
        <div className="mx-auto max-w-6xl px-4 py-8 md:px-8">
          <Outlet />
        </div>
      </main>
    </div>
  );
}