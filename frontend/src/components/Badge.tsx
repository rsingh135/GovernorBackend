interface BadgeProps {
  status: string;
  className?: string;
}

const configs: Record<string, { label: string; color: string; bg: string; border: string }> = {
  APPROVED:         { label: 'Approved',  color: 'var(--green)',  bg: 'var(--green-soft)',  border: 'var(--green-border)' },
  approved:         { label: 'Approved',  color: 'var(--green)',  bg: 'var(--green-soft)',  border: 'var(--green-border)' },
  DENIED:           { label: 'Denied',    color: 'var(--red)',    bg: 'var(--red-soft)',    border: 'var(--red-border)' },
  denied:           { label: 'Denied',    color: 'var(--red)',    bg: 'var(--red-soft)',    border: 'var(--red-border)' },
  PENDING_APPROVAL: { label: 'Pending',   color: 'var(--gold)',   bg: 'var(--gold-soft)',   border: 'var(--gold-border)' },
  pending_approval: { label: 'Pending',   color: 'var(--gold)',   bg: 'var(--gold-soft)',   border: 'var(--gold-border)' },
  active:           { label: 'Active',    color: 'var(--green)',  bg: 'var(--green-soft)',  border: 'var(--green-border)' },
  frozen:           { label: 'Frozen',    color: 'var(--muted)',  bg: 'rgba(0,0,0,0.04)',   border: 'var(--border)' },
  delivered:        { label: 'Delivered', color: 'var(--green)',  bg: 'var(--green-soft)',  border: 'var(--green-border)' },
  pending:          { label: 'Pending',   color: 'var(--gold)',   bg: 'var(--gold-soft)',   border: 'var(--gold-border)' },
  failed:           { label: 'Failed',    color: 'var(--red)',    bg: 'var(--red-soft)',    border: 'var(--red-border)' },
};

export function Badge({ status, className = '' }: BadgeProps) {
  const cfg = configs[status] ?? { label: status, color: 'var(--muted)', bg: 'rgba(0,0,0,0.04)', border: 'var(--border)' };
  return (
    <span
      className={`inline-flex items-center px-2.5 py-1 rounded-md text-xs font-medium ${className}`}
      style={{ color: cfg.color, backgroundColor: cfg.bg, border: `1px solid ${cfg.border}` }}
    >
      <span className="mr-1.5 w-1 h-1 rounded-full inline-block" style={{ backgroundColor: cfg.color }} />
      {cfg.label}
    </span>
  );
}
