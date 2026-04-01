'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { StatCard } from '@/components/StatCard';
import { Badge } from '@/components/Badge';
import { getAgents, getTransactions, formatCents } from '@/lib/api';
import { Transaction } from '@/lib/types';

export default function OverviewPage() {
  const [agentCount, setAgentCount] = useState<number | null>(null);
  const [pendingCount, setPendingCount] = useState<number | null>(null);
  const [todayApproved, setTodayApproved] = useState<number | null>(null);
  const [todayDenied, setTodayDenied] = useState<number | null>(null);
  const [recent, setRecent] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [noKey, setNoKey] = useState(false);

  useEffect(() => {
    async function load() {
      setLoading(true);
      const settings = JSON.parse(localStorage.getItem('governor_settings') || '{}');
      if (!settings.apiKey) setNoKey(true);

      const today = new Date();
      today.setHours(0, 0, 0, 0);

      const [agents, pending, todayTxns, recentTxns] = await Promise.allSettled([
        getAgents({ limit: 1 }),
        getTransactions({ status: 'PENDING_APPROVAL', limit: 1 }),
        getTransactions({ from_date: today.toISOString(), limit: 100 }),
        getTransactions({ limit: 10 }),
      ]);

      if (agents.status === 'fulfilled') setAgentCount(agents.value.total);
      if (pending.status === 'fulfilled') setPendingCount(pending.value.total);
      if (todayTxns.status === 'fulfilled') {
        const txns = todayTxns.value.transactions || [];
        setTodayApproved(txns.filter(t => t.status === 'APPROVED').reduce((s, t) => s + t.amount_cents, 0));
        setTodayDenied(txns.filter(t => t.status === 'DENIED').length);
      }
      if (recentTxns.status === 'fulfilled') setRecent(recentTxns.value.transactions || []);

      setLoading(false);
    }
    load();
  }, []);

  return (
    <div className="p-8 max-w-5xl mx-auto">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold mb-1" style={{ color: 'var(--text)' }}>Overview</h1>
        <p className="text-sm" style={{ color: 'var(--muted)' }}>AI agent spend activity at a glance</p>
      </div>

      {noKey && (
        <div className="mb-6 p-4 rounded-xl text-sm flex items-center gap-3" style={{ background: 'rgba(242,185,102,0.08)', border: '1px solid rgba(242,185,102,0.2)', color: 'var(--gold)' }}>
          <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v4m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
          </svg>
          No API key configured — some data may be unavailable. Open Settings to add your key.
        </div>
      )}

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard label="Total Agents" value={loading ? '—' : (agentCount ?? '—')} sub="provisioned" accent="blue" href="/agents" />
        <StatCard label="Pending Approvals" value={loading ? '—' : (pendingCount ?? '—')} sub="awaiting review" accent="gold" href="/queue" />
        <StatCard label="Today's Approved" value={loading ? '—' : (todayApproved != null ? formatCents(todayApproved) : '—')} sub="total spend" accent="green" />
        <StatCard label="Today's Denied" value={loading ? '—' : (todayDenied ?? '—')} sub="transactions blocked" accent="red" />
      </div>

      <div className="rounded-xl overflow-hidden" style={{ border: '1px solid var(--border)' }}>
        <div className="px-5 py-4 flex items-center justify-between" style={{ background: 'var(--panel)', borderBottom: '1px solid var(--border)' }}>
          <h2 className="text-sm font-semibold" style={{ color: 'var(--text)' }}>Recent Transactions</h2>
          <Link href="/transactions" className="text-xs" style={{ color: 'var(--accent)' }}>View all →</Link>
        </div>

        {loading ? (
          <div className="p-8 text-center text-sm" style={{ background: 'var(--panel)', color: 'var(--muted)' }}>Loading…</div>
        ) : recent.length === 0 ? (
          <div className="p-8 text-center text-sm" style={{ background: 'var(--panel)', color: 'var(--muted)' }}>No transactions yet.</div>
        ) : (
          <table className="w-full text-sm" style={{ background: 'var(--panel)' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--border)' }}>
                {['Vendor', 'Amount', 'Status', 'Reason', 'Date'].map(h => (
                  <th key={h} className="px-5 py-3 text-left text-xs font-medium tracking-wider" style={{ color: 'var(--muted)' }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {recent.map((t, i) => (
                <tr key={t.id} style={{ borderBottom: i < recent.length - 1 ? '1px solid var(--border)' : 'none' }}>
                  <td className="px-5 py-3 font-medium" style={{ color: 'var(--text)' }}>{t.vendor}</td>
                  <td className="px-5 py-3 font-medium" style={{ color: t.status === 'APPROVED' ? 'var(--green)' : t.status === 'DENIED' ? 'var(--red)' : 'var(--gold)' }}>
                    {formatCents(t.amount_cents)}
                  </td>
                  <td className="px-5 py-3"><Badge status={t.status} /></td>
                  <td className="px-5 py-3 text-xs" style={{ color: 'var(--muted)' }}>{t.reason}</td>
                  <td className="px-5 py-3 text-xs" style={{ color: 'var(--muted)' }}>{new Date(t.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
