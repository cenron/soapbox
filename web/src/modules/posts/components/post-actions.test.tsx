import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { MemoryRouter } from "react-router"
import { PostActions } from "./post-actions"
import type { PostsPostResponse } from "@/shared/api/generated/types.gen"

vi.mock("@/shared/auth/auth-context", () => ({
  useAuth: () => ({
    user: { id: "user-1", username: "viewer", displayName: "Viewer", role: "user", verified: false },
    isAuthenticated: true,
    isLoading: false,
    login: vi.fn(),
    logout: vi.fn(),
  }),
}))

let likeMutate: ReturnType<typeof vi.fn>
let unlikeMutate: ReturnType<typeof vi.fn>

vi.mock("@/shared/api/generated/@tanstack/react-query.gen", () => ({
  postPostsByIdLikeMutation: () => ({ mutationKey: ["like"] }),
  deletePostsByIdLikeMutation: () => ({ mutationKey: ["unlike"] }),
  postPostsByIdRepostMutation: () => ({ mutationKey: ["repost"] }),
  deletePostsByIdRepostMutation: () => ({ mutationKey: ["undoRepost"] }),
  deletePostsByIdMutation: () => ({ mutationKey: ["deletePost"] }),
  getPostsByIdQueryKey: () => ["post"],
}))

vi.mock("@tanstack/react-query", async () => {
  const actual = await vi.importActual("@tanstack/react-query")
  return {
    ...actual,
    useMutation: (opts: { mutationKey: string[] }) => {
      const mutate = vi.fn()
      if (opts.mutationKey[0] === "like") likeMutate = mutate
      if (opts.mutationKey[0] === "unlike") unlikeMutate = mutate
      return { mutate, isPending: false }
    },
  }
})

const basePost: PostsPostResponse = {
  id: "post-1",
  author_id: "author-1",
  author_username: "alice",
  body: "Hello",
  like_count: 3,
  repost_count: 1,
  reply_count: 2,
  liked_by_me: false,
  reposted_by_me: false,
}

function renderActions(post: PostsPostResponse = basePost) {
  return render(
    <QueryClientProvider client={new QueryClient()}>
      <MemoryRouter>
        <PostActions post={post} />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe("PostActions", () => {
  it("displays like count", () => {
    renderActions()
    expect(screen.getByText("3")).toBeInTheDocument()
  })

  it("displays reply count", () => {
    renderActions()
    expect(screen.getByText("2")).toBeInTheDocument()
  })

  it("displays repost count", () => {
    renderActions()
    expect(screen.getByText("1")).toBeInTheDocument()
  })

  it("calls like mutation when not liked", () => {
    renderActions()
    const likeButton = screen.getByTitle("Like")
    likeButton.click()
    expect(likeMutate).toHaveBeenCalledWith({ path: { id: "post-1" } })
  })

  it("calls unlike mutation when already liked", () => {
    renderActions({ ...basePost, liked_by_me: true })
    const likeButton = screen.getByTitle("Unlike")
    likeButton.click()
    expect(unlikeMutate).toHaveBeenCalledWith({ path: { id: "post-1" } })
  })

  it("does not show delete button when user is not the author", () => {
    renderActions()
    expect(screen.queryByTitle("Delete")).toBeNull()
  })

  it("shows delete button when user is the author", () => {
    renderActions({ ...basePost, author_id: "user-1" })
    expect(screen.getByTitle("Delete")).toBeInTheDocument()
  })
})
