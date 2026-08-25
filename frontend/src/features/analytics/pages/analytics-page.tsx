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
      <h1 className="fs-20 fw-bolder mb-0">Analytics</h1>
      <p className="fs-13 text-muted mb-0 mt-1">
        Deterministic summaries of your posted income and expenses.
      </p>

      <div className="row g-4 mt-4">
        <section className="col-12 col-lg-4">
          <div className="card h-100">
            <div className="card-body">
              <h2 className="fs-12 text-uppercase fw-semibold text-muted mb-3">Cashflow</h2>
              <div className="mt-4 d-flex flex-column gap-3">
            <Metric label="Income" value={cash ? formatAmountString(cash.income, currency) : '—'} color="text-success" />
            <Metric label="Expenses" value={cash ? formatAmountString(cash.expense, currency) : '—'} color="text-danger" />
            <Metric label="Net" value={cash ? formatAmountString(cash.net, currency) : '—'} />
          </div>
          <button
            type="button"
            onClick={() => void loadComparison()}
            className="mt-4 text-primary fw-medium"
            disabled={comparing}
          >
            {comparing ? 'Comparing…' : 'Compare with last month'}
          </button>
          {comparison ? (
            <div className="mt-3 bg-soft-secondary p-3 rounded fs-13">
              <p className="text-muted">
                Last month income:{' '}
                <span className="fw-medium text-dark">{formatAmountString(comparison.previous.income, currency)}</span>
                {comparison.income_delta_percent != null ? (
                  <ChangeBadge v={comparison.income_delta_percent} />
                ) : null}
              </p>
              <p className="mt-1 text-muted">
                Last month expenses:{' '}
                <span className="fw-medium text-dark">{formatAmountString(comparison.previous.expense, currency)}</span>
                {comparison.expense_delta_percent != null ? (
                  <ChangeBadge v={comparison.expense_delta_percent} />
                ) : null}
              </p>
            </div>
          ) : null}
            </div>
          </div>
        </section>

        <section className="col-12 col-lg-8">
          <div className="card h-100">
            <div className="card-body">
              <h2 className="fs-12 text-uppercase fw-semibold text-muted mb-3">
            Activity by category ({from} → {to})
          </h2>
          {rows.length === 0 ? (
            <div className="mt-4">
              <EmptyState title="No categorized activity yet" description="Add income or expenses to see your category breakdown." />
            </div>
          ) : (
            <ul className="mt-4 d-flex flex-column gap-3">
              {rows.map((row) => {
                const total = Number(row.total);
                const pct = Math.max((total / maxTotal) * 100, 2);
                return (
                  <li key={row.category_id || row.category_name}>
                    <div className="d-flex align-items-center justify-content-between fs-13">
                      <span className="fw-medium text-dark">{row.category_name}</span>
                      <span className="text-secondary">{formatAmountString(row.total, currency)}</span>
                    </div>
                    <div className="progress mt-2" style={{ height: 8 }}>
                      <div className="progress-bar bg-primary" style={{ width: `${pct}%` }} />
                    </div>
                  </li>
                );
              })}
            </ul>
          )}
            </div>
          </div>
        </section>
      </div>
    </div>
  );
}

function Metric({ label, value, color = 'text-dark' }: { label: string; value: string; color?: string }) {
  return (
    <div className="d-flex align-items-center justify-content-between">
      <span className="fs-13 text-muted">{label}</span>
      <span className={`fs-13 fw-semibold ${color}`}>{value}</span>
    </div>
  );
}

function ChangeBadge({ v }: { v: number }) {
  const dir = v > 0 ? 'up' : v < 0 ? 'down' : 'flat';
  const cls = dir === 'up' ? 'text-success' : dir === 'down' ? 'text-danger' : 'text-muted';
  return (
    <span className={`ms-1 fs-12 fw-medium ${cls}`}>
      ({v > 0 ? '+' : ''}
      {v.toFixed(1)}%)
    </span>
  );
}