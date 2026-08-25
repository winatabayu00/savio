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
import { useAccounts } from '@/features/accounts/hooks/use-accounts';
import { useCategories } from '@/features/categories/hooks/use-categories';
import {
  useOccurrences,
  useRecurring,
  useRecurringMutations,
} from '@/features/recurring/hooks/use-recurring';
import type { RecurringInput, RecurringRule } from '@/features/recurring/api/recurring.api';

const ruleSchema = z.object({
  type: z.enum(['INCOME', 'EXPENSE']),
  account_id: z.string().min(1, 'Account is required'),
  category_id: z.string(),
  amount: z.string().refine((v) => v.trim() !== '' && Number(v) > 0, 'Amount must be greater than zero'),
  frequency: z.enum(['DAILY', 'WEEKLY', 'MONTHLY', 'MONTH_END']),
  start_date: z.string().min(1, 'Start date is required'),
  end_date: z.string(),
  description: z.string().max(300),
  merchant: z.string().max(200),
});

export function RecurringPage() {
  const { auth } = useAuth();
  const currency = auth?.workspace.base_currency ?? 'IDR';
  const accounts = useAccounts('ACTIVE');
  const categories = useCategories();
  const { data: rules, isLoading, isError } = useRecurring();
  const { create, update, status, confirm, skip } = useRecurringMutations();

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<RecurringRule | undefined>();
  const [expanded, setExpanded] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [formError, setFormError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    setError,
    watch,
    formState: { errors, isSubmitting },
  } = useForm({
    resolver: zodResolver(ruleSchema),
    defaultValues: {
      type: 'EXPENSE',
      account_id: '',
      category_id: '',
      amount: '',
      frequency: 'MONTHLY',
      start_date: new Date().toISOString().slice(0, 10),
      end_date: '',
      description: '',
      merchant: '',
    },
  });

  const flash = (m: string) => {
    setNotice(m);
    window.setTimeout(() => setNotice(null), 4000);
  };

  const openCreate = () => {
    setEditing(undefined);
    reset();
    setFormOpen(true);
  };

  const openEdit = (rule: RecurringRule) => {
    setEditing(rule);
    reset({
      type: rule.type,
      account_id: rule.account_id,
      category_id: rule.category_id ?? '',
      amount: rule.amount,
      frequency: rule.frequency,
      start_date: rule.start_date,
      end_date: rule.end_date ?? '',
      description: rule.description ?? '',
      merchant: rule.merchant ?? '',
    });
    setFormOpen(true);
  };

  const txType = (watch('type') ?? 'EXPENSE') as 'INCOME' | 'EXPENSE';

  const onSubmit = handleSubmit(async (values) => {
    setFormError(null);
    const payload: RecurringInput = {
      type: values.type as 'INCOME' | 'EXPENSE',
      account_id: values.account_id,
      category_id: values.category_id || null,
      amount: values.amount,
      frequency: values.frequency as RecurringInput['frequency'],
      start_date: values.start_date,
      end_date: values.end_date || null,
      description: values.description,
      merchant: values.merchant || undefined,
    };
    try {
      if (editing) {
        await update.mutateAsync({ id: editing.id, input: { ...payload, version: editing.version } });
      } else {
        await create.mutateAsync(payload);
      }
      setFormOpen(false);
      flash(editing ? 'Recurring transaction updated.' : 'Recurring transaction created.');
    } catch (err) {
      const apiErr = err as ApiError;
      if (apiErr?.status === 422 && apiErr.details) {
        for (const [field, msgs] of Object.entries(apiErr.details)) {
          setError(field as keyof typeof values, { message: msgs[0] });
        }
      } else {
        setFormError(apiErr?.message ?? 'Could not save recurring transaction.');
      }
    }
  });

  const inputCls = 'form-select';
  const labelCls = 'form-label';

  return (
    <div>
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-3">
        <div>
          <h1 className="fs-20 fw-bolder mb-0">Recurring Transactions</h1>
          <p className="fs-13 text-muted mb-0 mt-1">
            Planned income and expenses. You confirm each scheduled occurrence before it becomes actual.
          </p>
        </div>
        <Button onClick={openCreate}>New Recurring</Button>
      </div>

      {notice ? <div className="mt-4 bg-soft-primary text-primary p-3 fs-13 rounded-3">{notice}</div> : null}
      {isLoading ? <p className="mt-4 fs-13 text-muted">Loading recurring transactions…</p> : null}
      {isError ? (
        <div className="mt-4 alert alert-danger p-3 fs-13 mb-0">
          We could not load recurring transactions.
        </div>
      ) : null}

      <div className="mt-4 d-flex flex-column gap-3">
        {rules && rules.length === 0 ? (
          <EmptyState
            title="No recurring transactions yet"
            description="Set up rent, salary or subscriptions to see what your cashflow looks like coming up."
            action={<Button onClick={openCreate}>Create your first recurring transaction</Button>}
          />
        ) : null}
        {rules?.map((rule) => (
          <RecurringRow
            key={rule.id}
            rule={rule}
            currency={currency}
            expanded={expanded === rule.id}
            onToggle={() => setExpanded(expanded === rule.id ? null : rule.id)}
            onEdit={() => openEdit(rule)}
            onStatus={(action) =>
              status
                .mutateAsync({ id: rule.id, action, version: rule.version })
                .then(() => flash(`Recurring transaction ${action === 'pause' ? 'paused' : action === 'resume' ? 'resumed' : 'ended'}.`))
                .catch((e) => flash((e as ApiError)?.message ?? 'Action failed.'))
            }
            onConfirm={(occ) =>
              confirm
                .mutateAsync({ id: occ.id, version: occ.version })
                .then(() => flash('Occurrence confirmed and posted.'))
                .catch((e) => flash((e as ApiError)?.message ?? 'Could not confirm.'))
            }
            onSkip={(occ) =>
              skip
                .mutateAsync({ id: occ.id, version: occ.version })
                .then(() => flash('Occurrence skipped.'))
                .catch((e) => flash((e as ApiError)?.message ?? 'Could not skip.'))
            }
          />
        ))}
      </div>

      <Modal open={formOpen} onClose={() => setFormOpen(false)} title={editing ? 'Edit recurring transaction' : 'New recurring transaction'}>
        <form onSubmit={onSubmit} className="d-flex flex-column gap-3" noValidate>
          <div className="row g-3">
            <div className="col-6">
              <label className={labelCls}>Type</label>
              <select className={inputCls} {...register('type')}>
                <option value="EXPENSE">Expense</option>
                <option value="INCOME">Income</option>
              </select>
            </div>
            <div className="col-6">
              <label className={labelCls}>Frequency</label>
              <select className={inputCls} {...register('frequency')}>
                <option value="MONTHLY">Monthly</option>
                <option value="MONTH_END">Monthly, end of month</option>
                <option value="WEEKLY">Weekly</option>
                <option value="DAILY">Daily</option>
              </select>
            </div>
          </div>
          <div>
            <label className={labelCls}>Account</label>
            <select className={inputCls} {...register('account_id')}>
              <option value="">Select account…</option>
              {(accounts.data?.data ?? []).map((a) => (
                <option key={a.id} value={a.id}>{a.name}</option>
              ))}
            </select>
            {errors.account_id ? <p className="mt-1 fs-12 text-danger">{errors.account_id.message}</p> : null}
          </div>
          <div>
            <label className={labelCls}>Category</label>
            <select className={inputCls} {...register('category_id')}>
              <option value="">No category</option>
              {(categories.data ?? [])
                .filter((c) => c.type === txType)
                .map((c) => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
            </select>
          </div>
          <div className="row g-3">
            <div className="col-6">
              <TextField label="Amount" inputMode="decimal" error={errors.amount?.message} {...register('amount')} />
            </div>
            <div className="col-6">
              <TextField label="Merchant (optional)" {...register('merchant')} />
            </div>
          </div>
          <div className="row g-3">
            <div className="col-6">
              <TextField label="Start date" type="date" error={errors.start_date?.message} {...register('start_date')} />
            </div>
            <div className="col-6">
              <TextField label="End date (optional)" type="date" {...register('end_date')} />
            </div>
          </div>
          <TextField label="Description (optional)" {...register('description')} />
          {formError ? <p role="alert" className="fs-13 text-danger mb-0">{formError}</p> : null}
          <div className="d-flex justify-content-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={() => setFormOpen(false)}>Cancel</Button>
            <Button type="submit" disabled={isSubmitting}>
              {editing ? 'Save changes' : 'Create recurring'}
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}

function RecurringRow({
  rule,
  currency,
  expanded,
  onToggle,
  onEdit,
  onStatus,
  onConfirm,
  onSkip,
}: {
  rule: RecurringRule;
  currency: string;
  expanded: boolean;
  onToggle: () => void;
  onEdit: () => void;
  onStatus: (action: 'pause' | 'resume' | 'end') => void;
  onConfirm: (occ: { id: string; version: number }) => void;
  onSkip: (occ: { id: string; version: number }) => void;
}) {
  const occurrences = useOccurrences(rule.id);
  const statusStyles: Record<string, string> = {
    ACTIVE: 'badge bg-soft-success text-success',
    PAUSED: 'badge bg-soft-warning text-warning',
    ENDED: 'badge bg-soft-secondary text-secondary',
  };
  return (
    <div className="card">
      <button type="button" onClick={onToggle} className="d-flex w-100 align-items-center justify-content-between p-3 text-start">
        <div className="text-truncate">
          <div className="d-flex align-items-center gap-2">
            <span className="fw-semibold text-dark">{rule.description || `${rule.type.toLowerCase()} rule`}</span>
            <span className={statusStyles[rule.status]}>{rule.status}</span>
          </div>
          <div className="mt-1 fs-13 text-muted">
            {rule.frequency} · {rule.account_name}
            {rule.category_name ? ` · ${rule.category_name}` : ''}
          </div>
        </div>
        <div className="d-flex align-items-center gap-3">
          <span className={`fw-semibold ${rule.type === 'EXPENSE' ? 'text-danger' : 'text-success'}`}>
            {rule.type === 'EXPENSE' ? '−' : '+'}
            {formatAmountString(rule.amount, currency)}
          </span>
          <span className="text-muted">{expanded ? '▾' : '▸'}</span>
        </div>
      </button>

      {expanded ? (
        <div className="border-top border-light p-3">
          <div className="d-flex flex-wrap gap-2">
            <Button variant="secondary" onClick={onEdit}>Edit</Button>
            {rule.status === 'ACTIVE' ? (
              <Button variant="secondary" onClick={() => onStatus('pause')}>Pause</Button>
            ) : null}
            {rule.status === 'PAUSED' ? (
              <Button variant="secondary" onClick={() => onStatus('resume')}>Resume</Button>
            ) : null}
            {rule.status !== 'ENDED' ? (
              <Button variant="ghost" onClick={() => onStatus('end')}>End</Button>
            ) : null}
          </div>

          <h3 className="mt-3 fs-13 fw-semibold text-secondary">Upcoming occurrences</h3>
          {occurrences.isLoading ? <p className="mt-2 fs-13 text-muted">Loading…</p> : null}
          <ul className="mt-2 list-unstyled d-flex flex-column gap-2 mb-0">
            {(occurrences.data ?? []).map((occ) => (
              <li key={occ.id} className="d-flex align-items-center justify-content-between border rounded p-2 fs-13">
                <div>
                  <span className="fw-medium text-dark">{occ.due_date}</span>
                  <span className="ms-2 fs-12 text-muted">{occ.status.toLowerCase()}</span>
                </div>
                <div className="d-flex gap-2">
                  {occ.status === 'PENDING' ? (
                    <>
                      <Button variant="secondary" onClick={() => onConfirm(occ)}>{rule.type === 'INCOME' ? 'Got it' : 'Pay'}</Button>
                      <Button variant="ghost" onClick={() => onSkip(occ)}>Skip</Button>
                    </>
                  ) : null}
                </div>
              </li>
            ))}
            {occurrences.data && occurrences.data.length === 0 ? (
              <p className="fs-13 text-muted mb-0">No scheduled occurrences.</p>
            ) : null}
          </ul>
        </div>
      ) : null}
    </div>
  );
}