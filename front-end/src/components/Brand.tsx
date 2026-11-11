interface BrandProps {
  compact?: boolean;
}

export default function Brand({ compact = false }: BrandProps) {
  return (
    <div className="brand" aria-label="Anywhere">
      <span className="brand-mark" aria-hidden="true">
        <svg viewBox="0 0 32 32" role="img" focusable="false">
          <defs>
            <linearGradient id="brandStroke" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stopColor="#4ade80" />
              <stop offset="100%" stopColor="#16a34a" />
            </linearGradient>
          </defs>
          <circle cx="8" cy="8" r="3.4" fill="url(#brandStroke)" />
          <circle cx="24" cy="24" r="3.4" fill="url(#brandStroke)" />
          <path
            d="M11 11c4.2 0 2.8 10 7 10h3"
            fill="none"
            stroke="url(#brandStroke)"
            strokeWidth="2.4"
            strokeLinecap="round"
          />
          <path
            d="M21 8h3v3"
            fill="none"
            stroke="rgba(148,163,184,.7)"
            strokeWidth="1.8"
            strokeLinecap="round"
          />
        </svg>
      </span>
      {!compact && (
        <span className="brand-text">
          <strong>Anywhere</strong>
          <small>隧道控制台</small>
        </span>
      )}
    </div>
  );
}
