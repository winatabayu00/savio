import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createTransfer, listTransfers, voidTransfer } from '@/features/transfers/api/transfer.api';

export const transferKeys = {
  all: ['transfers'] as const,
};

export function useTransfers(page = 1, limit = 20) {
  return useQuery({
    queryKey: [...transferKeys.all, page, limit],
    queryFn: () => listTransfers(page, limit),
  });
}

export function useTransferMutations() {
  const qc = useQueryClient();
  const invalidate = () => void qc.invalidateQueries({ queryKey: transferKeys.all });

  const create = useMutation({ mutationFn: createTransfer, onSuccess: invalidate });
  const voidT = useMutation({
    mutationFn: ({ id, version, reason }: { id: string; version: number; reason: string }) =>
      voidTransfer(id, version, reason),
    onSuccess: invalidate,
  });
  return { create, voidT };
}