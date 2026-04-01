interface StatCardProps {
  label: string;
  value: string | number;
  sub?: string;
  accent?: 'blue' | 'gold' | 'green' | 'red';
  href?: string;
  onClick?: () => void;
}

const accentColors = {
  blue:  'var(--accent)',
  gold:  'var(--gold)',
  green: 'var(--green)',
  red:   'var(--red)',
};

export function StatCard({ label, value, sub, accent = 'blue', href, onClick }: StatCardProps) {
  const color = accentColors[accent];
  const isClickable = href || onClick;

  const content = (
    <div
      className={`rounded-xl p-5 flex flex-col gap-3 transition-colors ${isClickable ? 'cursor-pointer' : ''}`}
      style={{
        background: 'var(--panel)',
        border: '1px solid var(--border)',
        boxShadow: `inset 0 0 24px rgba(62,130,255,0.04)`,
      }}
      onClick={onClick}
    >
      <p className="text-xs font-medium tracking-wider uppercase" style={{ color: 'var(--muted)' }}>
        {label}
      </p>
      <p className="text-3xl font-semibold" style={{ color }}>
        {value}
      </p>
      {sub && (
        <p className="text-xs" style={{ color: 'var(--muted)' }}>
          {sub}
        </p>
      )}
    </div>
  );

  if (href) {
    return <a href={href} className="no-underline">{content}</a>;
  }
  return content;
}
