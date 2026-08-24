import { api } from '@/shared/api/client';
import type { SuccessEnvelope } from '@/shared/api/types';

export type TransferStatus = 'POSTED' | 'VOIDED';

export interface Transfer {
  id: string;
  from_account_id: string;
  to_account_id: string;
  amount: string;
  transfer_date: string;
  description: string | null;
  status: TransferStatus;
  version: number;
  from_account_name: string;
  to_account_name: string;
  created_at: string;
  voided_at: string | null;
  void_reason: string | null;
}

export interface TransferInput {
  from_account_id: string;
  to_account_id: string;
  amount: string;
  transfer_date: string;
  description?: string;
}

export async function listTransfers(page = 1, limit = 20): Promise<{ data: Transfer[]; meta: { page: number; limit: number; total: number; total_pages: number } }> {
  const { data } = await api.get(`/transfers?page=${page}&limit=${limit}`);
  return data;
}

export async function createTransfer(input: TransferInput): Promise<Transfer> {
  const { data } = await api.post<SuccessEnvelope<Transfer>>('/transfers', input);
  return data.data;
}

export async function voidTransfer(id: string, version: number, reason: string): Promise<Transfer> {
  const { data } = await api.post<SuccessEnvelope<Transfer>>(`/transfers/${id}/void`, { version, reason });
  return data.data;
}