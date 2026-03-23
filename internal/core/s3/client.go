package s3

import (
	"context"
	"time"
)

// Client abstracts S3-compatible object storage operations.
type Client interface {
	// PresignPutObject returns a presigned URL for uploading an object.
	PresignPutObject(ctx context.Context, key, contentType string, expires time.Duration) (string, error)

	// ObjectURL returns the public URL for a stored object.
	ObjectURL(key string) string

	// EnsureBucket creates the bucket if it does not exist and sets a public read policy.
	EnsureBucket(ctx context.Context) error
}
