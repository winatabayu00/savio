import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createGoal, listGoals, setGoalStatus, updateGoal } from '@/features/goals/api/goal.api';
import type { GoalInput } from '@/features/goals/api/goal.api';

export const goalKeys = { all: ['goals'] as const };

export function useGoals() {
  return useQuery({ queryKey: goalKeys.all, queryFn: () => listGoals() });
}

export function useGoalMutations() {
  const qc = useQueryClient();
  const invalidate = () => void qc.invalidateQueries({ queryKey: goalKeys.all });
  const create = useMutation({ mutationFn: createGoal, onSuccess: invalidate });
  const update = useMutation({
    mutationFn: ({ id, input }: { id: string; input: GoalInput & { version: number } }) => updateGoal(id, input),
    onSuccess: invalidate,
  });
  const status = useMutation({
    mutationFn: ({ id, action, version }: { id: string; action: 'pause' | 'resume' | 'achieve' | 'cancel'; version: number }) =>
      setGoalStatus(id, action, version),
    onSuccess: invalidate,
  });
  return { create, update, status };
}