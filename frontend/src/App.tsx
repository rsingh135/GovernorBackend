import { useEffect, useMemo, useState } from 'react';
import type { FormEvent } from 'react';
import {
  approveTransaction,
  centsToDollars,
  denyTransaction,
  getAgentHistory,
  getAdminMe,
  getPolicyForAgent,
  listAgents,
  listPending,
  listTransactions,
  listUsers,
  loginAdmin,
  simulateSpend,
} from './api';
import type { Agent, Policy, SpendResponse, Transaction, User } from './types';
import './index.css';

const ADMIN_TOKEN_KEY = 'governor_admin_token';

function uidv4Like(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (ch) => {
    const rand = (Math.random() * 16) | 0;
    const value = ch === 'x' ? rand : (rand & 0x3) | 0x8;
    return value.toString(16);
  });
}

function statusClass(status: string): string {
  if (status.includes('approved')) return 'pill approved';
  if (status.includes('pending')) return 'pill pending';
  return 'pill denied';
}

export default function App() {
  const [email, setEmail] = useState('admin@governor.local');
  const [password, setPassword] = useState('governor_admin_123');
  const [token, setToken] = useState<string>(() => localStorage.getItem(ADMIN_TOKEN_KEY) || '');
  const [adminEmail, setAdminEmail] = useState<string>('');

  const [users, setUsers] = useState<User[]>([]);
  const [agents, setAgents] = useState<Agent[]>([]);
  const [pending, setPending] = useState<Transaction[]>([]);
  const [recentTransactions, setRecentTransactions] = useState<Transaction[]>([]);
  const [selectedAgentId, setSelectedAgentId] = useState<string>('');
  const [policy, setPolicy] = useState<Policy | null>(null);
  const [agentHistory, setAgentHistory] = useState<Transaction[]>([]);

  const [spendApiKey, setSpendApiKey] = useState('sk_test_agent_123');
  const [spendVendor, setSpendVendor] = useState('openai.com');
  const [spendAmount, setSpendAmount] = useState('500');
  const [spendResult, setSpendResult] = useState<SpendResponse | null>(null);

  const [loadingDashboard, setLoadingDashboard] = useState(false);
  const [busyTransactionId, setBusyTransactionId] = useState('');
  const [statusMessage, setStatusMessage] = useState<string>('');
  const [errorMessage, setErrorMessage] = useState<string>('');

  const totalPendingCents = useMemo(
    () => pending.reduce((sum, tx) => sum + tx.amount_cents, 0),
    [pending],
  );

  useEffect(() => {
    if (!token) return;

    let cancelled = false;

    const bootstrap = async () => {
      setLoadingDashboard(true);
      setErrorMessage('');
      try {
        const [me, usersRes, agentsRes, pendingRes, txRes] = await Promise.all([
          getAdminMe(token),
          listUsers(token, 20),
          listAgents(token, 30),
          listPending(token, 30),
          listTransactions(token, 30),
        ]);

        if (cancelled) return;

        setAdminEmail(me.admin.email);
        setUsers(usersRes.users);
        setAgents(agentsRes.agents);
        setPending(pendingRes.transactions);
        setRecentTransactions(txRes.transactions);

        const nextAgentId = selectedAgentId || pendingRes.transactions[0]?.agent_id || agentsRes.agents[0]?.id || '';
        setSelectedAgentId(nextAgentId);
      } catch (err) {
        if (cancelled) return;
        const message = err instanceof Error ? err.message : 'Failed to load dashboard data';
        setErrorMessage(message);
        if (message.toLowerCase().includes('invalid session') || message.toLowerCase().includes('unauthorized')) {
          logout();
        }
      } finally {
        if (!cancelled) setLoadingDashboard(false);
      }
    };

    bootstrap();

    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  useEffect(() => {
    if (!token || !selectedAgentId) {
      setPolicy(null);
      setAgentHistory([]);
      return;
    }

    let cancelled = false;

    const loadAgentContext = async () => {
      try {
        const [policyRes, historyRes] = await Promise.all([
          getPolicyForAgent(token, selectedAgentId),
          getAgentHistory(token, selectedAgentId, 10),
        ]);

        if (cancelled) return;
        setPolicy(policyRes.policy);
        setAgentHistory(historyRes.transactions);
      } catch (err) {
        if (cancelled) return;
        const message = err instanceof Error ? err.message : 'Failed to load agent context';
        setErrorMessage(message);
      }
    };

    loadAgentContext();

    return () => {
      cancelled = true;
    };
  }, [token, selectedAgentId]);

  const logout = () => {
    localStorage.removeItem(ADMIN_TOKEN_KEY);
    setToken('');
    setAdminEmail('');
    setUsers([]);
    setAgents([]);
    setPending([]);
    setRecentTransactions([]);
    setPolicy(null);
    setAgentHistory([]);
  };

  const refreshTransactions = async (sessionToken: string) => {
    const [pendingRes, txRes] = await Promise.all([
      listPending(sessionToken, 30),
      listTransactions(sessionToken, 30),
    ]);
    setPending(pendingRes.transactions);
    setRecentTransactions(txRes.transactions);
  };

  const handleLogin = async (event: FormEvent) => {
    event.preventDefault();
    setErrorMessage('');
    setStatusMessage('Signing in...');

    try {
      const res = await loginAdmin(email, password);
      localStorage.setItem(ADMIN_TOKEN_KEY, res.token);
      setToken(res.token);
      setAdminEmail(res.admin.email);
      setStatusMessage('Dashboard unlocked.');
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Login failed';
      setErrorMessage(message);
      setStatusMessage('');
    }
  };

  const handleReviewAction = async (transactionId: string, action: 'approve' | 'deny') => {
    if (!token) return;

    setBusyTransactionId(transactionId);
    setErrorMessage('');
    try {
      if (action === 'approve') {
        await approveTransaction(token, transactionId);
      } else {
        await denyTransaction(token, transactionId);
      }

      await refreshTransactions(token);
      if (selectedAgentId) {
        const [policyRes, historyRes] = await Promise.all([
          getPolicyForAgent(token, selectedAgentId),
          getAgentHistory(token, selectedAgentId, 10),
        ]);
        setPolicy(policyRes.policy);
        setAgentHistory(historyRes.transactions);
      }

      setStatusMessage(`Transaction ${action}d.`);
    } catch (err) {
      const message = err instanceof Error ? err.message : `Failed to ${action} transaction`;
      setErrorMessage(message);
    } finally {
      setBusyTransactionId('');
    }
  };

  const handleSpendSimulation = async (event: FormEvent) => {
    event.preventDefault();
    setErrorMessage('');
    setStatusMessage('Sending spend request...');

    const amount = Number.parseInt(spendAmount, 10);
    if (!Number.isInteger(amount) || amount <= 0) {
      setErrorMessage('Amount must be a positive integer in cents.');
      setStatusMessage('');
      return;
    }

    try {
      const response = await simulateSpend(spendApiKey, {
        request_id: uidv4Like(),
        amount,
        vendor: spendVendor.trim().toLowerCase(),
        meta: {
          source: 'governor-dashboard',
          initiated_at: new Date().toISOString(),
        },
      });
      setSpendResult(response);
      setStatusMessage(`Spend decision: ${response.status}.`);
      if (token) {
        await refreshTransactions(token);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Spend simulation failed';
      setErrorMessage(message);
      setStatusMessage('');
    }
  };

  return (
    <div className="app-shell">
      <div className="atmosphere-grid" />
      <header className="hero">
        <div>
          <p className="eyebrow">Governor Control Plane</p>
          <h1>Policy-First Purchasing for Agent Workflows</h1>
          <p>
            Simulate live spend requests, route high-risk transactions to human review, and keep every purchase
            decision explainable.
          </p>
        </div>
        <div className="hero-card">
          <p>Today&apos;s Queue</p>
          <strong>{pending.length}</strong>
          <span>{centsToDollars(totalPendingCents)} awaiting decisions</span>
        </div>
      </header>

      {!token ? (
        <section className="panel login-panel">
          <h2>Admin Sign In</h2>
          <p>Use the seeded local credentials to unlock dashboard controls.</p>
          <form onSubmit={handleLogin} className="stacked-form">
            <label>
              Email
              <input value={email} onChange={(e) => setEmail(e.target.value)} required autoComplete="email" />
            </label>
            <label>
              Password
              <input
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                type="password"
                required
                autoComplete="current-password"
              />
            </label>
            <button type="submit" className="button-primary">Unlock Dashboard</button>
          </form>
        </section>
      ) : (
        <main className="dashboard-grid">
          <section className="panel metrics-panel">
            <div className="panel-header-row">
              <div>
                <h2>Operations Snapshot</h2>
                <p>Signed in as {adminEmail || 'admin'}.</p>
              </div>
              <button className="button-ghost" onClick={logout}>Sign Out</button>
            </div>

            <div className="metrics-cards">
              <article>
                <p>Users</p>
                <strong>{users.length}</strong>
              </article>
              <article>
                <p>Agents</p>
                <strong>{agents.length}</strong>
              </article>
              <article>
                <p>Pending</p>
                <strong>{pending.length}</strong>
              </article>
              <article>
                <p>Recent Decisions</p>
                <strong>{recentTransactions.length}</strong>
              </article>
            </div>

            <label className="agent-selector">
              Focus Agent
              <select value={selectedAgentId} onChange={(e) => setSelectedAgentId(e.target.value)}>
                <option value="">Select agent</option>
                {agents.map((agent) => (
                  <option key={agent.id} value={agent.id}>
                    {agent.name} ({agent.status})
                  </option>
                ))}
              </select>
            </label>

            {policy && (
              <div className="policy-card">
                <h3>Live Policy Constraints</h3>
                <dl>
                  <div>
                    <dt>Daily Limit</dt>
                    <dd>{centsToDollars(policy.daily_limit_cents)}</dd>
                  </div>
                  <div>
                    <dt>Human Approval Over</dt>
                    <dd>{centsToDollars(policy.require_approval_above_cents)}</dd>
                  </div>
                  <div>
                    <dt>Allowed Vendors</dt>
                    <dd>{policy.allowed_vendors.join(', ') || 'None'}</dd>
                  </div>
                </dl>
              </div>
            )}
          </section>

          <section className="panel simulation-panel">
            <h2>Spend Simulation</h2>
            <p>Emulates an agent-side checkout attempt while hiding provider secrets from the agent.</p>

            <form className="stacked-form" onSubmit={handleSpendSimulation}>
              <label>
                Agent API Key
                <input value={spendApiKey} onChange={(e) => setSpendApiKey(e.target.value)} required />
              </label>
              <label>
                Vendor Domain
                <input value={spendVendor} onChange={(e) => setSpendVendor(e.target.value)} required />
              </label>
              <label>
                Amount (cents)
                <input
                  value={spendAmount}
                  onChange={(e) => setSpendAmount(e.target.value)}
                  inputMode="numeric"
                  pattern="[0-9]+"
                  required
                />
              </label>
              <button className="button-primary" type="submit">Run Purchase Decision</button>
            </form>

            {spendResult && (
              <div className="decision-card">
                <h3>Decision Outcome</h3>
                <p className={statusClass(spendResult.status)}>{spendResult.status}</p>
                <p><strong>Reason:</strong> {spendResult.reason}</p>
                {spendResult.provider_status && <p><strong>Provider:</strong> {spendResult.provider_status}</p>}
                {spendResult.checkout_url && (
                  <p>
                    <strong>Checkout:</strong>{' '}
                    <a href={spendResult.checkout_url} target="_blank" rel="noreferrer">Open Stripe Checkout</a>
                  </p>
                )}
              </div>
            )}
          </section>

          <section className="panel pending-panel">
            <h2>Pending Human Review</h2>
            <p>Transactions above threshold stay queued until approved or denied.</p>
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Vendor</th>
                    <th>Amount</th>
                    <th>Reason</th>
                    <th>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {pending.length === 0 && (
                    <tr>
                      <td colSpan={4} className="empty-row">No pending approvals.</td>
                    </tr>
                  )}
                  {pending.map((tx) => (
                    <tr key={tx.id}>
                      <td>
                        <button
                          className="link-button"
                          onClick={() => setSelectedAgentId(tx.agent_id)}
                          type="button"
                        >
                          {tx.vendor}
                        </button>
                      </td>
                      <td>{centsToDollars(tx.amount_cents)}</td>
                      <td>{tx.reason}</td>
                      <td className="actions-cell">
                        <button
                          type="button"
                          className="button-success"
                          disabled={busyTransactionId === tx.id}
                          onClick={() => handleReviewAction(tx.id, 'approve')}
                        >
                          Approve
                        </button>
                        <button
                          type="button"
                          className="button-danger"
                          disabled={busyTransactionId === tx.id}
                          onClick={() => handleReviewAction(tx.id, 'deny')}
                        >
                          Deny
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>

          <section className="panel history-panel">
            <h2>Selected Agent History (Last 10)</h2>
            <p>Use this context before deciding flagged payments.</p>
            <ul>
              {agentHistory.length === 0 && <li className="empty-row">Select an agent to view history.</li>}
              {agentHistory.map((tx) => (
                <li key={tx.id}>
                  <span>{new Date(tx.created_at).toLocaleString()}</span>
                  <span>{tx.vendor}</span>
                  <span>{centsToDollars(tx.amount_cents)}</span>
                  <span className={statusClass(tx.status)}>{tx.status}</span>
                </li>
              ))}
            </ul>
          </section>
        </main>
      )}

      {(loadingDashboard || statusMessage || errorMessage) && (
        <footer className="status-bar">
          {loadingDashboard && <span>Loading dashboard...</span>}
          {!loadingDashboard && statusMessage && <span>{statusMessage}</span>}
          {errorMessage && <span className="error">{errorMessage}</span>}
        </footer>
      )}
    </div>
  );
}
