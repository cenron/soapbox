import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react"
import { setAccessToken, refreshAccessToken } from "./token-storage"
import type { UsersProfileResponse } from "@/shared/api/generated/types.gen"

interface User {
  id: string
  username: string
  displayName: string
  role: string
  verified: boolean
  avatarUrl?: string
}

interface AuthState {
  user: User | null
  isLoading: boolean
  isAuthenticated: boolean
  login: (accessToken: string, profile: UsersProfileResponse) => void
  logout: () => void
}

interface JwtPayload {
  sub: string
  username: string
  role: string
  verified: boolean
  exp: number
  iat: number
}

function decodeJwtPayload(token: string): JwtPayload | null {
  try {
    const segments = token.split(".")
    if (segments.length !== 3) return null

    const payload = segments[1]
    const padded = payload + "=".repeat((4 - (payload.length % 4)) % 4)
    const decoded = atob(padded.replace(/-/g, "+").replace(/_/g, "/"))
    return JSON.parse(decoded) as JwtPayload
  } catch {
    return null
  }
}

function profileToUser(profile: UsersProfileResponse): User | null {
  if (!profile.id || !profile.username) return null

  return {
    id: profile.id,
    username: profile.username,
    displayName: profile.display_name ?? profile.username,
    role: "user",
    verified: profile.verified ?? false,
    avatarUrl: profile.avatar_url,
  }
}

function tokenToUser(token: string): User | null {
  const payload = decodeJwtPayload(token)
  if (!payload) return null

  return {
    id: payload.sub,
    username: payload.username,
    displayName: payload.username,
    role: payload.role,
    verified: payload.verified,
  }
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    refreshAccessToken()
      .then((token) => {
        if (!token) return
        const decoded = tokenToUser(token)
        if (decoded) setUser(decoded)
      })
      .finally(() => setIsLoading(false))
  }, [])

  const login = (accessToken: string, profile: UsersProfileResponse) => {
    setAccessToken(accessToken)

    const jwtUser = tokenToUser(accessToken)
    const mapped = profileToUser(profile)
    if (mapped) {
      mapped.role = jwtUser?.role ?? "user"
      setUser(mapped)
    }
  }

  const logout = () => {
    setAccessToken(null)
    setUser(null)
  }

  const value = useMemo<AuthState>(
    () => ({
      user,
      isLoading,
      isAuthenticated: user !== null,
      login,
      logout,
    }),
    [user, isLoading],
  )

  return <AuthContext value={value}>{children}</AuthContext>
}

export function useAuth(): AuthState {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider")
  }
  return context
}
