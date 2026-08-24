import { api } from '@/shared/api/client';
import type { SuccessEnvelope } from '@/shared/api/types';

export interface InsightDTO {
  headline: string;
  detail: string;
  signal: string;
  related_facts: string[];
}

export interface CopilotFact {
  tool: string;
  label: string;
  value: string;
}

export interface CopilotDTO {
  answer: string;
  facts: CopilotFact[];
  tool_used: string;
  sources: string[];
  actions: string[];
  clarification?: string;
}

function monthBounds(offset = 0): { from: string; to: string } {
  const now = new Date();
  const pad = (n: number) => String(n).padStart(2, '0');
  const first = new Date(now.getFullYear(), now.getMonth() + offset, 1);
  const last = new Date(now.getFullYear(), now.getMonth() + offset + 1, 0);
  return {
    from: `${first.getFullYear()}-${pad(first.getMonth() + 1)}-01`,
    to: `${last.getFullYear()}-${pad(last.getMonth() + 1)}-${pad(last.getDate())}`,
  };
}

export async function getInsight(): Promise<InsightDTO> {
  const cur = monthBounds(0);
  const prev = monthBounds(-1);
  const { data } = await api.post<SuccessEnvelope<InsightDTO>>('/ai/insight', {
    from: cur.from,
    to: cur.to,
    compare_from: prev.from,
    compare_to: prev.to,
  });
  return data.data;
}

export async function askCopilot(question: string): Promise<CopilotDTO> {
  const { data } = await api.post<SuccessEnvelope<CopilotDTO>>('/ai/copilot', { question, horizon: 90 });
  return data.data;
}

export interface CategorizeDTO {
  category_guess: string;
  confidence: number;
  matched_rule: string;
}

export async function suggestCategory(description: string, merchant = ''): Promise<CategorizeDTO> {
  const { data } = await api.post<SuccessEnvelope<CategorizeDTO>>('/ai/categorize', { description, merchant });
  return data.data;
}

export async function aiStatus(): Promise<{ enabled: boolean; state: string }> {
  const { data } = await api.get<SuccessEnvelope<{ enabled: boolean; state: string }>>('/ai/status');
  return data.data;
}