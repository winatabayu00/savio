import { useState } from 'react';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { useAuth } from '@/app/providers/auth-provider';
import { ApiError } from '@/shared/api/client';
import { Button } from '@/shared/components/ui/button';
import { EmptyState } from '@/shared/components/ui/empty-state';
import { Modal } from '@/shared/components/ui/modal';
import { TextField } from '@/shared/components/ui/text-field';
import { formatAmountString } from '@/shared/utils/money';
import { useScenarioMutations, useScenarios } from '@/features/scenarios/hooks/use-scenarios';
import { MOD_LABELS, type ModType, type Scenario } from '@/features/scenarios/api/scenario.api';

const modSchema = z.object({
  type: z.string().min(1),
  amount: z.string().refine((v) => v.toString() !== '' && Number(v) > 0, 'Amount must be positive'),
  frequency: z.string(),
  narrative: z.string().max(200),
});

export function ScenariosPage() {
  const { auth } = useAuth();
  const currency = auth?.workspace.base_currency ?? 'IDR';
  const { data: scenarios, isLoading, isError } = useScenarios();
  const { create, remove, addMod, removeMod, calc } = useScenarioMutations();
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [modOpen, setModOpen] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const [calcBusy, setCalcBusy] = useState(false);

  const { register, handleSubmit, reset, formState: { errors, isSubmitting } } = useForm<z.infer<typeof modSchema>>({
    resolver: zodResolver(modSchema),
    defaultValues: { type: 'ONE_TIME_EXPENSE', amount: '', frequency: 'MONTHLY', narrative: '' },
  });

  const selected = scenarios?.find((s) => s.id === selectedId);

  const flash = (m: string) => {
    setNotice(m);
    window.setTimeout(() => setNotice(null), 4000);
  };

  const onAddMod = handleSubmit(async (values) => {
    setFormError(null);
    if (!selectedId) return;
    try {
      await addMod.mutateAsync({
        id: selectedId,
        input: {
          type: values.type as ModType,
          amount: values.amount,
          frequency: values.type.startsWith('RECURRING') ? values.frequency : undefined,
          narrative: values.narrative || undefined,
        },
      });
      setModOpen(false);
      reset();
    } catch (err) {
      setFormError((err as ApiError)?.message ?? 'Could not add modification.');
    }
  });

  const onCalculate = async () => {
    if (!selectedId) return;
    setCalcBusy(true);
    try {
      await calc.mutateAsync({ id: selectedId });
      flash('Scenario calculated.');
    } catch (err) {
      flash((err as ApiError)?.message ?? 'Calculation failed.');
    } finally {
      setCalcBusy(false);
    }
  };

  return (
    <div>
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-3">
        <div>
          <h1 className="fs-20 fw-bolder mb-0">Scenario Simulator</h1>
          <p className="fs-13 text-muted mb-0 mt-1">
            Ask “what if?” without touching your real money. Scenarios are non-destructive overlays on your forecast.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>New Scenario</Button>
      </div>

      {notice ? <div className="mt-4 bg-soft-primary text-primary p-3 fs-13 rounded-3">{notice}</div> : null}
      {isLoading ? <p className="mt-4 fs-13 text-muted">Loading scenarios…</p> : null}
      {isError ? (
        <div className="mt-4 alert alert-danger p-3 fs-13 mb-0">
          We could not load your scenarios.
        </div>
      ) : null}

      <div className="row g-4 mt-1">
        <aside className="col-lg-3 d-flex flex-column gap-2">
          {scenarios && scenarios.length === 0 ? (
            <EmptyState title="No scenarios yet" description="Create one to simulate a purchase, raise, or cost reduction." />
          ) : null}
          {scenarios?.map((s) => (
            <div
              key={s.id}
              className={`card p-3 ${selectedId === s.id ? 'border-primary' : ''}`}
            >
              <button type="button" className="w-100 text-start" onClick={() => setSelectedId(s.id)}>
                <div className="d-flex align-items-center justify-content-between">
                  <span className="fw-semibold text-dark">{s.name}</span>
                  <span className={`badge ${s.status === 'CALCULATED' ? 'bg-soft-success text-success' : 'bg-soft-secondary text-secondary'}`}>
                    {s.status}
                  </span>
                </div>
                <div className="mt-1 fs-12 text-muted">{s.modifications.length} modification{s.modifications.length === 1 ? '' : 's'}</div>
                {s.is_stale ? <div className="mt-1 fs-12 fw-medium text-warning">stale — recalculate</div> : null}
              </button>
            </div>
          ))}
        </aside>

        <main className="col-lg-9">
          {selected ? (
            <div className="card">
              <div className="card-body">
              <div className="d-flex flex-wrap align-items-start justify-content-between gap-3">
                <div>
                  <h2 className="fs-18 fw-semibold text-dark">{selected.name}</h2>
                  {selected.description ? <p className="mt-1 fs-13 text-muted mb-0">{selected.description}</p> : null}
                </div>
                <div className="d-flex gap-2">
                  <Button variant="secondary" onClick={() => setModOpen(true)}>Add modification</Button>
                  <Button onClick={onCalculate} disabled={calcBusy}>
                    {calcBusy ? 'Calculating…' : 'Calculate'}
                  </Button>
                  <Button variant="ghost" onClick={() => remove.mutateAsync(selected.id).then(() => setSelectedId(null)).catch(() => undefined)}>
                    Delete
                  </Button>
                </div>
              </div>

              {selected.is_stale ? (
                <div className="mt-4 alert alert-warning p-3 fs-13 mb-0">
                  Your finance data changed since this result. Recalculate for a current comparison.
                </div>
              ) : null}

              <h3 className="mt-4 fs-12 text-uppercase fw-semibold text-muted mb-0">Modifications</h3>
              {selected.modifications.length === 0 ? (
                <p className="mt-2 fs-13 text-muted">No modifications yet. Add one to change the projection.</p>
              ) : (
                <ul className="mt-2 list-unstyled d-flex flex-column gap-2 mb-0">
                  {selected.modifications.map((m) => (
                    <li key={m.id} className="d-flex align-items-center justify-content-between border rounded p-2 fs-13">
                      <div>
                        <span className="fw-medium text-dark">{MOD_LABELS[m.type]}</span>
                        {m.narrative ? <span className="ms-2 text-secondary">· {m.narrative}</span> : null}
                        {m.frequency ? <span className="ms-2 fs-12 text-muted">{m.frequency}</span> : null}
                      </div>
                      <div className="d-flex align-items-center gap-3">
                        <span className={`fw-semibold ${m.type.includes('EXPENSE') || m.type.includes('REDUCTION') ? 'text-danger' : 'text-success'}`}>
                          {formatAmountString(m.amount, currency)}
                        </span>
                        <button className="text-muted" onClick={() => removeMod.mutateAsync({ scenarioId: selected.id, modId: m.id })} aria-label="Remove">
                          ×
                        </button>
                      </div>
                    </li>
                  ))}
                </ul>
              )}

              {selected.result ? (
                <ResultComparison scenario={selected} currency={currency} />
              ) : (
                <div className="mt-4 card border-dashed text-center">
                  <div className="card-body py-5">
                    <p className="fs-13 text-muted mb-0">Add modifications, then Calculate to see baseline vs scenario.</p>
                  </div>
                </div>
              )}
              </div>
            </div>
          ) : (
            <div className="card border-dashed text-center">
              <div className="card-body py-5">
              <p className="fs-13 text-muted mb-0">Select a scenario, or create a new one.</p>
              <Button className="mt-4" onClick={() => setCreateOpen(true)}>Create your first scenario</Button>
              </div>
            </div>
          )}
        </main>
      </div>

      <Modal open={createOpen} onClose={() => setCreateOpen(false)} title="New scenario">
        <NewScenarioForm
          onCancel={() => setCreateOpen(false)}
          onSubmit={async (name) => {
            await create.mutateAsync(name);
            setCreateOpen(false);
            flash('Scenario created.');
          }}
        />
      </Modal>

      <Modal open={modOpen} onClose={() => setModOpen(false)} title="Add modification">
        <form onSubmit={onAddMod} className="d-flex flex-column gap-3" noValidate>
          <div>
            <label className="form-label">What if…</label>
            <select className="form-select" {...register('type')}>
              {Object.entries(MOD_LABELS).map(([v, label]) => (
                <option key={v} value={v}>{label}</option>
              ))}
            </select>
          </div>
          <TextField label="Amount" inputMode="decimal" error={errors.amount?.message} {...register('amount')} />
          <TextField label="Note (optional)" placeholder="e.g. new car" error={errors.narrative?.message} {...register('narrative')} />
          {formError ? <p role="alert" className="fs-13 text-danger mb-0">{formError}</p> : null}
          <div className="d-flex justify-content-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={() => setModOpen(false)}>Cancel</Button>
            <Button type="submit" disabled={isSubmitting}>Add</Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}

function NewScenarioForm({ onCancel, onSubmit }: { onCancel: () => void; onSubmit: (name: string) => void }) {
  const [name, setName] = useState('');
  return (
    <div className="d-flex flex-column gap-3">
      <TextField label="Scenario name" placeholder="e.g. Buy a new car" value={name} onChange={(e) => setName(e.target.value)} />
      <div className="d-flex justify-content-end gap-2">
        <Button variant="secondary" onClick={onCancel}>Cancel</Button>
        <Button onClick={() => onSubmit(name)} disabled={name.trim() === ''}>Create</Button>
      </div>
    </div>
  );
}

function ResultComparison({ scenario, currency }: { scenario: Scenario; currency: string }) {
  const r = scenario.result!;
  const diffs = [
    { label: 'Ending balance', base: r.baseline_ending_balance, scen: r.scenario_ending_balance },
    { label: 'Minimum balance', base: r.baseline_minimum_balance, scen: r.scenario_minimum_balance },
    { label: 'Income', base: r.baseline_income, scen: r.scenario_income },
    { label: 'Expenses', base: r.baseline_expense, scen: r.scenario_expense },
  ];
  return (
    <div className="mt-4">
      <div className="d-flex align-items-center gap-2">
        <h3 className="fs-12 text-uppercase fw-semibold text-muted mb-0">Baseline vs scenario</h3>
        <span className="badge bg-soft-primary text-primary">
          net effect {formatAmountString(r.cashflow_difference, currency)}
        </span>
      </div>
      <div className="row g-3 mt-1">
        {diffs.map((d) => (
          <div key={d.label} className="col-12 col-md-3">
            <div className="card h-100">
              <div className="card-body p-3">
                <div className="fs-12 text-muted">{d.label}</div>
                <div className="mt-1 fs-12 text-muted">baseline: {formatAmountString(d.base, currency)}</div>
                <div className="fs-13 fw-semibold text-dark">{formatAmountString(d.scen, currency)}</div>
              </div>
            </div>
          </div>
        ))}
      </div>

      {r.timeline.length > 2 ? (
        <div className="mt-4">
          <h4 className="fs-13 fw-semibold text-secondary">Projection overlay</h4>
          <OverlayChart timeline={r.timeline} />
        </div>
      ) : null}
      <p className="mt-3 fs-12 text-muted mb-0">{r.assumption_note}</p>
    </div>
  );
}

function OverlayChart({ timeline }: { timeline: { date: string; baseline_balance: string; scenario_balance: string }[] }) {
  const pts = timeline;
  const W = 800;
  const H = 220;
  const vals = pts.flatMap((p) => [Number(p.baseline_balance), Number(p.scenario_balance)]);
  const min = Math.min(...vals, 0);
  const max = Math.max(...vals, 1);
  const span = max - min || 1;
  const line = (key: 'baseline_balance' | 'scenario_balance') =>
    pts.map((p, i) => {
      const x = (i / (pts.length - 1)) * W;
      const y = H - ((Number(p[key]) - min) / span) * (H - 20) - 10;
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    }).join(' ');
  return (
    <svg viewBox={`0 0 ${W} ${H}`} className="w-100 mt-3" role="img" aria-label="Baseline vs scenario projection">
      <polyline points={line('baseline_balance')} fill="none" stroke="#9ca3af" strokeWidth="2" strokeDasharray="6 4" />
      <polyline points={line('scenario_balance')} fill="none" stroke="#0e5b4e" strokeWidth="2.5" />
    </svg>
  );
}