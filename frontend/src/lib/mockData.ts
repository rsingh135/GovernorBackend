import { Agent, Transaction, ListAgentsResponse, ListTransactionsResponse } from './types';

export const DEMO_KEY = 'governor-demo';

const now = new Date();
const daysAgo = (d: number) => new Date(now.getTime() - d * 86400000).toISOString();
const hoursAgo = (h: number) => new Date(now.getTime() - h * 3600000).toISOString();
const minsAgo = (m: number) => new Date(now.getTime() - m * 60000).toISOString();

export const MOCK_AGENTS: Agent[] = [
  { id: 'agt-001', user_id: 'usr-001', name: 'travel-booking-agent', status: 'active',  api_key_prefix: 'agp_Tr7x', created_at: daysAgo(42) },
  { id: 'agt-002', user_id: 'usr-001', name: 'procurement-agent',    status: 'active',  api_key_prefix: 'agp_Pc2m', created_at: daysAgo(31) },
  { id: 'agt-003', user_id: 'usr-002', name: 'dev-tools-agent',      status: 'active',  api_key_prefix: 'agp_Dv9k', created_at: daysAgo(18) },
  { id: 'agt-004', user_id: 'usr-002', name: 'marketing-spend-agent',status: 'active',  api_key_prefix: 'agp_Mk4r', created_at: daysAgo(9)  },
  { id: 'agt-005', user_id: 'usr-003', name: 'research-agent',       status: 'frozen',  api_key_prefix: 'agp_Rs1q', created_at: daysAgo(5)  },
];

export const MOCK_TRANSACTIONS: Transaction[] = [
  // Today — approved
  { id: 'txn-001', request_id: 'req-001', agent_id: 'agt-001', amount_cents: 48900,  currency: 'usd', vendor: 'expedia',     status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: hoursAgo(1),  approved_at: hoursAgo(1),  approved_by_user_id: 'usr-001' },
  { id: 'txn-002', request_id: 'req-002', agent_id: 'agt-002', amount_cents: 12450,  currency: 'usd', vendor: 'stripe',      status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: hoursAgo(2),  approved_at: hoursAgo(2),  approved_by_user_id: 'usr-001' },
  { id: 'txn-003', request_id: 'req-003', agent_id: 'agt-003', amount_cents: 9900,   currency: 'usd', vendor: 'github',      status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: hoursAgo(3),  approved_at: hoursAgo(3),  approved_by_user_id: 'usr-001' },
  // Today — pending
  { id: 'txn-004', request_id: 'req-004', agent_id: 'agt-001', amount_cents: 189900, currency: 'usd', vendor: 'delta',       status: 'PENDING_APPROVAL', reason: 'requires_approval',         meta: {}, created_at: minsAgo(22) },
  { id: 'txn-005', request_id: 'req-005', agent_id: 'agt-002', amount_cents: 73500,  currency: 'usd', vendor: 'aws',         status: 'PENDING_APPROVAL', reason: 'requires_approval',         meta: {}, created_at: minsAgo(47) },
  { id: 'txn-006', request_id: 'req-006', agent_id: 'agt-004', amount_cents: 54000,  currency: 'usd', vendor: 'google ads',  status: 'PENDING_APPROVAL', reason: 'requires_approval',         meta: {}, created_at: hoursAgo(1) },
  // Today — denied
  { id: 'txn-007', request_id: 'req-007', agent_id: 'agt-005', amount_cents: 25000,  currency: 'usd', vendor: 'unknown-vendor', status: 'DENIED',        reason: 'vendor_not_allowed',        meta: {}, created_at: hoursAgo(4) },
  { id: 'txn-008', request_id: 'req-008', agent_id: 'agt-003', amount_cents: 310000, currency: 'usd', vendor: 'aws',         status: 'DENIED',           reason: 'daily_limit_exceeded',      meta: {}, created_at: hoursAgo(5) },
  // Yesterday
  { id: 'txn-009', request_id: 'req-009', agent_id: 'agt-001', amount_cents: 62400,  currency: 'usd', vendor: 'united',      status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: daysAgo(1), approved_at: daysAgo(1), approved_by_user_id: 'usr-001' },
  { id: 'txn-010', request_id: 'req-010', agent_id: 'agt-002', amount_cents: 8800,   currency: 'usd', vendor: 'notion',      status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: daysAgo(1), approved_at: daysAgo(1), approved_by_user_id: 'usr-001' },
  { id: 'txn-011', request_id: 'req-011', agent_id: 'agt-003', amount_cents: 4900,   currency: 'usd', vendor: 'vercel',      status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: daysAgo(1), approved_at: daysAgo(1), approved_by_user_id: 'usr-001' },
  { id: 'txn-012', request_id: 'req-012', agent_id: 'agt-004', amount_cents: 99000,  currency: 'usd', vendor: 'linkedin ads',status: 'DENIED',           reason: 'daily_limit_exceeded',      meta: {}, created_at: daysAgo(1) },
  // 2 days ago
  { id: 'txn-013', request_id: 'req-013', agent_id: 'agt-001', amount_cents: 37500,  currency: 'usd', vendor: 'marriott',    status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: daysAgo(2), approved_at: daysAgo(2), approved_by_user_id: 'usr-002' },
  { id: 'txn-014', request_id: 'req-014', agent_id: 'agt-002', amount_cents: 15000,  currency: 'usd', vendor: 'stripe',      status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: daysAgo(2), approved_at: daysAgo(2), approved_by_user_id: 'usr-002' },
  { id: 'txn-015', request_id: 'req-015', agent_id: 'agt-005', amount_cents: 200000, currency: 'usd', vendor: 'openai',      status: 'DENIED',           reason: 'agent_frozen',              meta: {}, created_at: daysAgo(2) },
  // 3 days ago
  { id: 'txn-016', request_id: 'req-016', agent_id: 'agt-003', amount_cents: 2900,   currency: 'usd', vendor: 'github',      status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: daysAgo(3), approved_at: daysAgo(3), approved_by_user_id: 'usr-001' },
  { id: 'txn-017', request_id: 'req-017', agent_id: 'agt-001', amount_cents: 112000, currency: 'usd', vendor: 'hertz',       status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: daysAgo(3), approved_at: daysAgo(3), approved_by_user_id: 'usr-001' },
  { id: 'txn-018', request_id: 'req-018', agent_id: 'agt-004', amount_cents: 45000,  currency: 'usd', vendor: 'meta ads',    status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: daysAgo(3), approved_at: daysAgo(3), approved_by_user_id: 'usr-002' },
  { id: 'txn-019', request_id: 'req-019', agent_id: 'agt-002', amount_cents: 5500,   currency: 'usd', vendor: 'zapier',      status: 'DENIED',           reason: 'vendor_not_allowed',        meta: {}, created_at: daysAgo(3) },
  { id: 'txn-020', request_id: 'req-020', agent_id: 'agt-003', amount_cents: 19900,  currency: 'usd', vendor: 'datadog',     status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: daysAgo(4), approved_at: daysAgo(4), approved_by_user_id: 'usr-001' },
];

