import { useState } from "react"
import { cn } from "@/shared/lib/utils"

interface ImageDisplayProps {
  src: string
  alt?: string
  className?: string
}

export function ImageDisplay({ src, alt = "", className }: ImageDisplayProps) {
  const [loaded, setLoaded] = useState(false)
  const [errored, setErrored] = useState(false)

  if (errored) {
    return (
      <div
        className={cn(
          "flex items-center justify-center rounded-md bg-muted text-sm text-muted-foreground",
          className,
        )}
      >
        Failed to load image
      </div>
    )
  }

  return (
    <div className={cn("relative overflow-hidden rounded-md", className)}>
      {!loaded && (
        <div className="absolute inset-0 animate-pulse bg-muted" />
      )}
      <img
        src={src}
        alt={alt}
        loading="lazy"
        onLoad={() => setLoaded(true)}
        onError={() => setErrored(true)}
        className={cn(
          "h-full w-full object-cover transition-opacity",
          loaded ? "opacity-100" : "opacity-0",
        )}
      />
    </div>
  )
}
