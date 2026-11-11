import { SwapOutlined } from "@ant-design/icons";
import EmptyState from "./EmptyState";
import type { BarItem } from "../types";

interface BarListProps {
  items?: BarItem[];
  valueFormatter?: (value: number) => string;
  tone?: "" | "info" | "danger";
  emptyText?: string;
  emptyDescription?: string;
}

export default function BarList({
  items,
  valueFormatter,
  tone = "",
  emptyText = "暂无数据",
  emptyDescription = "有流量经过代理后，这里会显示排名。",
}: BarListProps) {
  const hasValue = Boolean(items?.some((item) => Number(item.value) > 0));
  if (!items || items.length === 0 || !hasValue) {
    return (
      <EmptyState
        compact
        icon={<SwapOutlined />}
        title={emptyText}
        description={emptyDescription}
      />
    );
  }
  const max = Math.max(...items.map((item) => Number(item.value) || 0), 1);

  return (
    <div className="bar-list">
      {items.map((item, index) => {
        const value = Number(item.value) || 0;
        const width = value <= 0 ? 0 : Math.max(4, Math.round((value / max) * 100));
        return (
          <div className="bar-item" key={item.key ?? index}>
            <div className="bar-meta">
              <span className="name" title={item.label}>
                {item.label}
              </span>
              <span className="value">
                {valueFormatter ? valueFormatter(value) : value}
              </span>
            </div>
            <div
              className="bar-track"
              role="progressbar"
              aria-valuenow={width}
              aria-valuemin={0}
              aria-valuemax={100}
              aria-label={item.label}
            >
              <div
                className={`bar-fill${tone ? ` ${tone}` : ""}`}
                style={{ width: `${width}%` }}
              />
            </div>
          </div>
        );
      })}
    </div>
  );
}
