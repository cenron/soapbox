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

  it("returns Bearer-prefixed token when authenticated", () => {
    vi.mocked(getAccessToken).mockReturnValue("jwt-abc-123")

    const config = createClientConfig({})
    const authFn = config.auth as () => string
    expect(authFn()).toBe("Bearer jwt-abc-123")

    vi.mocked(getAccessToken).mockReturnValue(null)
  })

  it("returns empty string when no token is set", () => {
    vi.mocked(getAccessToken).mockReturnValue(null)

    const config = createClientConfig({})
    const authFn = config.auth as () => string
    expect(authFn()).toBe("")
  })

  it("never returns 'Bearer ' with no token (no dangling prefix)", () => {
    vi.mocked(getAccessToken).mockReturnValue(null)

    const config = createClientConfig({})
    const authFn = config.auth as () => string
    const result = authFn()

    expect(result).not.toMatch(/^Bearer\s*$/)
  })
})
