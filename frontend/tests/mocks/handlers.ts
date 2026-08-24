import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';

export const server = setupServer();

export const AUTH_BASE = 'http://localhost/api/v1';

export const mePayload = {
  user: {
    id: 'u1',
    name: 'Adi',
    email: 'adi@savio.test',
    timezone: 'Asia/Jakarta',
    default_currency: 'IDR',
  },
  workspace: {
    id: 'w1',
    name: 'Adi WS',
    base_currency: 'IDR',
    timezone: 'Asia/Jakarta',
  },
  role: 'OWNER',
  settings: {
    ai_insights_enabled: true,
    ai_copilot_enabled: true,
    notifications_enabled: true,
    budget_warning_threshold: 80,
    low_balance_threshold: null,
  },
  session_count: 1,
};

export const csrfHandler = http.get(`${AUTH_BASE}/auth/csrf`, () =>
  HttpResponse.json({ success: true, data: { csrf_token: 'csrf.test' } }),
);