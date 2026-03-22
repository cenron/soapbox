import { describe, it, expect, beforeEach, afterEach, vi } from "vitest"
import createClient, { type Middleware } from "openapi-fetch"
import type { paths } from "./schema"
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

function createTestClient() {
  const authMiddleware: Middleware = {
    async onRequest({ request }) {
      const token = tokenStorage.getAccessToken()
      if (token) {
        request.headers.set("Authorization", `Bearer ${token}`)
      }
      return request
    },
  }

  const client = createClient<paths>({
    baseUrl: "http://localhost/api/v1",
    credentials: "include",
  })

  client.use(authMiddleware)
  return client
}

describe("api client", () => {
  it("makes requests to /api/v1 prefix", async () => {
    mockFetch.mockResolvedValueOnce(
      new Response(JSON.stringify({ id: "1", username: "test" }), { status: 200 }),
    )

    const client = createTestClient()
    await client.GET("/users/{username}", { params: { path: { username: "test" } } })

    const request = mockFetch.mock.calls[0][0] as Request
    expect(request.url).toBe("http://localhost/api/v1/users/test")
  })

  it("includes auth header when token is set", async () => {
    vi.mocked(tokenStorage.getAccessToken).mockReturnValue("test-token")
    mockFetch.mockResolvedValueOnce(new Response(JSON.stringify({}), { status: 200 }))

    const client = createTestClient()
    await client.GET("/users/{username}", { params: { path: { username: "test" } } })

    const request = mockFetch.mock.calls[0][0] as Request
    expect(request.headers.get("Authorization")).toBe("Bearer test-token")
  })

  it("sends credentials include", async () => {
    mockFetch.mockResolvedValueOnce(new Response(JSON.stringify({}), { status: 200 }))

    const client = createTestClient()
    await client.GET("/users/{username}", { params: { path: { username: "test" } } })

    const request = mockFetch.mock.calls[0][0] as Request
    expect(request.credentials).toBe("include")
  })

  it("returns data on success", async () => {
    const body = { access_token: "tok", user: { id: "1", username: "test" } }
    mockFetch.mockResolvedValueOnce(new Response(JSON.stringify(body), { status: 200 }))

    const client = createTestClient()
    const { data, error } = await client.POST("/auth/login", {
      body: { email: "test@example.com", password: "password123" },
    })

    expect(error).toBeUndefined()
    expect(data).toEqual(body)
  })

  it("returns error on non-ok response", async () => {
    mockFetch.mockResolvedValueOnce(
      new Response(JSON.stringify({ message: "not found", detail: "user not found" }), { status: 404 }),
    )

    const client = createTestClient()
    const { data, error } = await client.GET("/users/{username}", {
      params: { path: { username: "nobody" } },
    })

    expect(data).toBeUndefined()
    expect(error).toEqual({ message: "not found", detail: "user not found" })
  })
})
