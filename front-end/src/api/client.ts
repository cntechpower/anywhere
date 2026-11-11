import axios, { AxiosError, AxiosRequestConfig } from "axios";
import qs from "qs";

// Shared HTTP client for the console. Response cookies carry the session, so
// every request must be sent with credentials.
const client = axios.create({
  withCredentials: true,
  timeout: 15000,
  headers: {
    "X-Requested-With": "XMLHttpRequest",
  },
});

client.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (
      error.response?.status === 401 &&
      typeof window !== "undefined" &&
      !window.location.pathname.startsWith("/user/login")
    ) {
      window.location.assign("/user/login");
    }
    return Promise.reject(error);
  }
);

export function apiGet<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return client.get<T>(url, config).then((response) => response.data);
}

export function apiPostForm<T = unknown>(
  url: string,
  data?: Record<string, unknown>,
  config?: AxiosRequestConfig
): Promise<T> {
  return client
    .post<T>(url, qs.stringify(data ?? {}), {
      headers: { "content-type": "application/x-www-form-urlencoded" },
      ...config,
    })
    .then((response) => response.data);
}

export default client;
