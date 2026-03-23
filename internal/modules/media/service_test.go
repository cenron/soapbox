package media

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/radni/soapbox/internal/core/config"
	"github.com/radni/soapbox/internal/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockS3 implements s3.Client for testing.
type mockS3 struct {
	presignURL string
	presignErr error
}

func (m *mockS3) PresignPutObject(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return m.presignURL, m.presignErr
}

func (m *mockS3) ObjectURL(key string) string {
	return "http://localhost:9000/soapbox/" + key
}

func (m *mockS3) EnsureBucket(_ context.Context) error {
	return nil
}

func TestValidateContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantExt     string
		wantErr     bool
	}{
		{"jpeg", "image/jpeg", ".jpg", false},
		{"png", "image/png", ".png", false},
		{"gif", "image/gif", ".gif", false},
		{"webp", "image/webp", ".webp", false},
		{"pdf rejected", "application/pdf", "", true},
		{"text rejected", "text/plain", "", true},
		{"empty rejected", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, err := validateContentType(tt.contentType)
			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := types.IsAppError(err)
				require.True(t, ok)
				assert.Equal(t, 422, appErr.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantExt, ext)
			}
		})
	}
}

func TestExtensionFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"photo.jpg", ".jpg"},
		{"image.PNG", ".png"},
		{"no-extension", ".bin"},
		{"", ".bin"},
		{"file.tar.gz", ".gz"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := extensionFromFilename(tt.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUploadToResponse(t *testing.T) {
	now := time.Now().UTC()
	id, err := types.NewID()
	require.NoError(t, err)

	upload := &Upload{
		ID:          id,
		FileKey:     "uploads/abc/def.jpg",
		ContentType: "image/jpeg",
		Size:        1024,
		Status:      StatusReady,
		CreatedAt:   now,
	}

	resp := uploadToResponse(upload, "http://localhost:9000/soapbox/uploads/abc/def.jpg")

	assert.Equal(t, id, resp.ID)
	assert.Equal(t, "uploads/abc/def.jpg", resp.FileKey)
	assert.Equal(t, "image/jpeg", resp.ContentType)
	assert.Equal(t, int64(1024), resp.Size)
	assert.Equal(t, StatusReady, resp.Status)
	assert.Equal(t, "http://localhost:9000/soapbox/uploads/abc/def.jpg", resp.URL)
	assert.Equal(t, now, resp.CreatedAt)
}

func TestRequestUpload_InvalidContentType(t *testing.T) {
	s3Client := &mockS3{presignURL: "http://example.com/presigned"}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := config.S3Config{Endpoint: "http://localhost:9000", Bucket: "soapbox"}

	svc := NewService(nil, s3Client, cfg, logger)

	userID, err := types.NewID()
	require.NoError(t, err)

	_, err = svc.RequestUpload(context.Background(), userID, UploadURLRequest{
		ContentType: "application/pdf",
		Filename:    "doc.pdf",
	})

	require.Error(t, err)
	appErr, ok := types.IsAppError(err)
	require.True(t, ok)
	assert.Equal(t, 422, appErr.Code)
}

func TestRequestUpload_S3Error(t *testing.T) {
	s3Client := &mockS3{presignErr: fmt.Errorf("s3 unavailable")}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := config.S3Config{Endpoint: "http://localhost:9000", Bucket: "soapbox"}

	svc := NewService(&Store{}, s3Client, cfg, logger)

	userID, err := types.NewID()
	require.NoError(t, err)

	_, err = svc.RequestUpload(context.Background(), userID, UploadURLRequest{
		ContentType: "image/jpeg",
		Filename:    "photo.jpg",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "presign")
}
