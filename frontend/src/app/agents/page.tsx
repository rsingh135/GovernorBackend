'use client';

import { useEffect, useState, Fragment } from 'react';
import { Badge } from '@/components/Badge';
import { getAgents, createAgent, createPolicy, setAgentStatus, getPolicy, updatePolicy } from '@/lib/api';
import { Agent, Policy } from '@/lib/types';

export default function AgentsPage() {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [freezingId, setFreezingId] = useState<string | null>(null);

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

  async function handleToggleStatus(agent: Agent) {
    setFreezingId(agent.id);
    try {
      const newStatus = agent.status === 'active' ? 'frozen' : 'active';
      await setAgentStatus(agent.id, newStatus);
      await load();
    } catch { /* ignore */ }
    setFreezingId(null);
  }

  function toggleExpanded(agentId: string) {
    setExpandedId(prev => prev === agentId ? null : agentId);
  }

  return (
    <div className="p-6 md:p-8 max-w-5xl mx-auto page-enter">
      <div className="mb-8 flex items-start justify-between">
        <div>
          <p className="text-xs font-medium tracking-wider uppercase mb-2" style={{ color: 'var(--muted)', letterSpacing: '2px' }}>Management</p>
          <h1 style={{ fontFamily: 'var(--font-serif)', fontSize: 'clamp(28px, 4vw, 36px)', fontWeight: 500, color: 'var(--black)', lineHeight: 1.1 }}>
            Agents
          </h1>
          <p className="text-sm mt-2" style={{ color: 'var(--mid)' }}>{total} agent{total !== 1 ? 's' : ''} provisioned</p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="btn-lift px-4 py-2.5 rounded-lg text-sm font-medium"
          style={{ background: 'var(--black)', color: 'var(--bg)' }}
        >
          + New Agent
        </button>
      </div>

      <div className="rounded-xl overflow-hidden fade-up" style={{ border: '1px solid var(--border)' }}>
        {loading ? (
          <div className="p-6" style={{ background: 'var(--bg-card)' }}>
            <div className="space-y-3">
              {[1,2,3].map(i => <div key={i} className="skeleton h-10 w-full" />)}
            </div>
          </div>
        ) : agents.length === 0 ? (
          <div className="p-12 text-center flex flex-col items-center gap-3" style={{ background: 'var(--bg-card)' }}>
            <div className="w-12 h-12 rounded-full flex items-center justify-center" style={{ background: 'rgba(0,0,0,0.05)' }}>
              <svg width="22" height="22" fill="none" stroke="var(--muted)" strokeWidth="1.8" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17H4a2 2 0 01-2-2V5a2 2 0 012-2h16a2 2 0 012 2v10a2 2 0 01-2 2h-1" />
              </svg>
            </div>
            <p className="text-sm font-medium" style={{ color: 'var(--black)' }}>No agents yet</p>
            <p className="text-xs" style={{ color: 'var(--muted)' }}>Create your first agent to get started.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm" style={{ background: 'var(--bg-card)' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border)' }}>
                  {['Name', 'Status', 'Key Prefix', 'User ID', 'Created', 'Actions'].map(h => (
                    <th key={h} className="px-5 py-3 text-left text-xs font-medium tracking-wider uppercase" style={{ color: 'var(--muted)', letterSpacing: '1px' }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {agents.map((a, i) => (
                  <Fragment key={a.id}>
                    <tr
                      className="table-row-hover row-animate cursor-pointer"
                      style={{
                        borderBottom: expandedId === a.id ? 'none' : (i < agents.length - 1 ? '1px solid var(--border)' : 'none'),
                        animationDelay: `${i * 0.04}s`,
                      }}
                      onClick={() => toggleExpanded(a.id)}
                    >
                      <td className="px-5 py-3.5 font-medium" style={{ color: 'var(--black)' }}>
                        <div className="flex items-center gap-2">
                          <svg width="12" height="12" fill="none" stroke="var(--muted)" strokeWidth="2" viewBox="0 0 24 24"
                            style={{ transform: expandedId === a.id ? 'rotate(90deg)' : 'rotate(0)', transition: 'transform 0.2s ease' }}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
                          </svg>
                          {a.name}
                        </div>
                      </td>
                      <td className="px-5 py-3.5"><Badge status={a.status} /></td>
                      <td className="px-5 py-3.5 text-xs" style={{ fontFamily: 'var(--font-mono)', color: 'var(--muted)' }}>{a.api_key_prefix}...</td>
                      <td className="px-5 py-3.5 text-xs" style={{ fontFamily: 'var(--font-mono)', color: 'var(--muted)' }}>{a.user_id.slice(0, 8)}...</td>
                      <td className="px-5 py-3.5 text-xs" style={{ color: 'var(--muted)' }}>{new Date(a.created_at).toLocaleDateString()}</td>
                      <td className="px-5 py-3.5" onClick={e => e.stopPropagation()}>
                        <button
                          onClick={() => handleToggleStatus(a)}
                          disabled={freezingId === a.id}
                          className="btn-lift px-3 py-1.5 rounded-md text-xs font-medium disabled:opacity-50"
                          style={{
                            background: a.status === 'active' ? 'var(--red-soft)' : 'var(--green-soft)',
                            color: a.status === 'active' ? 'var(--red)' : 'var(--green)',
                            border: `1px solid ${a.status === 'active' ? 'var(--red-border)' : 'var(--green-border)'}`,
                          }}
                        >
                          {freezingId === a.id ? '...' : a.status === 'active' ? 'Freeze' : 'Unfreeze'}
                        </button>
                      </td>
                    </tr>
                    {expandedId === a.id && (
                      <tr key={`${a.id}-policy`} className="row-animate" style={{ borderBottom: i < agents.length - 1 ? '1px solid var(--border)' : 'none' }}>
                        <td colSpan={6} className="px-5 pb-5 pt-0">
                          <PolicyPanel agentId={a.id} />
                        </td>
                      </tr>
                    )}
                  </Fragment>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {showCreate && <CreateAgentModal onClose={() => { setShowCreate(false); load(); }} />}
    </div>
  );
}

function PolicyPanel({ agentId }: { agentId: string }) {
  const [policy, setPolicy] = useState<Policy | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [dailyLimit, setDailyLimit] = useState('');
  const [vendors, setVendors] = useState('');
  const [threshold, setThreshold] = useState('');
  const [error, setError] = useState('');
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    async function load() {
      setLoading(true);
      try {
        const p = await getPolicy(agentId);
        setPolicy(p);
        if (p) {
          setDailyLimit((p.daily_limit_cents / 100).toString());
          setVendors(p.allowed_vendors.join(', '));
          setThreshold((p.require_approval_above_cents / 100).toString());
        }
      } catch { /* ignore */ }
      setLoading(false);
    }
    load();
  }, [agentId]);

  async function handleSave() {
    setError('');
    const dl = parseFloat(dailyLimit);
    const th = parseFloat(threshold);
    const v = vendors.split(',').map(s => s.trim()).filter(Boolean);

    if (isNaN(dl) || dl <= 0) { setError('Daily limit must be a positive number'); return; }
    if (isNaN(th) || th <= 0) { setError('Approval threshold must be a positive number'); return; }
    if (v.length === 0) { setError('At least one vendor is required'); return; }

    setSaving(true);
    try {
      await updatePolicy({
        agent_id: agentId,
        daily_limit_cents: Math.round(dl * 100),
        allowed_vendors: v,
        require_approval_above_cents: Math.round(th * 100),
      });
      setSaved(true);
      setTimeout(() => setSaved(false), 1500);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to save');
    }
    setSaving(false);
  }

  if (loading) {
    return <div className="p-4"><div className="skeleton h-24 w-full" /></div>;
  }

  return (
    <div className="rounded-lg p-5 mt-1 flex flex-col gap-4" style={{ background: 'var(--bg)', border: '1px solid var(--border)' }}>
      <h3 style={{ fontFamily: 'var(--font-serif)', fontSize: '16px', fontWeight: 500, color: 'var(--black)' }}>
        Spending Policy
      </h3>

      {!policy && !error && (
        <p className="text-xs" style={{ color: 'var(--muted)' }}>No policy configured. Set one below.</p>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-medium" style={{ color: 'var(--black)' }}>Daily Limit ($)</label>
          <input
            value={dailyLimit}
            onChange={e => setDailyLimit(e.target.value)}
            placeholder="1000.00"
            className="rounded-lg px-3 py-2.5 text-sm outline-none"
            style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', color: 'var(--black)', fontFamily: 'var(--font-mono)' }}
            onFocus={e => (e.target.style.borderColor = 'var(--border-hover)')}
            onBlur={e => (e.target.style.borderColor = 'var(--border)')}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-medium" style={{ color: 'var(--black)' }}>Approval Threshold ($)</label>
          <input
            value={threshold}
            onChange={e => setThreshold(e.target.value)}
            placeholder="500.00"
            className="rounded-lg px-3 py-2.5 text-sm outline-none"
            style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', color: 'var(--black)', fontFamily: 'var(--font-mono)' }}
            onFocus={e => (e.target.style.borderColor = 'var(--border-hover)')}
            onBlur={e => (e.target.style.borderColor = 'var(--border)')}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-medium" style={{ color: 'var(--black)' }}>Allowed Vendors</label>
          <input
            value={vendors}
            onChange={e => setVendors(e.target.value)}
            placeholder="stripe, aws, github"
            className="rounded-lg px-3 py-2.5 text-sm outline-none"
            style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', color: 'var(--black)' }}
            onFocus={e => (e.target.style.borderColor = 'var(--border-hover)')}
            onBlur={e => (e.target.style.borderColor = 'var(--border)')}
          />
        </div>
      </div>

      {error && (
        <p className="text-xs p-3 rounded-lg" style={{ background: 'var(--red-soft)', color: 'var(--red)', border: '1px solid var(--red-border)' }}>{error}</p>
      )}

      <div className="flex justify-end">
        <button
          onClick={handleSave}
          disabled={saving}
          className="btn-lift px-4 py-2 rounded-lg text-sm font-medium disabled:opacity-50"
          style={{ background: saved ? 'var(--green)' : 'var(--black)', color: 'var(--bg)' }}
        >
          {saving ? 'Saving...' : saved ? 'Saved!' : 'Save Policy'}
        </button>
      </div>
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
      className="fixed inset-0 z-50 flex items-center justify-center p-4 modal-backdrop"
      style={{ background: 'rgba(26,26,26,0.4)', backdropFilter: 'blur(4px)' }}
      onClick={(e) => e.target === e.currentTarget && !apiKey && onClose()}
    >
      <div className="w-full max-w-md rounded-xl p-6 flex flex-col gap-5 modal-content" style={{ background: 'var(--bg)', border: '1px solid var(--border)', boxShadow: '0 20px 60px rgba(0,0,0,0.12)' }}>
        {apiKey ? (
          <>
            <h2 style={{ fontFamily: 'var(--font-serif)', fontSize: '22px', fontWeight: 500, color: 'var(--black)' }}>Agent created</h2>
            <p className="text-sm" style={{ color: 'var(--muted)' }}>Copy this API key. It will not be shown again.</p>
            <div className="rounded-lg p-4 text-sm break-all" style={{ fontFamily: 'var(--font-mono)', background: 'var(--black)', color: 'var(--bg)', border: '1px solid var(--border)' }}>
              {apiKey}
            </div>
            <button onClick={onClose} className="btn-lift px-4 py-2.5 rounded-lg text-sm font-medium" style={{ background: 'var(--black)', color: 'var(--bg)' }}>
              Done
            </button>
          </>
        ) : (
          <>
            <div className="flex items-center justify-between">
              <h2 style={{ fontFamily: 'var(--font-serif)', fontSize: '22px', fontWeight: 500, color: 'var(--black)' }}>New Agent</h2>
              <button onClick={onClose} style={{ color: 'var(--muted)' }} className="p-1">
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

            {error && <p className="text-xs p-3 rounded-lg" style={{ background: 'var(--red-soft)', color: 'var(--red)', border: '1px solid var(--red-border)' }}>{error}</p>}

            <div className="flex gap-3">
              <button onClick={onClose} className="btn-lift flex-1 px-4 py-2.5 rounded-lg text-sm" style={{ color: 'var(--mid)', border: '1px solid var(--border)' }}>Cancel</button>
              <button onClick={submit} disabled={loading || !userId || !name} className="btn-lift flex-1 px-4 py-2.5 rounded-lg text-sm font-medium disabled:opacity-50" style={{ background: 'var(--black)', color: 'var(--bg)' }}>
                {loading ? 'Creating...' : 'Create Agent'}
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
      <label className="text-xs font-medium" style={{ color: 'var(--black)' }}>{label}</label>
      <input
        value={value}
        onChange={e => onChange(e.target.value)}
        placeholder={placeholder}
        className="rounded-lg px-3 py-2.5 text-sm outline-none"
        style={{
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          color: 'var(--black)',
          transition: 'border-color 0.2s ease',
        }}
        onFocus={e => (e.target.style.borderColor = 'var(--border-hover)')}
        onBlur={e => (e.target.style.borderColor = 'var(--border)')}
      />
    </div>
  );
}
