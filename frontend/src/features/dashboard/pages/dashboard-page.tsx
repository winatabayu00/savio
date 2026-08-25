import { Link } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';
import { Button } from '@/shared/components/ui/button';
import { EmptyState } from '@/shared/components/ui/empty-state';
import { PageHeader } from '@/app/layouts/page-header';
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
      <div className="row g-4">
        {[0, 1, 2].map((i) => (
          <div key={i} className="col-12 col-md-4">
            <div className="card">
              <div className="card-body placeholder-glow">
                <span className="placeholder col-8"></span>
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (isError) {
    return (
      <div className="alert alert-danger">
        <p className="mb-2">We could not load your dashboard.</p>
        <Button variant="secondary" onClick={() => void refetch()}>
          Try again
        </Button>
      </div>
    );
  }

  const accounts = dash?.accounts ?? [];
  const noData = accounts.length === 0;

  return (
    <>
      <PageHeader>
        <p className="fs-13 text-muted mb-0">
          Welcome back, {auth?.user.name}. Here is your current financial position.
        </p>
      </PageHeader>
      {noData ? (
        <div className="row">
          <div className="col-12">
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
        </div>
      ) : (
        <>
          <div className="row g-4 mb-4">
            <div className="col-12 col-md-4">
              <SummaryCard
                label="Total balance"
                value={formatMoney(dash?.total_balance ?? 0, currency)}
                sub={`across ${accounts.length} account${accounts.length > 1 ? 's' : ''}`}
              />
            </div>
            <div className="col-12 col-md-4">
              <SummaryCard
                label="Income this month"
                value={formatAmountString(cash?.income ?? '0', currency)}
                sub="posted income"
                tone="good"
              />
            </div>
            <div className="col-12 col-md-4">
              <SummaryCard
                label="Expenses this month"
                value={formatAmountString(cash?.expense ?? '0', currency)}
                sub="posted expenses"
                tone="bad"
              />
            </div>
          </div>

          <div className="row g-4">
            <div className="col-12 col-xl-6">
              <div className="card">
                <div className="card-body">
                  <div className="d-flex align-items-center justify-content-between mb-3">
                    <h5 className="mb-0 fs-15 fw-semibold">Upcoming scheduled activity</h5>
                    <Link to="/recurring" className="fs-13 text-primary fw-medium">
                      Manage recurring
                    </Link>
                  </div>
                  {dash && dash.upcoming.length === 0 ? (
                    <p className="fs-13 text-muted mb-0">No upcoming planned activity.</p>
                  ) : (
                    <ul className="list-group list-group-flush">
                      {(dash?.upcoming ?? []).map((u) => (
                        <li key={u.id} className="list-group-item d-flex align-items-center justify-content-between px-0">
                          <div>
                            <div className="fs-14 fw-medium text-dark">
                              {u.description || `${u.type.toLowerCase()} occurrence`}
                            </div>
                            <div className="fs-12 text-muted">
                              {u.due_date} · {u.account_name}
                            </div>
                          </div>
                          <span className={u.type === 'EXPENSE' ? 'fs-14 fw-semibold text-danger' : 'fs-14 fw-semibold text-success'}>
                            {u.type === 'EXPENSE' ? '−' : '+'}
                            {formatAmountString(u.amount, currency)}
                          </span>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              </div>
            </div>

            <div className="col-12 col-xl-6">
              <div className="card">
                <div className="card-body">
                  <h5 className="mb-3 fs-15 fw-semibold">Recent activity</h5>
                  {dash && dash.recent.length === 0 ? (
                    <p className="fs-13 text-muted mb-0">No transactions yet.</p>
                  ) : (
                    <ul className="list-group list-group-flush">
                      {(dash?.recent ?? []).map((r) => (
                        <li key={r.id} className="list-group-item d-flex align-items-center justify-content-between px-0">
                          <div className="min-w-0">
                            <div className="fs-14 fw-medium text-dark text-truncate">
                              {r.description || r.category_name || 'Transaction'}
                            </div>
                            <div className="fs-12 text-muted">
                              {r.transaction_date} · {r.account_name}
                            </div>
                          </div>
                          <span className={r.type === 'EXPENSE' ? 'fs-14 fw-semibold text-danger' : 'fs-14 fw-semibold text-success'}>
                            {r.type === 'EXPENSE' ? '−' : '+'}
                            {formatAmountString(r.amount, currency)}
                          </span>
                        </li>
                      ))}
                    </ul>
                  )}
                  <div className="mt-3">
                    <Link to="/transactions" className="fs-13 text-primary fw-medium">
                      View all transactions →
                    </Link>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {recurringRules && recurringRules.length > 0 ? (
            <div className="card mt-4">
              <div className="card-body">
                <h5 className="fs-15 fw-semibold mb-1">Planned cashflow</h5>
                <p className="fs-13 text-muted mb-0">
                  {recurringRules.length} recurring rule
                  {recurringRules.length > 1 ? 's' : ''} active. Confirm scheduled occurrences from the
                  Recurring page when they happen.
                </p>
              </div>
            </div>
          ) : null}
        </>
      )}
    </>
  );
}

function SummaryCard({ label, value, sub, tone }: { label: string; value: string; sub: string; tone?: 'good' | 'bad' }) {
  const color = tone === 'good' ? 'text-success' : tone === 'bad' ? 'text-danger' : 'text-dark';
  return (
    <div className="card h-100">
      <div className="card-body">
        <div className="fs-13 text-muted">{label}</div>
        <div className={`fs-24 fw-semibold mt-1 ${color}`}>{value}</div>
        <div className="fs-12 text-muted mt-1">{sub}</div>
      </div>
    </div>
  );
}