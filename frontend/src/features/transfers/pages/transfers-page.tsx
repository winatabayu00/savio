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
import { useTransferMutations, useTransfers } from '@/features/transfers/hooks/use-transfers';
import type { Transfer } from '@/features/transfers/api/transfer.api';

const transferSchema = z.object({
  from_account_id: z.string().min(1, 'Source account is required'),
  to_account_id: z.string().min(1, 'Destination account is required'),
  amount: z.string().refine((v) => v.trim() !== '' && Number(v) > 0, 'Amount must be greater than zero'),
  transfer_date: z.string().min(1, 'Date is required'),
  description: z.string().max(300),
});

export function TransfersPage() {
  const { auth } = useAuth();
  const currency = auth?.workspace.base_currency ?? 'IDR';
  const accounts = useAccounts('ACTIVE');
  const [page, setPage] = useState(1);
  const { data, isLoading, isError } = useTransfers(page, 20);
  const { create, voidT } = useTransferMutations();
  const [formOpen, setFormOpen] = useState(false);
  const [detail, setDetail] = useState<Transfer | undefined>();
  const [voidTx, setVoidTx] = useState<Transfer | undefined>();
  const [notice, setNotice] = useState<string | null>(null);
  const [voidReason, setVoidReason] = useState('');
  const [formError, setFormError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<z.infer<typeof transferSchema>>({
    resolver: zodResolver(transferSchema),
    defaultValues: { from_account_id: '', to_account_id: '', amount: '', transfer_date: new Date().toISOString().slice(0, 10), description: '' },
  });

  const flash = (m: string) => {
    setNotice(m);
    window.setTimeout(() => setNotice(null), 4000);
  };

  const onSubmit = handleSubmit(async (values) => {
    setFormError(null);
    try {
      await create.mutateAsync(values);
      setFormOpen(false);
      reset();
      flash('Transfer created.');
    } catch (err) {
      const apiErr = err as ApiError;
      if (apiErr?.status === 422 && apiErr.details) {
        for (const [field, msgs] of Object.entries(apiErr.details)) {
          setError(field as keyof typeof values, { message: msgs[0] });
        }
      } else {
        setFormError(apiErr?.message ?? 'Could not create transfer.');
      }
    }
  });

  const activeAccounts = accounts.data?.data ?? [];
  const inputCls =
    'w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:outline-none focus:ring-2 focus:ring-brand/30';
  const labelCls = 'mb-1.5 block text-sm font-medium text-gray-700';

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Transfers</h1>
          <p className="mt-1 text-sm text-gray-500">
            Move money between your accounts. Transfers never count as income or expense.
          </p>
        </div>
        <Button onClick={() => setFormOpen(true)}>New Transfer</Button>
      </div>

      {notice ? <div className="mt-4 rounded-lg bg-brand/10 p-3 text-sm text-brand">{notice}</div> : null}
      {isLoading ? <p className="mt-6 text-sm text-gray-500">Loading transfers…</p> : null}
      {isError ? (
        <div className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          We could not load transfers.
        </div>
      ) : null}

      <div className="mt-6">
        {data && data.data.length === 0 ? (
          <EmptyState
            title="No transfers yet"
            description="Transfers move money between your accounts without affecting income or expenses."
            action={<Button onClick={() => setFormOpen(true)}>Create your first transfer</Button>}
          />
        ) : (
          <div className="overflow-hidden rounded-xl border border-gray-200 bg-white">
            <table className="w-full text-left text-sm">
              <thead className="border-b border-gray-200 bg-gray-50 text-xs uppercase tracking-wide text-gray-500">
                <tr>
                  <th className="px-4 py-3 font-medium">Date</th>
                  <th className="px-4 py-3 font-medium">From → To</th>
                  <th className="hidden px-4 py-3 font-medium md:table-cell">Status</th>
                  <th className="px-4 py-3 text-right font-medium">Amount</th>
                </tr>
              </thead>
              <tbody>
                {data?.data.map((t) => (
                  <tr
                    key={t.id}
                    className="cursor-pointer border-b border-gray-100 last:border-0 hover:bg-gray-50"
                    onClick={() => setDetail(t)}
                  >
                    <td className="whitespace-nowrap px-4 py-3 text-gray-500">{t.transfer_date}</td>
                    <td className="px-4 py-3">
                      <span className="font-medium text-gray-900">{t.from_account_name}</span>
                      <span className="text-gray-400"> → </span>
                      <span className="font-medium text-gray-900">{t.to_account_name}</span>
                      {t.description ? (
                        <div className="text-xs text-gray-400">{t.description}</div>
                      ) : null}
                    </td>
                    <td className="hidden px-4 py-3 md:table-cell">
                      {t.status === 'POSTED' ? (
                        <span className="rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">POSTED</span>
                      ) : (
                        <span className="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-500">VOIDED</span>
                      )}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-right font-semibold text-gray-900">
                      {formatAmountString(t.amount, currency)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        {data && data.meta.total_pages > 1 ? (
          <div className="mt-4 flex gap-2">
            <Button variant="secondary" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>Previous</Button>
            <span className="self-center text-sm text-gray-600">Page {page} of {data.meta.total_pages}</span>
            <Button variant="secondary" disabled={page >= data.meta.total_pages} onClick={() => setPage((p) => p + 1)}>Next</Button>
          </div>
        ) : null}
      </div>

      <Modal open={formOpen} onClose={() => setFormOpen(false)} title="New transfer">
        <form onSubmit={onSubmit} className="space-y-4" noValidate>
          <div>
            <label className={labelCls}>From account</label>
            <select className={inputCls} {...register('from_account_id')}>
              <option value="">Select source…</option>
              {activeAccounts.map((a) => (
                <option key={a.id} value={a.id}>{a.name}</option>
              ))}
            </select>
            {errors.from_account_id ? <p className="mt-1 text-xs text-red-600">{errors.from_account_id.message}</p> : null}
          </div>
          <div>
            <label className={labelCls}>To account</label>
            <select className={inputCls} {...register('to_account_id')}>
              <option value="">Select destination…</option>
              {activeAccounts.map((a) => (
                <option key={a.id} value={a.id}>{a.name}</option>
              ))}
            </select>
            {errors.to_account_id ? <p className="mt-1 text-xs text-red-600">{errors.to_account_id.message}</p> : null}
          </div>
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Amount" inputMode="decimal" error={errors.amount?.message} {...register('amount')} />
            <TextField label="Date" type="date" error={errors.transfer_date?.message} {...register('transfer_date')} />
          </div>
          <TextField label="Description (optional)" {...register('description')} />
          {formError ? <p role="alert" className="text-sm text-red-600">{formError}</p> : null}
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={() => setFormOpen(false)}>Cancel</Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Creating…' : 'Create transfer'}
            </Button>
          </div>
        </form>
      </Modal>

      <Modal open={Boolean(detail)} onClose={() => setDetail(undefined)} title="Transfer details">
        {detail ? (
          <div className="space-y-3 text-sm">
            <div className="flex items-center justify-between">
              <span className="font-semibold text-gray-900">
                {detail.from_account_name} → {detail.to_account_name}
              </span>
              <span className="font-bold text-gray-900">{formatAmountString(detail.amount, currency)}</span>
            </div>
            <Row label="Date" value={detail.transfer_date} />
            <Row label="Status" value={detail.status} />
            {detail.description ? <Row label="Description" value={detail.description} /> : null}
            {detail.void_reason ? <Row label="Void reason" value={detail.void_reason} /> : null}
            <div className="flex justify-end pt-2">
              {detail.status === 'POSTED' ? (
                <Button variant="danger" onClick={() => { setVoidTx(detail); setDetail(undefined); }}>
                  Void transfer
                </Button>
              ) : null}
            </div>
          </div>
        ) : null}
      </Modal>

      <Modal open={Boolean(voidTx)} onClose={() => setVoidTx(undefined)} title="Void this transfer?">
        <div className="space-y-4 text-sm text-gray-700">
          <p>
            The transfer stays in history but stops moving balances. To redo the move, void it and create a new transfer.
          </p>
          <TextField label="Reason (optional)" value={voidReason} onChange={(e) => setVoidReason(e.target.value)} />
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setVoidTx(undefined)}>Cancel</Button>
            <Button
              variant="danger"
              onClick={async () => {
                if (!voidTx) return;
                try {
                  await voidT.mutateAsync({ id: voidTx.id, version: voidTx.version, reason: voidReason });
                  setVoidTx(undefined);
                  setVoidReason('');
                  flash('Transfer voided.');
                } catch (err) {
                  flash((err as ApiError)?.message ?? 'Could not void transfer.');
                }
              }}
            >
              Void transfer
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between gap-4 border-b border-gray-100 pb-2">
      <span className="text-gray-500">{label}</span>
      <span className="text-right text-gray-900">{value}</span>
    </div>
  );
}