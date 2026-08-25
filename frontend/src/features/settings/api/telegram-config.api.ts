import { api } from '@/shared/api/client';
import type { SuccessEnvelope, TelegramConfig, TelegramConfigInput } from '@/shared/api/types';

export async function getTelegramConfig(): Promise<TelegramConfig> {
  const { data } = await api.get<SuccessEnvelope<TelegramConfig>>('/telegram/config');
  return data.data;
}

export async function updateTelegramConfig(input: TelegramConfigInput): Promise<TelegramConfig> {
  const { data } = await api.patch<SuccessEnvelope<TelegramConfig>>('/telegram/config', input);
  return data.data;
}