export interface Agent {
  id: string;
  user_id: string;
  name: string;
  status: 'active' | 'frozen';
  api_key_prefix: string;
  created_at: string;
}

export interface ListAgentsResponse {
  agents: Agent[];
  total: number;
  limit: number;
  offset: number;
}

export interface Transaction {
  id: string;
  request_id: string;
  agent_id: string;
  amount_cents: number;
  currency: string;
  vendor: string;
  status: 'APPROVED' | 'DENIED' | 'PENDING_APPROVAL';
  reason: string;
  meta: Record<string, unknown>;
  created_at: string;
  approved_at?: string;
  approved_by_user_id?: string;
}

export interface ListTransactionsResponse {
  transactions: Transaction[];
  total: number;
  limit: number;
  offset: number;
}

export interface Policy {
  id: string;
  agent_id: string;
  daily_limit_cents: number;
  allowed_vendors: string[];
  require_approval_above_cents: number;
  created_at: string;
  updated_at: string;
}

export interface Settings {
  apiKey: string;
  approverUserId: string;
  apiBaseUrl: string;
}
