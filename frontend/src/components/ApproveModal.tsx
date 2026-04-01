'use client';

import { useState } from 'react';
import { Transaction } from '@/lib/types';
import { formatCents } from '@/lib/api';
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
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      style={{ background: 'rgba(5,6,8,0.8)', backdropFilter: 'blur(4px)' }}
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        className="w-full max-w-sm rounded-2xl p-6 flex flex-col gap-5"
        style={{ background: 'var(--panel)', border: '1px solid var(--border)' }}
      >
        <div>
          <h2 className="text-base font-semibold mb-1" style={{ color: 'var(--text)' }}>
            {isApprove ? 'Approve transaction?' : 'Deny transaction?'}
          </h2>
          <p className="text-sm" style={{ color: 'var(--muted)' }}>
            This action will be recorded in the audit trail.
          </p>
        </div>

        <div className="rounded-xl p-4 flex flex-col gap-2" style={{ background: '#0d0f15', border: '1px solid var(--border)' }}>
          <Row label="Amount" value={<span style={{ color: isApprove ? 'var(--green)' : 'var(--red)', fontWeight: 600 }}>{formatCents(transaction.amount_cents)}</span>} />
          <Row label="Vendor" value={transaction.vendor} />
          <Row label="Status" value={<Badge status={transaction.status} />} />
          <Row label="Reason" value={<span style={{ color: 'var(--muted)' }}>{transaction.reason}</span>} />
        </div>

        {!approverUserId && (
          <p className="text-xs p-3 rounded-lg" style={{ background: 'rgba(242,185,102,0.1)', color: 'var(--gold)', border: '1px solid rgba(242,185,102,0.2)' }}>
            No approver user ID configured. Open Settings and add your User ID.
          </p>
        )}

        {error && (
          <p className="text-xs p-3 rounded-lg" style={{ background: 'rgba(248,113,113,0.1)', color: 'var(--red)', border: '1px solid rgba(248,113,113,0.2)' }}>
            {error}
          </p>
        )}

        <div className="flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 px-4 py-2 rounded-lg text-sm transition-colors"
            style={{ color: 'var(--muted)', border: '1px solid var(--border)', background: 'transparent' }}
          >
            Cancel
          </button>
          <button
            onClick={handleConfirm}
            disabled={loading || !approverUserId}
            className="flex-1 px-4 py-2 rounded-lg text-sm font-medium transition-opacity disabled:opacity-50"
            style={{
              background: isApprove ? 'var(--green)' : 'var(--red)',
              color: isApprove ? '#0a0a0a' : 'white',
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
      <span style={{ color: 'var(--muted)' }}>{label}</span>
      <span style={{ color: 'var(--text)' }}>{value}</span>
    </div>
  );
}
