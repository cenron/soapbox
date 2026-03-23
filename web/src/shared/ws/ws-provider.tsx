import { createContext, useEffect, useSyncExternalStore } from "react"
import { useAuth } from "@/shared/auth/auth-context"
import { getAccessToken } from "@/shared/auth/token-storage"
import { WsClient } from "./ws-client"

export const WsContext = createContext<WsClient | null>(null)

// Module-level store for the WebSocket client. React subscribes to changes
// via useSyncExternalStore so the provider re-renders when the client changes.
let currentClient: WsClient | null = null
const listeners = new Set<() => void>()

function getClient() {
  return currentClient
}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

function setClient(client: WsClient | null) {
  currentClient = client
  for (const listener of listeners) {
    listener()
  }
}

export function WsProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth()
  const client = useSyncExternalStore(subscribe, getClient)

  useEffect(() => {
    if (!isAuthenticated) {
      if (currentClient) {
        currentClient.disconnect()
        setClient(null)
      }
      return
    }

    const token = getAccessToken()
    if (!token) return

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:"
    const url = `${protocol}//${window.location.host}/ws?token=${encodeURIComponent(token)}`

    const ws = new WsClient(url)
    ws.connect()
    setClient(ws)

    return () => {
      ws.disconnect()
      setClient(null)
    }
  }, [isAuthenticated])

  return (
    <WsContext.Provider value={client}>
      {children}
    </WsContext.Provider>
  )
}
