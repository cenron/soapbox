import { api } from "@/shared/api/client"

let accessToken: string | null = null

export function getAccessToken(): string | null {
  return accessToken
}

export function setAccessToken(token: string | null): void {
  accessToken = token
}

interface RefreshResponse {
  access_token: string
}

export async function refreshAccessToken(): Promise<string | null> {
  try {
    const data = await api.post<RefreshResponse>("/auth/refresh")
    setAccessToken(data.access_token)
    return data.access_token
  } catch {
    setAccessToken(null)
    return null
  }
}
