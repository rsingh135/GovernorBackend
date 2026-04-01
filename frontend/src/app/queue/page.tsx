'use client';

import { useEffect, useState, useCallback } from 'react';
import { getTransactions, approveTransaction, denyTransaction, formatCents } from '@/lib/api';
import { Transaction } from '@/lib/types';
import { ApproveModal } from '@/components/ApproveModal';

export default function QueuePage() {
  const [pending, setPending] = useState<Transaction[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState<Transaction | null>(null);
  const [action, setAction] = useState<'approve' | 'deny'>('approve');
  const [approverUserId, setApproverUserId] = useState('');
  const [noKey, setNoKey] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getTransactions({ status: 'PENDING_APPROVAL', limit: 100 });
      setPending(res.transactions || []);
      setTotal(res.total);
    } catch { /* ignore */ }
    setLoading(false);
  }, []);

  useEffect(() => {
    const settings = JSON.parse(localStorage.getItem('governor_settings') || '{}');
    setApproverUserId(settings.approverUserId || '');
    if (!settings.apiKey) setNoKey(true);
    load();
    const id = setInterval(load, 30_000);
    return () => clearInterval(id);
  }, [load]);

  async function handleDecision(txnId: string, act: 'approve' | 'deny', userId: string) {
    if (act === 'approve') {
      await approveTransaction(txnId, userId);
    } else {
      await denyTransaction(txnId, userId);
    }
    await load();
  }

  function openModal(txn: Transaction, act: 'approve' | 'deny') {
    setSelected(txn);
    setAction(act);
  }

  return (
    <div className="p-8 max-w-5xl mx-auto">
      <div className="mb-8 flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-semibold mb-1" style={{ color: 'var(--text)' }}>Approval Queue</h1>
          <p className="text-sm" style={{ color: 'var(--muted)' }}>
            {total} transaction{total !== 1 ? 's' : ''} pending review
          </p>
        </div>
        <button
          onClick={load}
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors"
          style={{ color: 'var(--muted)', border: '1px solid var(--border)', background: 'transparent' }}
        >
          <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Refresh
        </button>
      </div>

      {noKey && (
        <div className="mb-6 p-4 rounded-xl text-sm flex items-center gap-3" style={{ background: 'rgba(242,185,102,0.08)', border: '1px solid rgba(242,185,102,0.2)', color: 'var(--gold)' }}>
          <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v4m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
          </svg>
          No API key configured. Open Settings to add your key and approver User ID.
        </div>
      )}

      <div className="rounded-xl overflow-hidden" style={{ border: '1px solid var(--border)' }}>
        {loading ? (
          <div className="p-8 text-center text-sm" style={{ background: 'var(--panel)', color: 'var(--muted)' }}>Loading…</div>
        ) : pending.length === 0 ? (
          <div className="p-16 text-center flex flex-col items-center gap-3" style={{ background: 'var(--panel)' }}>
            <div className="w-12 h-12 rounded-full flex items-center justify-center" style={{ background: 'rgba(52,211,153,0.12)' }}>
              <svg width="22" height="22" fill="none" stroke="var(--green)" strokeWidth="2" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <p className="text-sm font-medium" style={{ color: 'var(--text)' }}>All clear</p>
            <p className="text-xs" style={{ color: 'var(--muted)' }}>No transactions pending approval</p>
          </div>
        ) : (
          <table className="w-full text-sm" style={{ background: 'var(--panel)' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--border)' }}>
                {['Vendor', 'Amount', 'Reason', 'Agent', 'Submitted', 'Actions'].map(h => (
                  <th key={h} className="px-5 py-3 text-left text-xs font-medium tracking-wider" style={{ color: 'var(--muted)' }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {pending.map((t, i) => (
                <tr key={t.id} style={{ borderBottom: i < pending.length - 1 ? '1px solid var(--border)' : 'none' }}>
                  <td className="px-5 py-3 font-medium" style={{ color: 'var(--text)' }}>{t.vendor}</td>
                  <td className="px-5 py-3 font-semibold" style={{ color: 'var(--gold)' }}>{formatCents(t.amount_cents)}</td>
                  <td className="px-5 py-3 text-xs" style={{ color: 'var(--muted)' }}>{t.reason}</td>
                  <td className="px-5 py-3 font-mono text-xs" style={{ color: 'var(--muted)' }}>{t.agent_id.slice(0, 8)}…</td>
                  <td className="px-5 py-3 text-xs" style={{ color: 'var(--muted)' }}>
                    {new Date(t.created_at).toLocaleString()}
                  </td>
                  <td className="px-5 py-3">
                    <div className="flex gap-2">
                      <button
                        onClick={() => openModal(t, 'approve')}
                        className="px-3 py-1 rounded-md text-xs font-medium transition-colors"
                        style={{ background: 'rgba(52,211,153,0.12)', color: 'var(--green)', border: '1px solid rgba(52,211,153,0.2)' }}
                      >
                        Approve
                      </button>
                      <button
                        onClick={() => openModal(t, 'deny')}
                        className="px-3 py-1 rounded-md text-xs font-medium transition-colors"
                        style={{ background: 'rgba(248,113,113,0.12)', color: 'var(--red)', border: '1px solid rgba(248,113,113,0.2)' }}
                      >
                        Deny
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <ApproveModal
        transaction={selected}
        action={action}
        approverUserId={approverUserId}
        onConfirm={handleDecision}
        onClose={() => setSelected(null)}
      />
    </div>
  );
}
