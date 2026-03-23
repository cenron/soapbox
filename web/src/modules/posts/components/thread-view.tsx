import { useQuery } from "@tanstack/react-query"
import { getPostsByIdRepliesOptions } from "@/shared/api/generated/@tanstack/react-query.gen"
import { PostCard } from "./post-card"

interface ThreadViewProps {
  postId: string
}

export function ThreadView({ postId }: ThreadViewProps) {
  const repliesQuery = useQuery(
    getPostsByIdRepliesOptions({ path: { id: postId } }),
  )

  if (repliesQuery.isLoading) {
    return <p className="p-4 text-sm text-muted-foreground">Loading replies...</p>
  }

  if (repliesQuery.isError) {
    return <p className="p-4 text-sm text-red-500">Failed to load replies.</p>
  }

  const replies = repliesQuery.data?.items ?? []

  if (replies.length === 0) {
    return <p className="p-4 text-sm text-muted-foreground">No replies yet.</p>
  }

  return (
    <div>
      {replies.map((reply) => (
        <PostCard key={reply.id} post={reply} />
      ))}
    </div>
  )
}
