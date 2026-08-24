import { useState } from 'react';
import { useAuth } from '@/app/providers/auth-provider';
import { EmptyState } from '@/shared/components/ui/empty-state';
import { formatAmountString } from '@/shared/utils/money';
import { useCashflow, useCategoryBreakdown, comparePeriods } from '@/features/analytics/api/analytics.api';
import type { PeriodComparisonDTO } from '@/features/analytics/api/analytics.api';

function monthBounds(offset = 0): { from: string; to: string } {
  const now = new Date();
  const first = new Date(now.getFullYear(), now.getMonth() + offset, 1);
  const last = new Date(now.getFullYear(), now.getMonth() + offset + 1, 0);
  return { from: iso(first), to: iso(last) };
}
function iso(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

export function AnalyticsPage() {
  const { auth } = useAuth();
  const currency = auth?.workspace.base_currency ?? 'IDR';
  const { from, to } = monthBounds();
  const { from: prevFrom, to: prevTo } = monthBounds(-1);
  const { data: cash } = useCashflow(from, to);
  const { data: breakdown } = useCategoryBreakdown(from, to);
  const [comparison, setComparison] = useState<PeriodComparisonDTO | null>(null);
  const [comparing, setComparing] = useState(false);

  const loadComparison = async () => {
    setComparing(true);
    try {
      const res = await comparePeriods(from, to, prevFrom, prevTo);
      setComparison(res);
    } finally {
      setComparing(false);
    }
  };

  const rows = breakdown ?? [];
  const maxTotal = Math.max(...rows.map((r) => Number(r.total)), 1);

  return (
    <div>
      <h1 className="text-2xl font-semibold">Analytics</h1>
      <p className="mt-1 text-sm text-gray-500">
        Deterministic summaries of your posted income and expenses.
      </p>

      <div className="mt-6 grid gap-6 lg:grid-cols-3">
        <section className="rounded-xl border border-gray-200 bg-white p-5">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-500">Cashflow</h2>
          <div className="mt-4 space-y-3">
            <Metric label="Income" value={cash ? formatAmountString(cash.income, currency) : '—'} color="text-green-600" />
            <Metric label="Expenses" value={cash ? formatAmountString(cash.expense, currency) : '—'} color="text-red-600" />
            <Metric label="Net" value={cash ? formatAmountString(cash.net, currency) : '—'} />
          </div>
          <button
            type="button"
            onClick={() => void loadComparison()}
            className="mt-5 text-sm font-medium text-brand hover:underline disabled:opacity-60"
            disabled={comparing}
          >
            {comparing ? 'Comparing…' : 'Compare with last month'}
          </button>
          {comparison ? (
            <div className="mt-3 rounded-lg bg-gray-50 p-3 text-sm">
              <p className="text-gray-500">
                Last month income:{' '}
                <span className="font-medium text-gray-900">{formatAmountString(comparison.previous.income, currency)}</span>
                {comparison.income_delta_percent != null ? (
                  <ChangeBadge v={comparison.income_delta_percent} />
                ) : null}
              </p>
              <p className="mt-1 text-gray-500">
                Last month expenses:{' '}
                <span className="font-medium text-gray-900">{formatAmountString(comparison.previous.expense, currency)}</span>
                {comparison.expense_delta_percent != null ? (
                  <ChangeBadge v={comparison.expense_delta_percent} />
                ) : null}
              </p>
            </div>
          ) : null}
        </section>

        <section className="rounded-xl border border-gray-200 bg-white p-5 lg:col-span-2">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-500">
            Activity by category ({from} → {to})
          </h2>
          {rows.length === 0 ? (
            <div className="mt-6">
              <EmptyState title="No categorized activity yet" description="Add income or expenses to see your category breakdown." />
            </div>
          ) : (
            <ul className="mt-4 space-y-3">
              {rows.map((row) => {
                const total = Number(row.total);
                const pct = Math.max((total / maxTotal) * 100, 2);
                return (
                  <li key={row.category_id || row.category_name}>
                    <div className="flex items-center justify-between text-sm">
                      <span className="font-medium text-gray-900">{row.category_name}</span>
                      <span className="text-gray-600">{formatAmountString(row.total, currency)}</span>
                    </div>
                    <div className="mt-1 h-2 w-full overflow-hidden rounded-full bg-gray-100">
                      <div className="h-full rounded-full bg-brand" style={{ width: `${pct}%` }} />
                    </div>
                  </li>
                );
              })}
            </ul>
          )}
        </section>
      </div>
    </div>
  );
}

function Metric({ label, value, color = 'text-gray-900' }: { label: string; value: string; color?: string }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-sm text-gray-500">{label}</span>
      <span className={`text-sm font-semibold ${color}`}>{value}</span>
    </div>
  );
}

function ChangeBadge({ v }: { v: number }) {
  const dir = v > 0 ? 'up' : v < 0 ? 'down' : 'flat';
  const cls = dir === 'up' ? 'text-green-600' : dir === 'down' ? 'text-red-600' : 'text-gray-500';
  return (
    <span className={`ml-1 text-xs font-medium ${cls}`}>
      ({v > 0 ? '+' : ''}
      {v.toFixed(1)}%)
    </span>
  );
}