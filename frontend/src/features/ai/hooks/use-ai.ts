import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { askCopilot, getInsight, aiStatus } from '@/features/ai/api/ai.api';

export const aiKeys = { all: ['ai'] as const };

export function useAIStatus() {
  return useQuery({ queryKey: [...aiKeys.all, 'status'], queryFn: aiStatus, retry: 1, staleTime: 60_000 });
}

export function useInsight() {
  return useQuery({ queryKey: [...aiKeys.all, 'insight'], queryFn: getInsight, retry: 1, staleTime: 60_000 });
}

export function useCopilot() {
  const qc = useQueryClient();
  const ask = useMutation({ mutationFn: (question: string) => askCopilot(question) });
  return { ask, clear: () => qc.removeQueries({ queryKey: aiKeys.all }) };
}