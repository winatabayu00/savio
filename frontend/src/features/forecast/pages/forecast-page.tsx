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
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Cashflow Forecast</h1>
          <p className="mt-1 text-sm text-gray-500">
            A deterministic projection from your balances, recurring activity and spending history — not an AI prediction.
          </p>
        </div>
        <div className="flex gap-1 rounded-lg border border-gray-200 bg-white p-1">
          {FORECAST_HORIZONS.map((h) => (
            <button
              key={h}
              type="button"
              onClick={() => setHorizon(h)}
              className={`rounded-md px-3 py-1.5 text-sm font-medium ${
                horizon === h ? 'bg-brand text-white' : 'text-gray-600 hover:bg-gray-100'
              }`}
            >
              {h}d
            </button>
          ))}
        </div>
      </div>

      {data?.stale ? (
        <div className="mt-4 rounded-lg bg-amber-50 p-3 text-sm text-amber-800">
          Your finance data has changed since this projection was last calculated. Results may be stale.
        </div>
      ) : null}

      {isLoading ? (
        <div className="mt-6 space-y-3">
          <div className="h-28 animate-pulse rounded-xl bg-gray-100" />
          <div className="h-64 animate-pulse rounded-xl bg-gray-100" />
        </div>
      ) : null}
      {isError ? (
        <div className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          <p>We could not calculate your forecast.</p>
          <button className="mt-2 text-brand underline" onClick={() => void refetch()}>
            Try again
          </button>
        </div>
      ) : null}

      {data ? (
        <>
          <div className="mt-6 grid gap-4 md:grid-cols-3 lg:grid-cols-4">
            <Card label="Opening balance" value={formatAmountString(data.opening_balance, currency)} />
            <Card label={`Ending balance (${horizon}d)`} value={formatAmountString(data.ending_balance, currency)} tone={Number(data.ending_balance) < 0 ? 'bad' : 'good'} />
            <Card
              label={`Minimum balance · ${data.minimum_balance_date}`}
              value={formatAmountString(data.minimum_balance, currency)}
              tone={Number(data.minimum_balance) < 0 ? 'bad' : undefined}
            />
            <Card label="Projected income / expense" value={`${formatAmountString(data.projected_income, currency)} / ${formatAmountString(data.projected_expense, currency)}`} />
          </div>

          <div className="mt-6 grid gap-6 lg:grid-cols-2">
            <section className="rounded-xl border border-gray-200 bg-white p-5">
              <div className="flex items-center justify-between">
                <h2 className="font-semibold text-gray-900">Balance projection</h2>
                <span className="text-xs font-medium text-gray-500">confidence: {confidenceLabel}</span>
              </div>
              <BalanceChart points={data.timeline} />
            </section>

            <section className="rounded-xl border border-gray-200 bg-white p-5">
              <h2 className="font-semibold text-gray-900">Events ({data.events.length})</h2>
              {data.events.length === 0 ? (
                <p className="mt-3 text-sm text-gray-500">
                  No upcoming activity scheduled. Add recurring income or expenses to see them projected here.
                </p>
              ) : (
                <ul className="mt-3 max-h-80 divide-y divide-gray-100 overflow-y-auto">
                  {data.events.slice(0, 80).map((e, i) => (
                    <li key={i} className="flex items-center justify-between py-2 text-sm">
                      <div className="min-w-0">
                        <div className="truncate font-medium text-gray-900">{e.description}</div>
                        <div className="text-xs text-gray-500">
                          {e.date} · {e.type}
                        </div>
                      </div>
                      <span className={e.kind === 'EXPENSE' ? 'font-semibold text-red-600' : 'font-semibold text-green-600'}>
                        {e.kind === 'EXPENSE' ? '−' : '+'}
                        {formatAmountString(e.amount, currency)}
                      </span>
                    </li>
                  ))}
                </ul>
              )}
            </section>
          </div>

          <section className="mt-6 rounded-xl border border-gray-200 bg-white p-5 text-sm">
            <h2 className="font-semibold text-gray-900">Assumptions</h2>
            <ul className="mt-3 grid gap-2 text-gray-600 md:grid-cols-2">
              <li>Estimated daily variable spending: <span className="font-medium text-gray-900">{formatAmountString(data.assumptions.variable_expense_daily, currency)}</span></li>
              <li>History used for baseline: <span className="font-medium text-gray-900">{data.assumptions.baseline_days} days</span></li>
              <li>Active recurring rules: <span className="font-medium text-gray-900">{data.assumptions.active_recurring_rules}</span></li>
              <li>Confidence basis: <span className="font-medium text-gray-900">{data.assumptions.confidence_basis}</span></li>
            </ul>
            <p className="mt-3 text-xs text-gray-400">Algorithm version {data.calculation_version}. Projections are estimates, not guarantees of future balances.</p>
          </section>
        </>
      ) : null}
    </div>
  );
}

function Card({ label, value, tone }: { label: string; value: string; tone?: 'good' | 'bad' }) {
  const color = tone === 'good' ? 'text-green-600' : tone === 'bad' ? 'text-red-600' : 'text-gray-900';
  return (
    <div className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
      <div className="text-sm text-gray-500">{label}</div>
      <div className={`mt-1 text-xl font-semibold ${color}`}>{value}</div>
    </div>
  );
}

function BalanceChart({ points }: { points: { date: string; balance: string }[] }) {
  if (points.length < 2) {
    return <p className="mt-4 text-sm text-gray-500">Not enough data to draw the projection.</p>;
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
    <div className="mt-4">
      <svg viewBox={`0 0 ${W} ${H}`} className="w-full" role="img" aria-label="Projected balance over time">
        <line x1="0" x2={W} y1={baselineY} y2={baselineY} stroke="#e5e7eb" strokeDasharray="4 4" />
        <polyline points={coords.join(' ')} fill="none" stroke="#0e5b4e" strokeWidth="2.5" />
        <circle
          cx={coords[coords.length - 1].split(',')[0]}
          cy={coords[coords.length - 1].split(',')[1]}
          r="4"
          fill="#0e5b4e"
        />
      </svg>
      <div className="mt-1 flex justify-between text-xs text-gray-400">
        <span>{points[0].date}</span>
        <span>{points[points.length - 1].date}</span>
      </div>
    </div>
  );
}