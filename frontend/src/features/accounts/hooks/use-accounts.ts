import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  archiveAccount,
  createAccount,
  deleteAccount,
  listAccounts,
  reconcileAccount,
  restoreAccount,
  updateAccount,
} from '@/features/accounts/api/account.api';
import type { AccountInput, AccountUpdateInput } from '@/features/accounts/types/account.types';

export const accountKeys = {
  all: ['accounts'] as const,
  lists: () => [...accountKeys.all, 'list'] as const,
};

export function useAccounts(status?: 'ACTIVE' | 'ARCHIVED') {
  return useQuery({
    queryKey: [...accountKeys.lists(), status ?? 'active'],
    queryFn: () => listAccounts(status),
  });
}

export function useAccountMutations() {
  const qc = useQueryClient();
  const invalidate = () => void qc.invalidateQueries({ queryKey: accountKeys.lists() });

  const create = useMutation({
    mutationFn: (input: AccountInput) => createAccount(input),
    onSuccess: invalidate,
  });
  const update = useMutation({
    mutationFn: ({ id, input }: { id: string; input: AccountUpdateInput }) =>
      updateAccount(id, input),
    onSuccess: invalidate,
  });
  const archive = useMutation({
    mutationFn: (id: string) => archiveAccount(id),
    onSuccess: invalidate,
  });
  const restore = useMutation({
    mutationFn: (id: string) => restoreAccount(id),
    onSuccess: invalidate,
  });
  const remove = useMutation({
    mutationFn: (id: string) => deleteAccount(id),
    onSuccess: invalidate,
  });
  const reconcile = useMutation({
    mutationFn: ({ id, actualBalance, reason }: { id: string; actualBalance: string; reason?: string }) =>
      reconcileAccount(id, actualBalance, reason),
    onSuccess: invalidate,
  });

  return { create, update, archive, restore, remove, reconcile };
}