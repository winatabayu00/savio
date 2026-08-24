import { api } from '@/shared/api/client';
import type { SuccessEnvelope } from '@/shared/api/types';

export type RecurringFrequency = 'DAILY' | 'WEEKLY' | 'MONTHLY' | 'MONTH_END';
export type RecurringStatus = 'ACTIVE' | 'PAUSED' | 'ENDED';

export interface RecurringRule {
  id: string;
  account_id: string;
  category_id: string | null;
  type: 'INCOME' | 'EXPENSE';
  amount: string;
  frequency: RecurringFrequency;
  start_date: string;
  end_date: string | null;
  description: string | null;
  merchant: string | null;
  notes: string | null;
  status: RecurringStatus;
  auto_post: boolean;
  version: number;
  account_name: string;
  category_name: string;
  created_at: string;
  updated_at: string;
}

export interface RecurringInput {
  account_id: string;
  category_id?: string | null;
  type: 'INCOME' | 'EXPENSE';
  amount: string;
  frequency: RecurringFrequency;
  start_date: string;
  end_date?: string | null;
  description?: string;
  merchant?: string;
  notes?: string;
  auto_post?: boolean;
}

export interface RecurringOccurrence {
  id: string;
  recurring_id: string;
  due_date: string;
  status: 'PENDING' | 'CONFIRMED' | 'SKIPPED' | 'FAILED';
  version: number;
  posted_transaction_id: string | null;
  recurring_type: string;
  recurring_amount: string;
  recurring_account: string;
}

export async function listRecurring(): Promise<RecurringRule[]> {
  const { data } = await api.get<SuccessEnvelope<RecurringRule[]>>('/recurring-transactions');
  return data.data;
}

export async function createRecurring(input: RecurringInput): Promise<RecurringRule> {
  const { data } = await api.post<SuccessEnvelope<RecurringRule>>('/recurring-transactions', input);
  return data.data;
}

export async function updateRecurring(id: string, input: RecurringInput & { version: number }): Promise<RecurringRule> {
  const { data } = await api.patch<SuccessEnvelope<RecurringRule>>(`/recurring-transactions/${id}`, input);
  return data.data;
}

export async function setRecurringStatus(id: string, action: 'pause' | 'resume' | 'end', version: number): Promise<RecurringRule> {
  const { data } = await api.post<SuccessEnvelope<RecurringRule>>(`/recurring-transactions/${id}/${action}`, { version });
  return data.data;
}

export async function listOccurrences(recurringId: string): Promise<RecurringOccurrence[]> {
  const { data } = await api.get<SuccessEnvelope<RecurringOccurrence[]>>(`/recurring-transactions/${recurringId}/occurrences?limit=100`);
  return data.data;
}

export async function confirmOccurrence(id: string, version: number): Promise<RecurringOccurrence> {
  const { data } = await api.post<SuccessEnvelope<RecurringOccurrence>>(`/recurring-occurrences/${id}/confirm`, { version });
  return data.data;
}

export async function skipOccurrence(id: string, version: number): Promise<RecurringOccurrence> {
  const { data } = await api.post<SuccessEnvelope<RecurringOccurrence>>(`/recurring-occurrences/${id}/skip`, { version });
  return data.data;
}