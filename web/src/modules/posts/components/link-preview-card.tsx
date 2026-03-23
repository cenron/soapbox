import type { PostsLinkPreviewResponse } from "@/shared/api/generated/types.gen"

interface LinkPreviewCardProps {
  preview: PostsLinkPreviewResponse
}

export function LinkPreviewCard({ preview }: LinkPreviewCardProps) {
  if (!preview.url) return null

  return (
    <a
      href={preview.url}
      target="_blank"
      rel="noopener noreferrer"
      onClick={(e) => e.stopPropagation()}
      className="mt-3 block overflow-hidden rounded-lg border border-border transition-colors hover:bg-muted/50"
    >
      {preview.image_url && (
        <img
          src={preview.image_url}
          alt={preview.title ?? ""}
          className="h-40 w-full object-cover"
        />
      )}

      <div className="p-3">
        {preview.title && (
          <p className="truncate text-sm font-medium">{preview.title}</p>
        )}
        {preview.description && (
          <p className="mt-0.5 line-clamp-2 text-sm text-muted-foreground">
            {preview.description}
          </p>
        )}
        <p className="mt-1 truncate text-xs text-muted-foreground">{preview.url}</p>
      </div>
    </a>
  )
}
