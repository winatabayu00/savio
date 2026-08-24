import { afterEach, describe, expect, it } from 'vitest';
import type { ReactNode } from 'react';
import { http, HttpResponse } from 'msw';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider, useAuth } from '@/app/providers/auth-provider';
import { ProtectedRoute } from '@/app/router/protected-route';
import { GuestRoute } from '@/app/router/guest-route';
import { server, AUTH_BASE, mePayload } from '../mocks/handlers';
import { api } from '@/shared/api/client';

api.defaults.baseURL = AUTH_BASE;

function renderWithProviders(ui: ReactNode, queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })) {
  return {
    queryClient,
    ...render(
      <QueryClientProvider client={queryClient}>
        <AuthProvider>{ui}</AuthProvider>
      </QueryClientProvider>,
    ),
  };
}

describe('route guards', () => {
  afterEach(() => server.resetHandlers());

  it('protected route shows loading then content when authenticated', async () => {
    server.use(
      http.get(`${AUTH_BASE}/auth/me`, () =>
        HttpResponse.json({ success: true, data: mePayload }),
      ),
    );
    renderWithProviders(
      <MemoryRouter initialEntries={['/dashboard']}>
        <Routes>
          <Route path="/login" element={<div>login page</div>} />
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <div>secret content</div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText(/loading/i)).toBeInTheDocument();
    await waitFor(() =>
      expect(screen.getByText('secret content')).toBeInTheDocument(),
    );
  });

  it('protected route redirects to login when unauthenticated', async () => {
    server.use(
      http.get(`${AUTH_BASE}/auth/me`, () => new HttpResponse(null, { status: 401 })),
    );
    renderWithProviders(
      <MemoryRouter initialEntries={['/dashboard']}>
        <Routes>
          <Route path="/login" element={<div>login page</div>} />
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <div>secret content</div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </MemoryRouter>,
    );
    await waitFor(() =>
      expect(screen.getByText('login page')).toBeInTheDocument(),
    );
    expect(screen.queryByText('secret content')).not.toBeInTheDocument();
  });

  it('guest route redirects authenticated user away from login', async () => {
    server.use(
      http.get(`${AUTH_BASE}/auth/me`, () =>
        HttpResponse.json({ success: true, data: mePayload }),
      ),
    );
    renderWithProviders(
      <MemoryRouter initialEntries={['/login']}>
        <Routes>
          <Route
            path="/dashboard"
            element={<div>dashboard page</div>}
          />
          <Route
            path="/login"
            element={
              <GuestRoute>
                <div>login form</div>
              </GuestRoute>
            }
          />
        </Routes>
      </MemoryRouter>,
    );
    await waitFor(() =>
      expect(screen.getByText('dashboard page')).toBeInTheDocument(),
    );
    expect(screen.queryByText('login form')).not.toBeInTheDocument();
  });
});

describe('auth provider lifecycle', () => {
  afterEach(() => server.resetHandlers());

  it('logout clears the query cache', async () => {
    server.use(
      http.get(`${AUTH_BASE}/auth/me`, () =>
        HttpResponse.json({ success: true, data: mePayload }),
      ),
      http.post(`${AUTH_BASE}/auth/logout`, () =>
        HttpResponse.json({ success: true, data: {} }),
      ),
      http.get(`${AUTH_BASE}/auth/csrf`, () =>
        HttpResponse.json({ success: true, data: { csrf_token: 't' } }),
      ),
    );

    const { queryClient } = renderWithProviders(
      <AuthHarness />,
    );

    await waitFor(() => expect(screen.getByText(/signed in/i)).toBeInTheDocument());
    queryClient.setQueryData(['financial', 'balance'], 1000);

    await userEvent.click(screen.getByRole('button', { name: /sign out/i }));

    await waitFor(() =>
      expect(screen.getByText(/signed out/i)).toBeInTheDocument(),
    );
    expect(queryClient.getQueryData(['financial', 'balance'])).toBeUndefined();
  });
});

function AuthHarness() {
  const { status, logout } = useAuth();
  const label =
    status === 'AUTHENTICATED'
      ? 'signed in'
      : status === 'UNAUTHENTICATED'
        ? 'signed out'
        : 'unknown';
  return (
    <div>
      <span>{label}</span>
      <button type="button" onClick={() => void logout()}>
        sign out
      </button>
    </div>
  );
}