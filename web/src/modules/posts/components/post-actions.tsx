import { useMutation, useQueryClient } from "@tanstack/react-query"
import { Heart, Repeat2, MessageCircle, Trash2 } from "lucide-react"
import { useNavigate } from "react-router"
import {
  postPostsByIdLikeMutation,
  deletePostsByIdLikeMutation,
  postPostsByIdRepostMutation,
  deletePostsByIdRepostMutation,
  deletePostsByIdMutation,
  getPostsByIdQueryKey,
} from "@/shared/api/generated/@tanstack/react-query.gen"
import type { PostsPostResponse } from "@/shared/api/generated/types.gen"
import { useAuth } from "@/shared/auth/auth-context"
import { Button } from "@/shared/ui/button"
import { cn } from "@/shared/lib/utils"

interface PostActionsProps {
  post: PostsPostResponse
}

export function PostActions({ post }: PostActionsProps) {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const postId = post.id ?? ""
  const queryKey = getPostsByIdQueryKey({ path: { id: postId } })

  function invalidatePost() {
    void queryClient.invalidateQueries({ queryKey })
  }

  const likeMutation = useMutation({
    ...postPostsByIdLikeMutation(),
    onSuccess: invalidatePost,
    onError: invalidatePost,
  })

  const unlikeMutation = useMutation({
    ...deletePostsByIdLikeMutation(),
    onSuccess: invalidatePost,
    onError: invalidatePost,
  })

  const repostMutation = useMutation({
    ...postPostsByIdRepostMutation(),
    onSuccess: invalidatePost,
    onError: invalidatePost,
  })

  const unrepostMutation = useMutation({
    ...deletePostsByIdRepostMutation(),
    onSuccess: invalidatePost,
    onError: invalidatePost,
  })

  const deleteMutation = useMutation({
    ...deletePostsByIdMutation(),
    onSuccess: () => {
      void queryClient.invalidateQueries()
      void navigate(-1)
    },
  })

  function handleLike(e: React.MouseEvent) {
    e.stopPropagation()
    if (!postId) return

    if (post.liked_by_me) {
      unlikeMutation.mutate({ path: { id: postId } })
    } else {
      likeMutation.mutate({ path: { id: postId } })
    }
  }

  function handleRepost(e: React.MouseEvent) {
    e.stopPropagation()
    if (!postId) return

    if (post.reposted_by_me) {
      unrepostMutation.mutate({ path: { id: postId } })
    } else {
      repostMutation.mutate({ path: { id: postId } })
    }
  }

  function handleReply(e: React.MouseEvent) {
    e.stopPropagation()
    void navigate(`/post/${postId}`)
  }

  function handleDelete(e: React.MouseEvent) {
    e.stopPropagation()
    if (!postId) return
    if (window.confirm("Delete this post?")) {
      deleteMutation.mutate({ path: { id: postId } })
    }
  }

  const isLikePending = likeMutation.isPending || unlikeMutation.isPending
  const isRepostPending = repostMutation.isPending || unrepostMutation.isPending
  const isOwnPost = user?.id === post.author_id

  return (
    <div className="mt-3 flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
      <Button
        variant="ghost"
        size="sm"
        className="h-8 gap-1.5 px-2 text-muted-foreground hover:text-foreground"
        title="Reply"
        onClick={handleReply}
      >
        <MessageCircle className="h-4 w-4" />
        <span className="text-xs">{post.reply_count ?? 0}</span>
      </Button>

      <Button
        variant="ghost"
        size="sm"
        className={cn(
          "h-8 gap-1.5 px-2",
          post.reposted_by_me
            ? "text-green-500 hover:text-green-600"
            : "text-muted-foreground hover:text-foreground",
        )}
        title={post.reposted_by_me ? "Undo repost" : "Repost"}
        disabled={isRepostPending}
        onClick={handleRepost}
      >
        <Repeat2 className="h-4 w-4" />
        <span className="text-xs">{post.repost_count ?? 0}</span>
      </Button>

      <Button
        variant="ghost"
        size="sm"
        className={cn(
          "h-8 gap-1.5 px-2",
          post.liked_by_me
            ? "text-red-500 hover:text-red-600"
            : "text-muted-foreground hover:text-foreground",
        )}
        title={post.liked_by_me ? "Unlike" : "Like"}
        disabled={isLikePending}
        onClick={handleLike}
      >
        <Heart className={cn("h-4 w-4", post.liked_by_me && "fill-current")} />
        <span className="text-xs">{post.like_count ?? 0}</span>
      </Button>

      {isOwnPost && (
        <Button
          variant="ghost"
          size="sm"
          className="ml-auto h-8 px-2 text-muted-foreground hover:text-destructive"
          title="Delete"
          disabled={deleteMutation.isPending}
          onClick={handleDelete}
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      )}
    </div>
  )
}
