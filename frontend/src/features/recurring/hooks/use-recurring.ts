import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  confirmOccurrence,
  createRecurring,
  listOccurrences,
  listRecurring,
  setRecurringStatus,
  skipOccurrence,
  updateRecurring,
} from '@/features/recurring/api/recurring.api';
import type { RecurringInput } from '@/features/recurring/api/recurring.api';

export const recurringKeys = {
  all: ['recurring'] as const,
  occurrences: (id: string) => [...recurringKeys.all, 'occurrences', id] as const,
};

export function useRecurring() {
  return useQuery({ queryKey: recurringKeys.all, queryFn: listRecurring });
}

export function useOccurrences(recurringId: string) {
  return useQuery({
    queryKey: recurringKeys.occurrences(recurringId),
    queryFn: () => listOccurrences(recurringId),
    enabled: Boolean(recurringId),
  });
}

export function useRecurringMutations() {
  const qc = useQueryClient();
  const invalidate = () => void qc.invalidateQueries({ queryKey: recurringKeys.all });

  const create = useMutation({ mutationFn: createRecurring, onSuccess: invalidate });
  const update = useMutation({
    mutationFn: ({ id, input }: { id: string; input: RecurringInput & { version: number } }) => updateRecurring(id, input),
    onSuccess: invalidate,
  });
  const status = useMutation({
    mutationFn: ({ id, action, version }: { id: string; action: 'pause' | 'resume' | 'end'; version: number }) =>
      setRecurringStatus(id, action, version),
    onSuccess: invalidate,
  });
  const confirm = useMutation({
    mutationFn: ({ id, version }: { id: string; version: number }) => confirmOccurrence(id, version),
    onSuccess: invalidate,
  });
  const skip = useMutation({
    mutationFn: ({ id, version }: { id: string; version: number }) => skipOccurrence(id, version),
    onSuccess: invalidate,
  });
  return { create, update, status, confirm, skip };
}