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
import { useCategories } from '@/features/categories/hooks/use-categories';
import { useBudgetMutations, useBudgets } from '@/features/budgets/hooks/use-budgets';
import type { Budget } from '@/features/budgets/api/budget.api';

const budgetSchema = z.object({
  category_id: z.string().min(1, 'Category is required'),
  amount: z.string().refine((v) => v.trim() !== '' && Number(v) > 0, 'Amount must be greater than zero'),
  period_start: z.string().min(1, 'Start date is required'),
  period_end: z.string().min(1, 'End date is required'),
});

function currentMonth(): { start: string; end: string } {
  const now = new Date();
  const pad = (n: number) => String(n).padStart(2, '0');
  const first = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-01`;
  const last = new Date(now.getFullYear(), now.getMonth() + 1, 0);
  return { start: first, end: `${last.getFullYear()}-${pad(last.getMonth() + 1)}-${pad(last.getDate())}` };
}

export function BudgetsPage() {
  const { auth } = useAuth();
  const currency = auth?.workspace.base_currency ?? 'IDR';
  const [tab, setTab] = useState<'ACTIVE' | 'CLOSED'>('ACTIVE');
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Budget | undefined>();
  const [notice, setNotice] = useState<string | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const categories = useCategories('EXPENSE');
  const { data: budgets, isLoading, isError } = useBudgets(tab);
  const { create, update, close } = useBudgetMutations();
  const m = currentMonth();

  const { register, handleSubmit, setError, reset, formState: { errors, isSubmitting } } = useForm<z.infer<typeof budgetSchema>>({
    resolver: zodResolver(budgetSchema),
    defaultValues: { category_id: '', amount: '', period_start: m.start, period_end: m.end },
  });

  const flash = (msg: string) => {
    setNotice(msg);
    window.setTimeout(() => setNotice(null), 4000);
  };

  const openCreate = () => {
    setEditing(undefined);
    reset({ category_id: '', amount: '', period_start: m.start, period_end: m.end });
    setFormOpen(true);
  };

  const openEdit = (b: Budget) => {
    setEditing(b);
    reset({ category_id: b.category_id, amount: b.amount, period_start: b.period_start, period_end: b.period_end });
    setFormOpen(true);
  };

  const onSubmit = handleSubmit(async (values) => {
    setFormError(null);
    try {
      if (editing) {
        await update.mutateAsync({ id: editing.id, input: { ...values, version: editing.version } });
        flash('Budget updated.');
      } else {
        await create.mutateAsync(values);
        flash('Budget created.');
      }
      setFormOpen(false);
    } catch (err) {
      const apiErr = err as ApiError;
      if (apiErr?.status === 422 && apiErr.details) {
        for (const [field, msgs] of Object.entries(apiErr.details)) {
          setError(field as keyof typeof values, { message: msgs[0] });
        }
      } else {
        setFormError(apiErr?.message ?? 'Could not save budget.');
      }
    }
  });

  const statusStyles: Record<string, string> = {
    ON_TRACK: 'bg-green-100 text-green-700',
    WARNING: 'bg-amber-100 text-amber-700',
    EXCEEDED: 'bg-red-100 text-red-700',
  };

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Budgets</h1>
          <p className="mt-1 text-sm text-gray-500">Plan monthly spending per category and track progress.</p>
        </div>
        <Button onClick={openCreate}>New Budget</Button>
      </div>

      <div className="mt-4 flex gap-1 border-b border-gray-200">
        {(['ACTIVE', 'CLOSED'] as const).map((t) => (
          <button
            key={t}
            type="button"
            onClick={() => setTab(t)}
            className={`rounded-t-lg px-4 py-2 text-sm font-medium ${
              tab === t ? 'border-b-2 border-brand text-brand' : 'text-gray-500 hover:text-gray-700'
            }`}
          >
            {t === 'ACTIVE' ? 'Active' : 'Closed'}
          </button>
        ))}
      </div>

      {notice ? <div className="mt-4 rounded-lg bg-brand/10 p-3 text-sm text-brand">{notice}</div> : null}
      {isLoading ? <p className="mt-6 text-sm text-gray-500">Loading budgets…</p> : null}
      {isError ? (
        <div className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          We could not load budgets.
        </div>
      ) : null}

      <div className="mt-6 grid gap-4 md:grid-cols-2">
        {budgets && budgets.length === 0 ? (
          <div className="md:col-span-2">
            <EmptyState
              title={tab === 'ACTIVE' ? 'No budgets yet' : 'No closed budgets'}
              description={
                tab === 'ACTIVE'
                  ? 'Create a monthly limit for an expense category to stay on track.'
                  : 'Closed budgets archive here for history.'
              }
              action={tab === 'ACTIVE' ? <Button onClick={openCreate}>Create your first budget</Button> : undefined}
            />
          </div>
        ) : null}
        {budgets?.map((b) => {
          const pct = Math.min(b.utilization_percent, 100);
          const barColor = b.computed_status === 'EXCEEDED' ? 'bg-red-500' : b.computed_status === 'WARNING' ? 'bg-amber-500' : 'bg-green-500';
          return (
            <div key={b.id} className="rounded-xl border border-gray-200 bg-white p-5">
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="font-semibold text-gray-900">{b.category_name}</h3>
                  <p className="text-xs text-gray-500">
                    {b.period_start} → {b.period_end}
                  </p>
                </div>
                <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${statusStyles[b.computed_status]}`}>
                  {b.computed_status}
                </span>
              </div>
              <div className="mt-4">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-600">
                    Spent {formatAmountString(b.spent, currency)} of {formatAmountString(b.amount, currency)}
                  </span>
                  <span className="font-medium text-gray-900">{b.utilization_percent.toFixed(1)}%</span>
                </div>
                <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-gray-100">
                  <div className={`h-full rounded-full ${barColor}`} style={{ width: `${pct}%` }} />
                </div>
                <div className="mt-2 text-xs text-gray-500">
                  Remaining: {formatAmountString(b.remaining, currency)}
                  {b.projected_overspend ? (
                    <span className="ml-2 text-red-600">
                      Projected overspend: {formatAmountString(b.projected_overspend, currency)}
                    </span>
                  ) : null}
                </div>
              </div>
              <div className="mt-4 flex gap-2">
                {b.status === 'ACTIVE' ? (
                  <>
                    <Button variant="secondary" onClick={() => openEdit(b)}>Edit</Button>
                    <Button
                      variant="ghost"
                      onClick={() => close.mutateAsync({ id: b.id, version: b.version }).then(() => flash('Budget closed.'))}
                    >
                      Close
                    </Button>
                  </>
                ) : null}
              </div>
            </div>
          );
        })}
      </div>

      <Modal open={formOpen} onClose={() => setFormOpen(false)} title={editing ? 'Edit budget' : 'New budget'}>
        <form onSubmit={onSubmit} className="space-y-4" noValidate>
          <div>
            <label htmlFor="b-cat" className="mb-1.5 block text-sm font-medium text-gray-700">Expense category</label>
            <select
              id="b-cat"
              className="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:outline-none focus:ring-2 focus:ring-brand/30"
              {...register('category_id')}
            >
              <option value="">Select category…</option>
              {(categories.data ?? []).map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
            {errors.category_id ? <p className="mt-1 text-xs text-red-600">{errors.category_id.message}</p> : null}
          </div>
          <TextField label="Amount" inputMode="decimal" error={errors.amount?.message} {...register('amount')} />
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Period start" type="date" error={errors.period_start?.message} {...register('period_start')} />
            <TextField label="Period end" type="date" error={errors.period_end?.message} {...register('period_end')} />
          </div>
          {formError ? <p role="alert" className="text-sm text-red-600">{formError}</p> : null}
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={() => setFormOpen(false)}>Cancel</Button>
            <Button type="submit" disabled={isSubmitting}>
              {editing ? 'Save changes' : 'Create budget'}
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}