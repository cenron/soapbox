import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { MemoryRouter } from "react-router"
import { PostCard } from "./post-card"
import type { PostsPostResponse } from "@/shared/api/generated/types.gen"

vi.mock("@/shared/auth/auth-context", () => ({
  useAuth: () => ({
    user: { id: "1", username: "viewer", displayName: "Viewer", role: "user", verified: false },
    isAuthenticated: true,
    isLoading: false,
    login: vi.fn(),
    logout: vi.fn(),
  }),
}))

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
    useMutation: () => ({
      mutate: vi.fn(),
      isPending: false,
    }),
  }
})

const basePost: PostsPostResponse = {
  id: "post-1",
  author_id: "author-1",
  author_username: "alice",
  author_display_name: "Alice",
  author_avatar_url: "",
  author_verified: false,
  body: "Hello world #test",
  media: [],
  hashtags: ["test"],
  like_count: 5,
  repost_count: 2,
  reply_count: 1,
  liked_by_me: false,
  reposted_by_me: false,
  created_at: new Date().toISOString(),
}

function renderPostCard(post: PostsPostResponse = basePost) {
  return render(
    <QueryClientProvider client={new QueryClient()}>
      <MemoryRouter>
        <PostCard post={post} />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe("PostCard", () => {
  it("displays author display name and username", () => {
    renderPostCard()
    expect(screen.getByText("Alice")).toBeInTheDocument()
    expect(screen.getByText("@alice")).toBeInTheDocument()
  })

  it("displays post body with hashtag links", () => {
    renderPostCard()
    expect(screen.getByText("Hello world")).toBeInTheDocument()
    expect(screen.getByText("#test")).toBeInTheDocument()
  })

  it("shows verified badge when author is verified", () => {
    renderPostCard({ ...basePost, author_verified: true })
    const badge = document.querySelector(".lucide-badge-check")
    expect(badge).toBeInTheDocument()
  })

  it("does not show verified badge when author is not verified", () => {
    renderPostCard()
    const badge = document.querySelector(".lucide-badge-check")
    expect(badge).toBeNull()
  })

  it("renders like count", () => {
    renderPostCard()
    expect(screen.getByText("5")).toBeInTheDocument()
  })

  it("shows link preview when present", () => {
    renderPostCard({
      ...basePost,
      link_preview: {
        url: "https://example.com",
        title: "Example Site",
        description: "A great site",
        image_url: "",
      },
    })
    expect(screen.getByText("Example Site")).toBeInTheDocument()
    expect(screen.getByText("A great site")).toBeInTheDocument()
  })
})
