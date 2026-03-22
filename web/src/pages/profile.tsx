import { useParams } from "react-router"
import { useQuery } from "@tanstack/react-query"
import {
  getUsersByUsernameOptions,
  getUsersByUsernameFollowersOptions,
  getUsersByUsernameFollowingOptions,
} from "@/shared/api/generated/@tanstack/react-query.gen"
import { useAuth } from "@/shared/auth/auth-context"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/shared/ui/tabs"
import { Separator } from "@/shared/ui/separator"
import { ProfileHeader } from "@/modules/users/components/profile-header"
import { UserCard } from "@/modules/users/components/user-card"

export function ProfilePage() {
  const { username: rawUsername = "" } = useParams()
  const username = rawUsername.replace(/^@/, "")
  const { user: currentUser } = useAuth()

  const profileQuery = useQuery({
    ...getUsersByUsernameOptions({ path: { username } }),
    enabled: username.length > 0,
  })

  const followersQuery = useQuery({
    ...getUsersByUsernameFollowersOptions({ path: { username } }),
    enabled: username.length > 0,
  })

  const followingQuery = useQuery({
    ...getUsersByUsernameFollowingOptions({ path: { username } }),
    enabled: username.length > 0,
  })

  if (profileQuery.isLoading) {
    return (
      <div className="p-6">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  if (profileQuery.isError || !profileQuery.data) {
    return (
      <div className="p-6">
        <h1 className="text-xl font-bold">User not found</h1>
        <p className="text-muted-foreground">@{username} does not exist.</p>
      </div>
    )
  }

  const profile = profileQuery.data
  const isOwnProfile = currentUser?.username === username

  const followers = followersQuery.data?.items ?? []
  const following = followingQuery.data?.items ?? []

  return (
    <div className="mx-auto max-w-xl p-6">
      <ProfileHeader profile={profile} isOwnProfile={isOwnProfile} />

      <Separator className="my-6" />

      <Tabs defaultValue="posts">
        <TabsList>
          <TabsTrigger value="posts">Posts</TabsTrigger>
          <TabsTrigger value="followers">
            Followers ({profile.follower_count ?? 0})
          </TabsTrigger>
          <TabsTrigger value="following">
            Following ({profile.following_count ?? 0})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="posts" className="mt-4">
          <p className="text-sm text-muted-foreground">No posts yet.</p>
        </TabsContent>

        <TabsContent value="followers" className="mt-4">
          {followersQuery.isLoading && (
            <p className="text-sm text-muted-foreground">Loading...</p>
          )}
          {!followersQuery.isLoading && followers.length === 0 && (
            <p className="text-sm text-muted-foreground">No followers yet.</p>
          )}
          {followers.map((u) => (
            <UserCard key={u.id} user={u} />
          ))}
        </TabsContent>

        <TabsContent value="following" className="mt-4">
          {followingQuery.isLoading && (
            <p className="text-sm text-muted-foreground">Loading...</p>
          )}
          {!followingQuery.isLoading && following.length === 0 && (
            <p className="text-sm text-muted-foreground">Not following anyone yet.</p>
          )}
          {following.map((u) => (
            <UserCard key={u.id} user={u} />
          ))}
        </TabsContent>
      </Tabs>
    </div>
  )
}
