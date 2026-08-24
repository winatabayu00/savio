import { useAuth } from '@/app/providers/auth-provider';

export function DashboardPage() {
  const { auth } = useAuth();
  return (
    <div>
      <h1 className="text-2xl font-semibold">Dashboard</h1>
      <p className="mt-1 text-sm text-gray-500">
        Welcome, {auth?.user.name}. Your financial dashboard appears here once
        you add your first account.
      </p>
      <div className="mt-8 rounded-xl border border-dashed border-gray-300 bg-white p-10 text-center">
        <p className="text-sm text-gray-500">
          Add your first account to start understanding your cashflow.
        </p>
      </div>
    </div>
  );
}