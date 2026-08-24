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
    <div className="rounded-xl border border-gray-200 bg-white p-5">
      <div className="text-sm text-gray-500">{label}</div>
      <div className="mt-1 text-xl font-semibold text-gray-900">{value}</div>
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
      <h1 className="text-2xl font-semibold">AI Insights</h1>
      <p className="mt-1 text-sm text-gray-500">
        Explained from real, deterministic numbers. The facts come first; AI only interprets them.
      </p>

      {dismissed ? (
        <div className="mt-6">
          <EmptyState
            title="Insight dismissed"
            description="Your cashflow facts below never depend on AI availability."
            action={<Button variant="secondary" onClick={() => setDismissed(false)}>Show insights again</Button>}
          />
        </div>
      ) : (
        <>
          {isLoading ? <p className="mt-6 text-sm text-gray-500">Summarizing your month…</p> : null}
          {isError && !aiUnavailable ? (
            <div className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
              <p>We could not generate this insight.</p>
              <Button variant="secondary" className="mt-2" onClick={() => void refetch()}>Try again</Button>
            </div>
          ) : null}
          {aiUnavailable ? (
            <div className="mt-6 rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">
              AI is unavailable right now. Your financial facts and features still work — the explanatory
              narrative is temporarily offline.
            </div>
          ) : null}

          {cash ? (
            <div className="mt-6 grid gap-4 md:grid-cols-3">
              <FactCard label="Income this month" value={formatAmountString(cash.income, currency)} />
              <FactCard label="Expenses this month" value={formatAmountString(cash.expense, currency)} />
              <FactCard label="Net cashflow" value={formatAmountString(cash.net, currency)} />
            </div>
          ) : null}

          {insight ? (
            <div className="mt-6 rounded-2xl border border-gray-200 bg-white p-6">
              <div className="flex items-start justify-between">
                <span className="rounded-full bg-brand/10 px-2.5 py-1 text-xs font-semibold text-brand">
                  AI-generated · verify against your facts
                </span>
                <Button variant="ghost" onClick={() => setDismissed(true)}>Dismiss</Button>
              </div>
              <h2 className="mt-4 text-xl font-semibold text-gray-900">{insight.headline}</h2>
              <p className="mt-2 text-sm text-gray-600">{insight.detail}</p>
              {insight.related_facts.length > 0 ? (
                <div className="mt-5 border-t border-gray-100 pt-4">
                  <h3 className="text-sm font-semibold uppercase tracking-wide text-gray-500">Supporting facts</h3>
                  <ul className="mt-2 list-inside list-disc space-y-1 text-sm text-gray-600">
                    {insight.related_facts.map((f, i) => (
                      <li key={i}>{f}</li>
                    ))}
                  </ul>
                </div>
              ) : null}
            </div>
          ) : null}

          {!isError && !insight ? (
            <div className="mt-6 rounded-2xl border border-dashed border-gray-300 bg-white p-8 text-center">
              <p className="mx-auto max-w-md text-sm text-gray-500">
                Insights need cashflow data. Add income and expenses, then revisit this page.
              </p>
            </div>
          ) : null}
        </>
      )}
    </div>
  );
}