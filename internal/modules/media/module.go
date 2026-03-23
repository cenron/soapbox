package media

import (
	"context"

	"github.com/radni/soapbox/internal/core"
	coreS3 "github.com/radni/soapbox/internal/core/s3"
	"github.com/radni/soapbox/internal/modules/media/migrations"
)

func Load(ctx context.Context, deps core.ModuleDeps) error {
	if err := deps.DB.Migrate(ctx, "media", migrations.Files); err != nil {
		return err
	}

	s3Client := coreS3.New(deps.Config.S3)

	if err := s3Client.EnsureBucket(ctx); err != nil {
		return err
	}

	store := NewStore(deps.DB)
	service := NewService(store, s3Client, deps.Config.S3, deps.Logger)
	handler := NewHandler(service, deps.Logger)

	if err := RegisterQueries(deps.Bus, service); err != nil {
		return err
	}

	handler.Routes(deps.Router, deps.AuthRequired)

	return nil
}
