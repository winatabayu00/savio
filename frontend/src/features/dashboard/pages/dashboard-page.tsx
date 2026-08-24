import { Link } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';
import { Button } from '@/shared/components/ui/button';
import { EmptyState } from '@/shared/components/ui/empty-state';
import { formatAmountString, formatMoney } from '@/shared/utils/money';
import { useDashboard, useCashflow } from '@/features/analytics/api/analytics.api';
import { useRecurring } from '@/features/recurring/hooks/use-recurring';

function monthBounds(offset = 0): { from: string; to: string } {
  const now = new Date();
  const first = new Date(now.getFullYear(), now.getMonth() + offset, 1);
  const last = new Date(now.getFullYear(), now.getMonth() + offset + 1, 0);
  return { from: isoDate(first), to: isoDate(last) };
}

function isoDate(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

export function DashboardPage() {
  const { auth } = useAuth();
  const currency = auth?.workspace.base_currency ?? 'IDR';
  const { data: dash, isLoading, isError, refetch } = useDashboard();
  const { from, to } = monthBounds();
  const { data: cash } = useCashflow(from, to);
  const { data: recurringRules } = useRecurring();

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="h-8 w-48 animate-pulse rounded bg-gray-200" />
        <div className="grid gap-4 md:grid-cols-3">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-28 animate-pulse rounded-xl bg-gray-100" />
          ))}
        </div>
        <div className="h-64 animate-pulse rounded-xl bg-gray-100" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="rounded-xl border border-red-200 bg-red-50 p-6 text-sm text-red-700">
        <p>We could not load your dashboard.</p>
        <Button variant="secondary" className="mt-3" onClick={() => void refetch()}>
          Try again
        </Button>
      </div>
    );
  }

  const accounts = dash?.accounts ?? [];
  const noData = accounts.length === 0;

  return (
    <div>
      <h1 className="text-2xl font-semibold">Dashboard</h1>
      <p className="mt-1 text-sm text-gray-500">
        Welcome back, {auth?.user.name}. Here is your current financial position.
      </p>

      {noData ? (
        <div className="mt-8">
          <EmptyState
            title="Set up your first account"
            description="Add income and expenses to see your cashflow, forecast and AI insights."
            action={
              <Link to="/accounts">
                <Button>Add an account</Button>
              </Link>
            }
          />
        </div>
      ) : (
        <>
          <div className="mt-6 grid gap-4 md:grid-cols-3">
            <SummaryCard
              label="Total balance"
              value={formatMoney(dash?.total_balance ?? 0, currency)}
              sub={`across ${accounts.length} account${accounts.length > 1 ? 's' : ''}`}
            />
            <SummaryCard
              label="Income this month"
              value={formatAmountString(cash?.income ?? '0', currency)}
              sub="posted income"
              tone="good"
            />
            <SummaryCard
              label="Expenses this month"
              value={formatAmountString(cash?.expense ?? '0', currency)}
              sub="posted expenses"
              tone="bad"
            />
          </div>

          <div className="mt-6 grid gap-6 lg:grid-cols-2">
            <section className="rounded-xl border border-gray-200 bg-white p-5">
              <div className="flex items-center justify-between">
                <h2 className="font-semibold text-gray-900">Upcoming scheduled activity</h2>
                <Link to="/recurring" className="text-sm font-medium text-brand hover:underline">
                  Manage recurring
                </Link>
              </div>
              {dash && dash.upcoming.length === 0 ? (
                <p className="mt-3 text-sm text-gray-500">No upcoming planned activity.</p>
              ) : (
                <ul className="mt-3 divide-y divide-gray-100">
                  {(dash?.upcoming ?? []).map((u) => (
                    <li key={u.id} className="flex items-center justify-between py-2.5 text-sm">
                      <div>
                        <div className="font-medium text-gray-900">
                          {u.description || `${u.type.toLowerCase()} occurrence`}
                        </div>
                        <div className="text-xs text-gray-500">
                          {u.due_date} · {u.account_name}
                        </div>
                      </div>
                      <span className={u.type === 'EXPENSE' ? 'font-semibold text-red-600' : 'font-semibold text-green-600'}>
                        {u.type === 'EXPENSE' ? '−' : '+'}
                        {formatAmountString(u.amount, currency)}
                      </span>
                    </li>
                  ))}
                </ul>
              )}
            </section>

            <section className="rounded-xl border border-gray-200 bg-white p-5">
              <h2 className="font-semibold text-gray-900">Recent activity</h2>
              {dash && dash.recent.length === 0 ? (
                <p className="mt-3 text-sm text-gray-500">No transactions yet.</p>
              ) : (
                <ul className="mt-3 divide-y divide-gray-100">
                  {(dash?.recent ?? []).map((r) => (
                    <li key={r.id} className="flex items-center justify-between py-2.5 text-sm">
                      <div className="min-w-0">
                        <div className="truncate font-medium text-gray-900">
                          {r.description || r.category_name || 'Transaction'}
                        </div>
                        <div className="text-xs text-gray-500">
                          {r.transaction_date} · {r.account_name}
                        </div>
                      </div>
                      <span className={r.type === 'EXPENSE' ? 'font-semibold text-red-600' : 'font-semibold text-green-600'}>
                        {r.type === 'EXPENSE' ? '−' : '+'}
                        {formatAmountString(r.amount, currency)}
                      </span>
                    </li>
                  ))}
                </ul>
              )}
              <div className="mt-4">
                <Link to="/transactions" className="text-sm font-medium text-brand hover:underline">
                  View all transactions →
                </Link>
              </div>
            </section>
          </div>

          {recurringRules && recurringRules.length > 0 ? (
            <section className="mt-6 rounded-xl border border-gray-200 bg-white p-5">
              <h2 className="font-semibold text-gray-900">Planned cashflow</h2>
              <p className="mt-1 text-sm text-gray-500">
                {recurringRules.length} recurring rule
                {recurringRules.length > 1 ? 's' : ''} active. Confirm scheduled occurrences from the
                Recurring page when they happen.
              </p>
            </section>
          ) : null}
        </>
      )}
    </div>
  );
}

function SummaryCard({ label, value, sub, tone }: { label: string; value: string; sub: string; tone?: 'good' | 'bad' }) {
  const color = tone === 'good' ? 'text-green-600' : tone === 'bad' ? 'text-red-600' : 'text-gray-900';
  return (
    <div className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
      <div className="text-sm text-gray-500">{label}</div>
      <div className={`mt-1 text-2xl font-semibold ${color}`}>{value}</div>
      <div className="mt-1 text-xs text-gray-400">{sub}</div>
    </div>
  );
}