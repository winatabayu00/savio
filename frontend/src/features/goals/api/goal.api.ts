import { api } from '@/shared/api/client';
import type { SuccessEnvelope } from '@/shared/api/types';

export interface Goal {
  id: string;
  name: string;
  target_amount: string;
  current_amount: string;
  target_date: string | null;
  priority: 'LOW' | 'MEDIUM' | 'HIGH';
  linked_account_id: string | null;
  status: 'ACTIVE' | 'PAUSED' | 'ACHIEVED' | 'CANCELLED';
  version: number;
  progress_percent: number;
  remaining: string;
  months_remaining: number;
  required_monthly: string;
  estimated_monthly_income: string;
  feasibility: 'ON_TRACK' | 'AT_RISK';
}

export interface GoalInput {
  name: string;
  target_amount: string;
  current_amount: string;
  target_date?: string | null;
  priority?: string;
}

export async function listGoals(status?: string): Promise<Goal[]> {
  const q = status ? `?status=${status}` : '';
  const { data } = await api.get<SuccessEnvelope<Goal[]>>(`/goals${q}`);
  return data.data;
}

export async function createGoal(input: GoalInput): Promise<Goal> {
  const { data } = await api.post<SuccessEnvelope<Goal>>('/goals', input);
  return data.data;
}

export async function updateGoal(id: string, input: GoalInput & { version: number }): Promise<Goal> {
  const { data } = await api.patch<SuccessEnvelope<Goal>>(`/goals/${id}`, input);
  return data.data;
}

export async function setGoalStatus(id: string, action: 'pause' | 'resume' | 'achieve' | 'cancel', version: number): Promise<Goal> {
  const { data } = await api.post<SuccessEnvelope<Goal>>(`/goals/${id}/${action}`, { version });
  return data.data;
}