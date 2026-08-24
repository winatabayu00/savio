export type AccountType = 'CASH' | 'BANK' | 'EWALLET' | 'SAVINGS' | 'OTHER';
export type AccountStatus = 'ACTIVE' | 'ARCHIVED';

export interface Account {
  id: string;
  name: string;
  type: AccountType;
  currency: string;
  opening_balance: number;
  derived_balance: number;
  institution_name: string | null;
  description: string | null;
  status: AccountStatus;
  version: number;
  archived_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface AccountInput {
  name: string;
  type: AccountType;
  currency: string;
  opening_balance: number;
  institution_name?: string;
  description?: string;
}

export interface AccountUpdateInput {
  name: string;
  type: AccountType;
  institution_name?: string;
  description?: string;
  version: number;
}