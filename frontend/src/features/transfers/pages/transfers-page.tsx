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
  const inputCls = 'form-select';
  const labelCls = 'form-label';

  return (
    <div>
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-3">
        <div>
          <h1 className="fs-20 fw-bolder mb-0">Transfers</h1>
          <p className="fs-13 text-muted mb-0">
            Move money between your accounts. Transfers never count as income or expense.
          </p>
        </div>
        <Button onClick={() => setFormOpen(true)}>New Transfer</Button>
      </div>

      {notice ? <div className="mt-4 bg-soft-primary text-primary fs-13 p-3 rounded">{notice}</div> : null}
      {isLoading ? <p className="mt-4 fs-13 text-muted">Loading transfers…</p> : null}
      {isError ? (
        <div className="mt-4 alert alert-danger">
          We could not load transfers.
        </div>
      ) : null}

      <div className="mt-4">
        {data && data.data.length === 0 ? (
          <EmptyState
            title="No transfers yet"
            description="Transfers move money between your accounts without affecting income or expenses."
            action={<Button onClick={() => setFormOpen(true)}>Create your first transfer</Button>}
          />
        ) : (
          <div className="card">
            <div className="table-responsive">
            <table className="table table-hover mb-0">
              <thead className="fs-12 text-uppercase text-muted">
                <tr>
                  <th className="fw-medium">Date</th>
                  <th className="fw-medium">From → To</th>
                  <th className="d-none d-md-table-cell fw-medium">Status</th>
                  <th className="text-end fw-medium">Amount</th>
                </tr>
              </thead>
              <tbody>
                {data?.data.map((t) => (
                  <tr
                    key={t.id}
                    className="border-bottom"
                    style={{ cursor: 'pointer' }}
                    onClick={() => setDetail(t)}
                  >
                    <td className="text-nowrap text-secondary">{t.transfer_date}</td>
                    <td>
                      <span className="fw-medium text-dark">{t.from_account_name}</span>
                      <span className="text-muted"> → </span>
                      <span className="fw-medium text-dark">{t.to_account_name}</span>
                      {t.description ? (
                        <div className="fs-12 text-muted">{t.description}</div>
                      ) : null}
                    </td>
                    <td className="d-none d-md-table-cell">
                      {t.status === 'POSTED' ? (
                        <span className="badge bg-soft-success text-success fw-medium">POSTED</span>
                      ) : (
                        <span className="badge bg-soft-secondary text-secondary fw-medium">VOIDED</span>
                      )}
                    </td>
                    <td className="text-nowrap text-end fw-semibold text-dark">
                      {formatAmountString(t.amount, currency)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            </div>
          </div>
        )}
        {data && data.meta.total_pages > 1 ? (
          <div className="mt-4 d-flex gap-2">
            <Button variant="secondary" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>Previous</Button>
            <span className="align-self-center fs-13 text-secondary">Page {page} of {data.meta.total_pages}</span>
            <Button variant="secondary" disabled={page >= data.meta.total_pages} onClick={() => setPage((p) => p + 1)}>Next</Button>
          </div>
        ) : null}
      </div>

      <Modal open={formOpen} onClose={() => setFormOpen(false)} title="New transfer">
        <form onSubmit={onSubmit} noValidate>
          <div className="mb-3">
            <label className={labelCls}>From account</label>
            <select className={inputCls} {...register('from_account_id')}>
              <option value="">Select source…</option>
              {activeAccounts.map((a) => (
                <option key={a.id} value={a.id}>{a.name}</option>
              ))}
            </select>
            {errors.from_account_id ? <p className="text-danger fs-12 mt-1">{errors.from_account_id.message}</p> : null}
          </div>
          <div className="mb-3">
            <label className={labelCls}>To account</label>
            <select className={inputCls} {...register('to_account_id')}>
              <option value="">Select destination…</option>
              {activeAccounts.map((a) => (
                <option key={a.id} value={a.id}>{a.name}</option>
              ))}
            </select>
            {errors.to_account_id ? <p className="text-danger fs-12 mt-1">{errors.to_account_id.message}</p> : null}
          </div>
          <div className="row g-3 mb-3">
            <div className="col-6">
              <TextField label="Amount" inputMode="decimal" error={errors.amount?.message} {...register('amount')} />
            </div>
            <div className="col-6">
              <TextField label="Date" type="date" error={errors.transfer_date?.message} {...register('transfer_date')} />
            </div>
          </div>
          <div className="mb-3">
            <TextField label="Description (optional)" {...register('description')} />
          </div>
          {formError ? <p role="alert" className="text-danger fs-13">{formError}</p> : null}
          <div className="d-flex justify-content-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={() => setFormOpen(false)}>Cancel</Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Creating…' : 'Create transfer'}
            </Button>
          </div>
        </form>
      </Modal>

      <Modal open={Boolean(detail)} onClose={() => setDetail(undefined)} title="Transfer details">
        {detail ? (
          <div className="d-flex flex-column gap-3 fs-13">
            <div className="d-flex align-items-center justify-content-between">
              <span className="fw-semibold text-dark">
                {detail.from_account_name} → {detail.to_account_name}
              </span>
              <span className="fw-bolder text-dark">{formatAmountString(detail.amount, currency)}</span>
            </div>
            <Row label="Date" value={detail.transfer_date} />
            <Row label="Status" value={detail.status} />
            {detail.description ? <Row label="Description" value={detail.description} /> : null}
            {detail.void_reason ? <Row label="Void reason" value={detail.void_reason} /> : null}
            <div className="d-flex justify-content-end pt-2">
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
        <div className="d-flex flex-column gap-3 fs-13 text-secondary">
          <p>
            The transfer stays in history but stops moving balances. To redo the move, void it and create a new transfer.
          </p>
          <TextField label="Reason (optional)" value={voidReason} onChange={(e) => setVoidReason(e.target.value)} />
          <div className="d-flex justify-content-end gap-2">
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
    <div className="d-flex justify-content-between gap-4 border-bottom pb-2">
      <span className="text-muted">{label}</span>
      <span className="text-end text-dark">{value}</span>
    </div>
  );
}