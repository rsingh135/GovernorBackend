'use client';

import { useEffect, useState } from 'react';
import { getTransactions, formatCents } from '@/lib/api';
import { Transaction } from '@/lib/types';

interface DayData {
  label: string;
  total: number;
}

interface VendorData {
  vendor: string;
  total: number;
}

export function SpendChart() {
  const [days, setDays] = useState<DayData[]>([]);
  const [vendors, setVendors] = useState<VendorData[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      setLoading(true);
      try {
        const sevenDaysAgo = new Date();
        sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);
        sevenDaysAgo.setHours(0, 0, 0, 0);

        const res = await getTransactions({
          from_date: sevenDaysAgo.toISOString(),
          limit: 200,
        });

        const txns = (res.transactions || []).filter((t: Transaction) => t.status === 'APPROVED');

        // Aggregate by day
        const dayMap: Record<string, number> = {};
        const vendorMap: Record<string, number> = {};

        // Initialize last 7 days
        for (let i = 6; i >= 0; i--) {
          const d = new Date();
          d.setDate(d.getDate() - i);
          const key = d.toISOString().slice(0, 10);
          dayMap[key] = 0;
        }

        txns.forEach((t: Transaction) => {
          const key = new Date(t.created_at).toISOString().slice(0, 10);
          if (dayMap[key] !== undefined) {
            dayMap[key] += t.amount_cents;
          }
          vendorMap[t.vendor] = (vendorMap[t.vendor] || 0) + t.amount_cents;
        });

        const dayData = Object.entries(dayMap).map(([date, total]) => ({
          label: new Date(date + 'T12:00:00').toLocaleDateString('en-US', { weekday: 'short' }),
          total,
        }));

        const vendorData = Object.entries(vendorMap)
          .sort((a, b) => b[1] - a[1])
          .slice(0, 5)
          .map(([vendor, total]) => ({ vendor, total }));

        setDays(dayData);
        setVendors(vendorData);
      } catch { /* ignore */ }
      setLoading(false);
    }
    load();
  }, []);

  if (loading) {
    return (
      <div className="rounded-xl p-6 card-hover" style={{ background: 'var(--bg-card)', border: '1px solid var(--border)' }}>
        <div className="space-y-3">
          <div className="skeleton h-6 w-32" />
          <div className="skeleton h-40 w-full" />
        </div>
      </div>
    );
  }

  const maxDay = Math.max(...days.map(d => d.total), 1);
  const maxVendor = Math.max(...vendors.map(v => v.total), 1);

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
      {/* Daily spend bar chart */}
      <div className="lg:col-span-2 rounded-xl p-5 card-hover" style={{ background: 'var(--bg-card)', border: '1px solid var(--border)' }}>
        <p className="text-xs font-medium tracking-wider uppercase mb-1" style={{ color: 'var(--muted)', letterSpacing: '2px' }}>
          Spend Trend
        </p>
        <h3 style={{ fontFamily: 'var(--font-serif)', fontSize: '18px', fontWeight: 500, color: 'var(--black)', marginBottom: '16px' }}>
          Daily Approved Spend
        </h3>

        <div className="flex items-end gap-2" style={{ height: '140px' }}>
          {days.map((d, i) => {
            const pct = maxDay > 0 ? (d.total / maxDay) * 100 : 0;
            return (
              <div key={i} className="flex-1 flex flex-col items-center gap-1.5">
                <span className="text-xs" style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: 'var(--muted)' }}>
                  {d.total > 0 ? formatCents(d.total) : '—'}
                </span>
                <div className="w-full relative" style={{ height: '100px' }}>
                  <div
                    className="w-full absolute bottom-0 rounded-t-md"
                    style={{
                      height: `${Math.max(pct, 2)}%`,
                      background: d.total > 0 ? 'var(--black)' : 'var(--border)',
                      opacity: d.total > 0 ? 0.85 : 0.4,
                      transition: 'height 0.6s cubic-bezier(0.22, 1, 0.36, 1)',
                      transitionDelay: `${i * 0.05}s`,
                    }}
                  />
                </div>
                <span className="text-xs" style={{ color: 'var(--muted)', fontSize: '11px' }}>{d.label}</span>
              </div>
            );
          })}
        </div>
      </div>

      {/* Top vendors */}
      <div className="rounded-xl p-5 card-hover" style={{ background: 'var(--bg-card)', border: '1px solid var(--border)' }}>
        <p className="text-xs font-medium tracking-wider uppercase mb-1" style={{ color: 'var(--muted)', letterSpacing: '2px' }}>
          Top Vendors
        </p>
        <h3 style={{ fontFamily: 'var(--font-serif)', fontSize: '18px', fontWeight: 500, color: 'var(--black)', marginBottom: '16px' }}>
          By Total Spend
        </h3>

        {vendors.length === 0 ? (
          <p className="text-xs" style={{ color: 'var(--muted)' }}>No approved transactions yet.</p>
        ) : (
          <div className="flex flex-col gap-3">
            {vendors.map((v, i) => {
              const pct = (v.total / maxVendor) * 100;
              return (
                <div key={i}>
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-sm font-medium" style={{ color: 'var(--black)' }}>{v.vendor}</span>
                    <span className="text-xs" style={{ fontFamily: 'var(--font-mono)', color: 'var(--muted)' }}>{formatCents(v.total)}</span>
                  </div>
                  <div className="w-full rounded-full" style={{ height: '4px', background: 'var(--border)' }}>
                    <div
                      className="rounded-full"
                      style={{
                        width: `${pct}%`,
                        height: '4px',
                        background: 'var(--black)',
                        opacity: 0.7,
                        transition: 'width 0.6s cubic-bezier(0.22, 1, 0.36, 1)',
                        transitionDelay: `${i * 0.08}s`,
                      }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
