import { Link, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';

const NAV = [
  {
    label: 'Ringkasan',
    items: [{ to: '/dashboard', label: 'Dashboard' }],
  },
  {
    label: 'Keuangan',
    items: [
      { to: '/accounts', label: 'Accounts' },
      { to: '/transactions', label: 'Transactions' },
      { to: '/transfers', label: 'Transfers' },
      { to: '/recurring', label: 'Recurring' },
      { to: '/categories', label: 'Categories' },
    ],
  },
  {
    label: 'Analisis & Perencanaan',
    items: [
      { to: '/analytics', label: 'Analytics' },
      { to: '/budgets', label: 'Budgets' },
      { to: '/goals', label: 'Goals' },
    ],
  },
  {
    label: 'Proyeksi',
    items: [
      { to: '/forecast', label: 'Forecast' },
      { to: '/scenarios', label: 'Scenarios' },
    ],
  },
  {
    label: 'AI',
    items: [
      { to: '/insights', label: 'Insights' },
      { to: '/copilot', label: 'Copilot' },
    ],
  },
  {
    label: 'Pengaturan',
    items: [{ to: '/settings', label: 'Pengaturan' }],
  },
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
        <nav className="flex-1 space-y-5 overflow-y-auto px-3">
          {NAV.map((group) => (
            <div key={group.label}>
              <div className="px-3 pb-1 text-xs font-semibold uppercase tracking-wide text-gray-400">
                {group.label}
              </div>
              <div className="space-y-1">
                {group.items.map((item) => (
                  <Link
                    key={item.to}
                    to={item.to}
                    className="block rounded-lg px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
                  >
                    {item.label}
                  </Link>
                ))}
              </div>
            </div>
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