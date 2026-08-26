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
    <div className="card h-100 insight-fact-card">
      <div className="card-body">
        <div className="insight-fact-label">{label}</div>
        <div className="insight-fact-value">{value}</div>
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
  const hasCashflow = cash && (cash.income !== '0' || cash.expense !== '0');

  return (
    <div>
      <section className="insights-intro" aria-labelledby="insights-heading">
        <div>
          <span className="insights-eyebrow">Cashflow intelligence</span>
          <h1 id="insights-heading">AI Insights</h1>
          <p>Read the month through verified cashflow facts. AI adds context, never changes the numbers.</p>
        </div>
        <span className={`insights-status ${aiUnavailable ? 'is-offline' : ''}`}>
          <span aria-hidden="true" />
          {aiUnavailable ? 'AI explanation offline' : 'Facts stay authoritative'}
        </span>
      </section>

      <>
          {isLoading ? (
            <div className="insights-skeleton mt-4" aria-label="Memuat insight">
              <span /><span /><span />
            </div>
          ) : null}
          {isError && !aiUnavailable ? (
            <div className="mt-4 alert alert-danger d-flex align-items-center justify-content-between gap-3" role="alert">
              <span>Insight tidak dapat dibuat. Data keuangan Anda tetap aman.</span>
              <Button variant="secondary" onClick={() => void refetch()}>Coba lagi</Button>
            </div>
          ) : null}
          {aiUnavailable ? (
            <div className="mt-4 alert alert-warning" role="status">
              AI sementara tidak tersedia. Ringkasan cashflow di bawah tetap berasal dari data keuangan Anda.
            </div>
          ) : null}

          {cash ? (
            <section className="mt-4" aria-labelledby="cashflow-facts-heading">
              <div className="insights-section-heading">
                <div>
                  <span>Actual</span>
                  <h2 id="cashflow-facts-heading">Cashflow bulan ini</h2>
                </div>
                <p>Transaksi POSTED saja</p>
              </div>
              <div className="row g-3">
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
            </section>
          ) : null}

          {insight && !dismissed ? (
            <section className="card mt-4 insights-narrative" aria-labelledby="insight-headline">
              <div className="card-body">
                <div className="d-flex align-items-start justify-content-between gap-3">
                  <div>
                    <span className="insights-eyebrow">AI interpretation</span>
                    <p className="insights-signal mb-0">{insight.signal}</p>
                  </div>
                  <Button variant="ghost" aria-label="Dismiss insight" onClick={() => setDismissed(true)}>Dismiss</Button>
                </div>
                <h2 id="insight-headline">{insight.headline}</h2>
                <p className="insight-detail">{insight.detail}</p>
              </div>
              {insight.related_facts.length > 0 ? (
                <div className="insight-supporting-facts">
                  <h3>Supporting facts</h3>
                  <ul>
                    {insight.related_facts.map((f, i) => (
                      <li key={i}>{f}</li>
                    ))}
                  </ul>
                </div>
              ) : null}
            </section>
          ) : null}

          {dismissed ? (
            <div className="mt-4 alert alert-light d-flex align-items-center justify-content-between gap-3" role="status">
              <span>Interpretasi AI disembunyikan. Cashflow aktual tetap terlihat di atas.</span>
              <Button variant="secondary" onClick={() => setDismissed(false)}>Tampilkan lagi</Button>
            </div>
          ) : null}

          {!isLoading && !isError && !insight && !aiUnavailable ? (
            <div className="mt-4">
              <EmptyState
                title={hasCashflow ? 'Insight belum tersedia' : 'Belum ada cashflow untuk dibaca'}
                description={hasCashflow ? 'Coba buat insight lagi untuk periode ini.' : 'Tambahkan income atau expense yang sudah diposting untuk mulai melihat pola cashflow.'}
                action={hasCashflow ? <Button variant="secondary" onClick={() => void refetch()}>Coba lagi</Button> : undefined}
              />
            </div>
          ) : null}
      </>
    </div>
  );
}
