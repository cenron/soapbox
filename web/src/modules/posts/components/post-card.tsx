import { Link, useNavigate } from "react-router"
import { BadgeCheck } from "lucide-react"
import type { PostsPostResponse } from "@/shared/api/generated/types.gen"
import { Avatar, AvatarFallback, AvatarImage } from "@/shared/ui/avatar"
import { ImageDisplay } from "@/modules/media/components/image-display"
import { HashtagText } from "./hashtag-text"
import { LinkPreviewCard } from "./link-preview-card"
import { PostActions } from "./post-actions"

function timeAgo(dateString: string | undefined): string {
  if (!dateString) return ""

  const now = Date.now()
  const then = new Date(dateString).getTime()
  const diff = Math.floor((now - then) / 1000)

  if (diff < 60) return `${diff}s`
  if (diff < 3600) return `${Math.floor(diff / 60)}m`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h`
  if (diff < 604800) return `${Math.floor(diff / 86400)}d`

  return new Date(dateString).toLocaleDateString()
}

interface PostCardProps {
  post: PostsPostResponse
}

export function PostCard({ post }: PostCardProps) {
  const navigate = useNavigate()

  const username = post.author_username ?? ""
  const displayName = post.author_display_name ?? username
  const initial = displayName.charAt(0).toUpperCase()
  const media = post.media ?? []

  function handleCardClick() {
    if (post.id) {
      void navigate(`/post/${post.id}`)
    }
  }

  return (
    <article
      className="cursor-pointer border-b border-border p-4 transition-colors hover:bg-muted/30"
      onClick={handleCardClick}
    >
      <div className="flex gap-3">
        <Link
          to={`/@${username}`}
          className="shrink-0"
          onClick={(e) => e.stopPropagation()}
        >
          <Avatar size="lg">
            {post.author_avatar_url && (
              <AvatarImage src={post.author_avatar_url} alt={displayName} />
            )}
            <AvatarFallback>{initial}</AvatarFallback>
          </Avatar>
        </Link>

        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <Link
              to={`/@${username}`}
              className="flex items-center gap-1 font-medium hover:underline"
              onClick={(e) => e.stopPropagation()}
            >
              <span className="truncate">{displayName}</span>
              {post.author_verified && (
                <BadgeCheck className="h-4 w-4 shrink-0 text-primary" />
              )}
            </Link>
            <span className="text-sm text-muted-foreground">@{username}</span>
            <span className="text-muted-foreground">·</span>
            <span className="shrink-0 text-sm text-muted-foreground">
              {timeAgo(post.created_at)}
            </span>
          </div>

          {post.body && (
            <p className="mt-1 whitespace-pre-wrap break-words text-sm leading-relaxed">
              <HashtagText body={post.body} />
            </p>
          )}

          {media.length > 0 && (
            <div
              className={`mt-3 grid gap-1 overflow-hidden rounded-lg ${media.length === 1 ? "grid-cols-1" : "grid-cols-2"}`}
            >
              {media
                .slice()
                .sort((a, b) => (a.position ?? 0) - (b.position ?? 0))
                .map((m, i) => (
                  <ImageDisplay
                    key={m.id}
                    src={m.media_url ?? ""}
                    alt={`Image attachment ${i + 1}`}
                    className="aspect-video"
                  />
                ))}
            </div>
          )}

          {post.link_preview && (
            <LinkPreviewCard preview={post.link_preview} />
          )}

          <PostActions post={post} />
        </div>
      </div>
    </article>
  )
}
