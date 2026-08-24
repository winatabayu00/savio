import { useState } from 'react';
import { useAuth } from '@/app/providers/auth-provider';
import { formatMoney } from '@/shared/utils/money';
import { Button } from '@/shared/components/ui/button';
import { ApiError } from '@/shared/api/client';
import type { Account } from '@/features/accounts/types/account.types';

interface AccountCardProps {
  account: Account;
  onEdit: (account: Account) => void;
  onReconcile: (account: Account) => void;
  onArchiveRestore: (id: string, archive: boolean) => Promise<void> | void;
  onChanged: (message: string) => void;
}

export function AccountCard({ account, onEdit, onReconcile, onArchiveRestore, onChanged }: AccountCardProps) {
  const { auth } = useAuth();
  const [busy, setBusy] = useState(false);
  const currency = account.currency || auth?.workspace.base_currency || 'IDR';

  const run = async (fn: () => Promise<unknown> | void) => {
    setBusy(true);
    try {
      await fn();
    } catch (err) {
      const apiErr = err as ApiError;
      onChanged(apiErr?.message ?? 'Something went wrong.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
      <div className="flex items-start justify-between">
        <div>
          <h3 className="font-semibold text-gray-900">{account.name}</h3>
          <p className="text-xs text-gray-500">
            {account.type}
            {account.institution_name ? ` · ${account.institution_name}` : ''}
          </p>
        </div>
        <span
          className={`rounded-full px-2 py-0.5 text-xs font-medium ${
            account.status === 'ACTIVE' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'
          }`}
        >
          {account.status}
        </span>
      </div>
      <div className="mt-4">
        <div className="text-2xl font-semibold text-gray-900">
          {formatMoney(account.derived_balance, currency)}
        </div>
        <p className="text-xs text-gray-500">{currency} · derived balance</p>
      </div>
      <div className="mt-4 flex flex-wrap gap-2">
        <Button variant="secondary" onClick={() => onEdit(account)} disabled={busy}>
          Edit
        </Button>
        <Button variant="secondary" onClick={() => onReconcile(account)} disabled={busy}>
          Reconcile
        </Button>
        {account.status === 'ACTIVE' ? (
          <Button
            variant="secondary"
            disabled={busy}
            onClick={() => void run(() => onArchiveRestore(account.id, true))}
          >
            Archive
          </Button>
        ) : (
          <Button
            variant="secondary"
            disabled={busy}
            onClick={() => void run(() => onArchiveRestore(account.id, false))}
          >
            Restore
          </Button>
        )}
      </div>
    </div>
  );
}