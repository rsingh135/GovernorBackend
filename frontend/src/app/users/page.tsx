'use client';

import { useEffect, useState } from 'react';
import { getUsers, createUser, formatCents } from '@/lib/api';
import { User } from '@/lib/types';

export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);

  async function load() {
    setLoading(true);
    try {
      const res = await getUsers();
      setUsers(res.users || []);
    } catch { /* ignore */ }
    setLoading(false);
  }

  useEffect(() => { load(); }, []);

  return (
    <div className="p-6 md:p-8 max-w-5xl mx-auto page-enter">
      <div className="mb-8 flex items-start justify-between">
        <div>
          <p className="text-xs font-medium tracking-wider uppercase mb-2" style={{ color: 'var(--muted)', letterSpacing: '2px' }}>Team</p>
          <h1 style={{ fontFamily: 'var(--font-serif)', fontSize: 'clamp(28px, 4vw, 36px)', fontWeight: 500, color: 'var(--black)', lineHeight: 1.1 }}>
            Users
          </h1>
          <p className="text-sm mt-2" style={{ color: 'var(--mid)' }}>{users.length} user{users.length !== 1 ? 's' : ''} registered</p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="btn-lift px-4 py-2.5 rounded-lg text-sm font-medium"
          style={{ background: 'var(--black)', color: 'var(--bg)' }}
        >
          + Create User
        </button>
      </div>

      <div className="rounded-xl overflow-hidden fade-up" style={{ border: '1px solid var(--border)' }}>
        {loading ? (
          <div className="p-6" style={{ background: 'var(--bg-card)' }}>
            <div className="space-y-3">
              {[1,2,3].map(i => <div key={i} className="skeleton h-10 w-full" />)}
            </div>
          </div>
        ) : users.length === 0 ? (
          <div className="p-12 text-center flex flex-col items-center gap-3" style={{ background: 'var(--bg-card)' }}>
            <div className="w-12 h-12 rounded-full flex items-center justify-center" style={{ background: 'rgba(0,0,0,0.05)' }}>
              <svg width="22" height="22" fill="none" stroke="var(--muted)" strokeWidth="1.8" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2M9 11a4 4 0 100-8 4 4 0 000 8z" />
              </svg>
            </div>
            <p className="text-sm font-medium" style={{ color: 'var(--black)' }}>No users yet</p>
            <p className="text-xs" style={{ color: 'var(--muted)' }}>Create your first user to get started.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm" style={{ background: 'var(--bg-card)' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border)' }}>
                  {['Name', 'Balance', 'Role', 'Created'].map(h => (
                    <th key={h} className="px-5 py-3 text-left text-xs font-medium tracking-wider uppercase" style={{ color: 'var(--muted)', letterSpacing: '1px' }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {users.map((u, i) => (
                  <tr
                    key={u.id}
                    className="table-row-hover row-animate"
                    style={{
                      borderBottom: i < users.length - 1 ? '1px solid var(--border)' : 'none',
                      animationDelay: `${i * 0.04}s`,
                    }}
                  >
                    <td className="px-5 py-3.5">
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded-full flex items-center justify-center text-xs font-medium" style={{ background: 'var(--bg)', border: '1px solid var(--border)', color: 'var(--black)' }}>
                          {u.name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()}
                        </div>
                        <div>
                          <p className="font-medium" style={{ color: 'var(--black)' }}>{u.name}</p>
                          <p className="text-xs" style={{ fontFamily: 'var(--font-mono)', color: 'var(--muted)', fontSize: '10px' }}>{u.id}</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-5 py-3.5">
                      <span style={{
                        fontFamily: 'var(--font-mono)',
                        fontSize: '13px',
                        fontWeight: 500,
                        color: u.balance_cents > 0 ? 'var(--green)' : 'var(--red)',
                      }}>
                        {formatCents(u.balance_cents)}
                      </span>
                    </td>
                    <td className="px-5 py-3.5">
                      <span
                        className="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-medium"
                        style={{
                          color: u.role === 'admin' ? 'var(--green)' : 'var(--muted)',
                          backgroundColor: u.role === 'admin' ? 'var(--green-soft)' : 'rgba(0,0,0,0.04)',
                          border: `1px solid ${u.role === 'admin' ? 'var(--green-border)' : 'var(--border)'}`,
                        }}
                      >
                        {u.role === 'admin' ? 'Admin' : 'Viewer'}
                      </span>
                    </td>
                    <td className="px-5 py-3.5 text-xs" style={{ color: 'var(--muted)' }}>
                      {new Date(u.created_at).toLocaleDateString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {showCreate && <CreateUserModal onClose={() => { setShowCreate(false); load(); }} />}
    </div>
  );
}

function CreateUserModal({ onClose }: { onClose: () => void }) {
  const [name, setName] = useState('');
  const [balance, setBalance] = useState('1000');
  const [role, setRole] = useState<'admin' | 'viewer'>('viewer');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  async function submit() {
    setError('');
    if (!name.trim()) { setError('Name is required'); return; }
    const bal = parseFloat(balance);
    if (isNaN(bal) || bal < 0) { setError('Balance must be a non-negative number'); return; }

    setLoading(true);
    try {
      await createUser({
        name: name.trim(),
        initial_balance_cents: Math.round(bal * 100),
      });
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create user');
    }
    setLoading(false);
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 modal-backdrop"
      style={{ background: 'rgba(26,26,26,0.4)', backdropFilter: 'blur(4px)' }}
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className="w-full max-w-md rounded-xl p-6 flex flex-col gap-5 modal-content" style={{ background: 'var(--bg)', border: '1px solid var(--border)', boxShadow: '0 20px 60px rgba(0,0,0,0.12)' }}>
        <div className="flex items-center justify-between">
          <h2 style={{ fontFamily: 'var(--font-serif)', fontSize: '22px', fontWeight: 500, color: 'var(--black)' }}>Create User</h2>
          <button onClick={onClose} style={{ color: 'var(--muted)' }} className="p-1">
            <svg width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium" style={{ color: 'var(--black)' }}>Name</label>
            <input
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="Full name"
              className="rounded-lg px-3 py-2.5 text-sm outline-none"
              style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', color: 'var(--black)' }}
              onFocus={e => (e.target.style.borderColor = 'var(--border-hover)')}
              onBlur={e => (e.target.style.borderColor = 'var(--border)')}
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium" style={{ color: 'var(--black)' }}>Initial Balance ($)</label>
            <input
              value={balance}
              onChange={e => setBalance(e.target.value)}
              placeholder="1000.00"
              className="rounded-lg px-3 py-2.5 text-sm outline-none"
              style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', color: 'var(--black)', fontFamily: 'var(--font-mono)' }}
              onFocus={e => (e.target.style.borderColor = 'var(--border-hover)')}
              onBlur={e => (e.target.style.borderColor = 'var(--border)')}
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium" style={{ color: 'var(--black)' }}>Role</label>
            <div className="flex gap-2">
              {(['admin', 'viewer'] as const).map(r => (
                <button
                  key={r}
                  onClick={() => setRole(r)}
                  className="flex-1 px-3 py-2.5 rounded-lg text-sm font-medium"
                  style={{
                    background: role === r ? (r === 'admin' ? 'var(--green-soft)' : 'rgba(0,0,0,0.04)') : 'var(--bg-card)',
                    color: role === r ? (r === 'admin' ? 'var(--green)' : 'var(--black)') : 'var(--muted)',
                    border: `1px solid ${role === r ? (r === 'admin' ? 'var(--green-border)' : 'var(--border-hover)') : 'var(--border)'}`,
                    transition: 'all 0.2s ease',
                  }}
                >
                  {r.charAt(0).toUpperCase() + r.slice(1)}
                </button>
              ))}
            </div>
          </div>
        </div>

        {error && <p className="text-xs p-3 rounded-lg" style={{ background: 'var(--red-soft)', color: 'var(--red)', border: '1px solid var(--red-border)' }}>{error}</p>}

        <div className="flex gap-3">
          <button onClick={onClose} className="btn-lift flex-1 px-4 py-2.5 rounded-lg text-sm" style={{ color: 'var(--mid)', border: '1px solid var(--border)' }}>Cancel</button>
          <button onClick={submit} disabled={loading || !name} className="btn-lift flex-1 px-4 py-2.5 rounded-lg text-sm font-medium disabled:opacity-50" style={{ background: 'var(--black)', color: 'var(--bg)' }}>
            {loading ? 'Creating...' : 'Create User'}
          </button>
        </div>
      </div>
    </div>
  );
}
