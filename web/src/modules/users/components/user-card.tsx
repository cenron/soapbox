import { Link } from "react-router"
import type { UsersProfileResponse } from "@/shared/api/generated/types.gen"
import { useAuth } from "@/shared/auth/auth-context"
import { Avatar, AvatarFallback, AvatarImage } from "@/shared/ui/avatar"
import { FollowButton } from "./follow-button"

interface UserCardProps {
  user: UsersProfileResponse
}

export function UserCard({ user }: UserCardProps) {
  const { user: currentUser } = useAuth()

  const username = user.username ?? ""
  const displayName = user.display_name ?? username
  const initial = displayName.charAt(0).toUpperCase()

  return (
    <div className="flex items-center gap-3 py-3">
      <Link to={`/@${username}`} className="shrink-0">
        <Avatar size="lg">
          {user.avatar_url && <AvatarImage src={user.avatar_url} alt={displayName} />}
          <AvatarFallback>{initial}</AvatarFallback>
        </Avatar>
      </Link>

      <div className="min-w-0 flex-1">
        <Link to={`/@${username}`} className="group flex flex-col">
          <span className="truncate font-medium leading-tight group-hover:underline">
            {displayName}
          </span>
          <span className="truncate text-sm text-muted-foreground">@{username}</span>
        </Link>

        {user.bio && (
          <p className="mt-0.5 line-clamp-1 text-sm text-muted-foreground">{user.bio}</p>
        )}
      </div>

      {currentUser?.username !== username && (
        <div className="shrink-0">
          <FollowButton username={username} isFollowing={user.is_following ?? false} />
        </div>
      )}
    </div>
  )
}
