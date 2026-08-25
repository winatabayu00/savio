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
    ON_TRACK: 'bg-soft-success text-success',
    WARNING: 'bg-soft-warning text-warning',
    EXCEEDED: 'bg-soft-danger text-danger',
  };

  return (
    <div>
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-3">
        <div>
          <h1 className="fs-20 fw-bolder mb-0">Budgets</h1>
          <p className="fs-13 text-muted mb-0 mt-1">Plan monthly spending per category and track progress.</p>
        </div>
        <Button onClick={openCreate}>New Budget</Button>
      </div>

      <div className="mt-4 d-flex gap-1 border-bottom">
        {(['ACTIVE', 'CLOSED'] as const).map((t) => (
          <button
            key={t}
            type="button"
            onClick={() => setTab(t)}
            className={`rounded-top px-4 py-2 fs-13 fw-medium ${
              tab === t ? 'border-bottom border-primary text-primary' : 'text-muted'
            }`}
          >
            {t === 'ACTIVE' ? 'Active' : 'Closed'}
          </button>
        ))}
      </div>

      {notice ? <div className="mt-4 bg-soft-primary text-primary p-3 fs-13 rounded">{notice}</div> : null}
      {isLoading ? <p className="mt-4 fs-13 text-muted">Loading budgets…</p> : null}
      {isError ? (
        <div className="mt-4 alert alert-danger">
          We could not load budgets.
        </div>
      ) : null}

      <div className="row g-4 mt-4">
        {budgets && budgets.length === 0 ? (
          <div className="col-12">
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
          const barColor = b.computed_status === 'EXCEEDED' ? 'bg-danger' : b.computed_status === 'WARNING' ? 'bg-warning' : 'bg-success';
          return (
            <div key={b.id} className="col-12 col-md-6">
              <div className="card h-100">
                <div className="card-body">
              <div className="d-flex align-items-start justify-content-between">
                <div>
                  <h3 className="fs-15 fw-semibold text-dark mb-3">{b.category_name}</h3>
                  <p className="fs-12 text-muted">
                    {b.period_start} → {b.period_end}
                  </p>
                </div>
                <span className={`badge ${statusStyles[b.computed_status]}`}>
                  {b.computed_status}
                </span>
              </div>
              <div className="mt-4">
                <div className="d-flex align-items-center justify-content-between fs-13">
                  <span className="text-secondary">
                    Spent {formatAmountString(b.spent, currency)} of {formatAmountString(b.amount, currency)}
                  </span>
                  <span className="fw-medium text-dark">{b.utilization_percent.toFixed(1)}%</span>
                </div>
                <div className="progress mt-2" style={{ height: 8 }}>
                  <div className={`progress-bar ${barColor}`} style={{ width: `${pct}%` }} />
                </div>
                <div className="mt-2 fs-12 text-muted">
                  Remaining: {formatAmountString(b.remaining, currency)}
                  {b.projected_overspend ? (
                    <span className="ms-2 text-danger">
                      Projected overspend: {formatAmountString(b.projected_overspend, currency)}
                    </span>
                  ) : null}
                </div>
              </div>
              <div className="mt-4 d-flex gap-2">
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
              </div>
            </div>
          );
        })}
      </div>

      <Modal open={formOpen} onClose={() => setFormOpen(false)} title={editing ? 'Edit budget' : 'New budget'}>
        <form onSubmit={onSubmit} className="d-flex flex-column gap-3" noValidate>
          <div>
            <label htmlFor="b-cat" className="form-label">Expense category</label>
            <select
              id="b-cat"
              className="form-select"
              {...register('category_id')}
            >
              <option value="">Select category…</option>
              {(categories.data ?? []).map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
            {errors.category_id ? <p className="mt-1 fs-12 text-danger">{errors.category_id.message}</p> : null}
          </div>
          <TextField label="Amount" inputMode="decimal" error={errors.amount?.message} {...register('amount')} />
          <div className="row g-3">
            <div className="col-6">
              <TextField label="Period start" type="date" error={errors.period_start?.message} {...register('period_start')} />
            </div>
            <div className="col-6">
              <TextField label="Period end" type="date" error={errors.period_end?.message} {...register('period_end')} />
            </div>
          </div>
          {formError ? <p role="alert" className="fs-13 text-danger">{formError}</p> : null}
          <div className="d-flex justify-content-end gap-2 pt-2">
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