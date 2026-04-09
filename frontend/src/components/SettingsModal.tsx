'use client';

import { useState, useEffect } from 'react';

interface SettingsModalProps {
  open: boolean;
  onClose: () => void;
}

export function SettingsModal({ open, onClose }: SettingsModalProps) {
  const [apiKey, setApiKey] = useState('');
  const [approverUserId, setApproverUserId] = useState('');
  const [apiBaseUrl, setApiBaseUrl] = useState('');
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (open) {
      try {
        const raw = localStorage.getItem('governor_settings');
        if (raw) {
          const s = JSON.parse(raw);
          setApiKey(s.apiKey || '');
          setApproverUserId(s.approverUserId || '');
          setApiBaseUrl(s.apiBaseUrl || '');
        }
      } catch { /* ignore */ }
    }
  }, [open]);

  function save() {
    localStorage.setItem('governor_settings', JSON.stringify({
      apiKey,
      approverUserId,
      apiBaseUrl: apiBaseUrl || (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'),
    }));
    setSaved(true);
    setTimeout(() => {
      setSaved(false);
      onClose();
    }, 800);
  }

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 modal-backdrop"
      style={{ background: 'rgba(26,26,26,0.4)', backdropFilter: 'blur(4px)' }}
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        className="w-full max-w-md rounded-xl p-6 flex flex-col gap-5 modal-content"
        style={{ background: 'var(--bg)', border: '1px solid var(--border)', boxShadow: '0 20px 60px rgba(0,0,0,0.12)' }}
      >
        <div className="flex items-center justify-between">
          <h2 style={{ fontFamily: 'var(--font-serif)', fontSize: '22px', fontWeight: 500, color: 'var(--black)' }}>Settings</h2>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 nav-item"
            style={{ color: 'var(--muted)' }}
          >
            <svg width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="flex flex-col gap-4">
          <Field
            label="API Base URL"
            hint={`Default: ${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}`}
            value={apiBaseUrl}
            onChange={setApiBaseUrl}
            placeholder={process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}
          />
          <Field
            label="Agent API Key"
            hint={'Enter "governor-demo" to explore with sample data, no backend needed'}
            value={apiKey}
            onChange={setApiKey}
            placeholder='agp_... or "governor-demo"'
            type="password"
          />
          <Field
            label="Your User ID (for approvals)"
            hint="UUID of the user account that can approve transactions"
            value={approverUserId}
            onChange={setApproverUserId}
            placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
          />
        </div>

        <div className="flex justify-end gap-3">
          <button
            onClick={onClose}
            className="btn-lift px-4 py-2.5 rounded-lg text-sm"
            style={{ color: 'var(--mid)', background: 'transparent', border: '1px solid var(--border)' }}
          >
            Cancel
          </button>
          <button
            onClick={save}
            className="btn-lift px-5 py-2.5 rounded-lg text-sm font-medium"
            style={{ background: saved ? 'var(--green)' : 'var(--black)', color: 'var(--bg)' }}
          >
            {saved ? 'Saved' : 'Save'}
          </button>
        </div>
      </div>
    </div>
  );
}

function Field({
  label, hint, value, onChange, placeholder, type = 'text',
}: {
  label: string; hint?: string; value: string; onChange: (v: string) => void; placeholder?: string; type?: string;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-xs font-medium" style={{ color: 'var(--black)' }}>{label}</label>
      {hint && <p className="text-xs" style={{ color: 'var(--muted)' }}>{hint}</p>}
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full rounded-lg px-3 py-2.5 text-sm outline-none"
        style={{
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          color: 'var(--black)',
          transition: 'border-color 0.2s ease',
        }}
        onFocus={(e) => (e.target.style.borderColor = 'var(--border-hover)')}
        onBlur={(e) => (e.target.style.borderColor = 'var(--border)')}
      />
    </div>
  );
}
