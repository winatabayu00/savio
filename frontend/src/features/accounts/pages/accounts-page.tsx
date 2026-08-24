import { useState } from 'react';
import { Button } from '@/shared/components/ui/button';
import { EmptyState } from '@/shared/components/ui/empty-state';
import { Modal } from '@/shared/components/ui/modal';
import { AccountForm } from '@/features/accounts/components/account-form';
import { AccountCard } from '@/features/accounts/components/account-card';
import { useAccountMutations, useAccounts } from '@/features/accounts/hooks/use-accounts';
import type { Account } from '@/features/accounts/types/account.types';

type Tab = 'ACTIVE' | 'ARCHIVED';

export function AccountsPage() {
  const [tab, setTab] = useState<Tab>('ACTIVE');
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Account | undefined>();
  const [notice, setNotice] = useState<string | null>(null);

  const { data, isLoading, isError, refetch } = useAccounts(tab);
  const { archive, restore } = useAccountMutations();

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
    </div>
  );
}