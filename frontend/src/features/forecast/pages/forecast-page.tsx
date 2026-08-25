import { useState } from 'react';
import { useAuth } from '@/app/providers/auth-provider';
import { formatAmountString } from '@/shared/utils/money';
import { useForecast } from '@/features/forecast/hooks/use-forecast';
import { FORECAST_HORIZONS } from '@/features/forecast/api/forecast.api';

export function ForecastPage() {
  const { auth } = useAuth();
  const currency = auth?.workspace.base_currency ?? 'IDR';
  const [horizon, setHorizon] = useState(90);
  const { data, isLoading, isError, refetch } = useForecast(horizon);

  const confidenceLabel = data?.confidence === 'HIGH' ? 'High' : data?.confidence === 'MEDIUM' ? 'Medium' : 'Low';

  return (
    <div>
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-3">
        <div>
          <h1 className="fs-20 fw-bolder mb-0">Cashflow Forecast</h1>
          <p className="fs-13 text-muted mb-0 mt-1">
            A deterministic projection from your balances, recurring activity and spending history — not an AI prediction.
          </p>
        </div>
        <div className="d-inline-flex gap-1 border rounded-3 bg-white p-1">
          {FORECAST_HORIZONS.map((h) => (
            <button
              key={h}
              type="button"
              onClick={() => setHorizon(h)}
              className={`rounded-2 px-2 py-1 fs-13 ${
                horizon === h ? 'bg-primary text-white' : 'text-secondary'
              }`}
            >
              {h}d
            </button>
          ))}
        </div>
      </div>

      {data?.stale ? (
        <div className="mt-4 alert alert-warning p-3 fs-13 mb-0">
          Your finance data has changed since this projection was last calculated. Results may be stale.
        </div>
      ) : null}

      {isLoading ? (
        <div className="mt-4 d-flex flex-column gap-3">
          <div className="placeholder w-100 rounded-3" style={{ height: 112 }} />
          <div className="placeholder w-100 rounded-3" style={{ height: 256 }} />
        </div>
      ) : null}
      {isError ? (
        <div className="mt-4 alert alert-danger p-3 fs-13 mb-0">
          <p className="mb-2">We could not calculate your forecast.</p>
          <button className="btn btn-link text-primary p-0" onClick={() => void refetch()}>
            Try again
          </button>
        </div>
      ) : null}

      {data ? (
        <>
          <div className="row g-4 mt-1">
            <div className="col-12 col-md-4 col-lg-3"><Card label="Opening balance" value={formatAmountString(data.opening_balance, currency)} /></div>
            <div className="col-12 col-md-4 col-lg-3"><Card label={`Ending balance (${horizon}d)`} value={formatAmountString(data.ending_balance, currency)} tone={Number(data.ending_balance) < 0 ? 'bad' : 'good'} /></div>
            <div className="col-12 col-md-4 col-lg-3">
              <Card
                label={`Minimum balance · ${data.minimum_balance_date}`}
                value={formatAmountString(data.minimum_balance, currency)}
                tone={Number(data.minimum_balance) < 0 ? 'bad' : undefined}
              />
            </div>
            <div className="col-12 col-md-4 col-lg-3"><Card label="Projected income / expense" value={`${formatAmountString(data.projected_income, currency)} / ${formatAmountString(data.projected_expense, currency)}`} /></div>
          </div>

          <div className="row g-4 mt-1">
            <section className="col-12 col-xl-6">
              <div className="card h-100"><div className="card-body">
              <div className="d-flex align-items-center justify-content-between">
                <h2 className="fs-16 fw-semibold text-dark">Balance projection</h2>
                <span className="fs-12 fw-medium text-muted">confidence: {confidenceLabel}</span>
              </div>
              <BalanceChart points={data.timeline} />
              </div></div>
            </section>

            <section className="col-12 col-xl-6">
              <div className="card h-100"><div className="card-body">
              <h2 className="fs-16 fw-semibold text-dark">Events ({data.events.length})</h2>
              {data.events.length === 0 ? (
                <p className="mt-3 fs-13 text-muted mb-0">
                  No upcoming activity scheduled. Add recurring income or expenses to see them projected here.
                </p>
              ) : (
                <ul className="list-group list-group-flush mt-3" style={{ maxHeight: 320, overflowY: 'auto' }}>
                  {data.events.slice(0, 80).map((e, i) => (
                    <li key={i} className="list-group-item d-flex align-items-center justify-content-between fs-13">
                      <div className="text-truncate me-2">
                        <div className="text-truncate fw-medium text-dark">{e.description}</div>
                        <div className="fs-12 text-muted">
                          {e.date} · {e.type}
                        </div>
                      </div>
                      <span className={e.kind === 'EXPENSE' ? 'fw-semibold text-danger' : 'fw-semibold text-success'}>
                        {e.kind === 'EXPENSE' ? '−' : '+'}
                        {formatAmountString(e.amount, currency)}
                      </span>
                    </li>
                  ))}
                </ul>
              )}
              </div></div>
            </section>
          </div>

          <section className="mt-4">
            <div className="card"><div className="card-body">
            <h2 className="fs-16 fw-semibold text-dark">Assumptions</h2>
            <ul className="row g-3 mt-1 text-secondary list-unstyled mb-0">
              <li className="col-md-6">Estimated daily variable spending: <span className="fw-medium text-dark">{formatAmountString(data.assumptions.variable_expense_daily, currency)}</span></li>
              <li className="col-md-6">History used for baseline: <span className="fw-medium text-dark">{data.assumptions.baseline_days} days</span></li>
              <li className="col-md-6">Active recurring rules: <span className="fw-medium text-dark">{data.assumptions.active_recurring_rules}</span></li>
              <li className="col-md-6">Confidence basis: <span className="fw-medium text-dark">{data.assumptions.confidence_basis}</span></li>
            </ul>
            <p className="mt-3 fs-12 text-muted mb-0">Algorithm version {data.calculation_version}. Projections are estimates, not guarantees of future balances.</p>
            </div></div>
          </section>
        </>
      ) : null}
    </div>
  );
}

