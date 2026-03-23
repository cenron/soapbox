import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { MemoryRouter } from "react-router"
import { Timeline } from "./timeline"

// Mock the generated API query options.
vi.mock("@/shared/api/generated/@tanstack/react-query.gen", () => ({
  getFeedInfiniteOptions: () => ({
    queryKey: ["getFeed"],
    queryFn: vi.fn().mockResolvedValue({
      items: [],
      has_more: false,
    }),
  }),
  getFeedInfiniteQueryKey: () => ["getFeed"],
}))

// Mock the WS provider.
vi.mock("@/shared/ws/use-ws-client", () => ({
  useWsClient: () => null,
}))

// Mock auth context.
vi.mock("@/shared/auth/auth-context", () => ({
  useAuth: () => ({ isAuthenticated: true, user: null }),
}))

function renderTimeline() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <Timeline />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe("Timeline", () => {
  it("shows empty state when no posts", async () => {
    renderTimeline()

    const emptyMessage = await screen.findByText(
      /timeline is empty/i,
    )
    expect(emptyMessage).toBeInTheDocument()
  })
})
