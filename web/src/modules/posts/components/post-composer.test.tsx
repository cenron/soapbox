import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, fireEvent, act } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { MemoryRouter } from "react-router"
import { PostComposer } from "./post-composer"

vi.mock("@/shared/auth/auth-context", () => ({
  useAuth: () => ({
    user: { id: "1", username: "testuser", displayName: "Test User", role: "user", verified: false },
    isAuthenticated: true,
    isLoading: false,
    login: vi.fn(),
    logout: vi.fn(),
  }),
}))

let capturedMutate: ReturnType<typeof vi.fn>
let capturedOnSuccess: (() => void) | undefined

vi.mock("@/shared/api/generated/@tanstack/react-query.gen", () => ({
  postPostsMutation: () => ({ mutationKey: ["createPost"] }),
  getPostsByIdRepliesQueryKey: (opts: { path: { id: string } }) => ["posts", opts.path.id, "replies"],
}))

vi.mock("@tanstack/react-query", async () => {
  const actual = await vi.importActual("@tanstack/react-query")
  return {
    ...actual,
    useMutation: (opts: { mutationKey: string[]; onSuccess?: () => void }) => {
      capturedMutate = vi.fn()
      capturedOnSuccess = opts.onSuccess
      return { mutate: capturedMutate, isPending: false, isError: false, error: null }
    },
  }
})

function renderComposer(props: { parentId?: string; onSuccess?: () => void } = {}) {
  return render(
    <QueryClientProvider client={new QueryClient()}>
      <MemoryRouter>
        <PostComposer {...props} />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe("PostComposer", () => {
  it("renders textarea and post button", () => {
    renderComposer()
    expect(screen.getByPlaceholderText("What's happening?")).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Post" })).toBeInTheDocument()
  })

  it("shows character count starting at 280", () => {
    renderComposer()
    expect(screen.getByText("280")).toBeInTheDocument()
  })

  it("updates character count as user types", () => {
    renderComposer()
    const textarea = screen.getByPlaceholderText("What's happening?")
    fireEvent.change(textarea, { target: { value: "Hello" } })
    expect(screen.getByText("275")).toBeInTheDocument()
  })

  it("disables post button when textarea is empty", () => {
    renderComposer()
    expect(screen.getByRole("button", { name: "Post" })).toBeDisabled()
  })

  it("enables post button when textarea has content", () => {
    renderComposer()
    const textarea = screen.getByPlaceholderText("What's happening?")
    fireEvent.change(textarea, { target: { value: "Hello world" } })
    expect(screen.getByRole("button", { name: "Post" })).toBeEnabled()
  })

  it("calls mutation with body on submit", () => {
    renderComposer()
    const textarea = screen.getByPlaceholderText("What's happening?")
    fireEvent.change(textarea, { target: { value: "My post" } })
    screen.getByRole("button", { name: "Post" }).click()
    expect(capturedMutate).toHaveBeenCalledWith({
      body: { body: "My post", media_ids: undefined, parent_id: undefined },
    })
  })

  it("includes parent_id when in reply mode", () => {
    renderComposer({ parentId: "abc-123" })
    const textarea = screen.getByPlaceholderText("What's happening?")
    fireEvent.change(textarea, { target: { value: "A reply" } })
    screen.getByRole("button", { name: "Post" }).click()
    expect(capturedMutate).toHaveBeenCalledWith({
      body: { body: "A reply", media_ids: undefined, parent_id: "abc-123" },
    })
  })

  it("clears textarea on successful post", () => {
    renderComposer()
    const textarea = screen.getByPlaceholderText("What's happening?")
    fireEvent.change(textarea, { target: { value: "My post" } })

    act(() => {
      capturedOnSuccess?.()
    })

    expect(textarea).toHaveValue("")
  })
})
