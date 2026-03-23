import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { Image } from "lucide-react"
import {
  postPostsMutation,
  getPostsByIdRepliesQueryKey,
} from "@/shared/api/generated/@tanstack/react-query.gen"
import { useAuth } from "@/shared/auth/auth-context"
import { Avatar, AvatarFallback, AvatarImage } from "@/shared/ui/avatar"
import { Button } from "@/shared/ui/button"
import { Textarea } from "@/shared/ui/textarea"
import { ImageUpload } from "@/modules/media/components/image-upload"
import type { MediaUploadResponse } from "@/shared/api/generated/types.gen"
import { cn } from "@/shared/lib/utils"

const MAX_CHARS = 280
const WARN_THRESHOLD = 20

interface PostComposerProps {
  parentId?: string
  onSuccess?: () => void
}

export function PostComposer({ parentId, onSuccess }: PostComposerProps) {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const [body, setBody] = useState("")
  const [mediaIds, setMediaIds] = useState<string[]>([])
  const [showImageUpload, setShowImageUpload] = useState(false)

  const mutation = useMutation({
    ...postPostsMutation(),
    onSuccess: () => {
      setBody("")
      setMediaIds([])
      setShowImageUpload(false)

      if (parentId) {
        void queryClient.invalidateQueries({
          queryKey: getPostsByIdRepliesQueryKey({ path: { id: parentId } }),
        })
      }

      onSuccess?.()
    },
  })

  const charsLeft = MAX_CHARS - body.length
  const isOverLimit = charsLeft < 0
  const isNearLimit = charsLeft <= WARN_THRESHOLD && !isOverLimit
  const isEmpty = body.trim().length === 0
  const isDisabled = isEmpty || isOverLimit || mutation.isPending

  function handleSubmit() {
    if (isDisabled) return

    mutation.mutate({
      body: {
        body: body.trim(),
        media_ids: mediaIds.length > 0 ? mediaIds : undefined,
        parent_id: parentId,
      },
    })
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
      handleSubmit()
    }
  }

  function handleUploadComplete(upload: MediaUploadResponse) {
    if (upload.id) {
      setMediaIds((prev) => [...prev, upload.id as string])
    }
  }

  if (!user) return null

  const initial = user.displayName.charAt(0).toUpperCase()

  return (
    <div className="border-b border-border p-4">
      <div className="flex gap-3">
        <Avatar size="lg">
          {user.avatarUrl && <AvatarImage src={user.avatarUrl} alt={user.displayName} />}
          <AvatarFallback>{initial}</AvatarFallback>
        </Avatar>

        <div className="flex-1 space-y-3">
          <Textarea
            value={body}
            onChange={(e) => setBody(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="What's happening?"
            className="min-h-[80px] resize-none border-none p-0 text-base shadow-none focus-visible:ring-0"
            disabled={mutation.isPending}
          />

          {showImageUpload && (
            <ImageUpload
              onUploadComplete={handleUploadComplete}
              className="mt-2"
            />
          )}

          <div className="flex items-center justify-between border-t border-border pt-3">
            <div className="flex items-center gap-2">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="h-8 w-8 p-0 text-primary hover:bg-primary/10"
                onClick={() => setShowImageUpload((v) => !v)}
                title="Add image"
              >
                <Image className="h-4 w-4" />
              </Button>
            </div>

            <div className="flex items-center gap-3">
              <span
                className={cn(
                  "text-sm tabular-nums",
                  isOverLimit && "text-destructive font-medium",
                  isNearLimit && "text-yellow-500",
                  !isOverLimit && !isNearLimit && "text-muted-foreground",
                )}
              >
                {charsLeft}
              </span>

              <Button
                size="sm"
                onClick={handleSubmit}
                disabled={isDisabled}
              >
                {mutation.isPending ? "Posting..." : "Post"}
              </Button>
            </div>
          </div>

          {mutation.isError && (
            <p className="text-sm text-destructive">
              {(mutation.error as { detail?: string; message?: string })?.detail ??
                (mutation.error as { detail?: string; message?: string })?.message ??
                "Failed to post. Please try again."}
            </p>
          )}
        </div>
      </div>
    </div>
  )
}
