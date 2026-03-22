import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { WsClient } from "./ws-client"

class MockWebSocket {
  static instances: MockWebSocket[] = []

  onopen: (() => void) | null = null
  onmessage: ((event: { data: string }) => void) | null = null
  onclose: (() => void) | null = null
  onerror: (() => void) | null = null

  readyState = 0

  url: string

  constructor(url: string) {
    this.url = url
    MockWebSocket.instances.push(this)
  }

  close() {
    this.readyState = 3
    this.onclose?.()
  }

  simulateOpen() {
    this.readyState = 1
    this.onopen?.()
  }

  simulateMessage(data: unknown) {
    this.onmessage?.({ data: JSON.stringify(data) })
  }

  simulateError() {
    this.onerror?.()
  }
}

beforeEach(() => {
  MockWebSocket.instances = []
  vi.stubGlobal("WebSocket", MockWebSocket)
  vi.useFakeTimers()
})

afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

describe("WsClient", () => {
  it("connects to the given URL", () => {
    const client = new WsClient("ws://localhost:8080/ws")
    client.connect()

    expect(MockWebSocket.instances).toHaveLength(1)
    expect(MockWebSocket.instances[0].url).toBe("ws://localhost:8080/ws")
  })

  it("dispatches messages to registered handlers", () => {
    const client = new WsClient("ws://localhost:8080/ws")
    const handler = vi.fn()

    client.on("new_post", handler)
    client.connect()

    const ws = MockWebSocket.instances[0]
    ws.simulateOpen()
    ws.simulateMessage({ type: "new_post", data: { id: "123" } })

    expect(handler).toHaveBeenCalledWith({ id: "123" })
  })

  it("removes handlers with off()", () => {
    const client = new WsClient("ws://localhost:8080/ws")
    const handler = vi.fn()

    client.on("new_post", handler)
    client.off("new_post", handler)
    client.connect()

    const ws = MockWebSocket.instances[0]
    ws.simulateOpen()
    ws.simulateMessage({ type: "new_post", data: {} })

    expect(handler).not.toHaveBeenCalled()
  })

  it("reconnects with exponential backoff", () => {
    const client = new WsClient("ws://localhost:8080/ws")
    client.connect()

    const ws = MockWebSocket.instances[0]
    ws.simulateOpen()
    ws.close()

    expect(MockWebSocket.instances).toHaveLength(1)

    // First reconnect at 1s
    vi.advanceTimersByTime(1000)
    expect(MockWebSocket.instances).toHaveLength(2)

    // Second disconnect → reconnect at 2s
    MockWebSocket.instances[1].close()
    vi.advanceTimersByTime(2000)
    expect(MockWebSocket.instances).toHaveLength(3)
  })

  it("ignores duplicate connect() calls", () => {
    const client = new WsClient("ws://localhost:8080/ws")
    client.connect()
    client.connect()
    client.connect()

    expect(MockWebSocket.instances).toHaveLength(1)
  })

  it("stops reconnecting after disconnect()", () => {
    const client = new WsClient("ws://localhost:8080/ws")
    client.connect()

    const ws = MockWebSocket.instances[0]
    ws.simulateOpen()
    client.disconnect()

    vi.advanceTimersByTime(60_000)
    expect(MockWebSocket.instances).toHaveLength(1)
  })
})
