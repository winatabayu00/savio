import { formatAmountString } from '@/shared/utils/money';
import type { Transaction } from '@/features/transactions/types/transaction.types';

const typeColor: Record<string, string> = {
  INCOME: 'text-green-600',
  EXPENSE: 'text-red-600',
  ADJUSTMENT: 'text-gray-700',
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
    <div className="overflow-hidden rounded-xl border border-gray-200 bg-white">
      <table className="w-full text-left text-sm">
        <thead className="border-b border-gray-200 bg-gray-50 text-xs uppercase tracking-wide text-gray-500">
          <tr>
            <th className="px-4 py-3 font-medium">Date</th>
            <th className="px-4 py-3 font-medium">Description</th>
            <th className="hidden px-4 py-3 font-medium md:table-cell">Category</th>
            <th className="hidden px-4 py-3 font-medium md:table-cell">Account</th>
            <th className="hidden px-4 py-3 font-medium lg:table-cell">Status</th>
            <th className="px-4 py-3 text-right font-medium">Amount</th>
          </tr>
        </thead>
        <tbody>
          {transactions.map((tx) => (
            <tr
              key={tx.id}
              className="cursor-pointer border-b border-gray-100 last:border-0 hover:bg-gray-50"
              onClick={() => onOpen(tx)}
            >
              <td className="whitespace-nowrap px-4 py-3 text-gray-500">
                <div>{tx.transaction_date}</div>
                <div className="text-xs text-gray-400">{timeLabel(tx) ?? '—'}</div>
              </td>
              <td className="px-4 py-3">
                <div className="font-medium text-gray-900">
                  {tx.description || tx.merchant || 'Untitled'}
                </div>
                {tx.merchant ? (
                  <div className="text-xs text-gray-400">{tx.merchant}</div>
                ) : null}
              </td>
              <td className="hidden px-4 py-3 text-gray-600 md:table-cell">
                {tx.category_name || '—'}
              </td>
              <td className="hidden px-4 py-3 text-gray-600 md:table-cell">{tx.account_name}</td>
              <td className="hidden px-4 py-3 lg:table-cell">
                {tx.status === 'POSTED' ? (
                  <span className="rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">POSTED</span>
                ) : tx.status === 'VOIDED' ? (
                  <span className="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-500">VOIDED</span>
                ) : (
                  <span className="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-medium text-amber-700">DRAFT</span>
                )}
              </td>
              <td
                className={`whitespace-nowrap px-4 py-3 text-right font-semibold ${
                  typeColor[tx.type] ?? 'text-gray-700'
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
  );
}