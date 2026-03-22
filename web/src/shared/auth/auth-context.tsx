import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react"
import { setAccessToken, refreshAccessToken } from "./token-storage"

interface User {
  id: string
  username: string
  role: string
}

interface AuthState {
  user: User | null
  isLoading: boolean
  isAuthenticated: boolean
  login: (accessToken: string, user: User) => void
  logout: () => void
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    refreshAccessToken()
      .then((token) => {
        if (!token) return
        // TODO: decode user from JWT or fetch /auth/me
      })
      .finally(() => setIsLoading(false))
  }, [])

  const login = (accessToken: string, newUser: User) => {
    setAccessToken(accessToken)
    setUser(newUser)
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
