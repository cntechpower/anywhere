import { useCallback, useEffect, useRef, useState } from "react";

export interface UseApiOptions {
  pollMs?: number;
  enabled?: boolean;
}

export interface UseApiState<T> {
  data: T | null;
  error: unknown;
  loading: boolean;
  updatedAt: number | null;
}

export interface UseApiResult<T> extends UseApiState<T> {
  reload: () => void;
}

/**
 * Small data-loading hook with optional polling.
 *
 * - keeps the previous data visible while a poll refreshes (no flicker)
 * - aborts the in-flight request when the component unmounts or reloads
 * - `reload()` forces an immediate fetch
 */
export default function useApi<T>(
  fetcher: (signal: AbortSignal) => Promise<T>,
  options: UseApiOptions = {}
): UseApiResult<T> {
  const { pollMs = 0, enabled = true } = options;
  const [state, setState] = useState<UseApiState<T>>({
    data: null,
    error: null,
    loading: true,
    updatedAt: null,
  });
  const [version, setVersion] = useState(0);
  const fetcherRef = useRef(fetcher);
  fetcherRef.current = fetcher;

  const reload = useCallback(() => setVersion((value) => value + 1), []);

  useEffect(() => {
    if (!enabled) {
      setState((previous) => ({ ...previous, loading: false }));
      return undefined;
    }
    let cancelled = false;
    const controller = new AbortController();
    setState((previous) => ({
      ...previous,
      loading: previous.data === null,
      error: null,
    }));

    fetcherRef
      .current(controller.signal)
      .then((data) => {
        if (!cancelled) {
          setState({
            data,
            error: null,
            loading: false,
            updatedAt: Date.now(),
          });
        }
      })
      .catch((error: unknown) => {
        if (cancelled || axiosIsCancel(error)) {
          return;
        }
        setState((previous) => ({
          data: previous.data,
          error,
          loading: false,
          updatedAt: previous.updatedAt,
        }));
      });

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [version, enabled]);

  useEffect(() => {
    if (!pollMs || !enabled) {
      return undefined;
    }
    const timer = window.setInterval(
      () => setVersion((value) => value + 1),
      pollMs
    );
    return () => window.clearInterval(timer);
  }, [pollMs, enabled]);

  return { ...state, reload };
}

function axiosIsCancel(error: unknown): boolean {
  if (!error || typeof error !== "object") {
    return false;
  }
  const value = error as { code?: string; name?: string };
  return (
    value.code === "ERR_CANCELED" ||
    value.name === "CanceledError" ||
    value.name === "AbortError"
  );
}
