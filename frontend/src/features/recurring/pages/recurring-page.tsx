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

  const inputCls =
    'w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:outline-none focus:ring-2 focus:ring-brand/30';
  const labelCls = 'mb-1.5 block text-sm font-medium text-gray-700';

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Recurring Transactions</h1>
          <p className="mt-1 text-sm text-gray-500">
            Planned income and expenses. You confirm each scheduled occurrence before it becomes actual.
          </p>
        </div>
        <Button onClick={openCreate}>New Recurring</Button>
      </div>

      {notice ? <div className="mt-4 rounded-lg bg-brand/10 p-3 text-sm text-brand">{notice}</div> : null}
      {isLoading ? <p className="mt-6 text-sm text-gray-500">Loading recurring transactions…</p> : null}
      {isError ? (
        <div className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          We could not load recurring transactions.
        </div>
      ) : null}

      <div className="mt-6 space-y-3">
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
        <form onSubmit={onSubmit} className="space-y-4" noValidate>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className={labelCls}>Type</label>
              <select className={inputCls} {...register('type')}>
                <option value="EXPENSE">Expense</option>
                <option value="INCOME">Income</option>
              </select>
            </div>
            <div>
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
            {errors.account_id ? <p className="mt-1 text-xs text-red-600">{errors.account_id.message}</p> : null}
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
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Amount" inputMode="decimal" error={errors.amount?.message} {...register('amount')} />
            <TextField label="Merchant (optional)" {...register('merchant')} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Start date" type="date" error={errors.start_date?.message} {...register('start_date')} />
            <TextField label="End date (optional)" type="date" {...register('end_date')} />
          </div>
          <TextField label="Description (optional)" {...register('description')} />
          {formError ? <p role="alert" className="text-sm text-red-600">{formError}</p> : null}
          <div className="flex justify-end gap-2 pt-2">
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
    ACTIVE: 'bg-green-100 text-green-700',
    PAUSED: 'bg-amber-100 text-amber-700',
    ENDED: 'bg-gray-100 text-gray-500',
  };
  return (
    <div className="rounded-xl border border-gray-200 bg-white">
      <button type="button" onClick={onToggle} className="flex w-full items-center justify-between px-5 py-4 text-left">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-gray-900">{rule.description || `${rule.type.toLowerCase()} rule`}</span>
            <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${statusStyles[rule.status]}`}>{rule.status}</span>
          </div>
          <div className="mt-0.5 text-sm text-gray-500">
            {rule.frequency} · {rule.account_name}
            {rule.category_name ? ` · ${rule.category_name}` : ''}
          </div>
        </div>
        <div className="flex items-center gap-3">
          <span className={`font-semibold ${rule.type === 'EXPENSE' ? 'text-red-600' : 'text-green-600'}`}>
            {rule.type === 'EXPENSE' ? '−' : '+'}
            {formatAmountString(rule.amount, currency)}
          </span>
          <span className="text-gray-400">{expanded ? '▾' : '▸'}</span>
        </div>
      </button>

      {expanded ? (
        <div className="border-t border-gray-100 px-5 py-4">
          <div className="flex flex-wrap gap-2">
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

          <h3 className="mt-4 text-sm font-semibold text-gray-700">Upcoming occurrences</h3>
          {occurrences.isLoading ? <p className="mt-2 text-sm text-gray-500">Loading…</p> : null}
          <ul className="mt-2 space-y-2">
            {(occurrences.data ?? []).map((occ) => (
              <li key={occ.id} className="flex items-center justify-between rounded-lg border border-gray-100 px-4 py-2.5 text-sm">
                <div>
                  <span className="font-medium text-gray-900">{occ.due_date}</span>
                  <span className="ml-2 text-xs text-gray-500">{occ.status.toLowerCase()}</span>
                </div>
                <div className="flex gap-2">
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
              <p className="text-sm text-gray-500">No scheduled occurrences.</p>
            ) : null}
          </ul>
        </div>
      ) : null}
    </div>
  );
}