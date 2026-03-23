package media

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/radni/soapbox/internal/core/config"
	coreS3 "github.com/radni/soapbox/internal/core/s3"
	"github.com/radni/soapbox/internal/core/types"
)

type Service struct {
	store    *Store
	s3       coreS3.Client
	s3Config config.S3Config
	logger   *slog.Logger
}

func NewService(store *Store, s3Client coreS3.Client, s3Config config.S3Config, logger *slog.Logger) *Service {
	return &Service{
		store:    store,
		s3:       s3Client,
		s3Config: s3Config,
		logger:   logger,
	}
}

func (s *Service) RequestUpload(ctx context.Context, userID types.ID, req UploadURLRequest) (*UploadURLResponse, error) {
	ext, err := validateContentType(req.ContentType)
	if err != nil {
		return nil, err
	}

	uploadID, err := types.NewID()
	if err != nil {
		return nil, fmt.Errorf("service: request upload: generate id: %w", err)
	}

	fileKeyID, err := types.NewID()
	if err != nil {
		return nil, fmt.Errorf("service: request upload: generate file key: %w", err)
	}

	fileKey := fmt.Sprintf("uploads/%s/%s%s", userID, fileKeyID, ext)

	presignedURL, err := s.s3.PresignPutObject(ctx, fileKey, req.ContentType, PresignTTL)
	if err != nil {
		return nil, fmt.Errorf("service: request upload: presign: %w", err)
	}

	upload := &Upload{
		ID:          uploadID,
		UserID:      userID,
		FileKey:     fileKey,
		ContentType: req.ContentType,
		Status:      StatusPending,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.store.Create(ctx, upload); err != nil {
		return nil, fmt.Errorf("service: request upload: create: %w", err)
	}

	return &UploadURLResponse{
		ID:        uploadID,
		UploadURL: presignedURL,
		FileKey:   fileKey,
	}, nil
}

func (s *Service) ConfirmUpload(ctx context.Context, userID, uploadID types.ID, size int64) (*UploadResponse, error) {
	upload, err := s.store.GetByID(ctx, uploadID)
	if err != nil {
		return nil, err
	}

	if upload.UserID != userID {
		return nil, types.ErrForbidden()
	}

	if upload.Status != StatusPending {
		return nil, types.NewConflict("upload already confirmed")
	}

	if err := s.store.UpdateStatus(ctx, uploadID, StatusReady, size); err != nil {
		return nil, fmt.Errorf("service: confirm upload: %w", err)
	}

	return &UploadResponse{
		ID:          upload.ID,
		FileKey:     upload.FileKey,
		ContentType: upload.ContentType,
		Size:        size,
		Status:      StatusReady,
		URL:         s.s3.ObjectURL(upload.FileKey),
		CreatedAt:   upload.CreatedAt,
	}, nil
}

func (s *Service) GetByIDs(ctx context.Context, ids []types.ID) ([]UploadResponse, error) {
	uploads, err := s.store.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	results := make([]UploadResponse, len(uploads))
	for i, u := range uploads {
		results[i] = uploadToResponse(&u, s.s3.ObjectURL(u.FileKey))
	}

	return results, nil
}

func uploadToResponse(u *Upload, url string) UploadResponse {
	return UploadResponse{
		ID:          u.ID,
		FileKey:     u.FileKey,
		ContentType: u.ContentType,
		Size:        u.Size,
		Status:      u.Status,
		URL:         url,
		CreatedAt:   u.CreatedAt,
	}
}

func validateContentType(ct string) (string, error) {
	ext, ok := AllowedContentTypes[ct]
	if !ok {
		return "", types.NewValidation("unsupported content type: " + ct + ". Allowed: image/jpeg, image/png, image/gif, image/webp")
	}

	return ext, nil
}

