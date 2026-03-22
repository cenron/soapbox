import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import {
  postUsersByUsernameFollowMutation,
  deleteUsersByUsernameFollowMutation,
  getUsersByUsernameQueryKey,
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

  const followMutation = useMutation({
    ...postUsersByUsernameFollowMutation(),
    onSuccess() {
      void queryClient.invalidateQueries({
        queryKey: getUsersByUsernameQueryKey({ path: { username } }),
      })
      onToggle?.()
    },
  })

  const unfollowMutation = useMutation({
    ...deleteUsersByUsernameFollowMutation(),
    onSuccess() {
      void queryClient.invalidateQueries({
        queryKey: getUsersByUsernameQueryKey({ path: { username } }),
      })
      onToggle?.()
    },
  })

  if (user?.username === username) return null

  const isPending = followMutation.isPending || unfollowMutation.isPending

  function handleClick() {
    if (isFollowing) {
      unfollowMutation.mutate({ path: { username } })
    } else {
      followMutation.mutate({ path: { username } })
    }
  }

  if (isFollowing) {
    return (
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
    )
  }

  return (
    <Button size="sm" disabled={isPending} onClick={handleClick}>
      Follow
    </Button>
  )
}
