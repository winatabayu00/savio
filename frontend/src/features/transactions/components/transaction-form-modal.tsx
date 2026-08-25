import { useEffect, useState } from 'react';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { ApiError } from '@/shared/api/client';
import { Button } from '@/shared/components/ui/button';
import { Modal } from '@/shared/components/ui/modal';
import { TextField } from '@/shared/components/ui/text-field';
import { useAccounts } from '@/features/accounts/hooks/use-accounts';
import { useCategories } from '@/features/categories/hooks/use-categories';
import { useTransactionMutations } from '@/features/transactions/hooks/use-transactions';
import type { Transaction, TransactionType } from '@/features/transactions/types/transaction.types';
import { suggestCategory } from '@/features/ai/api/ai.api';

const formSchema = z.object({
  type: z.enum(['INCOME', 'EXPENSE']),
  account_id: z.string().min(1, 'Account is required'),
  category_id: z.string(),
  amount: z.string().refine(
    (v) => v.trim() !== '' && !Number.isNaN(Number(v)) && Number(v) > 0,
    'Amount must be greater than zero',
  ),
  transaction_date: z.string().min(1, 'Date is required'),
  description: z.string().max(300),
  merchant: z.string().max(200),
  notes: z.string().max(1000),
});

interface Props {
  open: boolean;
  onClose: () => void;
  onError?: (message: string) => void;
  editTx?: Transaction;
}

