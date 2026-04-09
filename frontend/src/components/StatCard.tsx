interface StatCardProps {
  label: string;
  value: string | number;
  sub?: string;
  accent?: 'default' | 'gold' | 'green' | 'red';
  href?: string;
  onClick?: () => void;
}

export function StatCard({ label, value, sub, accent = 'default', href, onClick }: StatCardProps) {
  const isClickable = href || onClick;

  const accentStyles = {
    default: { color: 'var(--black)', indicator: 'var(--black)' },
    gold:    { color: 'var(--gold)',  indicator: 'var(--gold)' },
    green:   { color: 'var(--green)', indicator: 'var(--green)' },
    red:     { color: 'var(--red)',   indicator: 'var(--red)' },
  };

  const style = accentStyles[accent];

  const content = (
    <div
      className={`card-hover rounded-xl p-5 flex flex-col gap-3 ${isClickable ? 'cursor-pointer' : ''}`}
      style={{
        background: 'var(--bg-card)',
        border: '1px solid var(--border)',
      }}
      onClick={onClick}
    >
      <div className="flex items-center gap-2">
        <div className="w-1.5 h-1.5 rounded-full" style={{ background: style.indicator }} />
        <p className="text-xs font-medium tracking-wider uppercase" style={{ color: 'var(--muted)', letterSpacing: '1.5px' }}>
          {label}
        </p>
      </div>
      <p style={{ fontFamily: 'var(--font-serif)', fontSize: '32px', fontWeight: 500, color: style.color, lineHeight: 1 }}>
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
