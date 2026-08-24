import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  archiveCategory,
  createCategory,
  listCategories,
  restoreCategory,
  type CategoryType,
} from '@/features/categories/api/category.api';

export const categoryKeys = {
  all: ['categories'] as const,
};

export function useCategories(type?: CategoryType) {
  return useQuery({
    queryKey: [...categoryKeys.all, type ?? 'all'],
    queryFn: () => listCategories(type),
  });
}

export function useCategoryMutations() {
  const qc = useQueryClient();
  const invalidate = () => void qc.invalidateQueries({ queryKey: categoryKeys.all });

  const create = useMutation({ mutationFn: createCategory, onSuccess: invalidate });
  const archive = useMutation({ mutationFn: (id: string) => archiveCategory(id), onSuccess: invalidate });
  const restore = useMutation({ mutationFn: (id: string) => restoreCategory(id), onSuccess: invalidate });
  return { create, archive, restore };
}