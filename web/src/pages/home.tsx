import { useQueryClient } from "@tanstack/react-query"
import { PostComposer } from "@/modules/posts/components/post-composer"
import { Timeline } from "@/modules/feed/components/timeline"
import { getFeedInfiniteQueryKey } from "@/shared/api/generated/@tanstack/react-query.gen"
import { useAuth } from "@/shared/auth/auth-context"

export function HomePage() {
  const { isAuthenticated } = useAuth()
  const queryClient = useQueryClient()

  function handlePostCreated() {
    void queryClient.invalidateQueries({
      queryKey: getFeedInfiniteQueryKey(),
    })
  }

  return (
    <div>
      {isAuthenticated && <PostComposer onSuccess={handlePostCreated} />}
      <Timeline />
    </div>
  )
}
