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

  const inputStyle = {
    background: 'var(--bg-card)',
    border: '1px solid var(--border)',
    color: 'var(--black)',
    transition: 'border-color 0.2s ease',
  };

  return (
    <div className="p-6 md:p-8 max-w-5xl mx-auto page-enter">
      <div className="mb-8">
        <p className="text-xs font-medium tracking-wider uppercase mb-2" style={{ color: 'var(--muted)', letterSpacing: '2px' }}>History</p>
        <h1 style={{ fontFamily: 'var(--font-serif)', fontSize: 'clamp(28px, 4vw, 36px)', fontWeight: 500, color: 'var(--black)', lineHeight: 1.1 }}>
          Transactions
        </h1>
        <p className="text-sm mt-2" style={{ color: 'var(--mid)' }}>{total} total transactions</p>
      </div>

      {noKey && (
        <div className="mb-6 p-4 rounded-lg text-sm flex items-center gap-3 fade-up" style={{ background: 'var(--gold-soft)', border: '1px solid var(--gold-border)', color: 'var(--gold)' }}>
          <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v4m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
          </svg>
          API key required. Open Settings to configure it.
        </div>
      )}

      {/* Filters */}
      <div className="mb-5 flex flex-wrap gap-3 fade-up">
        <select
          value={status}
          onChange={e => { setStatus(e.target.value); setPage(0); }}
          className="rounded-lg px-3 py-2.5 text-sm outline-none cursor-pointer"
          style={{ ...inputStyle, color: status ? 'var(--black)' : 'var(--muted)' }}
        >
          {STATUS_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
        </select>

        <input
          type="text"
          value={vendor}
          onChange={e => { setVendor(e.target.value); setPage(0); }}
          placeholder="Filter by vendor..."
          className="rounded-lg px-3 py-2.5 text-sm outline-none"
          style={inputStyle}
          onFocus={e => (e.target.style.borderColor = 'var(--border-hover)')}
          onBlur={e => (e.target.style.borderColor = 'var(--border)')}
        />

        <input
          type="date"
          value={fromDate}
          onChange={e => { setFromDate(e.target.value); setPage(0); }}
          className="rounded-lg px-3 py-2.5 text-sm outline-none"
          style={{ ...inputStyle, color: fromDate ? 'var(--black)' : 'var(--muted)' }}
        />

        <input
          type="date"
          value={toDate}
          onChange={e => { setToDate(e.target.value); setPage(0); }}
          className="rounded-lg px-3 py-2.5 text-sm outline-none"
          style={{ ...inputStyle, color: toDate ? 'var(--black)' : 'var(--muted)' }}
        />

        {(status || vendor || fromDate || toDate) && (
          <button
            onClick={() => { setStatus(''); setVendor(''); setFromDate(''); setToDate(''); setPage(0); }}
            className="btn-lift px-3 py-2.5 rounded-lg text-sm"
            style={{ color: 'var(--mid)', border: '1px solid var(--border)' }}
          >
            Clear
          </button>
        )}
      </div>

      <div className="rounded-xl overflow-hidden fade-up" style={{ border: '1px solid var(--border)', animationDelay: '0.1s' }}>
        {loading ? (
          <div className="p-6" style={{ background: 'var(--bg-card)' }}>
            <div className="space-y-3">
              {[1,2,3,4].map(i => <div key={i} className="skeleton h-10 w-full" />)}
            </div>
          </div>
        ) : transactions.length === 0 ? (
          <div className="p-8 text-center text-sm" style={{ background: 'var(--bg-card)', color: 'var(--muted)' }}>No transactions match your filters.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm" style={{ background: 'var(--bg-card)' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border)' }}>
                  {['Vendor', 'Amount', 'Status', 'Reason', 'Agent', 'Date'].map(h => (
                    <th key={h} className="px-5 py-3 text-left text-xs font-medium tracking-wider uppercase" style={{ color: 'var(--muted)', letterSpacing: '1px' }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {transactions.map((t, i) => (
                  <tr
                    key={t.id}
                    className="table-row-hover row-animate"
                    style={{
                      borderBottom: i < transactions.length - 1 ? '1px solid var(--border)' : 'none',
                      animationDelay: `${i * 0.03}s`,
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
                    <td className="px-5 py-3.5 text-xs" style={{ fontFamily: 'var(--font-mono)', color: 'var(--muted)' }}>{t.agent_id.slice(0, 8)}...</td>
                    <td className="px-5 py-3.5 text-xs" style={{ color: 'var(--muted)' }}>{new Date(t.created_at).toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between text-sm fade-up" style={{ color: 'var(--muted)', animationDelay: '0.15s' }}>
          <span style={{ fontSize: '13px' }}>Page {page + 1} of {totalPages}</span>
          <div className="flex gap-2">
            <button
              onClick={() => setPage(p => Math.max(0, p - 1))}
              disabled={page === 0}
              className="btn-lift px-3 py-1.5 rounded-lg text-xs disabled:opacity-30"
              style={{ border: '1px solid var(--border)', color: 'var(--mid)' }}
            >
              Prev
            </button>
            <button
              onClick={() => setPage(p => Math.min(totalPages - 1, p + 1))}
              disabled={page >= totalPages - 1}
              className="btn-lift px-3 py-1.5 rounded-lg text-xs disabled:opacity-30"
              style={{ border: '1px solid var(--border)', color: 'var(--mid)' }}
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
