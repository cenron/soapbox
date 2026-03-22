import { Link } from "react-router"
import type { UsersProfileResponse } from "@/shared/api/generated/types.gen"
import { Avatar, AvatarFallback, AvatarImage } from "@/shared/ui/avatar"
import { Badge } from "@/shared/ui/badge"
import { Button } from "@/shared/ui/button"
import { FollowButton } from "./follow-button"

interface ProfileHeaderProps {
  profile: UsersProfileResponse
  isOwnProfile: boolean
}

export function ProfileHeader({ profile, isOwnProfile }: ProfileHeaderProps) {
  const username = profile.username ?? ""
  const displayName = profile.display_name ?? username
  const initial = displayName.charAt(0).toUpperCase()

  return (
    <div className="space-y-4">
      <div className="flex items-start justify-between gap-4">
        <Avatar className="size-20 text-2xl">
          {profile.avatar_url && (
            <AvatarImage src={profile.avatar_url} alt={displayName} />
          )}
          <AvatarFallback>{initial}</AvatarFallback>
        </Avatar>

        <div className="pt-1">
          {isOwnProfile ? (
            <Button variant="outline" size="sm" asChild>
              <Link to="/settings">Edit profile</Link>
            </Button>
          ) : (
            <FollowButton
              username={username}
              isFollowing={profile.is_following ?? false}
            />
          )}
        </div>
      </div>

      <div className="space-y-1">
        <div className="flex items-center gap-2">
          <h1 className="text-xl font-bold">{displayName}</h1>
          {profile.verified && <Badge variant="secondary">Verified</Badge>}
        </div>
        <p className="text-sm text-muted-foreground">@{username}</p>
      </div>

      {profile.bio && <p className="text-sm">{profile.bio}</p>}

      <div className="flex gap-4 text-sm">
        <span>
          <strong>{profile.follower_count ?? 0}</strong>{" "}
          <span className="text-muted-foreground">Followers</span>
        </span>
        <span>
          <strong>{profile.following_count ?? 0}</strong>{" "}
          <span className="text-muted-foreground">Following</span>
        </span>
      </div>
    </div>
  )
}
