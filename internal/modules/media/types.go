package media

import (
	"time"

	"github.com/radni/soapbox/internal/core/types"
)

// Upload status constants.
const (
	StatusPending = "pending"
	StatusReady   = "ready"
	StatusFailed  = "failed"
)

// Allowed content types for uploads.
var AllowedContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// PresignTTL is how long a presigned upload URL remains valid.
const PresignTTL = 15 * time.Minute

// Request types.

type UploadURLRequest struct {
	ContentType string `json:"content_type"`
	Filename    string `json:"filename"`
}

// Response types.

type UploadURLResponse struct {
	ID        types.ID `json:"id"`
	UploadURL string   `json:"upload_url"`
	FileKey   string   `json:"file_key"`
}

type UploadResponse struct {
	ID          types.ID  `json:"id"`
	FileKey     string    `json:"file_key"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	Status      string    `json:"status"`
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"created_at"`
}

// Bus query types.

type GetByIDsQuery struct {
	IDs []types.ID
}

// Query name constants.
const (
	QueryGetByIDs = "media.GetByIDs"
)
