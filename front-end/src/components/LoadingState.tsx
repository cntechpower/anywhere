interface LoadingStateProps {
  variant?: "table" | "cards" | "panels";
  rows?: number;
}

export default function LoadingState({
  variant = "table",
  rows = 5,
}: LoadingStateProps) {
  if (variant === "cards") {
    return (
      <div className="grid stats-grid" role="status" aria-label="加载中">
        {Array.from({ length: 5 }).map((_, index) => (
          <div key={index} className="skeleton skeleton-card" />
        ))}
      </div>
    );
  }
  if (variant === "panels") {
    return (
      <div className="grid two-col" role="status" aria-label="加载中">
        {Array.from({ length: 2 }).map((_, index) => (
          <div key={index} className="skeleton" style={{ height: 220 }} />
        ))}
      </div>
    );
  }
  return (
    <div role="status" aria-label="加载中">
      <div className="skeleton" style={{ height: 42, marginBottom: 12 }} />
      {Array.from({ length: rows }).map((_, index) => (
        <div key={index} className="skeleton skeleton-row" />
      ))}
    </div>
  );
}
