import { api } from '@/shared/api/client';
import type { AIConfig, AIConfigInput, SuccessEnvelope } from '@/shared/api/types';

export async function getAIConfig(): Promise<AIConfig> {
  const { data } = await api.get<SuccessEnvelope<AIConfig>>('/ai/config');
  return data.data;
}

export async function updateAIConfig(input: AIConfigInput): Promise<AIConfig> {
  const { data } = await api.patch<SuccessEnvelope<AIConfig>>('/ai/config', input);
  return data.data;
}