export function TransactionFormModal({ open, onClose, onError, editTx }: Props) {
  const accounts = useAccounts('ACTIVE');
  const categories = useCategories();
  const { create, update, post } = useTransactionMutations();
  const [formError, setFormError] = useState<string | null>(null);
  const [aiHint, setAiHint] = useState<string | null>(null);
  const [aiBusy, setAiBusy] = useState(false);
  // ponytail: single schema for create/edit flows
  void formSchema;

  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: editTx
      ? {
          type: editTx.type === 'ADJUSTMENT' ? 'EXPENSE' : editTx.type,
          account_id: editTx.account_id,
          category_id: editTx.category_id ?? '',
          amount: editTx.amount,
          transaction_date: editTx.transaction_date,
          description: editTx.description ?? '',
          merchant: editTx.merchant ?? '',
          notes: editTx.notes ?? '',
        }
      : {
          type: 'EXPENSE',
          account_id: '',
          category_id: '',
          amount: '',
          transaction_date: new Date().toISOString().slice(0, 10),
          description: '',
          merchant: '',
          notes: '',
        },
  });

  useEffect(() => {
    setFormError(null);
    if (open && !editTx) reset();
  }, [open, editTx, reset]);

  const txType = (watch('type') ?? 'EXPENSE') as TransactionType;
  const description = watch('description');
  const merchant = watch('merchant');
  const filteredCategories =
    categories.data?.filter((c) => (txType === 'INCOME' ? c.type === 'INCOME' : c.type === 'EXPENSE')) ?? [];

  const onSuggest = async () => {
    setAiBusy(true);
    setAiHint(null);
    try {
      const res = await suggestCategory(description || merchant || '', merchant || '');
      const match = filteredCategories.find(
        (c) => c.name.toLowerCase() === res.category_guess.toLowerCase(),
      );
      if (match) {
        setValue('category_id', match.id);
        setAiHint(`AI suggested: ${res.category_guess}`);
      } else {
        setAiHint(`AI guessed "${res.category_guess}" but it is not one of your categories yet.`);
      }
    } catch (err) {
      setFormError((err as ApiError)?.message ?? 'AI category suggestion is unavailable.');
    } finally {
      setAiBusy(false);
    }
  };

  const onSubmit = handleSubmit(async (values) => {
    setFormError(null);
    const payload = {
      type: values.type,
      account_id: values.account_id,
      category_id: values.category_id || null,
      amount: values.amount,
      transaction_date: values.transaction_date,
      description: values.description,
      merchant: values.merchant,
      notes: values.notes,
    };
    try {
      if (editTx) {
        await update.mutateAsync({ id: editTx.id, input: { ...payload, version: editTx.version } });
        onClose();
      } else {
        const created = await create.mutateAsync({ ...payload, status: 'POSTED' });
        // draft-to-post shortcut for instant effect; DRAFT refinement stays possible via edit
        void created;
        onClose();
      }
    } catch (err) {
      const apiErr = err as ApiError;
      const msg = apiErr?.status === 409 ? 'This record changed. Reload and try again.' : apiErr?.message ?? 'Something went wrong.';
      setFormError(msg);
      onError?.(msg);
    }
  });

  const inputCls = 'form-select';
  const labelCls = 'form-label';

  return (
    <Modal open={open} onClose={onClose} title={editTx ? 'Edit transaction' : 'Add income or expense'}>
      <form onSubmit={onSubmit} noValidate>
        <div className="row g-3 mb-3">
          <div className="col-6">
            <label htmlFor="tx-type" className={labelCls}>Type</label>
            <select id="tx-type" className={inputCls} {...register('type')}>
              <option value="EXPENSE">Expense</option>
              <option value="INCOME">Income</option>
            </select>
          </div>
          <div className="col-6">
            <TextField
              label="Amount"
              inputMode="decimal"
              placeholder="0.00"
              error={errors.amount?.message}
              {...register('amount')}
            />
          </div>
        </div>
        <div className="mb-3">
          <label htmlFor="tx-account" className={labelCls}>Account</label>
          <select id="tx-account" className={inputCls} {...register('account_id')}>
            <option value="">Select account…</option>
            {(accounts.data?.data ?? []).map((a) => (
              <option key={a.id} value={a.id}>{a.name}</option>
            ))}
          </select>
          {errors.account_id ? <p className="text-danger fs-12 mt-1">{errors.account_id.message}</p> : null}
        </div>
        <div className="mb-3">
          <label htmlFor="tx-cat" className={labelCls}>Category</label>
          <select id="tx-cat" className={inputCls} {...register('category_id')}>
            <option value="">No category</option>
            {filteredCategories.map((c) => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
          {errors.category_id ? <p className="text-danger fs-12 mt-1">{errors.category_id.message}</p> : null}
          <button
            type="button"
            onClick={() => void onSuggest()}
            disabled={aiBusy || (!description && !merchant)}
            className="btn btn-link p-0 mt-2 fs-12 fw-medium text-primary"
          >
            {aiBusy ? 'Asking AI…' : 'Suggest category with AI'}
          </button>
          {aiHint ? <p className="text-primary fs-12 mt-1">{aiHint}</p> : null}
        </div>
        <div className="mb-3">
          <TextField label="Date" type="date" error={errors.transaction_date?.message} {...register('transaction_date')} />
        </div>
        <div className="mb-3">
          <TextField label="Description" placeholder="e.g. Grocery shopping" error={errors.description?.message} {...register('description')} />
        </div>
        <div className="row g-3 mb-3">
          <div className="col-6">
            <TextField label="Merchant" placeholder="Optional" {...register('merchant')} />
          </div>
          <div className="col-6">
            <TextField label="Notes" placeholder="Optional" {...register('notes')} />
          </div>
        </div>
        {formError ? <p role="alert" className="text-danger fs-13">{formError}</p> : null}
        {editTx ? (
          <div className="d-flex gap-2 pt-1">
            <Button type="submit" disabled={isSubmitting || editTx.status !== 'DRAFT'} title={editTx.status === 'DRAFT' ? 'Save draft' : 'Posted transactions are immutable'}>
              Save
            </Button>
            {editTx.status === 'DRAFT' ? (
              <Button type="button" variant="secondary" disabled={isSubmitting} onClick={async () => {
                if (editTx.status === 'DRAFT') {
                  await post.mutateAsync({ id: editTx.id, version: editTx.version }).catch((e) => {
                    const apiErr = e as ApiError;
                    setFormError(apiErr?.message ?? 'Could not post.');
                  });
                  onClose();
                }
              }}>
                Post now
              </Button>
            ) : null}
          </div>
        ) : (
          <div className="d-flex justify-content-end gap-2 pt-1">
            <Button type="button" variant="secondary" onClick={onClose}>Cancel</Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Saving…' : 'Add transaction'}
            </Button>
          </div>
        )}
      </form>
    </Modal>
  );
}