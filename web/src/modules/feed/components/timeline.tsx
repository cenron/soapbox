import { useCallback, useEffect, useRef } from "react"
import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query"
import { Loader2 } from "lucide-react"
import type { PostsPostResponse } from "@/shared/api/generated/types.gen"
import {
  getFeedInfiniteOptions,
  getFeedInfiniteQueryKey,
} from "@/shared/api/generated/@tanstack/react-query.gen"
import { PostCard } from "@/modules/posts/components/post-card"
import { NewPostsBanner } from "./new-posts-banner"
import { useNewPosts } from "../hooks/use-new-posts"

export function Timeline() {
  const queryClient = useQueryClient()
  const { count: newPostsCount, reset: resetNewPosts } = useNewPosts()
  const sentinelRef = useRef<HTMLDivElement>(null)

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
    isError,
  } = useInfiniteQuery({
    ...getFeedInfiniteOptions(),
    initialPageParam: "" as string,
    getNextPageParam: (lastPage) =>
      lastPage.has_more ? (lastPage.next_cursor ?? "") : undefined,
  })

  // Infinite scroll via IntersectionObserver.
  const handleObserver = useCallback(
    (entries: IntersectionObserverEntry[]) => {
      const target = entries[0]
      if (target.isIntersecting && hasNextPage && !isFetchingNextPage) {
        void fetchNextPage()
      }
    },
    [fetchNextPage, hasNextPage, isFetchingNextPage],
  )

  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return

    const observer = new IntersectionObserver(handleObserver, {
      rootMargin: "200px",
    })
    observer.observe(sentinel)

    return () => observer.disconnect()
  }, [handleObserver])

  function handleNewPostsClick() {
    resetNewPosts()
    void queryClient.invalidateQueries({
      queryKey: getFeedInfiniteQueryKey(),
    })
  }

  const posts =
    data?.pages.flatMap((page) => page.items ?? []) ?? []

  if (isLoading) {
    return (
      <div className="flex justify-center p-8">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (isError) {
    return (
      <div className="p-6 text-center text-sm text-destructive">
        Failed to load timeline. Please try again.
      </div>
    )
  }

  if (posts.length === 0 && newPostsCount === 0) {
    return (
      <div className="p-6 text-center text-sm text-muted-foreground">
        Your timeline is empty. Follow some people to see their posts here.
      </div>
    )
  }

  return (
    <div>
      <NewPostsBanner count={newPostsCount} onClick={handleNewPostsClick} />

      {posts.map((post) => (
        <PostCard
          key={post.id}
          post={post as unknown as PostsPostResponse}
        />
      ))}

      <div ref={sentinelRef} className="h-1" />

      {isFetchingNextPage && (
        <div className="flex justify-center p-4">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </div>
      )}
    </div>
  )
}
