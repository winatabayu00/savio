import { api } from '@/shared/api/client';
import type { SuccessEnvelope } from '@/shared/api/types';
import type { Account, AccountInput, AccountUpdateInput } from '@/features/accounts/types/account.types';

export interface Paginated<T> {
  data: T[];
  meta: { page: number; limit: number; total: number; total_pages: number };
}

export async function listAccounts(status?: 'ACTIVE' | 'ARCHIVED'): Promise<Paginated<Account>> {
  const q = status ? `?status=${status}` : '';
  const { data } = await api.get<Paginated<Account>>(`/accounts${q}`);
  return data;
}

export async function getAccount(id: string): Promise<Account> {
  const { data } = await api.get<SuccessEnvelope<Account>>(`/accounts/${id}`);
  return data.data;
}

export async function createAccount(input: AccountInput): Promise<Account> {
  const { data } = await api.post<SuccessEnvelope<Account>>('/accounts', input);
  return data.data;
}

export async function updateAccount(id: string, input: AccountUpdateInput): Promise<Account> {
  const { data } = await api.patch<SuccessEnvelope<Account>>(`/accounts/${id}`, input);
  return data.data;
}

export async function archiveAccount(id: string): Promise<Account> {
  const { data } = await api.post<SuccessEnvelope<Account>>(`/accounts/${id}/archive`);
  return data.data;
}

export async function restoreAccount(id: string): Promise<Account> {
  const { data } = await api.post<SuccessEnvelope<Account>>(`/accounts/${id}/restore`);
  return data.data;
}

export async function deleteAccount(id: string): Promise<void> {
  await api.delete(`/accounts/${id}`);
}