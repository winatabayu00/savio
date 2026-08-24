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
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Scenario Simulator</h1>
          <p className="mt-1 text-sm text-gray-500">
            Ask “what if?” without touching your real money. Scenarios are non-destructive overlays on your forecast.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>New Scenario</Button>
      </div>

      {notice ? <div className="mt-4 rounded-lg bg-brand/10 p-3 text-sm text-brand">{notice}</div> : null}
      {isLoading ? <p className="mt-6 text-sm text-gray-500">Loading scenarios…</p> : null}
      {isError ? (
        <div className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          We could not load your scenarios.
        </div>
      ) : null}

      <div className="mt-6 grid gap-6 lg:grid-cols-[300px_1fr]">
        <aside className="space-y-2">
          {scenarios && scenarios.length === 0 ? (
            <EmptyState title="No scenarios yet" description="Create one to simulate a purchase, raise, or cost reduction." />
          ) : null}
          {scenarios?.map((s) => (
            <div
              key={s.id}
              className={`rounded-xl border bg-white p-4 ${selectedId === s.id ? 'border-brand ring-1 ring-brand' : 'border-gray-200'}`}
            >
              <button type="button" className="w-full text-left" onClick={() => setSelectedId(s.id)}>
                <div className="flex items-center justify-between">
                  <span className="font-semibold text-gray-900">{s.name}</span>
                  <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${s.status === 'CALCULATED' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                    {s.status}
                  </span>
                </div>
                <div className="mt-1 text-xs text-gray-500">{s.modifications.length} modification{s.modifications.length === 1 ? '' : 's'}</div>
                {s.is_stale ? <div className="mt-1 text-xs font-medium text-amber-600">stale — recalculate</div> : null}
              </button>
            </div>
          ))}
        </aside>

        <main>
          {selected ? (
            <div className="rounded-2xl border border-gray-200 bg-white p-6">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <h2 className="text-xl font-semibold">{selected.name}</h2>
                  {selected.description ? <p className="mt-1 text-sm text-gray-500">{selected.description}</p> : null}
                </div>
                <div className="flex gap-2">
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
                <div className="mt-4 rounded-lg bg-amber-50 p-3 text-sm text-amber-800">
                  Your finance data changed since this result. Recalculate for a current comparison.
                </div>
              ) : null}

              <h3 className="mt-6 text-sm font-semibold uppercase tracking-wide text-gray-500">Modifications</h3>
              {selected.modifications.length === 0 ? (
                <p className="mt-2 text-sm text-gray-500">No modifications yet. Add one to change the projection.</p>
              ) : (
                <ul className="mt-2 space-y-2">
                  {selected.modifications.map((m) => (
                    <li key={m.id} className="flex items-center justify-between rounded-lg border border-gray-100 px-4 py-2.5 text-sm">
                      <div>
                        <span className="font-medium text-gray-900">{MOD_LABELS[m.type]}</span>
                        {m.narrative ? <span className="ml-2 text-gray-500">· {m.narrative}</span> : null}
                        {m.frequency ? <span className="ml-2 text-xs text-gray-400">{m.frequency}</span> : null}
                      </div>
                      <div className="flex items-center gap-3">
                        <span className={`font-semibold ${m.type.includes('EXPENSE') || m.type.includes('REDUCTION') ? 'text-red-600' : 'text-green-600'}`}>
                          {formatAmountString(m.amount, currency)}
                        </span>
                        <button className="text-gray-400 hover:text-red-600" onClick={() => removeMod.mutateAsync({ scenarioId: selected.id, modId: m.id })} aria-label="Remove">
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
                <div className="mt-6 rounded-xl border border-dashed border-gray-300 p-8 text-center">
                  <p className="text-sm text-gray-500">Add modifications, then Calculate to see baseline vs scenario.</p>
                </div>
              )}
            </div>
          ) : (
            <div className="rounded-2xl border border-dashed border-gray-300 bg-white p-10 text-center">
              <p className="text-sm text-gray-500">Select a scenario, or create a new one.</p>
              <Button className="mt-4" onClick={() => setCreateOpen(true)}>Create your first scenario</Button>
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
        <form onSubmit={onAddMod} className="space-y-4" noValidate>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-gray-700">What if…</label>
            <select className="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm" {...register('type')}>
              {Object.entries(MOD_LABELS).map(([v, label]) => (
                <option key={v} value={v}>{label}</option>
              ))}
            </select>
          </div>
          <TextField label="Amount" inputMode="decimal" error={errors.amount?.message} {...register('amount')} />
          <TextField label="Note (optional)" placeholder="e.g. new car" error={errors.narrative?.message} {...register('narrative')} />
          {formError ? <p role="alert" className="text-sm text-red-600">{formError}</p> : null}
          <div className="flex justify-end gap-2 pt-2">
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
    <div className="space-y-4">
      <TextField label="Scenario name" placeholder="e.g. Buy a new car" value={name} onChange={(e) => setName(e.target.value)} />
      <div className="flex justify-end gap-2">
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
    <div className="mt-6">
      <div className="flex items-center gap-2">
        <h3 className="text-sm font-semibold uppercase tracking-wide text-gray-500">Baseline vs scenario</h3>
        <span className="rounded-full bg-brand/10 px-2 py-0.5 text-xs font-medium text-brand">
          net effect {formatAmountString(r.cashflow_difference, currency)}
        </span>
      </div>
      <div className="mt-3 grid gap-3 md:grid-cols-4">
        {diffs.map((d) => (
          <div key={d.label} className="rounded-lg border border-gray-100 p-3">
            <div className="text-xs text-gray-500">{d.label}</div>
            <div className="mt-1 text-xs text-gray-400">baseline: {formatAmountString(d.base, currency)}</div>
            <div className="text-sm font-semibold text-gray-900">{formatAmountString(d.scen, currency)}</div>
          </div>
        ))}
      </div>

      {r.timeline.length > 2 ? (
        <div className="mt-6">
          <h4 className="text-sm font-semibold text-gray-700">Projection overlay</h4>
          <OverlayChart timeline={r.timeline} />
        </div>
      ) : null}
      <p className="mt-4 text-xs text-gray-400">{r.assumption_note}</p>
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
    <svg viewBox={`0 0 ${W} ${H}`} className="w-full" role="img" aria-label="Baseline vs scenario projection">
      <polyline points={line('baseline_balance')} fill="none" stroke="#9ca3af" strokeWidth="2" strokeDasharray="6 4" />
      <polyline points={line('scenario_balance')} fill="none" stroke="#0e5b4e" strokeWidth="2.5" />
    </svg>
  );
}