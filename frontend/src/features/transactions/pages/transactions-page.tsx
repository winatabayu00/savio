import { useState } from 'react';
import { useAuth } from '@/app/providers/auth-provider';
import { Button } from '@/shared/components/ui/button';
import { EmptyState } from '@/shared/components/ui/empty-state';
import { Modal } from '@/shared/components/ui/modal';
import { TextField } from '@/shared/components/ui/text-field';
import { formatAmountString } from '@/shared/utils/money';
import { useAccounts } from '@/features/accounts/hooks/use-accounts';
import { useCategories } from '@/features/categories/hooks/use-categories';
import { TransactionFormModal } from '@/features/transactions/components/transaction-form-modal';
import { TransactionTable } from '@/features/transactions/components/transaction-table';
import { useTransactionMutations, useTransactions } from '@/features/transactions/hooks/use-transactions';
import type { Transaction, TransactionFilters } from '@/features/transactions/types/transaction.types';
import { ApiError } from '@/shared/api/client';

export function TransactionsPage() {
  const { auth } = useAuth();
  const currency = auth?.workspace.base_currency ?? 'IDR';
  const accounts = useAccounts('ACTIVE');
  const categories = useCategories();

  const [filters, setFilters] = useState<TransactionFilters>({
    search: '',
    type: undefined,
    account_id: '',
    category_id: '',
    status: undefined,
    from: '',
    to: '',
    page: 1,
    limit: 20,
  });
  const [formOpen, setFormOpen] = useState(false);
  const [editTx, setEditTx] = useState<Transaction | undefined>();
  const [detail, setDetail] = useState<Transaction | undefined>();
  const [voidTx, setVoidTx] = useState<Transaction | undefined>();
  const [notice, setNotice] = useState<string | null>(null);
  const { voidTx: doVoid, post } = useTransactionMutations();

  const { data, isLoading, isError } = useTransactions(filters);

  const flash = (m: string) => {
    setNotice(m);
    window.setTimeout(() => setNotice(null), 4000);
  };

  const patch = (part: Partial<TransactionFilters>) => setFilters((f) => ({ ...f, ...part, page: 1 }));

  const selectCls = 'form-select';
  const [voidReason, setVoidReason] = useState('');

  return (
    <div>
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-3">
        <div>
          <h1 className="fs-20 fw-bolder mb-0">Transactions</h1>
          <p className="fs-13 text-muted mb-0">
            Your income, expenses and adjustments.
          </p>
        </div>
        <Button onClick={() => { setEditTx(undefined); setFormOpen(true); }}>
          Add Transaction
        </Button>
      </div>

      <div className="mt-4 d-flex flex-wrap align-items-center gap-2">
        <input
          placeholder="Search descriptions, merchants…"
          className="form-control"
          style={{ width: '16rem' }}
          value={filters.search}
          onChange={(e) => patch({ search: e.target.value })}
        />
        <select className={selectCls} value={filters.type ?? ''} onChange={(e) => patch({ type: e.target.value as TransactionFilters['type'] | undefined })}>
          <option value="">All types</option>
          <option value="INCOME">Income</option>
          <option value="EXPENSE">Expense</option>
          <option value="ADJUSTMENT">Adjustment</option>
        </select>
        <select className={selectCls} value={filters.status ?? ''} onChange={(e) => patch({ status: e.target.value as TransactionFilters['status'] | undefined })}>
          <option value="">All statuses</option>
          <option value="POSTED">Posted</option>
          <option value="DRAFT">Draft</option>
          <option value="VOIDED">Voided</option>
        </select>
        <select className={selectCls} value={filters.account_id} onChange={(e) => patch({ account_id: e.target.value })}>
          <option value="">All accounts</option>
          {(accounts.data?.data ?? []).map((a) => (
            <option key={a.id} value={a.id}>{a.name}</option>
          ))}
        </select>
        <select className={selectCls} value={filters.category_id} onChange={(e) => patch({ category_id: e.target.value })}>
          <option value="">All categories</option>
          {(categories.data ?? []).map((c) => (
            <option key={c.id} value={c.id}>{c.name}</option>
          ))}
        </select>
        <div className="d-flex align-items-center gap-1 fs-13 text-secondary">
          <input type="date" className="form-control" value={filters.from} onChange={(e) => patch({ from: e.target.value })} />
          to
          <input type="date" className="form-control" value={filters.to} onChange={(e) => patch({ to: e.target.value })} />
        </div>
      </div>

      {notice ? <div className="mt-4 bg-soft-primary text-primary fs-13 p-3 rounded">{notice}</div> : null}
      {isLoading ? <p className="mt-4 fs-13 text-muted">Loading transactions…</p> : null}
      {isError ? (
        <div className="mt-4 alert alert-danger">
          We could not load your transactions. Please try again.
        </div>
      ) : null}

      <div className="mt-4">
        {data && data.data.length === 0 ? (
          <EmptyState
            title="No transactions found"
            description={
              filters.search || filters.type || filters.status || filters.account_id
                ? 'Try clearing a filter or changing the date range.'
                : 'Add your first income or expense to start understanding your cashflow.'
            }
            action={
              !filters.search && !filters.type && !filters.status && !filters.account_id ? (
                <Button onClick={() => setFormOpen(true)}>Add your first transaction</Button>
              ) : undefined
            }
          />
        ) : (
          <TransactionTable transactions={data?.data ?? []} currency={currency} onOpen={setDetail} />
        )}
        {data && data.meta.total_pages > 1 ? (
          <div className="mt-4 d-flex align-items-center justify-content-between fs-13 text-secondary">
            <span>
              Page {data.meta.page} of {data.meta.total_pages} · {data.meta.total} transactions
            </span>
            <div className="d-flex gap-2">
              <Button variant="secondary" disabled={data.meta.page <= 1} onClick={() => setFilters((f) => ({ ...f, page: f.page - 1 }))}>
                Previous
              </Button>
              <Button variant="secondary" disabled={data.meta.page >= data.meta.total_pages} onClick={() => setFilters((f) => ({ ...f, page: f.page + 1 }))}>
                Next
              </Button>
            </div>
          </div>
        ) : null}
      </div>

      <TransactionFormModal
        open={formOpen}
        onClose={() => { setFormOpen(false); setEditTx(undefined); }}
        editTx={editTx}
        onError={flash}
      />

      <Modal open={Boolean(detail)} onClose={() => setDetail(undefined)} title="Transaction details">
        {detail ? (
          <div className="d-flex flex-column gap-3 fs-13">
            <div className="d-flex align-items-center justify-content-between">
              <span className="fw-semibold text-dark">
                {detail.description || detail.merchant || 'Untitled'}
              </span>
              <span className={`fw-bolder ${detail.type === 'EXPENSE' ? 'text-danger' : detail.type === 'INCOME' ? 'text-success' : 'text-secondary'}`}>
                {detail.type === 'EXPENSE' ? '−' : '+'}
                {formatAmountString(detail.amount, currency)}
              </span>
            </div>
            <Row label="Type" value={detail.type} />
            <Row label="Status" value={detail.status} />
            <Row label="Date" value={detail.transaction_date} />
            <Row label="Account" value={detail.account_name} />
            <Row label="Category" value={detail.category_name || '—'} />
            <Row label="Merchant" value={detail.merchant || '—'} />
            {detail.notes ? <Row label="Notes" value={detail.notes} /> : null}
            {detail.void_reason ? <Row label="Void reason" value={detail.void_reason} /> : null}
            <div className="d-flex flex-wrap justify-content-end gap-2 pt-2">
              {detail.status === 'DRAFT' ? (
                <Button variant="secondary" onClick={() => {
                  post.mutateAsync({ id: detail.id, version: detail.version }).then(() => setDetail(undefined))
                    .catch((e) => flash((e as ApiError).message ?? 'Could not post.'));
                }}>
                  Post now
                </Button>
              ) : null}
              <Button variant="secondary" onClick={() => { setEditTx(detail); setDetail(undefined); setFormOpen(true); }}>
                Edit
              </Button>
              {detail.status === 'POSTED' ? (
                <Button variant="danger" onClick={() => { setVoidTx(detail); setDetail(undefined); }}>
                  Void
                </Button>
              ) : null}
            </div>
          </div>
        ) : null}
      </Modal>

      <Modal open={Boolean(voidTx)} onClose={() => setVoidTx(undefined)} title="Void this transaction?">
        <div className="d-flex flex-column gap-3 fs-13 text-secondary">
          <p>
            The transaction remains in history but stops counting toward balances and analytics. To fix an error, void it and create a replacement.
          </p>
          <TextField label="Reason (optional)" value={voidReason} onChange={(e) => setVoidReason(e.target.value)} />
          <div className="d-flex justify-content-end gap-2">
            <Button variant="secondary" onClick={() => setVoidTx(undefined)}>Cancel</Button>
            <Button variant="danger" disabled={!voidTx} onClick={async () => {
              if (!voidTx) return;
              try {
                await doVoid.mutateAsync({ id: voidTx.id, version: voidTx.version, reason: voidReason });
                setVoidTx(undefined);
                setVoidReason('');
                flash('Transaction voided.');
              } catch (err) {
                flash((err as ApiError).message ?? 'Could not void transaction.');
              }
            }}>
              Void transaction
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