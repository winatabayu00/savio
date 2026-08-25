import { api } from '@/shared/api/client';
import type { SuccessEnvelope, UserSettings } from '@/shared/api/types';

export async function getSettings(): Promise<UserSettings> {
  const { data } = await api.get<SuccessEnvelope<UserSettings>>('/settings');
  return data.data;
}

export async function updateSettings(input: Partial<UserSettings>): Promise<UserSettings> {
  const { data } = await api.patch<SuccessEnvelope<UserSettings>>('/settings', input);
  return data.data;
}
