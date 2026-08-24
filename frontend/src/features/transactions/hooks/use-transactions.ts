import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  createTransaction,
  listTransactions,
  postTransaction,
  updateTransaction,
  voidTransaction,
} from '@/features/transactions/api/transaction.api';
import type {
  TransactionFilters,
  TransactionUpdateInput,
} from '@/features/transactions/types/transaction.types';

export const transactionKeys = {
  all: ['transactions'] as const,
  lists: () => [...transactionKeys.all, 'list'] as const,
};

export function useTransactions(filters: TransactionFilters) {
  return useQuery({
    queryKey: [...transactionKeys.lists(), filters],
    queryFn: () => listTransactions(filters),
    placeholderData: (prev) => prev,
  });
}

export function useTransactionMutations() {
  const qc = useQueryClient();
  const invalidate = () => void qc.invalidateQueries({ queryKey: transactionKeys.lists() });

  const create = useMutation({ mutationFn: createTransaction, onSuccess: invalidate });
  const update = useMutation({
    mutationFn: ({ id, input }: { id: string; input: TransactionUpdateInput }) =>
      updateTransaction(id, input),
    onSuccess: invalidate,
  });
  const post = useMutation({
    mutationFn: ({ id, version }: { id: string; version: number }) => postTransaction(id, version),
    onSuccess: invalidate,
  });
  const voidTx = useMutation({
    mutationFn: ({ id, version, reason }: { id: string; version: number; reason: string }) =>
      voidTransaction(id, version, reason),
    onSuccess: invalidate,
  });
  return { create, update, post, voidTx };
}