import { api } from '@/shared/api/client';

export interface AuditLog {
  id: string;
  actor_user_id: string | null;
  actor_name: string | null;
  actor_type: 'USER' | 'AI' | 'SYSTEM';
  action: string;
  resource_type: string;
  resource_id: string | null;
  reason: string | null;
  before_data: Record<string, unknown> | null;
  after_data: Record<string, unknown> | null;
  occurred_at: string;
}

export interface AuditLogPage {
  data: AuditLog[];
  meta: { page: number; limit: number; total: number };
}

export async function listAuditLogs(page = 1, limit = 20): Promise<AuditLogPage> {
  const { data } = await api.get<AuditLogPage>(`/audit-logs?page=${page}&limit=${limit}`);
  return data;
}
