import { PostComposer } from "./post-composer"

interface ReplyComposerProps {
  parentId: string
  parentUsername: string
  onSuccess?: () => void
}

export function ReplyComposer({ parentId, parentUsername, onSuccess }: ReplyComposerProps) {
  return (
    <div>
      <p className="px-4 pt-4 text-sm text-muted-foreground">
        Replying to{" "}
        <span className="text-primary">@{parentUsername}</span>
      </p>
      <PostComposer parentId={parentId} onSuccess={onSuccess} />
    </div>
  )
}
