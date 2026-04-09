'use client';

import { useState } from 'react';
import { Transaction } from '@/lib/types';
import { formatCents, isDemoMode } from '@/lib/api';
import { Badge } from './Badge';

interface ApproveModalProps {
  transaction: Transaction | null;
  action: 'approve' | 'deny';
  approverUserId: string;
  onConfirm: (txnId: string, action: 'approve' | 'deny', approverUserId: string) => Promise<void>;
  onClose: () => void;
}

export function ApproveModal({ transaction, action, approverUserId, onConfirm, onClose }: ApproveModalProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  if (!transaction) return null;

  async function handleConfirm() {
    if (!transaction) return;
    setLoading(true);
    setError('');
    try {
      await onConfirm(transaction.id, action, approverUserId);
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Something went wrong');
    } finally {
      setLoading(false);
    }
  }

  const isApprove = action === 'approve';

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 modal-backdrop"
      style={{ background: 'rgba(26,26,26,0.4)', backdropFilter: 'blur(4px)' }}
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        className="w-full max-w-sm rounded-xl p-6 flex flex-col gap-5 modal-content"
        style={{ background: 'var(--bg)', border: '1px solid var(--border)', boxShadow: '0 20px 60px rgba(0,0,0,0.12)' }}
      >
        <div>
          <h2 style={{ fontFamily: 'var(--font-serif)', fontSize: '22px', fontWeight: 500, color: 'var(--black)', marginBottom: '4px' }}>
            {isApprove ? 'Approve transaction?' : 'Deny transaction?'}
          </h2>
          <p className="text-sm" style={{ color: 'var(--muted)' }}>
            This action will be recorded in the audit trail.
          </p>
        </div>

        <div className="rounded-lg p-4 flex flex-col gap-2.5" style={{ background: 'var(--bg-card)', border: '1px solid var(--border)' }}>
          <Row label="Amount" value={<span style={{ color: isApprove ? 'var(--green)' : 'var(--red)', fontWeight: 600 }}>{formatCents(transaction.amount_cents)}</span>} />
          <Row label="Vendor" value={transaction.vendor} />
          <Row label="Status" value={<Badge status={transaction.status} />} />
          <Row label="Reason" value={<span style={{ color: 'var(--muted)' }}>{transaction.reason}</span>} />
        </div>

        {!approverUserId && !isDemoMode() && (
          <p className="text-xs p-3 rounded-lg" style={{ background: 'var(--gold-soft)', color: 'var(--gold)', border: '1px solid var(--gold-border)' }}>
            No approver user ID configured. Open Settings and add your User ID.
          </p>
        )}

        {error && (
          <p className="text-xs p-3 rounded-lg" style={{ background: 'var(--red-soft)', color: 'var(--red)', border: '1px solid var(--red-border)' }}>
            {error}
          </p>
        )}

        <div className="flex gap-3">
          <button
            onClick={onClose}
            className="btn-lift flex-1 px-4 py-2.5 rounded-lg text-sm"
            style={{ color: 'var(--mid)', border: '1px solid var(--border)', background: 'transparent' }}
          >
            Cancel
          </button>
          <button
            onClick={handleConfirm}
            disabled={loading || (!approverUserId && !isDemoMode())}
            className="btn-lift flex-1 px-4 py-2.5 rounded-lg text-sm font-medium disabled:opacity-50"
            style={{
              background: isApprove ? 'var(--green)' : 'var(--red)',
              color: 'var(--bg)',
            }}
          >
            {loading ? '...' : isApprove ? 'Approve' : 'Deny'}
          </button>
        </div>
      </div>
    </div>
  );
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between text-sm">
      <span style={{ color: 'var(--muted)', fontSize: '13px' }}>{label}</span>
      <span style={{ color: 'var(--black)' }}>{value}</span>
    </div>
  );
}
