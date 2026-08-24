import { api } from '@/shared/api/client';
import type { SuccessEnvelope } from '@/shared/api/types';
import type {
  Transaction,
  TransactionFilters,
  TransactionInput,
  TransactionUpdateInput,
} from '@/features/transactions/types/transaction.types';

export interface TransactionListResult {
  data: Transaction[];
  meta: { page: number; limit: number; total: number; total_pages: number };
}

function buildParams(f: TransactionFilters): URLSearchParams {
  const p = new URLSearchParams();
  p.set('page', String(f.page));
  p.set('limit', String(f.limit));
  if (f.search && f.search.trim()) p.set('search', f.search.trim());
  if (f.type) p.set('type', f.type);
  if (f.account_id) p.set('account_id', f.account_id);
  if (f.category_id) p.set('category_id', f.category_id);
  if (f.status) p.set('status', f.status);
  if (f.from) p.set('from', f.from);
  if (f.to) p.set('to', f.to);
  return p;
}

export async function listTransactions(filters: TransactionFilters): Promise<TransactionListResult> {
  const params = buildParams(filters);
  const { data } = await api.get<TransactionListResult>(`/transactions?${params.toString()}`);
  return data;
}

export async function getTransaction(id: string): Promise<Transaction> {
  const { data } = await api.get<SuccessEnvelope<Transaction>>(`/transactions/${id}`);
  return data.data;
}

export async function createTransaction(input: TransactionInput): Promise<Transaction> {
  const { data } = await api.post<SuccessEnvelope<Transaction>>('/transactions', input);
  return data.data;
}

export async function updateTransaction(
  id: string,
  input: TransactionUpdateInput,
): Promise<Transaction> {
  const { data } = await api.patch<SuccessEnvelope<Transaction>>(`/transactions/${id}`, input);
  return data.data;
}

export async function postTransaction(id: string, version: number): Promise<Transaction> {
  const { data } = await api.post<SuccessEnvelope<Transaction>>(`/transactions/${id}/post`, { version });
  return data.data;
}

export async function voidTransaction(
  id: string,
  version: number,
  reason: string,
): Promise<Transaction> {
  const { data } = await api.post<SuccessEnvelope<Transaction>>(`/transactions/${id}/void`, {
    version,
    reason,
  });
  return data.data;
}