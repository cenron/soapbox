import { describe, it, expect, beforeEach, afterEach, vi } from "vitest"
import { api } from "./client"
import { ApiError } from "./errors"
import * as tokenStorage from "@/shared/auth/token-storage"

const mockFetch = vi.fn()

vi.spyOn(tokenStorage, "getAccessToken")

beforeEach(() => {
  vi.stubGlobal("fetch", mockFetch)
  mockFetch.mockReset()
  vi.mocked(tokenStorage.getAccessToken).mockReturnValue(null)
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe("api client", () => {
  it("makes GET requests to /api/v1 prefix", async () => {
    mockFetch.mockResolvedValue(new Response(JSON.stringify({ ok: true }), { status: 200 }))

    const result = await api.get("/users/me")

    expect(mockFetch).toHaveBeenCalledWith("/api/v1/users/me", expect.objectContaining({ method: "GET" }))
    expect(result).toEqual({ ok: true })
  })

  it("includes auth header when token is set", async () => {
    vi.mocked(tokenStorage.getAccessToken).mockReturnValue("test-token")
    mockFetch.mockResolvedValue(new Response(JSON.stringify({}), { status: 200 }))

    await api.get("/users/me")

    const [, options] = mockFetch.mock.calls[0] as [string, RequestInit]
    const headers = options.headers as Headers
    expect(headers.get("Authorization")).toBe("Bearer test-token")
  })

  it("sends credentials include", async () => {
    mockFetch.mockResolvedValue(new Response(JSON.stringify({}), { status: 200 }))

    await api.get("/test")

    expect(mockFetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ credentials: "include" }),
    )
  })

  it("sends JSON body on POST", async () => {
    mockFetch.mockResolvedValue(new Response(JSON.stringify({}), { status: 200 }))

    await api.post("/auth/login", { email: "test@example.com", password: "secret" })

    const [, options] = mockFetch.mock.calls[0] as [string, RequestInit]
    expect(options.method).toBe("POST")
    expect(options.body).toBe(JSON.stringify({ email: "test@example.com", password: "secret" }))
  })

  it("throws ApiError on non-ok response", async () => {
    mockFetch.mockResolvedValue(
      new Response(JSON.stringify({ message: "not found", detail: "user not found" }), { status: 404 }),
    )

    try {
      await api.get("/users/999")
      expect.fail("should have thrown")
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError)
      const apiErr = err as ApiError
      expect(apiErr.status).toBe(404)
      expect(apiErr.kind).toBe("not found")
      expect(apiErr.message).toBe("user not found")
      expect(apiErr.isNotFound).toBe(true)
    }
  })

  it("returns undefined for 204 responses", async () => {
    mockFetch.mockResolvedValue(new Response(null, { status: 204 }))

    const result = await api.delete("/users/me")

    expect(result).toBeUndefined()
  })
})