function Card({ label, value, tone }: { label: string; value: string; tone?: 'good' | 'bad' }) {
  const color = tone === 'good' ? 'text-success' : tone === 'bad' ? 'text-danger' : 'text-dark';
  return (
    <div className="card shadow-sm h-100">
      <div className="card-body">
        <div className="fs-13 text-muted">{label}</div>
        <div className={`mt-1 fs-18 fw-semibold ${color}`}>{value}</div>
      </div>
    </div>
  );
}

function BalanceChart({ points }: { points: { date: string; balance: string }[] }) {
  if (points.length < 2) {
    return <p className="mt-3 fs-13 text-muted">Not enough data to draw the projection.</p>;
  }
  const W = 600;
  const H = 220;
  const values = points.map((p) => Number(p.balance));
  const min = Math.min(...values, 0);
  const max = Math.max(...values, 1);
  const span = max - min || 1;
  const coords = points.map((p, i) => {
    const x = (i / (points.length - 1)) * W;
    const y = H - ((Number(p.balance) - min) / span) * (H - 20) - 10;
    return `${x.toFixed(1)},${y.toFixed(1)}`;
  });
  const baselineY = H - ((0 - min) / span) * (H - 20) - 10;
  return (
    <div className="mt-3">
      <svg viewBox={`0 0 ${W} ${H}`} className="w-100" role="img" aria-label="Projected balance over time">
        <line x1="0" x2={W} y1={baselineY} y2={baselineY} stroke="#e5e7eb" strokeDasharray="4 4" />
        <polyline points={coords.join(' ')} fill="none" stroke="#0e5b4e" strokeWidth="2.5" />
        <circle
          cx={coords[coords.length - 1].split(',')[0]}
          cy={coords[coords.length - 1].split(',')[1]}
          r="4"
          fill="#0e5b4e"
        />
      </svg>
      <div className="mt-1 d-flex justify-content-between fs-12 text-muted">
        <span>{points[0].date}</span>
        <span>{points[points.length - 1].date}</span>
      </div>
    </div>
  );
}