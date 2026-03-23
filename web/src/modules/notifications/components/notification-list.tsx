import { useCallback, useEffect, useRef } from "react"
import { useInfiniteQuery } from "@tanstack/react-query"
import { Loader2 } from "lucide-react"
import { getNotificationsInfiniteOptions } from "@/shared/api/generated/@tanstack/react-query.gen"
import { Button } from "@/shared/ui/button"
import { NotificationItem } from "./notification-item"
import { useMarkRead } from "../hooks/use-mark-read"

export function NotificationList() {
  const sentinelRef = useRef<HTMLDivElement>(null)
  const { markRead, markAllRead } = useMarkRead()

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
    isError,
  } = useInfiniteQuery({
    ...getNotificationsInfiniteOptions(),
    initialPageParam: "" as string,
    getNextPageParam: (lastPage) =>
      lastPage.has_more ? (lastPage.next_cursor ?? "") : undefined,
  })

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

  function handleMarkRead(id: string) {
    markRead.mutate({
      path: { id },
    })
  }

  function handleMarkAllRead() {
    markAllRead.mutate({})
  }

  const notifications =
    data?.pages.flatMap((page) => page.items ?? []) ?? []

  const hasUnread = notifications.some((n) => !n.read)

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
        Failed to load notifications. Please try again.
      </div>
    )
  }

  if (notifications.length === 0) {
    return (
      <div className="p-6 text-center text-sm text-muted-foreground">
        No notifications yet. Interactions from other users will appear here.
      </div>
    )
  }

  return (
    <div>
      {hasUnread && (
        <div className="flex justify-end border-b px-4 py-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleMarkAllRead}
            disabled={markAllRead.isPending}
          >
            Mark all as read
          </Button>
        </div>
      )}

      {notifications.map((notification) => (
        <NotificationItem
          key={notification.id}
          notification={notification}
          onMarkRead={handleMarkRead}
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
