import { api } from '@/shared/api/client';
import type { SuccessEnvelope } from '@/shared/api/types';

export interface ForecastEvent {
  date: string;
  type: 'KNOWN' | 'SCHEDULED' | 'ESTIMATED' | 'ASSUMED';
  kind: 'INCOME' | 'EXPENSE';
  amount: string;
  description: string;
}

export interface ForecastDTO {
  opening_balance: string;
  ending_balance: string;
  minimum_balance: string;
  minimum_balance_date: string;
  projected_income: string;
  projected_expense: string;
  timeline: { date: string; balance: string }[];
  events: ForecastEvent[];
  confidence: 'LOW' | 'MEDIUM' | 'HIGH';
  assumptions: {
    variable_expense_daily: string;
    baseline_days: number;
    active_recurring_rules: number;
    confidence_basis: string;
  };
  calculation_version: string;
  stale: boolean;
}

export const FORECAST_HORIZONS = [30, 60, 90, 180, 365];

export async function getForecast(horizon: number): Promise<ForecastDTO> {
  const { data } = await api.get<SuccessEnvelope<ForecastDTO>>(`/forecast?horizon=${horizon}`);
  return data.data;
}