import { useCallback, useRef, useState } from "react"
import { useMutation } from "@tanstack/react-query"
import {
  postMediaUploadUrlMutation,
  postMediaByIdConfirmMutation,
} from "@/shared/api/generated/@tanstack/react-query.gen"
import type { MediaUploadResponse } from "@/shared/api/generated/types.gen"
import { Button } from "@/shared/ui/button"
import { cn } from "@/shared/lib/utils"

type UploadStatus = "idle" | "uploading" | "confirming" | "complete" | "error"

interface ImageUploadProps {
  onUploadComplete: (upload: MediaUploadResponse) => void
  accept?: string[]
  maxSizeMB?: number
  className?: string
}

const DEFAULT_ACCEPT = ["image/jpeg", "image/png", "image/gif", "image/webp"]
const DEFAULT_MAX_SIZE_MB = 10

export function ImageUpload({
  onUploadComplete,
  accept = DEFAULT_ACCEPT,
  maxSizeMB = DEFAULT_MAX_SIZE_MB,
  className,
}: ImageUploadProps) {
  const [file, setFile] = useState<File | null>(null)
  const [preview, setPreview] = useState<string | null>(null)
  const [status, setStatus] = useState<UploadStatus>("idle")
  const [progress, setProgress] = useState(0)
  const [error, setError] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const requestUrl = useMutation(postMediaUploadUrlMutation())
  const confirmUpload = useMutation(postMediaByIdConfirmMutation())

  const resetState = useCallback(() => {
    setFile(null)
    setPreview(null)
    setStatus("idle")
    setProgress(0)
    setError(null)
  }, [])

  function validateFile(f: File): string | null {
    if (!accept.includes(f.type)) {
      return `File type "${f.type}" is not supported. Use: ${accept.join(", ")}`
    }

    if (f.size > maxSizeMB * 1024 * 1024) {
      return `File is too large. Maximum size is ${maxSizeMB} MB.`
    }

    return null
  }

  function handleFileSelect(f: File) {
    const validationError = validateFile(f)
    if (validationError) {
      setError(validationError)
      return
    }

    setError(null)
    setFile(f)
    setPreview(URL.createObjectURL(f))
  }

  function handleInputChange(e: React.ChangeEvent<HTMLInputElement>) {
    const f = e.target.files?.[0]
    if (f) handleFileSelect(f)
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault()
    const f = e.dataTransfer.files[0]
    if (f) handleFileSelect(f)
  }

  function handleDragOver(e: React.DragEvent) {
    e.preventDefault()
  }

  async function uploadToS3(url: string, f: File): Promise<void> {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()

      xhr.upload.addEventListener("progress", (e) => {
        if (e.lengthComputable) {
          setProgress(Math.round((e.loaded / e.total) * 100))
        }
      })

      xhr.addEventListener("load", () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          resolve()
        } else {
          reject(new Error(`S3 upload failed with status ${xhr.status}`))
        }
      })

      xhr.addEventListener("error", () => reject(new Error("Network error during upload")))

      xhr.open("PUT", url)
      xhr.setRequestHeader("Content-Type", f.type)
      xhr.send(f)
    })
  }

  async function handleUpload() {
    if (!file) return

    setError(null)
    setStatus("uploading")
    setProgress(0)

    try {
      const urlResp = await requestUrl.mutateAsync({
        body: { content_type: file.type, filename: file.name },
      })

      if (!urlResp.upload_url || !urlResp.id) {
        throw new Error("Missing upload URL or ID in response")
      }

      await uploadToS3(urlResp.upload_url, file)

      setStatus("confirming")
      const confirmed = await confirmUpload.mutateAsync({
        path: { id: urlResp.id as string },
      })

      setStatus("complete")
      onUploadComplete(confirmed)
    } catch (err) {
      setStatus("error")
      const msg = err instanceof Error ? err.message : "Upload failed"
      setError(msg)
    }
  }

  const isUploading = status === "uploading" || status === "confirming"

  return (
    <div className={cn("space-y-3", className)}>
      <div
        role="button"
        tabIndex={0}
        onDrop={handleDrop}
        onDragOver={handleDragOver}
        onClick={() => inputRef.current?.click()}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") inputRef.current?.click()
        }}
        className={cn(
          "flex min-h-[120px] cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed p-4 transition-colors",
          "border-muted-foreground/25 hover:border-muted-foreground/50",
          isUploading && "pointer-events-none opacity-50",
        )}
      >
        {preview ? (
          <img
            src={preview}
            alt="Upload preview"
            className="max-h-[200px] rounded-md object-contain"
          />
        ) : (
          <div className="text-center text-sm text-muted-foreground">
            <p>Drop an image here or click to browse</p>
            <p className="mt-1 text-xs">
              {accept.map((t) => t.replace("image/", "")).join(", ")} &middot; max {maxSizeMB} MB
            </p>
          </div>
        )}

        <input
          ref={inputRef}
          type="file"
          accept={accept.join(",")}
          onChange={handleInputChange}
          className="hidden"
          data-testid="file-input"
        />
      </div>

      {isUploading && (
        <div className="space-y-1">
          <div className="h-2 overflow-hidden rounded-full bg-muted">
            <div
              className="h-full bg-primary transition-all"
              style={{ width: `${progress}%` }}
            />
          </div>
          <p className="text-center text-xs text-muted-foreground">
            {status === "confirming" ? "Confirming..." : `Uploading... ${progress}%`}
          </p>
        </div>
      )}

      {status === "complete" && (
        <p className="text-center text-sm text-green-600">Upload complete</p>
      )}

      {error && <p className="text-center text-sm text-red-500">{error}</p>}

      <div className="flex gap-2">
        {file && status !== "complete" && (
          <Button
            type="button"
            onClick={handleUpload}
            disabled={isUploading}
            className="flex-1"
          >
            {isUploading ? "Uploading..." : "Upload"}
          </Button>
        )}

        {(file || error) && (
          <Button type="button" variant="outline" onClick={resetState} disabled={isUploading}>
            Clear
          </Button>
        )}
      </div>
    </div>
  )
}
