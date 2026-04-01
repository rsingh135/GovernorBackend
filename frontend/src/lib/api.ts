import { ListAgentsResponse, ListTransactionsResponse, Transaction } from './types';

const DEFAULT_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

function getSettings() {
  if (typeof window === 'undefined') return { apiKey: '', approverUserId: '', apiBaseUrl: DEFAULT_BASE_URL };
  try {
    const raw = localStorage.getItem('governor_settings');
    if (!raw) return { apiKey: '', approverUserId: '', apiBaseUrl: DEFAULT_BASE_URL };
    return JSON.parse(raw);
  } catch {
    return { apiKey: '', approverUserId: '', apiBaseUrl: DEFAULT_BASE_URL };
  }
}

async function apiFetch<T>(path: string, options: RequestInit = {}, authenticated = false): Promise<T> {
  const settings = getSettings();
  const base = settings.apiBaseUrl || DEFAULT_BASE_URL;

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> || {}),
  };

  if (authenticated && settings.apiKey) {
    headers['X-API-Key'] = settings.apiKey;
  }

  const res = await fetch(`${base}${path}`, { ...options, headers });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || `HTTP ${res.status}`);
  }
  return res.json();
}

export async function getAgents(params?: { user_id?: string; status?: string; limit?: number; offset?: number }): Promise<ListAgentsResponse> {
  const q = new URLSearchParams();
  if (params?.user_id) q.set('user_id', params.user_id);
  if (params?.status) q.set('status', params.status);
  if (params?.limit != null) q.set('limit', String(params.limit));
  if (params?.offset != null) q.set('offset', String(params.offset));
  const qs = q.toString() ? `?${q}` : '';
  return apiFetch<ListAgentsResponse>(`/agents${qs}`);
}

export async function getTransactions(params?: {
  status?: string;
  from_date?: string;
  to_date?: string;
  limit?: number;
  offset?: number;
}): Promise<ListTransactionsResponse> {
  const q = new URLSearchParams();
  if (params?.status) q.set('status', params.status);
  if (params?.from_date) q.set('from_date', params.from_date);
  if (params?.to_date) q.set('to_date', params.to_date);
  if (params?.limit != null) q.set('limit', String(params.limit));
  if (params?.offset != null) q.set('offset', String(params.offset));
  const qs = q.toString() ? `?${q}` : '';
  return apiFetch<ListTransactionsResponse>(`/transactions${qs}`, {}, true);
}

export async function approveTransaction(txnId: string, approverUserId: string): Promise<Transaction> {
  return apiFetch<Transaction>(`/transactions/${txnId}/approve`, {
    method: 'POST',
    body: JSON.stringify({ approver_user_id: approverUserId }),
  });
}

export async function denyTransaction(txnId: string, approverUserId: string): Promise<Transaction> {
  return apiFetch<Transaction>(`/transactions/${txnId}/deny`, {
    method: 'POST',
    body: JSON.stringify({ approver_user_id: approverUserId }),
  });
}

export async function createAgent(body: { user_id: string; name: string }) {
  return apiFetch('/agents', { method: 'POST', body: JSON.stringify(body) });
}

export async function createPolicy(body: {
  agent_id: string;
  daily_limit_cents: number;
  allowed_vendors: string[];
  require_approval_above_cents: number;
}) {
  return apiFetch('/policies', { method: 'POST', body: JSON.stringify(body) });
}

export function formatCents(cents: number): string {
  return `$${(cents / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

export function getSettings_public() {
  return getSettings();
}
