import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  closeBudget,
  createBudget,
  listBudgets,
  updateBudget,
} from '@/features/budgets/api/budget.api';
import type { BudgetInput } from '@/features/budgets/api/budget.api';

export const budgetKeys = { all: ['budgets'] as const };

export function useBudgets(status?: 'ACTIVE' | 'CLOSED') {
  return useQuery({
    queryKey: [...budgetKeys.all, status ?? 'all'],
    queryFn: () => listBudgets(status),
  });
}

export function useBudgetMutations() {
  const qc = useQueryClient();
  const invalidate = () => void qc.invalidateQueries({ queryKey: budgetKeys.all, refetchType: 'active' });
  const create = useMutation({ mutationFn: createBudget, onSuccess: invalidate });
  const update = useMutation({
    mutationFn: ({ id, input }: { id: string; input: BudgetInput & { version: number } }) => updateBudget(id, input),
    onSuccess: invalidate,
  });
  const close = useMutation({
    mutationFn: ({ id, version }: { id: string; version: number }) => closeBudget(id, version),
    onSuccess: invalidate,
  });
  return { create, update, close };
}