import type {
  LoginResponse,
  Policy,
  SpendResponse,
  Transaction,
  User,
  Agent,
} from './types';

const API_BASE = import.meta.env.VITE_API_BASE || '/api';

function buildUrl(path: string): string {
  return `${API_BASE}${path}`;
}

async function parseJson<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Request failed with status ${response.status}`);
  }
  return (await response.json()) as T;
}

export async function loginAdmin(email: string, password: string): Promise<LoginResponse> {
  const response = await fetch(buildUrl('/admin/login'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  return parseJson<LoginResponse>(response);
}

export async function getAdminMe(token: string): Promise<{ admin: LoginResponse['admin'] }> {
  const response = await fetch(buildUrl('/admin/me'), {
    headers: { Authorization: `Bearer ${token}` },
  });
  return parseJson<{ admin: LoginResponse['admin'] }>(response);
}

export async function listPending(token: string, limit = 25): Promise<{ transactions: Transaction[] }> {
  const response = await fetch(buildUrl(`/admin/transactions/pending?limit=${limit}`), {
    headers: { Authorization: `Bearer ${token}` },
  });
  return parseJson<{ transactions: Transaction[] }>(response);
}

export async function listTransactions(token: string, limit = 30): Promise<{ transactions: Transaction[] }> {
  const response = await fetch(buildUrl(`/admin/transactions?limit=${limit}`), {
    headers: { Authorization: `Bearer ${token}` },
  });
  return parseJson<{ transactions: Transaction[] }>(response);
}

export async function listUsers(token: string, limit = 20): Promise<{ users: User[] }> {
  const response = await fetch(buildUrl(`/admin/users?limit=${limit}`), {
    headers: { Authorization: `Bearer ${token}` },
  });
  return parseJson<{ users: User[] }>(response);
}

export async function listAgents(token: string, limit = 30): Promise<{ agents: Agent[] }> {
  const response = await fetch(buildUrl(`/admin/agents?limit=${limit}`), {
    headers: { Authorization: `Bearer ${token}` },
  });
  return parseJson<{ agents: Agent[] }>(response);
}

export async function getPolicyForAgent(token: string, agentId: string): Promise<{ policy: Policy }> {
  const response = await fetch(buildUrl(`/admin/policies?agent_id=${agentId}`), {
    headers: { Authorization: `Bearer ${token}` },
  });
  return parseJson<{ policy: Policy }>(response);
}

export async function getAgentHistory(token: string, agentId: string, limit = 10): Promise<{ transactions: Transaction[] }> {
  const response = await fetch(buildUrl(`/admin/agents/${agentId}/history?limit=${limit}`), {
    headers: { Authorization: `Bearer ${token}` },
  });
  return parseJson<{ transactions: Transaction[] }>(response);
}

export async function approveTransaction(token: string, transactionId: string): Promise<{ transaction: Transaction }> {
  const response = await fetch(buildUrl(`/admin/transactions/${transactionId}/approve`), {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  });
  return parseJson<{ transaction: Transaction }>(response);
}

export async function denyTransaction(token: string, transactionId: string): Promise<{ transaction: Transaction }> {
  const response = await fetch(buildUrl(`/admin/transactions/${transactionId}/deny`), {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  });
  return parseJson<{ transaction: Transaction }>(response);
}

export async function simulateSpend(apiKey: string, payload: { request_id: string; amount: number; vendor: string; meta: Record<string, unknown> }): Promise<SpendResponse> {
  const response = await fetch(buildUrl('/spend'), {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      apiKey,
    },
    body: JSON.stringify(payload),
  });
  return parseJson<SpendResponse>(response);
}

export function centsToDollars(cents: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(cents / 100);
}
