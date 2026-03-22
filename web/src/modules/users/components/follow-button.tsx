import { useCallback, useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import {
  postUsersByUsernameFollowMutation,
  deleteUsersByUsernameFollowMutation,
  getUsersByUsernameQueryKey,
  getUsersByUsernameFollowersQueryKey,
  getUsersByUsernameFollowingQueryKey,
} from "@/shared/api/generated/@tanstack/react-query.gen"
import { useAuth } from "@/shared/auth/auth-context"
import { Button } from "@/shared/ui/button"

interface FollowButtonProps {
  username: string
  isFollowing: boolean
  onToggle?: () => void
}

export function FollowButton({ username, isFollowing, onToggle }: FollowButtonProps) {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const [isHoveringFollow, setIsHoveringFollow] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const invalidateAndToggle = useCallback(() => {
    setError(null)
    const path = { path: { username } }
    void queryClient.invalidateQueries({ queryKey: getUsersByUsernameQueryKey(path) })
    void queryClient.invalidateQueries({ queryKey: getUsersByUsernameFollowersQueryKey(path) })
    void queryClient.invalidateQueries({ queryKey: getUsersByUsernameFollowingQueryKey(path) })
    onToggle?.()
  }, [queryClient, username, onToggle])

  const handleError = useCallback((err: { detail?: string; message?: string }) => {
    setError(err?.detail ?? err?.message ?? "Something went wrong.")
    const path = { path: { username } }
    void queryClient.invalidateQueries({ queryKey: getUsersByUsernameQueryKey(path) })
    void queryClient.invalidateQueries({ queryKey: getUsersByUsernameFollowersQueryKey(path) })
    void queryClient.invalidateQueries({ queryKey: getUsersByUsernameFollowingQueryKey(path) })
  }, [queryClient, username])

  const followMutation = useMutation({
    ...postUsersByUsernameFollowMutation(),
    onSuccess: invalidateAndToggle,
    onError: handleError,
  })

  const unfollowMutation = useMutation({
    ...deleteUsersByUsernameFollowMutation(),
    onSuccess: invalidateAndToggle,
    onError: handleError,
  })

  if (!user || user.username === username) return null

  const isPending = followMutation.isPending || unfollowMutation.isPending

  function handleClick() {
    if (isFollowing) {
      unfollowMutation.mutate({ path: { username } })
    } else {
      followMutation.mutate({ path: { username } })
    }
  }

  return (
    <div className="flex flex-col items-end gap-1">
      {isFollowing ? (
        <Button
          variant={isHoveringFollow ? "destructive" : "outline"}
          size="sm"
          disabled={isPending}
          onClick={handleClick}
          onMouseEnter={() => setIsHoveringFollow(true)}
          onMouseLeave={() => setIsHoveringFollow(false)}
        >
          {isHoveringFollow ? "Unfollow" : "Following"}
        </Button>
      ) : (
        <Button size="sm" disabled={isPending} onClick={handleClick}>
          Follow
        </Button>
      )}
      {error && <p className="text-xs text-red-500">{error}</p>}
    </div>
  )
}
