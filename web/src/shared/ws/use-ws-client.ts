import { useContext } from "react"
import type { WsClient } from "./ws-client"
import { WsContext } from "./ws-provider"

export function useWsClient(): WsClient | null {
  return useContext(WsContext)
}
