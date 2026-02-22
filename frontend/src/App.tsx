import { useEffect, useMemo, useState } from 'react';
import type { FormEvent } from 'react';
import {
  ApiError,
  approveTransaction,
  centsToDollars,
  denyTransaction,
  freezeAgent,
  freezeUser,
  getAgentHistory,
  getAdminMe,
  getPolicyForAgent,
  listAgents,
  listPending,
  listTransactions,
  listUsers,
  loginAdmin,
  simulateSpend,
  unfreezeAgent,
  unfreezeUser,
  upsertPolicyForAgent,
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

function isAdminSessionError(err: unknown): boolean {
  if (err instanceof ApiError) {
    return err.status === 401;
  }
  if (!(err instanceof Error)) {
    return false;
  }
  const message = err.message.toLowerCase();
  return message.includes('invalid session') || message.includes('unauthorized');
}

function normalizePolicy(policy: Policy): Policy {
  const dailyLimit = Number.isFinite(policy.daily_limit_cents) ? policy.daily_limit_cents : 0;
  const perTxnLimit = Number.isFinite(policy.per_transaction_limit_cents) && policy.per_transaction_limit_cents > 0
    ? policy.per_transaction_limit_cents
    : dailyLimit;

  return {
    ...policy,
    daily_limit_cents: dailyLimit,
    per_transaction_limit_cents: perTxnLimit,
    allowed_vendors: policy.allowed_vendors || [],
    allowed_mccs: policy.allowed_mccs || [],
    allowed_weekdays_utc: policy.allowed_weekdays_utc || [],
    allowed_hours_utc: policy.allowed_hours_utc || [],
  };
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
  const [guidelineDraft, setGuidelineDraft] = useState('');
  const [agentHistory, setAgentHistory] = useState<Transaction[]>([]);

  const [spendApiKey, setSpendApiKey] = useState('sk_test_agent_123');
  const [spendVendor, setSpendVendor] = useState('openai.com');
  const [spendMcc, setSpendMcc] = useState('5734');
  const [spendAmount, setSpendAmount] = useState('5.00');
  const [spendResult, setSpendResult] = useState<SpendResponse | null>(null);

  const [loadingDashboard, setLoadingDashboard] = useState(false);
  const [savingGuideline, setSavingGuideline] = useState(false);
  const [needsReauth, setNeedsReauth] = useState(false);
  const [busyKillSwitch, setBusyKillSwitch] = useState(false);
  const [busyTransactionId, setBusyTransactionId] = useState('');
  const [statusMessage, setStatusMessage] = useState<string>('');
  const [errorMessage, setErrorMessage] = useState<string>('');

  const selectedAgent = useMemo(
    () => agents.find((agent) => agent.id === selectedAgentId) || null,
    [agents, selectedAgentId],
  );

  const markReauthRequired = (message: string) => {
    setNeedsReauth(true);
    setStatusMessage('Session expired. Re-authentication required.');
    setErrorMessage(message);
  };

  const handleAdminError = (err: unknown, fallbackMessage: string) => {
    const message = err instanceof Error ? err.message : fallbackMessage;
    if (isAdminSessionError(err)) {
      markReauthRequired(message);
      return;
    }
    setErrorMessage(message);
  };

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
        handleAdminError(err, 'Failed to load dashboard data');
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
      setGuidelineDraft('');
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
        const normalizedPolicy = normalizePolicy(policyRes.policy);
        setPolicy(normalizedPolicy);
        setGuidelineDraft(normalizedPolicy.purchase_guideline || '');
        setAgentHistory(historyRes.transactions);
      } catch (err) {
        if (cancelled) return;
        handleAdminError(err, 'Failed to load agent context');
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
    setNeedsReauth(false);
    setAdminEmail('');
    setUsers([]);
    setAgents([]);
    setPending([]);
    setRecentTransactions([]);
    setPolicy(null);
    setGuidelineDraft('');
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
      setNeedsReauth(false);
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
        const normalizedPolicy = normalizePolicy(policyRes.policy);
        setPolicy(normalizedPolicy);
        setGuidelineDraft(normalizedPolicy.purchase_guideline || '');
        setAgentHistory(historyRes.transactions);
      }

      setStatusMessage(`Transaction ${action}d.`);
    } catch (err) {
      handleAdminError(err, `Failed to ${action} transaction`);
    } finally {
      setBusyTransactionId('');
    }
  };

  const handleSaveGuideline = async (event: FormEvent) => {
    event.preventDefault();
    if (!token || !policy) return;

    setSavingGuideline(true);
    setErrorMessage('');
    setStatusMessage('Saving purchase guideline...');

    try {
      await upsertPolicyForAgent(token, {
        agent_id: policy.agent_id,
        daily_limit_cents: policy.daily_limit_cents,
        per_transaction_limit_cents: policy.per_transaction_limit_cents,
        allowed_vendors: policy.allowed_vendors,
        allowed_mccs: policy.allowed_mccs || [],
        allowed_weekdays_utc: policy.allowed_weekdays_utc || [],
        allowed_hours_utc: policy.allowed_hours_utc || [],
        require_approval_above_cents: policy.require_approval_above_cents,
        purchase_guideline: guidelineDraft.trim(),
      });

      const policyRes = await getPolicyForAgent(token, policy.agent_id);
      const normalizedPolicy = normalizePolicy(policyRes.policy);
      setPolicy(normalizedPolicy);
      setGuidelineDraft(normalizedPolicy.purchase_guideline || '');
      setStatusMessage('Purchase guideline saved.');
    } catch (err) {
      handleAdminError(err, 'Failed to save purchase guideline');
    } finally {
      setSavingGuideline(false);
    }
  };

  const handleSpendSimulation = async (event: FormEvent) => {
    event.preventDefault();
    setErrorMessage('');
    setStatusMessage('Sending spend request...');

    const amountDollars = Number.parseFloat(spendAmount);
    if (!Number.isFinite(amountDollars) || amountDollars <= 0) {
      setErrorMessage('Amount must be a positive dollar value.');
      setStatusMessage('');
      return;
    }
    const amountCents = Math.round(amountDollars * 100);

    try {
      const response = await simulateSpend(spendApiKey, {
        request_id: uidv4Like(),
        amount: amountCents,
        vendor: spendVendor.trim().toLowerCase(),
        meta: {
          source: 'governor-dashboard',
          initiated_at: new Date().toISOString(),
        },
        mcc: spendMcc.trim(),
      });
      setSpendResult(response);
      setStatusMessage(`Spend decision: ${response.status}.`);
      if (token) {
        try {
          await refreshTransactions(token);
        } catch (err) {
          handleAdminError(err, 'Spend processed but dashboard refresh failed');
        }
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Spend simulation failed';
      setErrorMessage(message);
      setStatusMessage('');
    }
  };

  const handleReauthenticate = () => {
    logout();
    setStatusMessage('Session cleared. Please sign in again.');
    setErrorMessage('');
  };

  const handleAgentKillSwitch = async (action: 'freeze' | 'unfreeze') => {
    if (!token || !selectedAgentId) return;
    setBusyKillSwitch(true);
    setErrorMessage('');
    setStatusMessage(`${action === 'freeze' ? 'Freezing' : 'Unfreezing'} agent...`);
    try {
      if (action === 'freeze') {
        await freezeAgent(token, selectedAgentId);
      } else {
        await unfreezeAgent(token, selectedAgentId);
      }
      const [agentsRes, usersRes] = await Promise.all([listAgents(token, 30), listUsers(token, 20)]);
      setAgents(agentsRes.agents);
      setUsers(usersRes.users);
      await refreshTransactions(token);
      setStatusMessage(`Agent ${action}d.`);
    } catch (err) {
      handleAdminError(err, `Failed to ${action} agent`);
    } finally {
      setBusyKillSwitch(false);
    }
  };

  const handleOrgKillSwitch = async (action: 'freeze' | 'unfreeze') => {
    if (!token || !selectedAgent) return;
    setBusyKillSwitch(true);
    setErrorMessage('');
    setStatusMessage(`${action === 'freeze' ? 'Freezing' : 'Unfreezing'} organization...`);
    try {
      if (action === 'freeze') {
        await freezeUser(token, selectedAgent.user_id);
      } else {
        await unfreezeUser(token, selectedAgent.user_id);
      }
      const [agentsRes, usersRes] = await Promise.all([listAgents(token, 30), listUsers(token, 20)]);
      setAgents(agentsRes.agents);
      setUsers(usersRes.users);
      await refreshTransactions(token);
      setStatusMessage(`Organization ${action}d.`);
    } catch (err) {
      handleAdminError(err, `Failed to ${action} organization`);
    } finally {
      setBusyKillSwitch(false);
    }
  };

  return (
    <div className={`app-shell ${token ? 'app-shell--dashboard' : 'app-shell--auth'}`}>
      <div className="atmosphere-grid" />

      {!token ? (
        <main className="auth-page">
          <section className="auth-hero">
            <p className="eyebrow">Applied AI for Internet Money</p>
            <h1>
              Let your agents spend.
              <br />
              Keep them inside the bounds.
            </h1>
            <p>
              Governor is a deterministic policy engine that sits between your AI agents and real-world payments.
              Agents request spend; Governor enforces hard limits with a clean audit trail for every decision.
            </p>
            <div className="hero-actions">
              <a className="hero-cta" href="https://usegovernor.vercel.app/" target="_blank" rel="noreferrer">Get early access</a>
              <a className="hero-link" href="https://usegovernor.vercel.app/" target="_blank" rel="noreferrer">See how it works →</a>
            </div>
          </section>

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
        </main>
      ) : (
        <>
          {needsReauth && (
            <section className="panel reauth-panel">
              <h2>Re-authentication Required</h2>
              <p>Your admin session is no longer valid. Sign in again to continue reviewing and updating policies.</p>
              <button type="button" className="button-primary" onClick={handleReauthenticate}>
                Re-authenticate
              </button>
            </section>
          )}

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

            {selectedAgent && (
              <div className="kill-switch-row">
                <button
                  type="button"
                  className="button-danger"
                  disabled={busyKillSwitch || needsReauth}
                  onClick={() => handleAgentKillSwitch('freeze')}
                >
                  Freeze Agent
                </button>
                <button
                  type="button"
                  className="button-ghost"
                  disabled={busyKillSwitch || needsReauth}
                  onClick={() => handleAgentKillSwitch('unfreeze')}
                >
                  Unfreeze Agent
                </button>
                <button
                  type="button"
                  className="button-danger"
                  disabled={busyKillSwitch || needsReauth}
                  onClick={() => handleOrgKillSwitch('freeze')}
                >
                  Freeze Org
                </button>
                <button
                  type="button"
                  className="button-ghost"
                  disabled={busyKillSwitch || needsReauth}
                  onClick={() => handleOrgKillSwitch('unfreeze')}
                >
                  Unfreeze Org
                </button>
              </div>
            )}

            {policy && (
              <div className="policy-card">
                <h3>Live Policy Constraints</h3>
                <dl>
                  <div>
                    <dt>Daily Limit</dt>
                    <dd>{centsToDollars(policy.daily_limit_cents)}</dd>
                  </div>
                  <div>
                    <dt>Per Transaction Limit</dt>
                    <dd>{centsToDollars(policy.per_transaction_limit_cents)}</dd>
                  </div>
                  <div>
                    <dt>Human Approval Over</dt>
                    <dd>{centsToDollars(policy.require_approval_above_cents)}</dd>
                  </div>
                  <div>
                    <dt>Allowed Vendors</dt>
                    <dd>{policy.allowed_vendors.join(', ') || 'None'}</dd>
                  </div>
                  <div>
                    <dt>Allowed MCCs</dt>
                    <dd>{(policy.allowed_mccs || []).join(', ') || 'Any'}</dd>
                  </div>
                  <div>
                    <dt>Allowed UTC Weekdays</dt>
                    <dd>{(policy.allowed_weekdays_utc || []).join(', ') || 'Any'}</dd>
                  </div>
                </dl>

                <form className="stacked-form guideline-form" onSubmit={handleSaveGuideline}>
                  <label>
                    Purchase Guideline Prompt
                    <textarea
                      rows={2}
                      value={guidelineDraft}
                      onChange={(e) => setGuidelineDraft(e.target.value)}
                      placeholder="Example: AI and engineering tooling subscriptions only"
                    />
                  </label>
                  <button
                    type="submit"
                    className="button-primary"
                    disabled={savingGuideline || needsReauth}
                  >
                    {savingGuideline ? 'Saving...' : 'Save Guideline'}
                  </button>
                </form>
                <p className="guideline-note">
                  Spend requests are denied with <code>guideline_mismatch</code> when a vendor domain is not relevant
                  to this prompt.
                </p>
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
                MCC (optional)
                <input value={spendMcc} onChange={(e) => setSpendMcc(e.target.value)} />
              </label>
              <label>
                Amount (USD)
                <input
                  value={spendAmount}
                  onChange={(e) => setSpendAmount(e.target.value)}
                  type="number"
                  inputMode="decimal"
                  min="0.01"
                  step="0.01"
                  placeholder="5.00"
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
                          disabled={busyTransactionId === tx.id || needsReauth}
                          onClick={() => handleReviewAction(tx.id, 'approve')}
                        >
                          Approve
                        </button>
                        <button
                          type="button"
                          className="button-danger"
                          disabled={busyTransactionId === tx.id || needsReauth}
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
        </>
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
