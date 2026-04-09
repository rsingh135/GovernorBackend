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
      <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
      </svg>
    ),
  },
  {
    href: '/agents',
    label: 'Agents',
    icon: (
      <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17H4a2 2 0 01-2-2V5a2 2 0 012-2h16a2 2 0 012 2v10a2 2 0 01-2 2h-1" />
      </svg>
    ),
  },
  {
    href: '/users',
    label: 'Users',
    icon: (
      <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2M9 11a4 4 0 100-8 4 4 0 000 8zM23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75" />
      </svg>
    ),
  },
  {
    href: '/queue',
    label: 'Approval Queue',
    icon: (
      <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
    badge: true,
  },
  {
    href: '/transactions',
    label: 'Transactions',
    icon: (
      <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
    ),
  },
  {
    href: '/webhooks',
    label: 'Webhooks',
    icon: (
      <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101M10.172 13.828a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
      </svg>
    ),
  },
];

export function Sidebar() {
  const pathname = usePathname();
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [pendingCount, setPendingCount] = useState<number | null>(null);
  const [isDemo, setIsDemo] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

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

  useEffect(() => {
    setMobileOpen(false);
  }, [pathname]);

  const sidebarContent = (
    <>
      {/* Logo */}
      <div className="px-5 py-5" style={{ borderBottom: '1px solid var(--border)' }}>
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: 'var(--black)' }}>
            <svg width="16" height="16" viewBox="0 0 56 56" fill="none">
              <path d="M18 28C18 22.477 22.477 18 28 18C33.523 18 38 22.477 38 28C38 33.523 33.523 38 28 38" stroke="var(--bg)" strokeWidth="4" strokeLinecap="round"/>
              <path d="M28 24V28L31 31" stroke="var(--bg)" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"/>
              <circle cx="21" cy="37" r="3.5" fill="var(--bg)" opacity="0.6"/>
            </svg>
          </div>
          <span style={{ fontFamily: 'var(--font-serif)', fontSize: '18px', fontWeight: 500, color: 'var(--black)', letterSpacing: '-0.3px' }}>
            governor
          </span>
        </div>
      </div>

      {/* Demo banner */}
      {isDemo && (
        <div
          className="mx-3 mt-3 px-3 py-2 rounded-lg text-center"
          style={{
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            fontFamily: 'var(--font-mono)',
            fontSize: '10px',
            fontWeight: 500,
            letterSpacing: '1.5px',
            textTransform: 'uppercase' as const,
            color: 'var(--muted)',
          }}
        >
          demo mode
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
              className={`nav-item flex items-center gap-2.5 px-3 py-2.5 rounded-lg text-sm ${active ? 'nav-item-active' : ''}`}
              style={{
                color: active ? 'var(--black)' : 'var(--muted)',
                background: active ? 'var(--bg-card)' : 'transparent',
                fontWeight: active ? 500 : 400,
              }}
            >
              <span style={{ color: active ? 'var(--black)' : 'var(--light)', opacity: active ? 1 : 0.8 }}>{item.icon}</span>
              <span className="flex-1">{item.label}</span>
              {item.badge && pendingCount !== null && pendingCount > 0 && (
                <span
                  className="badge-pop text-xs font-semibold rounded-full px-1.5 py-0.5 min-w-5 text-center"
                  style={{ background: 'var(--red)', color: 'var(--bg)', fontSize: '10px' }}
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
          className="nav-item w-full flex items-center gap-2.5 px-3 py-2.5 rounded-lg text-sm"
          style={{ color: 'var(--muted)' }}
        >
          <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          Settings
        </button>
      </div>
    </>
  );

  return (
    <>
      {/* Mobile menu button */}
      <button
        onClick={() => setMobileOpen(true)}
        className="fixed top-4 left-4 z-50 p-2 rounded-lg md:hidden no-print"
        style={{ background: 'var(--bg-card)', border: '1px solid var(--border)' }}
      >
        <svg width="20" height="20" fill="none" stroke="var(--black)" strokeWidth="1.8" viewBox="0 0 24 24">
          <path strokeLinecap="round" d="M4 6h16M4 12h16M4 18h16"/>
        </svg>
      </button>

      {/* Desktop sidebar */}
      <aside
        className="sidebar-desktop w-56 flex-shrink-0 flex-col h-full hidden md:flex no-print"
        style={{ background: 'var(--bg)', borderRight: '1px solid var(--border)' }}
      >
        {sidebarContent}
      </aside>

      {/* Mobile sidebar overlay */}
      {mobileOpen && (
        <div
          className="fixed inset-0 z-40 md:hidden modal-backdrop"
          style={{ background: 'rgba(0,0,0,0.3)' }}
          onClick={() => setMobileOpen(false)}
        >
          <aside
            className="w-56 h-full flex flex-col slide-left"
            style={{ background: 'var(--bg)', borderRight: '1px solid var(--border)' }}
            onClick={(e) => e.stopPropagation()}
          >
            {sidebarContent}
          </aside>
        </div>
      )}

      <SettingsModal open={settingsOpen} onClose={() => setSettingsOpen(false)} />
    </>
  );
}
