import { api, resetCsrf } from '@/shared/api/client';
import type {
  AuthState,
  LoginInput,
  RegisterInput,
  SuccessEnvelope,
} from '@/shared/api/types';

export async function fetchMe(): Promise<AuthState> {
  const { data } = await api.get<SuccessEnvelope<AuthState>>('/auth/me');
  return data.data;
}

export async function login(input: LoginInput): Promise<AuthState> {
  const { data } = await api.post<SuccessEnvelope<AuthState>>('/auth/login', input);
  return data.data;
}

export async function register(input: RegisterInput): Promise<AuthState> {
  const { data } = await api.post<SuccessEnvelope<AuthState>>('/auth/register', input);
  return data.data;
}

export async function logout(): Promise<void> {
  await api.post('/auth/logout').catch(() => undefined);
  resetCsrf();
}

export async function logoutAll(): Promise<void> {
  await api.post('/auth/logout-all').catch(() => undefined);
  resetCsrf();
}