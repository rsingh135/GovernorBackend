'use client';

import { useEffect, useState, useCallback } from 'react';
import { getTransactions, formatCents } from '@/lib/api';
import { Transaction } from '@/lib/types';
import { Badge } from '@/components/Badge';

const PAGE_SIZE = 20;

const STATUS_OPTIONS = [
  { value: '', label: 'All statuses' },
  { value: 'APPROVED', label: 'Approved' },
  { value: 'DENIED', label: 'Denied' },
  { value: 'PENDING_APPROVAL', label: 'Pending' },
];

export default function TransactionsPage() {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [status, setStatus] = useState('');
  const [vendor, setVendor] = useState('');
  const [fromDate, setFromDate] = useState('');
  const [toDate, setToDate] = useState('');
  const [page, setPage] = useState(0);
  const [noKey, setNoKey] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getTransactions({
        status: status || undefined,
        from_date: fromDate ? new Date(fromDate).toISOString() : undefined,
        to_date: toDate ? new Date(toDate + 'T23:59:59').toISOString() : undefined,
        limit: PAGE_SIZE,
        offset: page * PAGE_SIZE,
      });
      let txns = res.transactions || [];
      if (vendor) txns = txns.filter(t => t.vendor.toLowerCase().includes(vendor.toLowerCase()));
      setTransactions(txns);
      setTotal(res.total);
    } catch { /* ignore */ }
    setLoading(false);
  }, [status, fromDate, toDate, page, vendor]);

  useEffect(() => {
    const settings = JSON.parse(localStorage.getItem('governor_settings') || '{}');
    if (!settings.apiKey) setNoKey(true);
    load();
  }, [load]);

  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <div className="p-8 max-w-5xl mx-auto">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold mb-1" style={{ color: 'var(--text)' }}>Transactions</h1>
        <p className="text-sm" style={{ color: 'var(--muted)' }}>{total} total transactions</p>
      </div>

      {noKey && (
        <div className="mb-6 p-4 rounded-xl text-sm flex items-center gap-3" style={{ background: 'rgba(242,185,102,0.08)', border: '1px solid rgba(242,185,102,0.2)', color: 'var(--gold)' }}>
          <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v4m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
          </svg>
          API key required. Open Settings to configure it.
        </div>
      )}

      {/* Filters */}
      <div className="mb-5 flex flex-wrap gap-3">
        <select
          value={status}
          onChange={e => { setStatus(e.target.value); setPage(0); }}
          className="rounded-lg px-3 py-2 text-sm outline-none"
          style={{ background: 'var(--panel)', border: '1px solid var(--border)', color: status ? 'var(--text)' : 'var(--muted)' }}
        >
          {STATUS_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
        </select>

        <input
          type="text"
          value={vendor}
          onChange={e => { setVendor(e.target.value); setPage(0); }}
          placeholder="Filter by vendor…"
          className="rounded-lg px-3 py-2 text-sm outline-none"
          style={{ background: 'var(--panel)', border: '1px solid var(--border)', color: 'var(--text)' }}
          onFocus={e => (e.target.style.borderColor = 'var(--accent)')}
          onBlur={e => (e.target.style.borderColor = 'var(--border)')}
        />

        <input
          type="date"
          value={fromDate}
          onChange={e => { setFromDate(e.target.value); setPage(0); }}
          className="rounded-lg px-3 py-2 text-sm outline-none"
          style={{ background: 'var(--panel)', border: '1px solid var(--border)', color: fromDate ? 'var(--text)' : 'var(--muted)', colorScheme: 'dark' }}
        />

        <input
          type="date"
          value={toDate}
          onChange={e => { setToDate(e.target.value); setPage(0); }}
          className="rounded-lg px-3 py-2 text-sm outline-none"
          style={{ background: 'var(--panel)', border: '1px solid var(--border)', color: toDate ? 'var(--text)' : 'var(--muted)', colorScheme: 'dark' }}
        />

        {(status || vendor || fromDate || toDate) && (
          <button
            onClick={() => { setStatus(''); setVendor(''); setFromDate(''); setToDate(''); setPage(0); }}
            className="px-3 py-2 rounded-lg text-sm transition-colors"
            style={{ color: 'var(--muted)', border: '1px solid var(--border)' }}
          >
            Clear
          </button>
        )}
      </div>

      <div className="rounded-xl overflow-hidden" style={{ border: '1px solid var(--border)' }}>
        {loading ? (
          <div className="p-8 text-center text-sm" style={{ background: 'var(--panel)', color: 'var(--muted)' }}>Loading…</div>
        ) : transactions.length === 0 ? (
          <div className="p-8 text-center text-sm" style={{ background: 'var(--panel)', color: 'var(--muted)' }}>No transactions match your filters.</div>
        ) : (
          <table className="w-full text-sm" style={{ background: 'var(--panel)' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--border)' }}>
                {['Vendor', 'Amount', 'Status', 'Reason', 'Agent', 'Date'].map(h => (
                  <th key={h} className="px-5 py-3 text-left text-xs font-medium tracking-wider" style={{ color: 'var(--muted)' }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {transactions.map((t, i) => (
                <tr key={t.id} style={{ borderBottom: i < transactions.length - 1 ? '1px solid var(--border)' : 'none' }}>
                  <td className="px-5 py-3 font-medium" style={{ color: 'var(--text)' }}>{t.vendor}</td>
                  <td className="px-5 py-3 font-medium" style={{ color: t.status === 'APPROVED' ? 'var(--green)' : t.status === 'DENIED' ? 'var(--red)' : 'var(--gold)' }}>
                    {formatCents(t.amount_cents)}
                  </td>
                  <td className="px-5 py-3"><Badge status={t.status} /></td>
                  <td className="px-5 py-3 text-xs" style={{ color: 'var(--muted)' }}>{t.reason}</td>
                  <td className="px-5 py-3 font-mono text-xs" style={{ color: 'var(--muted)' }}>{t.agent_id.slice(0, 8)}…</td>
                  <td className="px-5 py-3 text-xs" style={{ color: 'var(--muted)' }}>{new Date(t.created_at).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between text-sm" style={{ color: 'var(--muted)' }}>
          <span>Page {page + 1} of {totalPages}</span>
          <div className="flex gap-2">
            <button
              onClick={() => setPage(p => Math.max(0, p - 1))}
              disabled={page === 0}
              className="px-3 py-1.5 rounded-lg text-xs disabled:opacity-30 transition-colors"
              style={{ border: '1px solid var(--border)', color: 'var(--muted)' }}
            >
              ← Prev
            </button>
            <button
              onClick={() => setPage(p => Math.min(totalPages - 1, p + 1))}
              disabled={page >= totalPages - 1}
              className="px-3 py-1.5 rounded-lg text-xs disabled:opacity-30 transition-colors"
              style={{ border: '1px solid var(--border)', color: 'var(--muted)' }}
            >
              Next →
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
