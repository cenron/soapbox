import { useCallback, useEffect, useState } from "react"
import { useWsClient } from "@/shared/ws/use-ws-client"

export function useNotificationBadge() {
  const wsClient = useWsClient()
  const [count, setCount] = useState(0)

  useEffect(() => {
    if (!wsClient) return

    const handler = () => {
      setCount((prev) => prev + 1)
    }

    wsClient.on("new_notification", handler)

    return () => {
      wsClient.off("new_notification", handler)
    }
  }, [wsClient])

  const reset = useCallback(() => {
    setCount(0)
  }, [])

  return { count, reset }
}
