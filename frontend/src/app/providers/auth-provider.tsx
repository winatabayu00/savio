import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { UNAUTHORIZED_EVENT } from '@/shared/api/client';
import type { AuthState, LoginInput, RegisterInput } from '@/shared/api/types';
import * as authApi from '@/features/auth/api/auth.api';

export type AuthStatus = 'UNKNOWN' | 'AUTHENTICATED' | 'UNAUTHENTICATED';

interface AuthContextValue {
  status: AuthStatus;
  auth?: AuthState;
  login: (input: LoginInput) => Promise<AuthState>;
  register: (input: RegisterInput) => Promise<AuthState>;
  logout: () => Promise<void>;
  logoutAll: () => Promise<void>;
  refresh: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const [status, setStatus] = useState<AuthStatus>('UNKNOWN');
  const [auth, setAuth] = useState<AuthState | undefined>();

  const clearAll = useCallback(() => {
    setAuth(undefined);
    setStatus('UNAUTHENTICATED');
    queryClient.clear();
  }, [queryClient]);

  useEffect(() => {
    let active = true;
    authApi
      .fetchMe()
      .then((a) => {
        if (!active) return;
        setAuth(a);
        setStatus('AUTHENTICATED');
      })
      .catch(() => {
        if (active) setStatus('UNAUTHENTICATED');
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    const onUnauthorized = (): void => clearAll();
    window.addEventListener(UNAUTHORIZED_EVENT, onUnauthorized);
    return () => window.removeEventListener(UNAUTHORIZED_EVENT, onUnauthorized);
  }, [clearAll]);

  const login = useCallback(async (input: LoginInput) => {
    const a = await authApi.login(input);
    setAuth(a);
    setStatus('AUTHENTICATED');
    return a;
  }, []);

  const register = useCallback(async (input: RegisterInput) => {
    const a = await authApi.register(input);
    setAuth(a);
    setStatus('AUTHENTICATED');
    return a;
  }, []);

  const logout = useCallback(async () => {
    await authApi.logout();
    clearAll();
  }, [clearAll]);

  const logoutAll = useCallback(async () => {
    await authApi.logoutAll();
    clearAll();
  }, [clearAll]);

  const refresh = useCallback(async () => {
    const a = await authApi.fetchMe();
    setAuth(a);
    setStatus('AUTHENTICATED');
  }, []);

  const value = useMemo(
    () => ({ status, auth, login, register, logout, logoutAll, refresh }),
    [status, auth, login, register, logout, logoutAll, refresh],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}