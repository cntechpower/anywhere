import type { ReactNode } from "react";

interface StatusBadgeProps {
  tone?: "" | "ok" | "warn" | "bad" | "info";
  children: ReactNode;
}

export default function StatusBadge({
  tone = "",
  children,
}: StatusBadgeProps) {
  return (
    <span className={`status${tone ? ` ${tone}` : ""}`}>
      <span className="dot" aria-hidden="true" />
      {children}
    </span>
  );
}
