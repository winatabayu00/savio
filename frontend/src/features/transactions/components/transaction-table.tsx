import { formatAmountString } from '@/shared/utils/money';
import type { Transaction } from '@/features/transactions/types/transaction.types';

const typeColor: Record<string, string> = {
  INCOME: 'text-success',
  EXPENSE: 'text-danger',
  ADJUSTMENT: 'text-secondary',
};

function timeLabel(tx: Transaction): string | null {
  const raw = tx.posted_at ?? tx.created_at;
  if (!raw) return null;
  const d = new Date(raw);
  if (Number.isNaN(d.getTime())) return null;
  return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
}

interface Props {
  transactions: Transaction[];
  currency: string;
  onOpen: (tx: Transaction) => void;
}

export function TransactionTable({ transactions, currency, onOpen }: Props) {
  if (transactions.length === 0) return null;
  return (
    <div className="card">
      <div className="table-responsive">
        <table className="table table-hover mb-0">
        <thead className="fs-12 text-uppercase text-muted">
          <tr>
            <th className="fw-medium">Date</th>
            <th className="fw-medium">Description</th>
            <th className="d-none d-md-table-cell fw-medium">Category</th>
            <th className="d-none d-md-table-cell fw-medium">Account</th>
            <th className="d-none d-lg-table-cell fw-medium">Status</th>
            <th className="text-end fw-medium">Amount</th>
          </tr>
        </thead>
        <tbody>
          {transactions.map((tx) => (
            <tr
              key={tx.id}
              className="border-bottom"
              style={{ cursor: 'pointer' }}
              onClick={() => onOpen(tx)}
            >
              <td className="text-nowrap text-secondary">
                <div>{tx.transaction_date}</div>
                <div className="fs-12 text-muted">{timeLabel(tx) ?? '—'}</div>
              </td>
              <td>
                <div className="fw-medium text-dark">
                  {tx.description || tx.merchant || 'Untitled'}
                </div>
                {tx.merchant ? (
                  <div className="fs-12 text-muted">{tx.merchant}</div>
                ) : null}
              </td>
              <td className="d-none d-md-table-cell text-secondary">
                {tx.category_name || '—'}
              </td>
              <td className="d-none d-md-table-cell text-secondary">{tx.account_name}</td>
              <td className="d-none d-lg-table-cell">
                {tx.status === 'POSTED' ? (
                  <span className="badge bg-soft-success text-success fw-medium">POSTED</span>
                ) : tx.status === 'VOIDED' ? (
                  <span className="badge bg-soft-secondary text-secondary fw-medium">VOIDED</span>
                ) : (
                  <span className="badge bg-soft-warning text-warning fw-medium">DRAFT</span>
                )}
              </td>
              <td
                className={`text-nowrap text-end fw-semibold ${
                  typeColor[tx.type] ?? 'text-secondary'
                }`}
              >
                {tx.type === 'EXPENSE' ? '−' : '+'}
                {formatAmountString(tx.amount, currency)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      </div>
    </div>
  );
}