import { api } from '@/shared/api/client';
import type { SuccessEnvelope } from '@/shared/api/types';

export interface Budget {
  id: string;
  category_id: string;
  category_name: string;
  amount: string;
  period_start: string;
  period_end: string;
  status: 'ACTIVE' | 'CLOSED';
  version: number;
  spent: string;
  remaining: string;
  utilization_percent: number;
  computed_status: 'ON_TRACK' | 'WARNING' | 'EXCEEDED';
  projected_spend: string;
  projected_overspend: string | null;
}

export interface BudgetInput {
  category_id: string;
  amount: string;
  period_start: string;
  period_end: string;
}

export async function listBudgets(status?: 'ACTIVE' | 'CLOSED'): Promise<Budget[]> {
  const q = status ? `?status=${status}` : '';
  const { data } = await api.get<SuccessEnvelope<Budget[]>>(`/budgets${q}`);
  return data.data;
}

export async function createBudget(input: BudgetInput): Promise<Budget> {
  const { data } = await api.post<SuccessEnvelope<Budget>>('/budgets', input);
  return data.data;
}

export async function updateBudget(id: string, input: BudgetInput & { version: number }): Promise<Budget> {
  const { data } = await api.patch<SuccessEnvelope<Budget>>(`/budgets/${id}`, input);
  return data.data;
}

export async function closeBudget(id: string, version: number): Promise<Budget> {
  const { data } = await api.post<SuccessEnvelope<Budget>>(`/budgets/${id}/close`, { version });
  return data.data;
}