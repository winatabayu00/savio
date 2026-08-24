export type TransactionType = 'INCOME' | 'EXPENSE' | 'ADJUSTMENT';
export type TransactionStatus = 'DRAFT' | 'POSTED' | 'VOIDED';

export interface Transaction {
  id: string;
  account_id: string;
  category_id: string | null;
  type: TransactionType;
  amount: string;
  transaction_date: string;
  description: string | null;
  merchant: string | null;
  notes: string | null;
  source: string;
  status: TransactionStatus;
  version: number;
  account_name: string;
  category_name: string;
  category_type: string;
  created_by_name: string;
  posted_at: string | null;
  voided_at: string | null;
  void_reason: string | null;
  created_at: string;
  updated_at: string;
}

export interface TransactionInput {
  account_id: string;
  category_id?: string | null;
  type: TransactionType;
  amount: string;
  transaction_date: string;
  description?: string;
  merchant?: string;
  notes?: string;
  status?: 'DRAFT' | 'POSTED';
}

export interface TransactionUpdateInput {
  category_id?: string | null;
  type: TransactionType;
  amount: string;
  transaction_date: string;
  description?: string;
  merchant?: string;
  notes?: string;
  version: number;
}

export interface TransactionFilters {
  search?: string;
  type?: TransactionType;
  account_id?: string;
  category_id?: string;
  status?: TransactionStatus;
  from?: string;
  to?: string;
  page: number;
  limit: number;
}