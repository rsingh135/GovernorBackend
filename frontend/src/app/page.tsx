'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { StatCard } from '@/components/StatCard';
import { Badge } from '@/components/Badge';
import { SpendChart } from '@/components/SpendChart';
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

  const loadingPlaceholder = '\u2014';

  return (
    <div className="p-6 md:p-8 max-w-5xl mx-auto page-enter">
      <div className="mb-8">
        <p className="text-xs font-medium tracking-wider uppercase mb-2" style={{ color: 'var(--muted)', letterSpacing: '2px' }}>Dashboard</p>
        <h1 style={{ fontFamily: 'var(--font-serif)', fontSize: 'clamp(28px, 4vw, 36px)', fontWeight: 500, color: 'var(--black)', lineHeight: 1.1 }}>
          Overview
        </h1>
        <p className="text-sm mt-2" style={{ color: 'var(--mid)' }}>AI agent spend activity at a glance.</p>
      </div>

      {noKey && (
        <div className="mb-6 p-4 rounded-lg text-sm flex items-center gap-3 fade-up" style={{ background: 'var(--gold-soft)', border: '1px solid var(--gold-border)', color: 'var(--gold)' }}>
          <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v4m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
          </svg>
          No API key configured. Open Settings to add your key.
        </div>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <div className="fade-up fade-up-d1">
          <StatCard label="Total Agents" value={loading ? loadingPlaceholder : (agentCount ?? loadingPlaceholder)} sub="provisioned" accent="default" href="/agents" />
        </div>
        <div className="fade-up fade-up-d2">
          <StatCard label="Pending Approvals" value={loading ? loadingPlaceholder : (pendingCount ?? loadingPlaceholder)} sub="awaiting review" accent="gold" href="/queue" />
        </div>
        <div className="fade-up fade-up-d3">
          <StatCard label="Today's Approved" value={loading ? loadingPlaceholder : (todayApproved != null ? formatCents(todayApproved) : loadingPlaceholder)} sub="total spend" accent="green" />
        </div>
        <div className="fade-up fade-up-d4">
          <StatCard label="Today's Denied" value={loading ? loadingPlaceholder : (todayDenied ?? loadingPlaceholder)} sub="transactions blocked" accent="red" />
        </div>
      </div>

      {/* Spend Analytics Charts */}
      <div className="mb-8 fade-up" style={{ animationDelay: '0.15s' }}>
        <SpendChart />
      </div>

      <div className="rounded-xl overflow-hidden fade-up" style={{ border: '1px solid var(--border)', animationDelay: '0.2s' }}>
        <div className="px-5 py-4 flex items-center justify-between" style={{ background: 'var(--bg-card)', borderBottom: '1px solid var(--border)' }}>
          <h2 className="text-sm font-semibold" style={{ color: 'var(--black)' }}>Recent Transactions</h2>
          <Link href="/transactions" className="text-xs font-medium" style={{ color: 'var(--mid)', textDecoration: 'none' }}>
            View all &rarr;
          </Link>
        </div>

        {loading ? (
          <div className="p-6" style={{ background: 'var(--bg-card)' }}>
            <div className="space-y-3">
              {[1,2,3].map(i => <div key={i} className="skeleton h-10 w-full" />)}
            </div>
          </div>
        ) : recent.length === 0 ? (
          <div className="p-8 text-center text-sm" style={{ background: 'var(--bg-card)', color: 'var(--muted)' }}>No transactions yet.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm" style={{ background: 'var(--bg-card)' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border)' }}>
                  {['Vendor', 'Amount', 'Status', 'Reason', 'Date'].map(h => (
                    <th key={h} className="px-5 py-3 text-left text-xs font-medium tracking-wider uppercase" style={{ color: 'var(--muted)', letterSpacing: '1px' }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {recent.map((t, i) => (
                  <tr
                    key={t.id}
                    className="table-row-hover row-animate"
                    style={{
                      borderBottom: i < recent.length - 1 ? '1px solid var(--border)' : 'none',
                      animationDelay: `${i * 0.04}s`,
                    }}
                  >
                    <td className="px-5 py-3.5 font-medium" style={{ color: 'var(--black)' }}>{t.vendor}</td>
                    <td className="px-5 py-3.5 font-medium" style={{
                      fontFamily: 'var(--font-mono)',
                      fontSize: '13px',
                      color: t.status === 'APPROVED' ? 'var(--green)' : t.status === 'DENIED' ? 'var(--red)' : 'var(--gold)',
                    }}>
                      {formatCents(t.amount_cents)}
                    </td>
                    <td className="px-5 py-3.5"><Badge status={t.status} /></td>
                    <td className="px-5 py-3.5 text-xs" style={{ color: 'var(--muted)' }}>{t.reason}</td>
                    <td className="px-5 py-3.5 text-xs" style={{ color: 'var(--muted)' }}>{new Date(t.created_at).toLocaleDateString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
