import { api } from '@/shared/api/client';
import type { SuccessEnvelope } from '@/shared/api/types';

export type CategoryType = 'INCOME' | 'EXPENSE';

export interface Category {
  id: string;
  workspace_id: string | null;
  name: string;
  type: CategoryType;
  is_system: boolean;
  status: string;
  icon: string | null;
  description: string | null;
  created_at: string;
}

export async function listCategories(type?: CategoryType): Promise<Category[]> {
  const q = type ? `?type=${type}` : '';
  const { data } = await api.get<SuccessEnvelope<Category[]>>(`/categories${q}`);
  return data.data;
}

export async function createCategory(input: { name: string; type: CategoryType }): Promise<Category> {
  const { data } = await api.post<SuccessEnvelope<Category>>('/categories', input);
  return data.data;
}

export async function archiveCategory(id: string): Promise<Category> {
  const { data } = await api.post<SuccessEnvelope<Category>>(`/categories/${id}/archive`);
  return data.data;
}

export async function restoreCategory(id: string): Promise<Category> {
  const { data } = await api.post<SuccessEnvelope<Category>>(`/categories/${id}/restore`);
  return data.data;
}