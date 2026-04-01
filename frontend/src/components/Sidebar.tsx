'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useState, useEffect } from 'react';
import { SettingsModal } from './SettingsModal';
import { DEMO_KEY } from '@/lib/mockData';

const navItems = [
  {
    href: '/',
    label: 'Overview',
    icon: (
      <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
      </svg>
    ),
  },
  {
    href: '/agents',
    label: 'Agents',
    icon: (
      <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17H4a2 2 0 01-2-2V5a2 2 0 012-2h16a2 2 0 012 2v10a2 2 0 01-2 2h-1" />
      </svg>
    ),
  },
  {
    href: '/queue',
    label: 'Approval Queue',
    icon: (
      <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
    badge: true,
  },
  {
    href: '/transactions',
    label: 'Transactions',
    icon: (
      <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
    ),
  },
];

export function Sidebar() {
  const pathname = usePathname();
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [pendingCount, setPendingCount] = useState<number | null>(null);
  const [isDemo, setIsDemo] = useState(false);

  useEffect(() => {
    async function fetchPending() {
      try {
        const settings = JSON.parse(localStorage.getItem('governor_settings') || '{}');
        setIsDemo(settings.apiKey === DEMO_KEY);
        if (!settings.apiKey) return;
        if (settings.apiKey === DEMO_KEY) { setPendingCount(3); return; }
        const base = settings.apiBaseUrl || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
        const res = await fetch(`${base}/transactions?status=PENDING_APPROVAL&limit=1`, {
          headers: { 'X-API-Key': settings.apiKey },
        });
        if (res.ok) {
          const data = await res.json();
          setPendingCount(data.total ?? 0);
        }
      } catch { /* ignore */ }
    }
    fetchPending();
    const id = setInterval(fetchPending, 30_000);
    return () => clearInterval(id);
  }, []);

  return (
    <>
      <aside
        className="w-56 flex-shrink-0 flex flex-col h-full"
        style={{ background: 'var(--panel)', borderRight: '1px solid var(--border)' }}
      >
        {/* Logo */}
        <div className="px-5 py-5" style={{ borderBottom: '1px solid var(--border)' }}>
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-lg flex items-center justify-center" style={{ background: 'var(--accent)' }}>
              <svg width="14" height="14" fill="white" viewBox="0 0 24 24">
                <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="white" strokeWidth="2" fill="none" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </div>
            <span className="font-semibold text-sm" style={{ color: 'var(--text)' }}>Governor</span>
          </div>
        </div>

        {/* Demo banner */}
        {isDemo && (
          <div className="mx-3 mt-3 px-3 py-2 rounded-lg text-xs text-center font-medium" style={{ background: 'rgba(62,130,255,0.1)', color: 'var(--accent)', border: '1px solid rgba(62,130,255,0.2)' }}>
            Demo mode
          </div>
        )}

        {/* Nav */}
        <nav className="flex-1 px-3 py-4 flex flex-col gap-0.5">
          {navItems.map((item) => {
            const active = pathname === item.href || (item.href !== '/' && pathname.startsWith(item.href));
            return (
              <Link
                key={item.href}
                href={item.href}
                className="flex items-center gap-2.5 px-2.5 py-2 rounded-lg text-sm transition-colors"
                style={{
                  color: active ? 'var(--text)' : 'var(--muted)',
                  background: active ? 'rgba(62,130,255,0.12)' : 'transparent',
                  fontWeight: active ? '500' : '400',
                }}
              >
                <span style={{ color: active ? 'var(--accent)' : 'var(--muted)' }}>{item.icon}</span>
                <span className="flex-1">{item.label}</span>
                {item.badge && pendingCount !== null && pendingCount > 0 && (
                  <span
                    className="text-xs font-semibold rounded-full px-1.5 py-0.5 min-w-5 text-center"
                    style={{ background: 'var(--gold)', color: '#0a0a0a' }}
                  >
                    {pendingCount > 99 ? '99+' : pendingCount}
                  </span>
                )}
              </Link>
            );
          })}
        </nav>

        {/* Settings footer */}
        <div className="px-3 py-4" style={{ borderTop: '1px solid var(--border)' }}>
          <button
            onClick={() => setSettingsOpen(true)}
            className="w-full flex items-center gap-2.5 px-2.5 py-2 rounded-lg text-sm transition-colors hover:text-white"
            style={{ color: 'var(--muted)' }}
          >
            <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            Settings
          </button>
        </div>
      </aside>

      <SettingsModal open={settingsOpen} onClose={() => setSettingsOpen(false)} />
    </>
  );
}
