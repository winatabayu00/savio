import { describe, expect, it } from 'vitest';
import { AxiosError } from 'axios';
import { ApiError, errorMessage, toApiError, type ApiErrorBody } from '@/shared/api/client';
import { formatAmountString } from '@/shared/utils/money';

function axiosErrorWith(status: number, body?: unknown): AxiosError<ApiErrorBody> {
  return {
    response: {
      status,
      data: body,
    },
    isAxiosError: true,
    toJSON: () => ({}),
    name: 'AxiosError',
    message: 'Request failed',
  } as unknown as AxiosError<ApiErrorBody>;
}

describe('ApiError mapping', () => {
  it('maps 422 field details', () => {
    const err = toApiError(axiosErrorWith(422, {
      success: false,
      error: { code: 'VALIDATION_ERROR', details: { amount: ['Amount must be positive'] } },
      message: 'Please correct the highlighted fields.',
    }));
    expect(err).toBeInstanceOf(ApiError);
    expect(err.status).toBe(422);
    expect(err.code).toBe('VALIDATION_ERROR');
    expect(err.details?.amount).toEqual(['Amount must be positive']);
  });

  it('maps 409 version conflict for reload UI copy', () => {
    const err = toApiError(axiosErrorWith(409, {
      success: false,
      error: { code: 'VERSION_CONFLICT', details: null },
      message: 'This record changed since you last opened it. Reload the latest version.',
    }));
    expect(err.status).toBe(409);
    expect(err.message).toContain('changed since you last opened it');
  });

  it('does not expose internal network internals', () => {
    const err = toApiError({} as AxiosError<ApiErrorBody>);
    expect(err.status).toBe(0);
    expect(err.message).toContain('Unable to reach the server');
    expect(err.message).not.toContain('AxiosError');
  });

  it('keeps rate-limit status distinct from auth errors', () => {
    const err = toApiError(axiosErrorWith(429, {
      success: false,
      error: { code: 'RATE_LIMITED', details: null },
      message: 'Too many requests',
    }));
    expect(err.status).toBe(429);
    expect(err.code).toBe('RATE_LIMITED');
  });

  it('uses actionable UI copy for version conflicts', () => {
    expect(errorMessage(new ApiError(409, 'VERSION_CONFLICT', 'ignored'))).toContain('Reload');
  });

  it('keeps the server request ID for support correlation', () => {
    const err = toApiError(axiosErrorWith(500, {
      success: false,
      error: { code: 'INTERNAL_ERROR', details: null },
      message: 'An unexpected error occurred',
      request_id: 'req_123',
    }));
    expect(err.requestId).toBe('req_123');
  });

  it('formats amounts and never renders NaN', () => {
    expect(formatAmountString('1500000.00', 'IDR')).toMatch(/Rp\s?1\.500\.000/);
    expect(formatAmountString('not-money', 'IDR')).toBe('—');
  });
});