// Filter mock transactions by the same params the real API accepts
export function filterTransactions(params?: {
  status?: string;
  from_date?: string;
  to_date?: string;
  limit?: number;
  offset?: number;
}): ListTransactionsResponse {
  let txns = [...MOCK_TRANSACTIONS];

  if (params?.status) txns = txns.filter(t => t.status === params.status);

  if (params?.from_date) {
    const from = new Date(params.from_date).getTime();
    txns = txns.filter(t => new Date(t.created_at).getTime() >= from);
  }
  if (params?.to_date) {
    const to = new Date(params.to_date).getTime();
    txns = txns.filter(t => new Date(t.created_at).getTime() <= to);
  }

  const total = txns.length;
  const offset = params?.offset ?? 0;
  const limit = params?.limit ?? 20;
  txns = txns.slice(offset, offset + limit);

  return { transactions: txns, total, limit, offset };
}

export function filterAgents(params?: {
  status?: string;
  limit?: number;
  offset?: number;
}): ListAgentsResponse {
  let agents = [...MOCK_AGENTS];
  if (params?.status) agents = agents.filter(a => a.status === params.status);
  const total = agents.length;
  const offset = params?.offset ?? 0;
  const limit = params?.limit ?? 20;
  agents = agents.slice(offset, offset + limit);
  return { agents, total, limit, offset };
}

export function mockApprove(txnId: string): Transaction {
  const txn = MOCK_TRANSACTIONS.find(t => t.id === txnId);
  if (!txn) throw new Error('Transaction not found');
  txn.status = 'APPROVED';
  txn.approved_at = new Date().toISOString();
  txn.approved_by_user_id = 'usr-001';
  return txn;
}

export function mockDeny(txnId: string): Transaction {
  const txn = MOCK_TRANSACTIONS.find(t => t.id === txnId);
  if (!txn) throw new Error('Transaction not found');
  txn.status = 'DENIED';
  return txn;
}
