import { useState } from 'react';
import { Button } from '@/shared/components/ui/button';
import { EmptyState } from '@/shared/components/ui/empty-state';
import { Modal } from '@/shared/components/ui/modal';
import { TextField } from '@/shared/components/ui/text-field';
import { formatAmountString } from '@/shared/utils/money';
import { AccountForm } from '@/features/accounts/components/account-form';
import { AccountCard } from '@/features/accounts/components/account-card';
import { useAccountMutations, useAccounts } from '@/features/accounts/hooks/use-accounts';
import type { Account } from '@/features/accounts/types/account.types';
import { ApiError } from '@/shared/api/client';

type Tab = 'ACTIVE' | 'ARCHIVED';

export function AccountsPage() {
  const [tab, setTab] = useState<Tab>('ACTIVE');
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Account | undefined>();
  const [reconciling, setReconciling] = useState<Account | undefined>();
  const [actualBalance, setActualBalance] = useState('');
  const [reconcileError, setReconcileError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const { data, isLoading, isError, refetch } = useAccounts(tab);
  const { archive, restore, reconcile } = useAccountMutations();

  const flash = (message: string) => {
    setNotice(message);
    window.setTimeout(() => setNotice(null), 4000);
  };

  return (
    <div>
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-3">
        <div>
          <h1 className="fs-20 fw-bolder mb-0">Accounts</h1>
          <p className="fs-13 text-muted mb-0">
            Your cash, bank accounts, e-wallets and savings.
          </p>
        </div>
        <Button onClick={() => { setEditing(undefined); setFormOpen(true); }}>
          Add Account
        </Button>
      </div>

      <div className="mt-4 nav nav-tabs border-bottom gap-1">
        {(['ACTIVE', 'ARCHIVED'] as const).map((t) => (
          <button
            key={t}
            type="button"
            onClick={() => setTab(t)}
            className={`nav-link px-4 py-2 fs-13 fw-medium ${
              tab === t ? 'active text-primary' : 'text-muted'
            }`}
          >
            {t === 'ACTIVE' ? 'Active' : 'Archived'}
          </button>
        ))}
      </div>

      {notice ? (
        <div className="mt-4 bg-soft-primary text-primary fs-13 p-3 rounded">{notice}</div>
      ) : null}

      <div className="mt-4">
        {isLoading ? <p className="fs-13 text-muted">Loading accounts…</p> : null}
        {isError ? (
          <div className="alert alert-danger">
            <p>We could not load your accounts.</p>
            <Button variant="secondary" className="mt-2" onClick={() => void refetch()}>
              Try again
            </Button>
          </div>
        ) : null}
        {data && data.data.length === 0 ? (
          <EmptyState
            title={tab === 'ACTIVE' ? 'No accounts yet' : 'No archived accounts'}
            description={
              tab === 'ACTIVE'
                ? 'Add your first account to start tracking your cashflow.'
                : 'Archived accounts appear here and stay visible for history.'
            }
            action={
              tab === 'ACTIVE' ? (
                <Button onClick={() => { setEditing(undefined); setFormOpen(true); }}>
                  Add your first account
                </Button>
              ) : undefined
            }
          />
        ) : null}
        <div className="row g-4">
          {data?.data.map((account) => (
            <div className="col-12 col-md-6 col-lg-4" key={account.id}>
              <AccountCard
                account={account}
              onEdit={(a) => { setEditing(a); setFormOpen(true); }}
              onReconcile={(a) => { setReconciling(a); setActualBalance((a.derived_balance / 100).toFixed(2).replace(/\.?0+$/, '')); setReconcileError(null); }}
              onArchiveRestore={(id, isArchive) =>
                isArchive
                  ? archive.mutateAsync(id).then(() => flash('Account archived.'))
                  : restore.mutateAsync(id).then(() => flash('Account restored.'))
              }
              onChanged={flash}
              />
            </div>
          ))}
        </div>
      </div>

      <Modal open={formOpen} onClose={() => setFormOpen(false)} title={editing ? 'Edit account' : 'Add account'}>
        <AccountForm account={editing} onDone={() => { setFormOpen(false); setEditing(undefined); }} />
      </Modal>

      <Modal open={Boolean(reconciling)} onClose={() => setReconciling(undefined)} title={`Reconcile ${reconciling?.name ?? ''}`}>
        {reconciling ? (
          <div className="d-flex flex-column gap-3 fs-13 text-secondary">
            <p>
              Tracked balance:{' '}
              <span className="fw-semibold text-dark">
                {formatAmountString((reconciling.derived_balance / 100).toFixed(2), reconciling.currency)}
              </span>
              . Enter what the account actually holds to create a signed adjustment.
            </p>
            <TextField
              label={`Actual balance (${reconciling.currency})`}
              inputMode="decimal"
              value={actualBalance}
              onChange={(e) => setActualBalance(e.target.value)}
            />
            {reconcileError ? <p role="alert" className="text-danger fs-13">{reconcileError}</p> : null}
            <div className="d-flex justify-content-end gap-2">
              <Button variant="secondary" onClick={() => setReconciling(undefined)}>Cancel</Button>
              <Button
                onClick={async () => {
                  setReconcileError(null);
                  try {
                    const res = await reconcile.mutateAsync({ id: reconciling.id, actualBalance });
                    setReconciling(undefined);
                    flash(`Reconciled. Adjustment: ${res.difference}.`);
                  } catch (err) {
                    const apiErr = err as ApiError;
                    setReconcileError(apiErr?.message ?? 'Reconciliation failed.');
                  }
                }}
              >
                Reconcile
              </Button>
            </div>
          </div>
        ) : null}
      </Modal>
    </div>
  );
}