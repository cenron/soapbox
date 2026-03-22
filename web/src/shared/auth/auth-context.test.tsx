import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, act } from "@testing-library/react"
import { MemoryRouter } from "react-router"
import { AuthProvider, useAuth } from "./auth-context"
import * as tokenStorage from "./token-storage"

vi.mock("./token-storage", () => ({
  getAccessToken: vi.fn(() => null),
  setAccessToken: vi.fn(),
  refreshAccessToken: vi.fn(() => Promise.resolve(null)),
}))

function TestConsumer() {
  const { isAuthenticated, isLoading, user } = useAuth()
  return (
    <div>
      <span data-testid="loading">{String(isLoading)}</span>
      <span data-testid="authenticated">{String(isAuthenticated)}</span>
      <span data-testid="user">{user?.username ?? "none"}</span>
    </div>
  )
}

function LoginButton() {
  const { login } = useAuth()
  return (
    <button
      onClick={() => login("token-123", { id: "1", username: "alice" })}
    >
      login
    </button>
  )
}

function LogoutButton() {
  const { logout } = useAuth()
  return <button onClick={logout}>logout</button>
}

function renderWithProviders(ui: React.ReactNode) {
  return render(
    <MemoryRouter>
      <AuthProvider>{ui}</AuthProvider>
    </MemoryRouter>,
  )
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe("AuthProvider", () => {
  it("starts in loading state and resolves to unauthenticated", async () => {
    vi.mocked(tokenStorage.refreshAccessToken).mockResolvedValue(null)

    renderWithProviders(<TestConsumer />)

    await act(async () => {})

    expect(screen.getByTestId("loading")).toHaveTextContent("false")
    expect(screen.getByTestId("authenticated")).toHaveTextContent("false")
    expect(screen.getByTestId("user")).toHaveTextContent("none")
  })

  it("restores user from refresh token on mount", async () => {
    // JWT with sub=1, username=bob, role=user, verified=false
    const fakeJwt = buildFakeJwt({ sub: "1", username: "bob", role: "user", verified: false })
    vi.mocked(tokenStorage.refreshAccessToken).mockResolvedValue(fakeJwt)

    renderWithProviders(<TestConsumer />)

    await act(async () => {})

    expect(screen.getByTestId("authenticated")).toHaveTextContent("true")
    expect(screen.getByTestId("user")).toHaveTextContent("bob")
  })

  it("sets user after login", async () => {
    vi.mocked(tokenStorage.refreshAccessToken).mockResolvedValue(null)

    renderWithProviders(
      <>
        <TestConsumer />
        <LoginButton />
      </>,
    )

    await act(async () => {})

    await act(async () => {
      screen.getByText("login").click()
    })

    expect(screen.getByTestId("authenticated")).toHaveTextContent("true")
    expect(screen.getByTestId("user")).toHaveTextContent("alice")
    expect(tokenStorage.setAccessToken).toHaveBeenCalledWith("token-123")
  })

  it("clears user and token on logout", async () => {
    vi.mocked(tokenStorage.refreshAccessToken).mockResolvedValue(null)

    renderWithProviders(
      <>
        <TestConsumer />
        <LoginButton />
        <LogoutButton />
      </>,
    )

    await act(async () => {})

    // Login first
    await act(async () => {
      screen.getByText("login").click()
    })
    expect(screen.getByTestId("authenticated")).toHaveTextContent("true")

    // Logout
    await act(async () => {
      screen.getByText("logout").click()
    })

    expect(screen.getByTestId("authenticated")).toHaveTextContent("false")
    expect(screen.getByTestId("user")).toHaveTextContent("none")
    expect(tokenStorage.setAccessToken).toHaveBeenCalledWith(null)
  })

  it("stays unauthenticated when refresh returns null", async () => {
    vi.mocked(tokenStorage.refreshAccessToken).mockResolvedValue(null)

    renderWithProviders(<TestConsumer />)

    await act(async () => {})

    expect(screen.getByTestId("loading")).toHaveTextContent("false")
    expect(screen.getByTestId("authenticated")).toHaveTextContent("false")
    expect(screen.getByTestId("user")).toHaveTextContent("none")
  })

  it("throws if useAuth is used outside provider", () => {
    const spy = vi.spyOn(console, "error").mockImplementation(() => {})

    expect(() =>
      render(
        <MemoryRouter>
          <TestConsumer />
        </MemoryRouter>,
      ),
    ).toThrow("useAuth must be used within an AuthProvider")

    spy.mockRestore()
  })
})

/**
 * Builds a fake JWT (header.payload.signature) with the given payload fields.
 * The signature is not valid — this is only for testing token decoding.
 */
function buildFakeJwt(payload: Record<string, unknown>): string {
  const header = btoa(JSON.stringify({ alg: "HS256", typ: "JWT" }))
  const body = btoa(JSON.stringify({ exp: Math.floor(Date.now() / 1000) + 3600, iat: Math.floor(Date.now() / 1000), ...payload }))
  const signature = "fake-signature"
  return `${header}.${body}.${signature}`
}
