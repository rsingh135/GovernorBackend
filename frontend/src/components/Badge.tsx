interface BadgeProps {
  status: string;
  className?: string;
}

const configs: Record<string, { label: string; color: string; bg: string }> = {
  APPROVED:        { label: 'Approved',        color: '#34d399', bg: 'rgba(52,211,153,0.12)' },
  approved:        { label: 'Approved',        color: '#34d399', bg: 'rgba(52,211,153,0.12)' },
  DENIED:          { label: 'Denied',          color: '#f87171', bg: 'rgba(248,113,113,0.12)' },
  denied:          { label: 'Denied',          color: '#f87171', bg: 'rgba(248,113,113,0.12)' },
  PENDING_APPROVAL:{ label: 'Pending',         color: '#f2b966', bg: 'rgba(242,185,102,0.12)' },
  pending_approval:{ label: 'Pending',         color: '#f2b966', bg: 'rgba(242,185,102,0.12)' },
  active:          { label: 'Active',          color: '#34d399', bg: 'rgba(52,211,153,0.12)' },
  frozen:          { label: 'Frozen',          color: '#9fa3ae', bg: 'rgba(159,163,174,0.12)' },
  delivered:       { label: 'Delivered',       color: '#34d399', bg: 'rgba(52,211,153,0.12)' },
  pending:         { label: 'Pending',         color: '#f2b966', bg: 'rgba(242,185,102,0.12)' },
  failed:          { label: 'Failed',          color: '#f87171', bg: 'rgba(248,113,113,0.12)' },
};

export function Badge({ status, className = '' }: BadgeProps) {
  const cfg = configs[status] ?? { label: status, color: '#9fa3ae', bg: 'rgba(159,163,174,0.12)' };
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${className}`}
      style={{ color: cfg.color, backgroundColor: cfg.bg }}
    >
      <span className="mr-1.5 w-1.5 h-1.5 rounded-full inline-block" style={{ backgroundColor: cfg.color }} />
      {cfg.label}
    </span>
  );
}
