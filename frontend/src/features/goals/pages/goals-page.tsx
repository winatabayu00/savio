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
import { useGoalMutations, useGoals } from '@/features/goals/hooks/use-goals';
import type { Goal } from '@/features/goals/api/goal.api';

const goalSchema = z.object({
  name: z.string().min(1, 'Goal name is required'),
  target_amount: z.string().refine((v) => v.trim() !== '' && Number(v) > 0, 'Target must be positive'),
  current_amount: z.string().refine((v) => v === '' || Number(v) >= 0, 'Current amount cannot be negative'),
  target_date: z.string(),
  priority: z.enum(['LOW', 'MEDIUM', 'HIGH']),
});

const TWO_YEARS = new Date();
TWO_YEARS.setFullYear(TWO_YEARS.getFullYear() + 2);

export function GoalsPage() {
  const { auth } = useAuth();
  const currency = auth?.workspace.base_currency ?? 'IDR';
  const { data: goals, isLoading, isError } = useGoals();
  const { create, update, status } = useGoalMutations();
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Goal | undefined>();
  const [formError, setFormError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const { register, handleSubmit, setError, reset, formState: { errors, isSubmitting } } = useForm<z.infer<typeof goalSchema>>({
    resolver: zodResolver(goalSchema),
    defaultValues: {
      name: '',
      target_amount: '',
      current_amount: '',
      target_date: TWO_YEARS.toISOString().slice(0, 10),
      priority: 'MEDIUM',
    },
  });

  const flash = (m: string) => {
    setNotice(m);
    window.setTimeout(() => setNotice(null), 4000);
  };

  const onSubmit = handleSubmit(async (values) => {
    setFormError(null);
    const payload = {
      name: values.name,
      target_amount: values.target_amount,
      current_amount: values.current_amount || '0',
      target_date: values.target_date || null,
      priority: values.priority,
    };
    try {
      if (editing) {
        await update.mutateAsync({ id: editing.id, input: { ...payload, version: editing.version } });
        flash('Goal updated.');
      } else {
        await create.mutateAsync(payload);
        flash('Goal created.');
      }
      setFormOpen(false);
    } catch (err) {
      const apiErr = err as ApiError;
      if (apiErr?.status === 422 && apiErr.details) {
        for (const [field, msgs] of Object.entries(apiErr.details)) {
          setError(field as keyof typeof values, { message: msgs[0] });
        }
      } else {
        setFormError(apiErr?.message ?? 'Could not save goal.');
      }
    }
  });

  const openCreate = () => {
    setEditing(undefined);
    reset();
    setFormOpen(true);
  };
  const openEdit = (g: Goal) => {
    setEditing(g);
    reset({ name: g.name, target_amount: g.target_amount, current_amount: g.current_amount, target_date: g.target_date ?? '', priority: g.priority });
    setFormOpen(true);
  };

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Goals</h1>
          <p className="mt-1 text-sm text-gray-500">
            Tracked progress toward your target. Progress is what you maintain manually — it never claims money is reserved in an account.
          </p>
        </div>
        <Button onClick={openCreate}>New Goal</Button>
      </div>

      {notice ? <div className="mt-4 rounded-lg bg-brand/10 p-3 text-sm text-brand">{notice}</div> : null}
      {isLoading ? <p className="mt-6 text-sm text-gray-500">Loading goals…</p> : null}
      {isError ? (
        <div className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          We could not load your goals.
        </div>
      ) : null}

      <div className="mt-6 grid gap-4 md:grid-cols-2">
        {goals && goals.length === 0 ? (
          <div className="md:col-span-2">
            <EmptyState
              title="No goals yet"
              description="Set a savings target, like an emergency fund or a trip, and track progress over time."
              action={<Button onClick={openCreate}>Create your first goal</Button>}
            />
          </div>
        ) : null}
        {goals?.map((g) => {
          const pct = Math.min(g.progress_percent, 100);
          const feasCls = g.feasibility === 'ON_TRACK' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700';
          return (
            <div key={g.id} className="rounded-xl border border-gray-200 bg-white p-5">
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="font-semibold text-gray-900">{g.name}</h3>
                  <p className="text-xs text-gray-500">
                    Goal for {formatAmountString(g.target_amount, currency)} · {g.status.toLowerCase()}
                    {g.target_date ? ` · by ${g.target_date}` : ''}
                  </p>
                </div>
                <div className="flex gap-2">
                  <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${feasCls}`}>{g.feasibility}</span>
                  <span className="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600">{g.priority}</span>
                </div>
              </div>
              <div className="mt-4">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-600">
                    {formatAmountString(g.current_amount, currency)} of {formatAmountString(g.target_amount, currency)}
                  </span>
                  <span className="font-medium text-gray-900">{g.progress_percent.toFixed(0)}%</span>
                </div>
                <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-gray-100">
                  <div className={`h-full rounded-full ${pct >= 100 ? 'bg-green-500' : 'bg-brand'}`} style={{ width: `${pct}%` }} />
                </div>
              </div>
              <div className="mt-3 grid grid-cols-3 gap-2 text-xs text-gray-600">
                <div>
                  <div>Remaining</div>
                  <div className="font-medium text-gray-900">{formatAmountString(g.remaining, currency)}</div>
                </div>
                <div>
                  <div>Needed / month</div>
                  <div className="font-medium text-gray-900">
                    {g.months_remaining <= 0 ? '—' : formatAmountString(g.required_monthly, currency)}
                  </div>
                </div>
                <div>
                  <div>Est. free cashflow</div>
                  <div className="font-medium text-gray-900">{formatAmountString(g.estimated_monthly_income, currency)}</div>
                </div>
              </div>
              <div className="mt-4 flex flex-wrap gap-2">
                {g.status === 'ACTIVE' || g.status === 'PAUSED' ? (
                  <>
                    <Button variant="secondary" onClick={() => openEdit(g)}>Edit</Button>
                    {g.status === 'ACTIVE' ? (
                      <Button variant="ghost" onClick={() => status.mutateAsync({ id: g.id, action: 'pause', version: g.version }).then(() => flash('Goal paused.'))}>
                        Pause
                      </Button>
                    ) : (
                      <Button variant="ghost" onClick={() => status.mutateAsync({ id: g.id, action: 'resume', version: g.version }).then(() => flash('Goal resumed.'))}>
                        Resume
                      </Button>
                    )}
                    <Button variant="ghost" onClick={() => status.mutateAsync({ id: g.id, action: 'achieve', version: g.version }).then(() => flash('Goal achieved!'))}>
                      Mark achieved
                    </Button>
                  </>
                ) : null}
              </div>
            </div>
          );
        })}
      </div>

      <Modal open={formOpen} onClose={() => setFormOpen(false)} title={editing ? 'Edit goal' : 'New goal'}>
        <form onSubmit={onSubmit} className="space-y-4" noValidate>
          <TextField label="Goal name" placeholder="e.g. Emergency Fund" error={errors.name?.message} {...register('name')} />
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Target amount" inputMode="decimal" error={errors.target_amount?.message} {...register('target_amount')} />
            <TextField label="Current progress" inputMode="decimal" error={errors.current_amount?.message} {...register('current_amount')} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Target date" type="date" error={errors.target_date?.message} {...register('target_date')} />
            <div>
              <label className="mb-1.5 block text-sm font-medium text-gray-700">Priority</label>
              <select className="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm" {...register('priority')}>
                <option value="LOW">Low</option>
                <option value="MEDIUM">Medium</option>
                <option value="HIGH">High</option>
              </select>
            </div>
          </div>
          {formError ? <p role="alert" className="text-sm text-red-600">{formError}</p> : null}
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={() => setFormOpen(false)}>Cancel</Button>
            <Button type="submit" disabled={isSubmitting}>
              {editing ? 'Save changes' : 'Create goal'}
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}