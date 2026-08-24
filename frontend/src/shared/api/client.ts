import axios, {
  AxiosError,
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from 'axios';

export interface ApiErrorBody {
  success: false;
  error: { code?: string; details?: Record<string, string> | null };
  message?: string;
}

export class ApiError extends Error {
  readonly status: number;
  readonly code?: string;
  readonly details?: Record<string, string> | null;

  constructor(
    status: number,
    code: string | undefined,
    message: string,
    details?: Record<string, string> | null,
  ) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

const CSRF = { token: '', promise: null as Promise<string> | null };

export const api: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
});

export function resetCsrf(): void {
  CSRF.token = '';
  CSRF.promise = null;
}

async function fetchCsrf(): Promise<string> {
  if (CSRF.token) return CSRF.token;
  if (!CSRF.promise) {
    CSRF.promise = api
      .get('/auth/csrf')
      .then((r) => {
        CSRF.token = (r.data?.data?.csrf_token as string) ?? '';
        return CSRF.token;
      })
      .finally(() => {
        CSRF.promise = null;
      });
  }
  return CSRF.promise;
}

// Single-flight refresh: concurrent 401s share one refresh, never more.
let refreshPromise: Promise<boolean> | null = null;
let refreshFailureNotified = false;

const AUTH_PATHS = ['/auth/login', '/auth/register', '/auth/refresh'];

function isAuthPath(url: string): boolean {
  return AUTH_PATHS.some((p) => url.startsWith(p));
}

async function refreshSession(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = api
      .post('/auth/refresh')
      .then(() => {
        refreshFailureNotified = false;
        return true;
      })
      .catch(() => {
        // notify once per consecutive failure streak; retries after a
        // successful refresh re-arm the flag
        if (!refreshFailureNotified) {
          refreshFailureNotified = true;
          notifyUnauthorized();
        }
        return false;
      })
      .finally(() => {
        refreshPromise = null;
      });
  }
  return refreshPromise;
}

export const UNAUTHORIZED_EVENT = 'savio:unauthorized';

// Notify the app that the session can no longer be refreshed. The auth
// provider listens and clears auth + query caches.
export function notifyUnauthorized(): void {
  window.dispatchEvent(new CustomEvent(UNAUTHORIZED_EVENT));
}

api.interceptors.request.use(async (config) => {
  const method = (config.method ?? 'get').toLowerCase();
  if (method === 'post' || method === 'put' || method === 'patch' || method === 'delete') {
    try {
      const token = await fetchCsrf();
      if (token) config.headers.set('X-CSRF-Token', token);
    } catch {
      // server still enforces CSRF; request will fail with a safe error
    }
  }
  return config;
});

interface RetryableRequest extends InternalAxiosRequestConfig {
  _retried?: boolean;
}

api.interceptors.response.use(
  (res) => res,
  async (error: AxiosError<ApiErrorBody>) => {
    const response = error.response;
    const status = response?.status ?? 0;
    const config = error.config as RetryableRequest | undefined;

if (
      status === 401 &&
      config &&
      !config._retried &&
      !isAuthPath(config.url ?? '')
    ) {
      config._retried = true;
      const refreshed = await refreshSession();
      if (refreshed) return api(config);
    }
    throw toApiError(error);
  },
);

export function toApiError(error: AxiosError<ApiErrorBody>): ApiError {
  const status = error.response?.status ?? 0;
  const data = error.response?.data;
  let message = data?.message;
  if (!message) {
    message =
      status === 0
        ? 'Unable to reach the server. Check your connection and try again.'
        : 'Something went wrong. Please try again.';
  }
  return new ApiError(status, data?.error?.code, message, data?.error?.details);
}