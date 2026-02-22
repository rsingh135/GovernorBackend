export type UUID = string;

export interface Admin {
  id: UUID;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: UUID;
  name: string;
  status: string;
  balance_cents: number;
  created_at: string;
  updated_at: string;
}

export interface Agent {
  id: UUID;
  user_id: UUID;
  name: string;
  status: string;
  api_key_prefix: string;
  created_at: string;
}

export interface Policy {
  id: UUID;
  agent_id: UUID;
  daily_limit_cents: number;
  per_transaction_limit_cents: number;
  allowed_vendors: string[];
  allowed_mccs: string[];
  allowed_weekdays_utc: number[];
  allowed_hours_utc: number[];
  require_approval_above_cents: number;
  purchase_guideline: string;
  created_at: string;
  updated_at: string;
}

export interface Transaction {
  id: UUID;
  request_id: UUID;
  agent_id: UUID;
  amount_cents: number;
  currency: string;
  vendor: string;
  status: string;
  reason: string;
  provider?: string;
  provider_session_id?: string;
  provider_payment_intent_id?: string;
  provider_status?: string;
  provider_checkout_url?: string;
  meta: Record<string, unknown>;
  created_at: string;
}

export interface SpendResponse {
  status: string;
  reason: string;
  checkout_url?: string;
  provider_status?: string;
  transaction_id?: string;
}

export interface LoginResponse {
  token: string;
  expires_at: string;
  admin: Admin;
}
