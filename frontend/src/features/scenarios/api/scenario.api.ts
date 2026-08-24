import { api } from '@/shared/api/client';
import type { SuccessEnvelope } from '@/shared/api/types';

export type ModType =
  | 'ONE_TIME_EXPENSE'
  | 'ONE_TIME_INCOME'
  | 'RECURRING_EXPENSE'
  | 'RECURRING_INCOME'
  | 'INCOME_REDUCTION'
  | 'INCOME_REMOVAL'
  | 'EXPENSE_REDUCTION';

export interface ScenarioResult {
  baseline_ending_balance: string;
  scenario_ending_balance: string;
  baseline_minimum_balance: string;
  scenario_minimum_balance: string;
  baseline_income: string;
  scenario_income: string;
  baseline_expense: string;
  scenario_expense: string;
  cashflow_difference: string;
  modified_events: number;
  assumption_note: string;
  calculation_version: string;
  timeline: { date: string; baseline_balance: string; scenario_balance: string }[];
}

export interface ScenarioMod {
  id: string;
  type: ModType;
  amount: string;
  frequency: string | null;
  narrative: string | null;
  version: number;
  updated_at: string;
}

export interface Scenario {
  id: string;
  name: string;
  description: string | null;
  status: 'DRAFT' | 'CALCULATED';
  is_stale: boolean;
  version: number;
  calculation_version: string;
  modifications: ScenarioMod[];
  result?: ScenarioResult;
  created_at: string;
  updated_at: string;
}

export const MOD_LABELS: Record<ModType, string> = {
  ONE_TIME_EXPENSE: 'One-time expense',
  ONE_TIME_INCOME: 'One-time income',
  RECURRING_EXPENSE: 'Recurring expense',
  RECURRING_INCOME: 'Recurring income',
  INCOME_REDUCTION: 'Reduce income',
  INCOME_REMOVAL: 'Remove income',
  EXPENSE_REDUCTION: 'Reduce expense',
};

export async function listScenarios(): Promise<Scenario[]> {
  const { data } = await api.get<SuccessEnvelope<Scenario[]>>('/scenarios');
  return data.data;
}

export async function createScenario(name: string): Promise<Scenario> {
  const { data } = await api.post<SuccessEnvelope<Scenario>>('/scenarios', { name });
  return data.data;
}

export async function deleteScenario(id: string): Promise<void> {
  await api.delete(`/scenarios/${id}`);
}

export async function calculateScenario(id: string, horizon = 90): Promise<Scenario> {
  const { data } = await api.post<SuccessEnvelope<Scenario>>(`/scenarios/${id}/calculate?horizon=${horizon}`);
  return data.data;
}

export async function addModification(id: string, input: { type: ModType; amount: string; frequency?: string; narrative?: string }): Promise<ScenarioMod> {
  const { data } = await api.post<SuccessEnvelope<ScenarioMod>>(`/scenarios/${id}/modifications`, input);
  return data.data;
}

export async function removeModification(scenarioId: string, modId: string): Promise<void> {
  await api.delete(`/scenarios/${scenarioId}/modifications/${modId}`);
}