import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { MemoryRouter } from "react-router"
import { Sidebar } from "./sidebar"

// Mock useMediaQuery to always return desktop
vi.mock("@/shared/hooks/use-media-query", () => ({
  useMediaQuery: vi.fn(() => true),
}))

// Mock useAuth — toggle between authenticated and unauthenticated
const mockUser = { id: "1", username: "alice", displayName: "Alice", role: "user", verified: false }
let currentUser: typeof mockUser | null = null

vi.mock("@/shared/auth/auth-context", () => ({
  useAuth: () => ({
    user: currentUser,
    isAuthenticated: currentUser !== null,
    isLoading: false,
    login: vi.fn(),
    logout: vi.fn(),
  }),
}))

function renderSidebar() {
  return render(
    <MemoryRouter>
      <Sidebar />
    </MemoryRouter>,
  )
}

describe("Sidebar nav items", () => {
  it("shows only Home and Search when logged out", () => {
    currentUser = null
    renderSidebar()

    expect(screen.getByRole("link", { name: "Home" })).toBeInTheDocument()
    expect(screen.getByRole("link", { name: "Search" })).toBeInTheDocument()
    expect(screen.queryByRole("link", { name: "Notifications" })).not.toBeInTheDocument()
    expect(screen.queryByRole("link", { name: "Profile" })).not.toBeInTheDocument()
    expect(screen.queryByRole("link", { name: "Settings" })).not.toBeInTheDocument()
  })

  it("shows all nav links when logged in", () => {
    currentUser = mockUser
    renderSidebar()

    expect(screen.getByRole("link", { name: "Home" })).toBeInTheDocument()
    expect(screen.getByRole("link", { name: "Search" })).toBeInTheDocument()
    expect(screen.getByRole("link", { name: "Notifications" })).toBeInTheDocument()
    expect(screen.getByRole("link", { name: "Profile" })).toBeInTheDocument()
    expect(screen.getByRole("link", { name: "Settings" })).toBeInTheDocument()
  })

  it("Profile link points to /@username when logged in", () => {
    currentUser = mockUser
    renderSidebar()

    const profileLink = screen.getByRole("link", { name: "Profile" })
    expect(profileLink).toHaveAttribute("href", "/@alice")
  })
})
