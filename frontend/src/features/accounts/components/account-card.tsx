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
    <div className="card h-100 shadow-sm">
      <div className="card-body">
        <div className="d-flex align-items-start justify-content-between">
          <div>
            <h6 className="fw-semibold text-dark mb-0">{account.name}</h6>
            <p className="fs-12 text-muted mb-0">
              {account.type}
              {account.institution_name ? ` · ${account.institution_name}` : ''}
            </p>
          </div>
          <span
            className={`badge ${
              account.status === 'ACTIVE' ? 'bg-soft-success text-success' : 'bg-soft-secondary text-secondary'
            } fw-medium`}
          >
            {account.status}
          </span>
        </div>
        <div className="mt-3">
          <div className="fs-22 fw-semibold text-dark">
            {formatMoney(account.derived_balance, currency)}
          </div>
          <p className="fs-12 text-muted mb-0">{currency} · derived balance</p>
        </div>
        <div className="mt-3 d-flex flex-wrap gap-2">
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
    </div>
  );
}