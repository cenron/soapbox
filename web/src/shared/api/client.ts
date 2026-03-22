import { ApiError } from "./errors"
import { getAccessToken } from "@/shared/auth/token-storage"

const BASE_URL = "/api/v1"

interface RequestOptions extends Omit<RequestInit, "body"> {
  body?: unknown
}

async function parseErrorResponse(res: Response): Promise<ApiError> {
  try {
    const data = await res.json()
    return new ApiError(res.status, data.message ?? "unknown", data.detail ?? data.message ?? res.statusText)
  } catch {
    return new ApiError(res.status, "unknown", res.statusText)
  }
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { body, headers: customHeaders, ...rest } = options

  const headers = new Headers(customHeaders)
  headers.set("Content-Type", "application/json")

  const token = getAccessToken()
  if (token) {
    headers.set("Authorization", `Bearer ${token}`)
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    ...rest,
    headers,
    credentials: "include",
    body: body != null ? JSON.stringify(body) : undefined,
  })

  if (!res.ok) {
    throw await parseErrorResponse(res)
  }

  if (res.status === 204) {
    return undefined as T
  }

  return res.json() as Promise<T>
}

export const api = {
  get: <T>(path: string, options?: RequestOptions) =>
    request<T>(path, { ...options, method: "GET" }),

  post: <T>(path: string, body?: unknown, options?: RequestOptions) =>
    request<T>(path, { ...options, method: "POST", body }),

  put: <T>(path: string, body?: unknown, options?: RequestOptions) =>
    request<T>(path, { ...options, method: "PUT", body }),

  patch: <T>(path: string, body?: unknown, options?: RequestOptions) =>
    request<T>(path, { ...options, method: "PATCH", body }),

  delete: <T>(path: string, options?: RequestOptions) =>
    request<T>(path, { ...options, method: "DELETE" }),
}
