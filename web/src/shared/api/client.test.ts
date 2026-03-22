import { describe, it, expect, vi } from "vitest"

vi.mock("@/shared/auth/token-storage", () => ({
  getAccessToken: vi.fn(() => null),
}))

import { getAccessToken } from "@/shared/auth/token-storage"
import { createClientConfig } from "./hey-api"

describe("hey-api client config", () => {
  it("sets base URL to /api/v1", () => {
    const config = createClientConfig({})
    expect(config.baseUrl).toBe("/api/v1")
  })

  it("includes credentials", () => {
    const config = createClientConfig({})
    expect(config.credentials).toBe("include")
  })

  it("provides auth function that returns token when set", () => {
    vi.mocked(getAccessToken).mockReturnValue("test-token")

    const config = createClientConfig({})
    const authFn = config.auth as () => string
    expect(authFn()).toBe("test-token")

    vi.mocked(getAccessToken).mockReturnValue(null)
  })

  it("provides auth function that returns empty string when no token", () => {
    vi.mocked(getAccessToken).mockReturnValue(null)

    const config = createClientConfig({})
    const authFn = config.auth as () => string
    expect(authFn()).toBe("")
  })
})
