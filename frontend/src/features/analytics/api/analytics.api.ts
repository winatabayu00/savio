import { useQuery } from '@tanstack/react-query';
import { api } from '@/shared/api/client';
import type { Account } from '@/features/accounts/types/account.types';

export interface CashflowDTO {
  income: string;
  expense: string;
  net: string;
}

export interface DashboardDTO {
  total_balance: number;
  accounts: Account[];
  cashflow: CashflowDTO;
  upcoming: {
    id: string;
    recurring_id: string;
    due_date: string;
    type: string;
    amount: string;
    account_name: string;
    description: string | null;
  }[];
  recent: {
    id: string;
    type: string;
    amount: string;
    transaction_date: string;
    description: string | null;
    account_name: string;
    category_name: string;
    status: string;
  }[];
}

async function fetchDashboard(): Promise<DashboardDTO> {
  const { data } = await api.get<{ success: true; data: DashboardDTO }>('/dashboard');
  return data.data;
}

export function useDashboard() {
  return useQuery({ queryKey: ['dashboard'], queryFn: fetchDashboard });
}

const analyticsKeys = { all: ['analytics'] as const };

export async function fetchCashflow(from: string, to: string): Promise<CashflowDTO> {
  const { data } = await api.get(`/analytics/cashflow?from=${from}&to=${to}`);
  return data.data;
}

export function useCashflow(from: string, to: string) {
  return useQuery({
    queryKey: [...analyticsKeys.all, 'cashflow', from, to],
    queryFn: () => fetchCashflow(from, to),
  });
}

export interface CategoryBreakdownRow {
  category_id: string;
  category_name: string;
  total: string;
  items: number;
}

export async function fetchCategoryBreakdown(from: string, to: string): Promise<CategoryBreakdownRow[]> {
  const { data } = await api.get(`/analytics/categories?from=${from}&to=${to}`);
  return data.data;
}

export function useCategoryBreakdown(from: string, to: string) {
  return useQuery({
    queryKey: [...analyticsKeys.all, 'categories', from, to],
    queryFn: () => fetchCategoryBreakdown(from, to),
  });
}

export interface PeriodComparisonDTO {
  current: CashflowDTO;
  previous: CashflowDTO;
  income_delta_percent: number | null;
  expense_delta_percent: number | null;
}

export async function comparePeriods(from: string, to: string, prevFrom: string, prevTo: string): Promise<PeriodComparisonDTO> {
  const { data } = await api.get(
    `/analytics/period-comparison?from=${from}&to=${to}&compare_from=${prevFrom}&compare_to=${prevTo}`,
  );
  return data.data;
}

export { analyticsKeys };