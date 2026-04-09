import { Agent, Transaction, ListAgentsResponse, ListTransactionsResponse, Policy, Webhook, User, ListWebhooksResponse, ListUsersResponse } from './types';

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
  // 5 days ago
  { id: 'txn-021', request_id: 'req-021', agent_id: 'agt-001', amount_cents: 28500,  currency: 'usd', vendor: 'expedia',     status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: daysAgo(5), approved_at: daysAgo(5), approved_by_user_id: 'usr-001' },
  { id: 'txn-022', request_id: 'req-022', agent_id: 'agt-002', amount_cents: 6700,   currency: 'usd', vendor: 'notion',      status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: daysAgo(5), approved_at: daysAgo(5), approved_by_user_id: 'usr-001' },
  // 6 days ago
  { id: 'txn-023', request_id: 'req-023', agent_id: 'agt-003', amount_cents: 14200,  currency: 'usd', vendor: 'github',      status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: daysAgo(6), approved_at: daysAgo(6), approved_by_user_id: 'usr-002' },
  { id: 'txn-024', request_id: 'req-024', agent_id: 'agt-004', amount_cents: 38000,  currency: 'usd', vendor: 'google ads',  status: 'APPROVED',         reason: 'approved',                  meta: {}, created_at: daysAgo(6), approved_at: daysAgo(6), approved_by_user_id: 'usr-002' },
];

export const MOCK_POLICIES: Record<string, Policy> = {
  'agt-001': {
    id: 'pol-001', agent_id: 'agt-001', daily_limit_cents: 200000, allowed_vendors: ['expedia', 'delta', 'united', 'marriott', 'hertz'], require_approval_above_cents: 100000, created_at: daysAgo(42), updated_at: daysAgo(10),
  },
  'agt-002': {
    id: 'pol-002', agent_id: 'agt-002', daily_limit_cents: 50000, allowed_vendors: ['stripe', 'notion', 'aws', 'zapier'], require_approval_above_cents: 25000, created_at: daysAgo(31), updated_at: daysAgo(5),
  },
  'agt-003': {
    id: 'pol-003', agent_id: 'agt-003', daily_limit_cents: 30000, allowed_vendors: ['github', 'vercel', 'datadog', 'aws'], require_approval_above_cents: 15000, created_at: daysAgo(18), updated_at: daysAgo(2),
  },
  'agt-004': {
    id: 'pol-004', agent_id: 'agt-004', daily_limit_cents: 100000, allowed_vendors: ['google ads', 'meta ads', 'linkedin ads'], require_approval_above_cents: 50000, created_at: daysAgo(9), updated_at: daysAgo(1),
  },
  'agt-005': {
    id: 'pol-005', agent_id: 'agt-005', daily_limit_cents: 10000, allowed_vendors: ['openai'], require_approval_above_cents: 5000, created_at: daysAgo(5), updated_at: daysAgo(5),
  },
};

export const MOCK_WEBHOOKS: Webhook[] = [
  { id: 'wh-001', agent_id: 'agt-001', url: 'https://hooks.example.com/governor/approved', events: ['transaction.approved'], active: true, created_at: daysAgo(30) },
  { id: 'wh-002', agent_id: 'agt-001', url: 'https://hooks.example.com/governor/denied', events: ['transaction.denied'], active: true, created_at: daysAgo(28) },
  { id: 'wh-003', agent_id: 'agt-002', url: 'https://api.internal.co/webhooks/spend', events: ['transaction.approved', 'transaction.denied', 'transaction.pending'], active: false, created_at: daysAgo(14) },
];

export const MOCK_USERS: User[] = [
  { id: 'usr-001', name: 'Ranveer Singh', balance_cents: 5000000, role: 'admin', created_at: daysAgo(60), updated_at: daysAgo(1) },
  { id: 'usr-002', name: 'Priya Patel', balance_cents: 2500000, role: 'admin', created_at: daysAgo(45), updated_at: daysAgo(3) },
  { id: 'usr-003', name: 'Arjun Mehta', balance_cents: 0, role: 'viewer', created_at: daysAgo(10), updated_at: daysAgo(10) },
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

export function mockSetAgentStatus(agentId: string, status: 'active' | 'frozen'): Agent {
  const agent = MOCK_AGENTS.find(a => a.id === agentId);
  if (!agent) throw new Error('Agent not found');
  agent.status = status;
  return { ...agent };
}

export function mockGetPolicy(agentId: string): Policy | null {
  return MOCK_POLICIES[agentId] || null;
}

export function mockUpdatePolicy(body: { agent_id: string; daily_limit_cents: number; allowed_vendors: string[]; require_approval_above_cents: number }): Policy {
  const existing = MOCK_POLICIES[body.agent_id];
  const policy: Policy = {
    id: existing?.id || `pol-${Date.now()}`,
    agent_id: body.agent_id,
    daily_limit_cents: body.daily_limit_cents,
    allowed_vendors: body.allowed_vendors,
    require_approval_above_cents: body.require_approval_above_cents,
    created_at: existing?.created_at || new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };
  MOCK_POLICIES[body.agent_id] = policy;
  return policy;
}

export function mockGetWebhooks(): ListWebhooksResponse {
  return { webhooks: [...MOCK_WEBHOOKS] };
}

export function mockRegisterWebhook(body: { url: string; events: string[] }): { id: string; url: string; events: string[]; secret: string; created_at: string } {
  const wh: Webhook = {
    id: `wh-${Date.now()}`,
    agent_id: 'agt-001',
    url: body.url,
    events: body.events,
    active: true,
    created_at: new Date().toISOString(),
  };
  MOCK_WEBHOOKS.push(wh);
  return { ...wh, secret: 'whsec_demo_' + Math.random().toString(36).slice(2, 18) };
}

export function mockDeleteWebhook(id: string): void {
  const idx = MOCK_WEBHOOKS.findIndex(w => w.id === id);
  if (idx === -1) throw new Error('Webhook not found');
  MOCK_WEBHOOKS.splice(idx, 1);
}

export function mockGetUsers(): ListUsersResponse {
  return { users: [...MOCK_USERS] };
}

export function mockCreateUser(body: { name: string; initial_balance_cents: number; role?: string }): User {
  const user: User = {
    id: `usr-${Date.now()}`,
    name: body.name,
    balance_cents: body.initial_balance_cents,
    role: (body.role as 'admin' | 'viewer') || 'viewer',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };
  MOCK_USERS.push(user);
  return user;
}
