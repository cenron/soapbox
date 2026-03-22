import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, act } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { MemoryRouter } from "react-router"
import { FollowButton } from "./follow-button"

// Mock auth — logged in as a different user than the target
vi.mock("@/shared/auth/auth-context", () => ({
  useAuth: () => ({
    user: { id: "1", username: "viewer", displayName: "Viewer", role: "user", verified: false },
    isAuthenticated: true,
    isLoading: false,
    login: vi.fn(),
    logout: vi.fn(),
  }),
}))

// Track which mutations fire and capture their callbacks
let followMutate: ReturnType<typeof vi.fn>
let unfollowMutate: ReturnType<typeof vi.fn>
let capturedFollowOnSuccess: (() => void) | undefined
let capturedFollowOnError: ((err: { detail?: string }) => void) | undefined

vi.mock("@/shared/api/generated/@tanstack/react-query.gen", () => ({
  postUsersByUsernameFollowMutation: () => ({ mutationKey: ["follow"] }),
  deleteUsersByUsernameFollowMutation: () => ({ mutationKey: ["unfollow"] }),
  getUsersByUsernameQueryKey: (opts: { path: { username: string } }) => ["users", opts.path.username],
  getUsersByUsernameFollowersQueryKey: (opts: { path: { username: string } }) => ["users", opts.path.username, "followers"],
  getUsersByUsernameFollowingQueryKey: (opts: { path: { username: string } }) => ["users", opts.path.username, "following"],
}))

// Intercept useMutation to capture onSuccess/onError callbacks
vi.mock("@tanstack/react-query", async () => {
  const actual = await vi.importActual("@tanstack/react-query")
  return {
    ...actual,
    useMutation: (opts: { mutationKey: string[]; onSuccess?: () => void; onError?: (err: { detail?: string }) => void }) => {
      if (opts.mutationKey[0] === "follow") {
        followMutate = vi.fn()
        capturedFollowOnSuccess = opts.onSuccess
        capturedFollowOnError = opts.onError
        return { mutate: followMutate, isPending: false }
      }
      unfollowMutate = vi.fn()
      return { mutate: unfollowMutate, isPending: false }
    },
  }
})

let queryClient: QueryClient

function renderFollowButton(props: { username: string; isFollowing: boolean; onToggle?: () => void }) {
  queryClient = new QueryClient()
  const spy = vi.spyOn(queryClient, "invalidateQueries")

  render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <FollowButton {...props} />
      </MemoryRouter>
    </QueryClientProvider>,
  )

  return spy
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe("FollowButton", () => {
  it("shows Follow button when not following", () => {
    renderFollowButton({ username: "admin", isFollowing: false })
    expect(screen.getByRole("button", { name: "Follow" })).toBeInTheDocument()
  })

  it("shows Following button when following", () => {
    renderFollowButton({ username: "admin", isFollowing: true })
    expect(screen.getByRole("button", { name: "Following" })).toBeInTheDocument()
  })

  it("does not render when viewing own profile", () => {
    const { container } = render(
      <QueryClientProvider client={new QueryClient()}>
        <MemoryRouter>
          <FollowButton username="viewer" isFollowing={false} />
        </MemoryRouter>
      </QueryClientProvider>,
    )
    expect(container.innerHTML).toBe("")
  })

  it("calls follow mutation when clicking Follow", () => {
    renderFollowButton({ username: "admin", isFollowing: false })
    screen.getByRole("button", { name: "Follow" }).click()
    expect(followMutate).toHaveBeenCalledWith({ path: { username: "admin" } })
  })

  it("calls unfollow mutation when clicking Following", () => {
    renderFollowButton({ username: "admin", isFollowing: true })
    screen.getByRole("button", { name: "Following" }).click()
    expect(unfollowMutate).toHaveBeenCalledWith({ path: { username: "admin" } })
  })

  it("invalidates profile, followers, and following queries on success", () => {
    const spy = renderFollowButton({ username: "admin", isFollowing: false })

    act(() => {
      capturedFollowOnSuccess?.()
    })

    const invalidatedKeys = spy.mock.calls.map((call) => call[0].queryKey)
    expect(invalidatedKeys).toContainEqual(["users", "admin"])
    expect(invalidatedKeys).toContainEqual(["users", "admin", "followers"])
    expect(invalidatedKeys).toContainEqual(["users", "admin", "following"])
  })

  it("invalidates profile, followers, and following queries on error", () => {
    const spy = renderFollowButton({ username: "admin", isFollowing: false })

    act(() => {
      capturedFollowOnError?.({ detail: "already following" })
    })

    const invalidatedKeys = spy.mock.calls.map((call) => call[0].queryKey)
    expect(invalidatedKeys).toContainEqual(["users", "admin"])
    expect(invalidatedKeys).toContainEqual(["users", "admin", "followers"])
    expect(invalidatedKeys).toContainEqual(["users", "admin", "following"])
  })

  it("displays error message on mutation error", () => {
    renderFollowButton({ username: "admin", isFollowing: false })

    act(() => {
      capturedFollowOnError?.({ detail: "already following" })
    })

    expect(screen.getByText("already following")).toBeInTheDocument()
  })
})
