import axios, {
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from 'axios';

export interface ApiErrorBody {
  success: false;
  error: { code?: string; details?: Record<string, string[]> | null };
  message?: string;
  request_id?: string;
}

export class ApiError extends Error {
  readonly status: number;
  readonly code?: string;
  readonly details?: Record<string, string[]> | null;
  readonly requestId?: string;

  constructor(
    status: number,
    code: string | undefined,
    message: string,
    details?: Record<string, string[]> | null,
    requestId?: string,
  ) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.details = details;
    this.requestId = requestId;
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
  async (error: unknown) => {
    if (!axios.isAxiosError<ApiErrorBody>(error)) {
      throw error;
    }
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

export function toApiError(error: unknown): ApiError {
  const response = (error as { response?: { status?: number; data?: ApiErrorBody } } | null)?.response;
  const status = response?.status ?? 0;
  const data = response?.data;
  let message = data?.message;
  if (!message) {
    message =
      status === 0
        ? 'Unable to reach the server. Check your connection and try again.'
        : 'Something went wrong. Please try again.';
  }
  return new ApiError(status, data?.error?.code, message, data?.error?.details, data?.request_id);
}

export function errorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.status === 401) return 'Your session has expired. Please sign in again.';
    if (error.status === 403) return 'You do not have permission to perform this action.';
    if (error.status === 404) return 'The requested record is no longer available.';
    if (error.status === 409) return 'This record changed. Reload the latest data, then try again.';
    // Form fields retain the server-provided validation message and details.
    if (error.status === 422) return error.message || 'Please correct the highlighted fields.';
    if (error.status === 429) return 'Too many requests. Please wait a moment before trying again.';
    if (error.status >= 500) return 'The server could not complete your request. Your data is safe. Try again shortly.';
    return error.message;
  }
  return 'Something went wrong. Please try again.';
}
