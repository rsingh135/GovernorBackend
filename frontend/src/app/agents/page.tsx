'use client';

import { useEffect, useState } from 'react';
import { Badge } from '@/components/Badge';
import { getAgents, createAgent, createPolicy } from '@/lib/api';
import { Agent } from '@/lib/types';

export default function AgentsPage() {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);

  async function load() {
    setLoading(true);
    try {
      const res = await getAgents({ limit: 100 });
      setAgents(res.agents || []);
      setTotal(res.total);
    } catch { /* ignore */ }
    setLoading(false);
  }

  useEffect(() => { load(); }, []);

  return (
    <div className="p-8 max-w-5xl mx-auto">
      <div className="mb-8 flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-semibold mb-1" style={{ color: 'var(--text)' }}>Agents</h1>
          <p className="text-sm" style={{ color: 'var(--muted)' }}>{total} agent{total !== 1 ? 's' : ''} provisioned</p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          style={{ background: 'var(--accent)', color: 'white' }}
        >
          + New Agent
        </button>
      </div>

      <div className="rounded-xl overflow-hidden" style={{ border: '1px solid var(--border)' }}>
        {loading ? (
          <div className="p-8 text-center text-sm" style={{ background: 'var(--panel)', color: 'var(--muted)' }}>Loading…</div>
        ) : agents.length === 0 ? (
          <div className="p-8 text-center text-sm" style={{ background: 'var(--panel)', color: 'var(--muted)' }}>No agents yet.</div>
        ) : (
          <table className="w-full text-sm" style={{ background: 'var(--panel)' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--border)' }}>
                {['Name', 'Status', 'Key Prefix', 'User ID', 'Created', 'Actions'].map(h => (
                  <th key={h} className="px-5 py-3 text-left text-xs font-medium tracking-wider" style={{ color: 'var(--muted)' }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {agents.map((a, i) => (
                <tr key={a.id} style={{ borderBottom: i < agents.length - 1 ? '1px solid var(--border)' : 'none' }}>
                  <td className="px-5 py-3 font-medium" style={{ color: 'var(--text)' }}>{a.name}</td>
                  <td className="px-5 py-3"><Badge status={a.status} /></td>
                  <td className="px-5 py-3 font-mono text-xs" style={{ color: 'var(--muted)' }}>{a.api_key_prefix}…</td>
                  <td className="px-5 py-3 font-mono text-xs" style={{ color: 'var(--muted)' }}>{a.user_id.slice(0, 8)}…</td>
                  <td className="px-5 py-3 text-xs" style={{ color: 'var(--muted)' }}>{new Date(a.created_at).toLocaleDateString()}</td>
                  <td className="px-5 py-3">
                    <button
                      disabled
                      title="Coming soon — freeze/unfreeze endpoint not yet implemented"
                      className="px-3 py-1 rounded-md text-xs opacity-30 cursor-not-allowed"
                      style={{ border: '1px solid var(--border)', color: 'var(--muted)' }}
                    >
                      {a.status === 'active' ? 'Freeze' : 'Unfreeze'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {showCreate && <CreateAgentModal onClose={() => { setShowCreate(false); load(); }} />}
    </div>
  );
}

function CreateAgentModal({ onClose }: { onClose: () => void }) {
  const [userId, setUserId] = useState('');
  const [name, setName] = useState('');
  const [dailyLimit, setDailyLimit] = useState('10000');
  const [vendors, setVendors] = useState('');
  const [approvalThreshold, setApprovalThreshold] = useState('5000');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [apiKey, setApiKey] = useState('');

  async function submit() {
    setLoading(true);
    setError('');
    try {
      const agent = await createAgent({ user_id: userId, name }) as { id: string; api_key: string };
      await createPolicy({
        agent_id: agent.id,
        daily_limit_cents: Math.round(parseFloat(dailyLimit) * 100),
        allowed_vendors: vendors.split(',').map(v => v.trim()).filter(Boolean),
        require_approval_above_cents: Math.round(parseFloat(approvalThreshold) * 100),
      });
      setApiKey(agent.api_key);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create agent');
    }
    setLoading(false);
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      style={{ background: 'rgba(5,6,8,0.8)', backdropFilter: 'blur(4px)' }}
      onClick={(e) => e.target === e.currentTarget && !apiKey && onClose()}
    >
      <div className="w-full max-w-md rounded-2xl p-6 flex flex-col gap-5" style={{ background: 'var(--panel)', border: '1px solid var(--border)' }}>
        {apiKey ? (
          <>
            <h2 className="text-base font-semibold" style={{ color: 'var(--text)' }}>Agent created!</h2>
            <p className="text-sm" style={{ color: 'var(--muted)' }}>Copy this API key — it won't be shown again.</p>
            <div className="rounded-xl p-4 font-mono text-sm break-all" style={{ background: '#0d0f15', border: '1px solid var(--border)', color: 'var(--accent)' }}>
              {apiKey}
            </div>
            <button onClick={onClose} className="px-4 py-2 rounded-lg text-sm font-medium" style={{ background: 'var(--accent)', color: 'white' }}>
              Done
            </button>
          </>
        ) : (
          <>
            <div className="flex items-center justify-between">
              <h2 className="text-base font-semibold" style={{ color: 'var(--text)' }}>New Agent</h2>
              <button onClick={onClose} style={{ color: 'var(--muted)' }}>
                <svg width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div className="flex flex-col gap-4">
              <InputField label="User ID" value={userId} onChange={setUserId} placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" />
              <InputField label="Agent Name" value={name} onChange={setName} placeholder="e.g. travel-booking-agent" />
              <InputField label="Daily Limit ($)" value={dailyLimit} onChange={setDailyLimit} placeholder="100.00" />
              <InputField label="Allowed Vendors (comma-separated)" value={vendors} onChange={setVendors} placeholder="stripe, expedia, aws" />
              <InputField label="Approval Threshold ($)" value={approvalThreshold} onChange={setApprovalThreshold} placeholder="50.00" />
            </div>

            {error && <p className="text-xs p-3 rounded-lg" style={{ background: 'rgba(248,113,113,0.1)', color: 'var(--red)', border: '1px solid rgba(248,113,113,0.2)' }}>{error}</p>}

            <div className="flex gap-3">
              <button onClick={onClose} className="flex-1 px-4 py-2 rounded-lg text-sm" style={{ color: 'var(--muted)', border: '1px solid var(--border)' }}>Cancel</button>
              <button onClick={submit} disabled={loading || !userId || !name} className="flex-1 px-4 py-2 rounded-lg text-sm font-medium disabled:opacity-50" style={{ background: 'var(--accent)', color: 'white' }}>
                {loading ? 'Creating…' : 'Create Agent'}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

function InputField({ label, value, onChange, placeholder }: { label: string; value: string; onChange: (v: string) => void; placeholder?: string }) {
  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-xs font-medium" style={{ color: 'var(--text)' }}>{label}</label>
      <input
        value={value}
        onChange={e => onChange(e.target.value)}
        placeholder={placeholder}
        className="rounded-lg px-3 py-2 text-sm outline-none"
        style={{ background: '#0d0f15', border: '1px solid var(--border)', color: 'var(--text)' }}
        onFocus={e => (e.target.style.borderColor = 'var(--accent)')}
        onBlur={e => (e.target.style.borderColor = 'var(--border)')}
      />
    </div>
  );
}
