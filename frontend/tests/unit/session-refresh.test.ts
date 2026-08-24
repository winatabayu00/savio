import { afterEach, describe, expect, it, vi } from 'vitest';
import { http, HttpResponse } from 'msw';
import {
  api,
  resetCsrf,
  UNAUTHORIZED_EVENT,
} from '@/shared/api/client';
import { server, AUTH_BASE } from '../mocks/handlers';

const base = AUTH_BASE;
api.defaults.baseURL = base;

describe('auth session refresh', () => {
  afterEach(() => {
    server.resetHandlers();
    resetCsrf();
  });

  it('single-flight: concurrent 401s trigger exactly one refresh, each retries once', async () => {
    let refreshCalls = 0;
    let sessionCalls = 0;
    server.use(
      http.get(`${base}/session`, () => {
        sessionCalls++;
        if (sessionCalls <= 3) {
          return new HttpResponse(null, { status: 401 });
        }
        return HttpResponse.json({ success: true, data: { ok: true } });
      }),
      http.post(`${base}/auth/refresh`, async () => {
        refreshCalls++;
        await new Promise((r) => setTimeout(r, 25));
        return HttpResponse.json({ success: true, data: {} });
      }),
      http.get(`${base}/auth/csrf`, () =>
        HttpResponse.json({ success: true, data: { csrf_token: 'x' } }),
      ),
    );

    const results = await Promise.allSettled([
      api.get('/session'),
      api.get('/session'),
      api.get('/session'),
    ]);

    expect(refreshCalls).toBe(1);
    expect(sessionCalls).toBe(6); // 3 original + 3 retried
    for (const r of results) {
      expect(r.status).toBe('fulfilled');
    }
  });

  it('refresh failure emits unauthorized once and does not loop', async () => {
    server.use(
      http.get(`${base}/session`, () => new HttpResponse(null, { status: 401 })),
      http.post(`${base}/auth/refresh`, () => new HttpResponse(null, { status: 401 })),
      http.get(`${base}/auth/csrf`, () =>
        HttpResponse.json({ success: true, data: { csrf_token: 'x' } }),
      ),
    );

    const unauthorized = vi.fn();
    window.addEventListener(UNAUTHORIZED_EVENT, unauthorized);

    let remaining = 5;
    while (remaining > 0) {
      try {
        await api.get('/session');
      } catch {
        /* expected */
      }
      remaining--;
    }

    expect(unauthorized).toHaveBeenCalledTimes(1);
    window.removeEventListener(UNAUTHORIZED_EVENT, unauthorized);
  });
});