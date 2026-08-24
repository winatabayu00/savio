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
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Accounts</h1>
          <p className="mt-1 text-sm text-gray-500">
            Your cash, bank accounts, e-wallets and savings.
          </p>
        </div>
        <Button onClick={() => { setEditing(undefined); setFormOpen(true); }}>
          Add Account
        </Button>
      </div>

      <div className="mt-4 flex gap-1 border-b border-gray-200">
        {(['ACTIVE', 'ARCHIVED'] as const).map((t) => (
          <button
            key={t}
            type="button"
            onClick={() => setTab(t)}
            className={`rounded-t-lg px-4 py-2 text-sm font-medium ${
              tab === t ? 'border-b-2 border-brand text-brand' : 'text-gray-500 hover:text-gray-700'
            }`}
          >
            {t === 'ACTIVE' ? 'Active' : 'Archived'}
          </button>
        ))}
      </div>

      {notice ? (
        <div className="mt-4 rounded-lg bg-brand/10 p-3 text-sm text-brand">{notice}</div>
      ) : null}

      <div className="mt-6">
        {isLoading ? <p className="text-sm text-gray-500">Loading accounts…</p> : null}
        {isError ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
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
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {data?.data.map((account) => (
            <AccountCard
              key={account.id}
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
          ))}
        </div>
      </div>

      <Modal open={formOpen} onClose={() => setFormOpen(false)} title={editing ? 'Edit account' : 'Add account'}>
        <AccountForm account={editing} onDone={() => { setFormOpen(false); setEditing(undefined); }} />
      </Modal>

      <Modal open={Boolean(reconciling)} onClose={() => setReconciling(undefined)} title={`Reconcile ${reconciling?.name ?? ''}`}>
        {reconciling ? (
          <div className="space-y-4 text-sm text-gray-700">
            <p>
              Tracked balance:{' '}
              <span className="font-semibold text-gray-900">
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
            {reconcileError ? <p role="alert" className="text-sm text-red-600">{reconcileError}</p> : null}
            <div className="flex justify-end gap-2">
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