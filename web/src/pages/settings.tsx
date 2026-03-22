import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  putUsersMeMutation,
  getUsersByUsernameOptions,
  getUsersByUsernameQueryKey,
} from "@/shared/api/generated/@tanstack/react-query.gen"
import type { UsersProfileResponse, UsersUpdateProfileRequest } from "@/shared/api/generated/types.gen"
import { useAuth } from "@/shared/auth/auth-context"
import { Button } from "@/shared/ui/button"
import { Input } from "@/shared/ui/input"
import { Label } from "@/shared/ui/label"
import { Textarea } from "@/shared/ui/textarea"
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card"

interface ProfileFormProps {
  profile: UsersProfileResponse
  username: string
}

function ProfileForm({ profile, username }: ProfileFormProps) {
  const queryClient = useQueryClient()

  const [displayName, setDisplayName] = useState(profile.display_name ?? "")
  const [bio, setBio] = useState(profile.bio ?? "")
  const [avatarUrl, setAvatarUrl] = useState(profile.avatar_url ?? "")
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { mutate, isPending } = useMutation({
    ...putUsersMeMutation(),
    onSuccess() {
      setSuccess(true)
      setError(null)
      void queryClient.invalidateQueries({
        queryKey: getUsersByUsernameQueryKey({ path: { username } }),
      })
    },
    onError(err: { detail?: string; message?: string }) {
      setError(err?.detail ?? err?.message ?? "Something went wrong.")
      setSuccess(false)
    },
  })

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSuccess(false)
    setError(null)

    const body: UsersUpdateProfileRequest = {
      display_name: displayName,
      bio,
      avatar_url: avatarUrl,
    }

    mutate({ body })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-1.5">
        <Label htmlFor="displayName">Display name</Label>
        <Input
          id="displayName"
          type="text"
          autoComplete="name"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="bio">Bio</Label>
        <Textarea
          id="bio"
          rows={3}
          placeholder="Tell people a little about yourself"
          value={bio}
          onChange={(e) => setBio(e.target.value)}
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="avatarUrl">Avatar URL</Label>
        <Input
          id="avatarUrl"
          type="url"
          placeholder="https://example.com/avatar.jpg"
          value={avatarUrl}
          onChange={(e) => setAvatarUrl(e.target.value)}
        />
      </div>

      {error && <p className="text-sm text-red-500">{error}</p>}
      {success && <p className="text-sm text-green-600">Profile updated.</p>}

      <Button type="submit" className="w-full" disabled={isPending}>
        {isPending ? "Saving..." : "Save changes"}
      </Button>
    </form>
  )
}

export function SettingsPage() {
  const { user } = useAuth()
  const username = user?.username ?? ""

  const profileQuery = useQuery({
    ...getUsersByUsernameOptions({ path: { username } }),
    enabled: username.length > 0,
  })

  return (
    <div className="flex min-h-[60vh] items-center justify-center p-6">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle className="text-xl">Edit profile</CardTitle>
        </CardHeader>
        <CardContent>
          {profileQuery.isLoading && (
            <p className="text-sm text-muted-foreground">Loading...</p>
          )}
          {profileQuery.isError && (
            <div className="space-y-2">
              <p className="text-sm text-red-500">Failed to load profile.</p>
              <Button variant="outline" size="sm" onClick={() => profileQuery.refetch()}>
                Retry
              </Button>
            </div>
          )}
          {profileQuery.data && (
            <ProfileForm
              key={profileQuery.data.id}
              profile={profileQuery.data}
              username={username}
            />
          )}
        </CardContent>
      </Card>
    </div>
  )
}
