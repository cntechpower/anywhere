import type { ReactNode } from "react";

interface StatCardProps {
  label: string;
  value: ReactNode;
  unit?: string;
  hint?: string;
  icon?: ReactNode;
  tone?: "" | "info" | "warning" | "danger";
}

export default function StatCard({
  label,
  value,
  unit,
  hint,
  icon,
  tone = "",
}: StatCardProps) {
  return (
    <div className={`stat-card${tone ? ` tone-${tone}` : ""}`}>
      {icon && (
        <span className="stat-icon" aria-hidden="true">
          {icon}
        </span>
      )}
      <span className="stat-label">{label}</span>
      <div className="stat-value">
        {value}
        {unit && <span className="stat-unit">{unit}</span>}
      </div>
      {hint && <div className="stat-hint">{hint}</div>}
    </div>
  );
}
