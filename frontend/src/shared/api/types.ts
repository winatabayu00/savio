export interface User {
  id: string;
  name: string;
  email: string;
  timezone: string;
  default_currency: string;
}

export interface Workspace {
  id: string;
  name: string;
  base_currency: string;
  timezone: string;
}

export interface UserSettings {
  ai_insights_enabled: boolean;
  ai_copilot_enabled: boolean;
  notifications_enabled: boolean;
  budget_warning_threshold: number;
  low_balance_threshold: number | null;
}

export interface AuthState {
  user: User;
  workspace: Workspace;
  role: string;
  settings: UserSettings;
  session_count: number;
}

export interface LoginInput {
  email: string;
  password: string;
}

export interface RegisterInput {
  name: string;
  email: string;
  password: string;
}

export interface SuccessEnvelope<T> {
  success: true;
  data: T;
}