'use client';

import { useEffect, useState } from 'react';
import { Badge } from '@/components/Badge';
import { getWebhooks, registerWebhook, deleteWebhook } from '@/lib/api';
import { Webhook } from '@/lib/types';

const EVENT_OPTIONS = [
  { value: 'transaction.approved', label: 'transaction.approved' },
  { value: 'transaction.denied', label: 'transaction.denied' },
  { value: 'transaction.pending', label: 'transaction.pending' },
];

export default function WebhooksPage() {
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [loading, setLoading] = useState(true);
  const [showRegister, setShowRegister] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  async function load() {
    setLoading(true);
    try {
      const res = await getWebhooks();
      setWebhooks(res.webhooks || []);
    } catch { /* ignore */ }
    setLoading(false);
  }

  useEffect(() => { load(); }, []);

  async function handleDelete(id: string) {
    if (confirmDeleteId !== id) {
      setConfirmDeleteId(id);
      return;
    }
    setDeletingId(id);
    try {
      await deleteWebhook(id);
      await load();
    } catch { /* ignore */ }
    setDeletingId(null);
    setConfirmDeleteId(null);
  }

  return (
    <div className="p-6 md:p-8 max-w-5xl mx-auto page-enter">
      <div className="mb-8 flex items-start justify-between">
        <div>
          <p className="text-xs font-medium tracking-wider uppercase mb-2" style={{ color: 'var(--muted)', letterSpacing: '2px' }}>Integrations</p>
          <h1 style={{ fontFamily: 'var(--font-serif)', fontSize: 'clamp(28px, 4vw, 36px)', fontWeight: 500, color: 'var(--black)', lineHeight: 1.1 }}>
            Webhooks
          </h1>
          <p className="text-sm mt-2" style={{ color: 'var(--mid)' }}>{webhooks.length} webhook{webhooks.length !== 1 ? 's' : ''} registered</p>
        </div>
        <button
          onClick={() => setShowRegister(true)}
          className="btn-lift px-4 py-2.5 rounded-lg text-sm font-medium"
          style={{ background: 'var(--black)', color: 'var(--bg)' }}
        >
          + Register Webhook
        </button>
      </div>

      <div className="rounded-xl overflow-hidden fade-up" style={{ border: '1px solid var(--border)' }}>
        {loading ? (
          <div className="p-6" style={{ background: 'var(--bg-card)' }}>
            <div className="space-y-3">
              {[1,2,3].map(i => <div key={i} className="skeleton h-10 w-full" />)}
            </div>
          </div>
        ) : webhooks.length === 0 ? (
          <div className="p-12 text-center flex flex-col items-center gap-3" style={{ background: 'var(--bg-card)' }}>
            <div className="w-12 h-12 rounded-full flex items-center justify-center" style={{ background: 'rgba(0,0,0,0.05)' }}>
              <svg width="22" height="22" fill="none" stroke="var(--muted)" strokeWidth="1.8" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101M10.172 13.828a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
            </div>
            <p className="text-sm font-medium" style={{ color: 'var(--black)' }}>No webhooks registered</p>
            <p className="text-xs" style={{ color: 'var(--muted)' }}>Register a webhook to receive transaction events.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm" style={{ background: 'var(--bg-card)' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border)' }}>
                  {['URL', 'Events', 'Status', 'Created', 'Actions'].map(h => (
                    <th key={h} className="px-5 py-3 text-left text-xs font-medium tracking-wider uppercase" style={{ color: 'var(--muted)', letterSpacing: '1px' }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {webhooks.map((w, i) => (
                  <tr
                    key={w.id}
                    className="table-row-hover row-animate"
                    style={{
                      borderBottom: i < webhooks.length - 1 ? '1px solid var(--border)' : 'none',
                      animationDelay: `${i * 0.04}s`,
                    }}
                  >
                    <td className="px-5 py-3.5 font-medium" style={{ color: 'var(--black)', fontFamily: 'var(--font-mono)', fontSize: '12px', maxWidth: '280px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {w.url}
                    </td>
                    <td className="px-5 py-3.5">
                      <div className="flex flex-wrap gap-1">
                        {w.events.map(e => (
                          <span key={e} className="px-2 py-0.5 rounded text-xs" style={{ background: 'var(--bg)', border: '1px solid var(--border)', color: 'var(--muted)', fontFamily: 'var(--font-mono)', fontSize: '10px' }}>
                            {e}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="px-5 py-3.5">
                      <Badge status={w.active ? 'active' : 'frozen'} />
                    </td>
                    <td className="px-5 py-3.5 text-xs" style={{ color: 'var(--muted)' }}>
                      {new Date(w.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-5 py-3.5">
                      <button
                        onClick={() => handleDelete(w.id)}
                        disabled={deletingId === w.id}
                        className="btn-lift px-3 py-1.5 rounded-md text-xs font-medium disabled:opacity-50"
                        style={{
                          background: confirmDeleteId === w.id ? 'var(--red)' : 'var(--red-soft)',
                          color: confirmDeleteId === w.id ? 'var(--bg)' : 'var(--red)',
                          border: `1px solid ${confirmDeleteId === w.id ? 'var(--red)' : 'var(--red-border)'}`,
                        }}
                      >
                        {deletingId === w.id ? '...' : confirmDeleteId === w.id ? 'Confirm' : 'Delete'}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {showRegister && <RegisterWebhookModal onClose={() => { setShowRegister(false); load(); }} />}
    </div>
  );
}

function RegisterWebhookModal({ onClose }: { onClose: () => void }) {
  const [url, setUrl] = useState('');
  const [events, setEvents] = useState<string[]>(['transaction.approved']);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [secret, setSecret] = useState('');

  function toggleEvent(ev: string) {
    setEvents(prev => prev.includes(ev) ? prev.filter(e => e !== ev) : [...prev, ev]);
  }

  async function submit() {
    setError('');

    if (!url.startsWith('https://')) {
      setError('URL must begin with https://');
      return;
    }
    if (events.length === 0) {
      setError('Select at least one event type');
      return;
    }

    setLoading(true);
    try {
      const generatedSecret = 'whsec_' + crypto.randomUUID().replace(/-/g, '').slice(0, 32);
      const res = await registerWebhook({ url, events, secret: generatedSecret }) as { secret?: string };
      setSecret(res?.secret || generatedSecret);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to register');
    }
    setLoading(false);
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 modal-backdrop"
      style={{ background: 'rgba(26,26,26,0.4)', backdropFilter: 'blur(4px)' }}
      onClick={(e) => e.target === e.currentTarget && !secret && onClose()}
    >
      <div className="w-full max-w-md rounded-xl p-6 flex flex-col gap-5 modal-content" style={{ background: 'var(--bg)', border: '1px solid var(--border)', boxShadow: '0 20px 60px rgba(0,0,0,0.12)' }}>
        {secret ? (
          <>
            <h2 style={{ fontFamily: 'var(--font-serif)', fontSize: '22px', fontWeight: 500, color: 'var(--black)' }}>Webhook registered</h2>
            <p className="text-sm" style={{ color: 'var(--muted)' }}>Copy this signing secret. It will not be shown again.</p>
            <div className="rounded-lg p-4 text-sm break-all" style={{ fontFamily: 'var(--font-mono)', background: 'var(--black)', color: 'var(--bg)', border: '1px solid var(--border)' }}>
              {secret}
            </div>
            <button onClick={onClose} className="btn-lift px-4 py-2.5 rounded-lg text-sm font-medium" style={{ background: 'var(--black)', color: 'var(--bg)' }}>
              Done
            </button>
          </>
        ) : (
          <>
            <div className="flex items-center justify-between">
              <h2 style={{ fontFamily: 'var(--font-serif)', fontSize: '22px', fontWeight: 500, color: 'var(--black)' }}>Register Webhook</h2>
              <button onClick={onClose} style={{ color: 'var(--muted)' }} className="p-1">
                <svg width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium" style={{ color: 'var(--black)' }}>Endpoint URL</label>
                <input
                  value={url}
                  onChange={e => setUrl(e.target.value)}
                  placeholder="https://your-server.com/webhooks"
                  className="rounded-lg px-3 py-2.5 text-sm outline-none"
                  style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', color: 'var(--black)', fontFamily: 'var(--font-mono)' }}
                  onFocus={e => (e.target.style.borderColor = 'var(--border-hover)')}
                  onBlur={e => (e.target.style.borderColor = 'var(--border)')}
                />
              </div>

              <div className="flex flex-col gap-2">
                <label className="text-xs font-medium" style={{ color: 'var(--black)' }}>Event Types</label>
                {EVENT_OPTIONS.map(opt => (
                  <label key={opt.value} className="flex items-center gap-2.5 cursor-pointer text-sm" style={{ color: 'var(--black)' }}>
                    <input
                      type="checkbox"
                      checked={events.includes(opt.value)}
                      onChange={() => toggleEvent(opt.value)}
                      className="rounded"
                      style={{ accentColor: 'var(--black)' }}
                    />
                    <span style={{ fontFamily: 'var(--font-mono)', fontSize: '12px' }}>{opt.label}</span>
                  </label>
                ))}
              </div>
            </div>

            {error && <p className="text-xs p-3 rounded-lg" style={{ background: 'var(--red-soft)', color: 'var(--red)', border: '1px solid var(--red-border)' }}>{error}</p>}

            <div className="flex gap-3">
              <button onClick={onClose} className="btn-lift flex-1 px-4 py-2.5 rounded-lg text-sm" style={{ color: 'var(--mid)', border: '1px solid var(--border)' }}>Cancel</button>
              <button onClick={submit} disabled={loading || !url} className="btn-lift flex-1 px-4 py-2.5 rounded-lg text-sm font-medium disabled:opacity-50" style={{ background: 'var(--black)', color: 'var(--bg)' }}>
                {loading ? 'Registering...' : 'Register'}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
