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
      onClick={() => login("token-123", { id: "1", username: "alice", role: "user" })}
    >
      login
    </button>
  )
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe("AuthProvider", () => {
  it("starts in loading state and resolves", async () => {
    vi.mocked(tokenStorage.refreshAccessToken).mockResolvedValue(null)

    render(
      <MemoryRouter>
        <AuthProvider>
          <TestConsumer />
        </AuthProvider>
      </MemoryRouter>,
    )

    // Wait for the refresh to complete
    await act(async () => {})

    expect(screen.getByTestId("loading")).toHaveTextContent("false")
    expect(screen.getByTestId("authenticated")).toHaveTextContent("false")
    expect(screen.getByTestId("user")).toHaveTextContent("none")
  })

  it("sets user after login", async () => {
    vi.mocked(tokenStorage.refreshAccessToken).mockResolvedValue(null)

    render(
      <MemoryRouter>
        <AuthProvider>
          <TestConsumer />
          <LoginButton />
        </AuthProvider>
      </MemoryRouter>,
    )

    await act(async () => {})

    await act(async () => {
      screen.getByText("login").click()
    })

    expect(screen.getByTestId("authenticated")).toHaveTextContent("true")
    expect(screen.getByTestId("user")).toHaveTextContent("alice")
    expect(tokenStorage.setAccessToken).toHaveBeenCalledWith("token-123")
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
