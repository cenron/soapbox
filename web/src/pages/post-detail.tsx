import { useParams } from "react-router"
import { useQuery } from "@tanstack/react-query"
import { getPostsByIdOptions } from "@/shared/api/generated/@tanstack/react-query.gen"
import { PostCard } from "@/modules/posts/components/post-card"
import { ThreadView } from "@/modules/posts/components/thread-view"
import { ReplyComposer } from "@/modules/posts/components/reply-composer"
import { Separator } from "@/shared/ui/separator"
import { useAuth } from "@/shared/auth/auth-context"

export function PostDetailPage() {
  const { id = "" } = useParams()
  const { isAuthenticated } = useAuth()

  const postQuery = useQuery({
    ...getPostsByIdOptions({ path: { id } }),
    enabled: id.length > 0,
  })

  if (postQuery.isLoading) {
    return <p className="p-6 text-muted-foreground">Loading...</p>
  }

  if (postQuery.isError) {
    const errorMessage = (postQuery.error as { message?: string })?.message
    const isNotFound = errorMessage === "not found"

    return (
      <div className="p-6">
        <h1 className="text-xl font-bold">
          {isNotFound ? "Post not found" : "Something went wrong"}
        </h1>
        <p className="text-muted-foreground">
          {isNotFound
            ? "This post does not exist or has been deleted."
            : "Failed to load post. Please try again."}
        </p>
      </div>
    )
  }

  if (!postQuery.data) {
    return (
      <div className="p-6">
        <h1 className="text-xl font-bold">Post not found</h1>
        <p className="text-muted-foreground">This post does not exist or has been deleted.</p>
      </div>
    )
  }

  const post = postQuery.data
  const authorUsername = post.author_username ?? ""

  return (
    <div>
      <PostCard post={post} />

      <Separator />

      {isAuthenticated && (
        <ReplyComposer parentId={id} parentUsername={authorUsername} />
      )}

      <Separator />

      <ThreadView postId={id} />
    </div>
  )
}
