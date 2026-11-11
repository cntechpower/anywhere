const UNITS = ["B", "KB", "MB", "GB", "TB", "PB"];

export function formatBytes(bytes?: number | null): string {
  const value = Number(bytes ?? 0);
  if (!value || value < 0) {
    return "0 B";
  }
  const exponent = Math.min(
    Math.floor(Math.log(value) / Math.log(1024)),
    UNITS.length - 1
  );
  const scaled = value / 1024 ** exponent;
  const digits = scaled >= 100 || exponent === 0 ? 0 : scaled >= 10 ? 1 : 2;
  return `${scaled.toFixed(digits)} ${UNITS[exponent]}`;
}

export function formatNumber(value?: number | null): string {
  const number = Number(value ?? 0);
  if (!Number.isFinite(number)) {
    return "0";
  }
  return number.toLocaleString("zh-CN");
}

// Server timestamps use "2006-01-02 15:04:05.000"; Date.parse needs a "T" on
// Safari and older WebKit.
export function parseServerTime(value?: string | number | null): Date | null {
  if (value === null || value === undefined || value === "") {
    return null;
  }
  if (typeof value === "number") {
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? null : date;
  }
  const normalized = value.includes("T") ? value : value.replace(" ", "T");
  const date = new Date(normalized);
  return Number.isNaN(date.getTime()) ? null : date;
}

export function formatDateTime(
  value?: string | number | Date | null
): string {
  let date: Date | null;
  if (value instanceof Date) {
    date = value;
  } else {
    date = parseServerTime(value);
  }
  if (!date || Number.isNaN(date.getTime())) {
    return "-";
  }
  return date.toLocaleString("zh-CN", { hour12: false });
}

export function timeAgo(
  value?: string | number | Date | null,
  now = Date.now()
): string {
  let date: Date | null;
  if (value instanceof Date) {
    date = value;
  } else {
    date = parseServerTime(value);
  }
  if (!date) {
    return "未知";
  }
  const seconds = Math.max(0, Math.round((now - date.getTime()) / 1000));
  if (seconds < 10) {
    return "刚刚";
  }
  if (seconds < 60) {
    return `${seconds} 秒前`;
  }
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) {
    return `${minutes} 分钟前`;
  }
  const hours = Math.floor(minutes / 60);
  if (hours < 48) {
    return `${hours} 小时前`;
  }
  return formatDateTime(date);
}

interface ErrorLike {
  response?: { data?: unknown };
  code?: string;
  message?: string;
}

export function apiErrorMessage(error: unknown): string {
  if (!error) {
    return "请求失败，请稍后重试";
  }
  const err = error as ErrorLike;
  const data = err.response?.data;
  if (typeof data === "string") {
    return data;
  }
  if (data && typeof data === "object") {
    const payload = data as { data?: unknown; message?: unknown };
    if (typeof payload.data === "string") {
      return payload.data;
    }
    if (typeof payload.message === "string") {
      return payload.message;
    }
  }
  if (err.code === "ECONNABORTED") {
    return "请求超时，请检查网络或服务端状态";
  }
  return err.message || "请求失败，请稍后重试";
}

export function splitHostPort(addr?: string): { host: string; port: string } {
  if (!addr) {
    return { host: "-", port: "" };
  }
  const index = addr.lastIndexOf(":");
  if (index <= 0) {
    return { host: addr, port: "" };
  }
  return { host: addr.slice(0, index), port: addr.slice(index + 1) };
}
