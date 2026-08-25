import { useState } from 'react';
import { useAuth } from '@/app/providers/auth-provider';
import { ApiError } from '@/shared/api/client';
import { Button } from '@/shared/components/ui/button';
import { EmptyState } from '@/shared/components/ui/empty-state';
import { formatAmountString } from '@/shared/utils/money';
import { useAIStatus, useInsight } from '@/features/ai/hooks/use-ai';
import { useCashflow } from '@/features/analytics/api/analytics.api';

function monthBounds(): { from: string; to: string } {
  const now = new Date();
  const pad = (n: number) => String(n).padStart(2, '0');
  const first = new Date(now.getFullYear(), now.getMonth(), 1);
  const last = new Date(now.getFullYear(), now.getMonth() + 1, 0);
  return {
    from: `${first.getFullYear()}-${pad(first.getMonth() + 1)}-01`,
    to: `${last.getFullYear()}-${pad(last.getMonth() + 1)}-${pad(last.getDate())}`,
  };
}

function FactCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="card h-100">
      <div className="card-body">
        <div className="fs-13 text-muted">{label}</div>
        <div className="mt-1 fs-16 fw-semibold text-dark">{value}</div>
      </div>
    </div>
  );
}

export function InsightsPage() {
  const { data: status } = useAIStatus();
  const { data: insight, isLoading, isError, error, refetch } = useInsight();
  const { auth } = useAuth();
  const currency = auth?.workspace.base_currency ?? 'IDR';
  const [dismissed, setDismissed] = useState(false);
  const { from, to } = monthBounds();
  const { data: cash } = useCashflow(from, to);

  const aiUnavailable =
    status?.enabled === false ||
    error instanceof ApiError && error.status === 503 ||
    (error instanceof ApiError && error.status === 422);

  return (
    <div>
      <h1 className="fs-20 fw-bolder mb-0">AI Insights</h1>
      <p className="fs-13 text-muted mb-0 mt-1">
        Explained from real, deterministic numbers. The facts come first; AI only interprets them.
      </p>

      {dismissed ? (
        <div className="mt-4">
          <EmptyState
            title="Insight dismissed"
            description="Your cashflow facts below never depend on AI availability."
            action={<Button variant="secondary" onClick={() => setDismissed(false)}>Show insights again</Button>}
          />
        </div>
      ) : (
        <>
          {isLoading ? <p className="mt-4 fs-13 text-muted">Summarizing your month…</p> : null}
          {isError && !aiUnavailable ? (
            <div className="mt-4 alert alert-danger">
              <p>We could not generate this insight.</p>
              <Button variant="secondary" className="mt-2" onClick={() => void refetch()}>Try again</Button>
            </div>
          ) : null}
          {aiUnavailable ? (
            <div className="mt-4 alert alert-warning">
              AI is unavailable right now. Your financial facts and features still work — the explanatory
              narrative is temporarily offline.
            </div>
          ) : null}

          {cash ? (
            <div className="row g-4 mt-4">
              <div className="col-12 col-md-4">
                <FactCard label="Income this month" value={formatAmountString(cash.income, currency)} />
              </div>
              <div className="col-12 col-md-4">
                <FactCard label="Expenses this month" value={formatAmountString(cash.expense, currency)} />
              </div>
              <div className="col-12 col-md-4">
                <FactCard label="Net cashflow" value={formatAmountString(cash.net, currency)} />
              </div>
            </div>
          ) : null}

          {insight ? (
            <div className="card card-body mt-4">
              <div className="d-flex align-items-start justify-content-between">
                <span className="badge bg-soft-primary text-primary">
                  AI-generated · verify against your facts
                </span>
                <Button variant="ghost" onClick={() => setDismissed(true)}>Dismiss</Button>
              </div>
              <h2 className="mt-4 fs-16 fw-semibold text-dark">{insight.headline}</h2>
              <p className="mt-2 fs-13 text-secondary">{insight.detail}</p>
              {insight.related_facts.length > 0 ? (
                <div className="mt-4 border-top pt-4">
                  <h3 className="fs-12 text-uppercase fw-semibold text-muted mb-3">Supporting facts</h3>
                  <ul className="mt-2 d-flex flex-column gap-1 fs-13 text-secondary">
                    {insight.related_facts.map((f, i) => (
                      <li key={i}>{f}</li>
                    ))}
                  </ul>
                </div>
              ) : null}
            </div>
          ) : null}

          {!isError && !insight ? (
            <div className="card card-body mt-4 p-5 text-center">
              <p className="mx-auto fs-13 text-muted">
                Insights need cashflow data. Add income and expenses, then revisit this page.
              </p>
            </div>
          ) : null}
        </>
      )}
    </div>
  );
}