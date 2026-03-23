package users

import (
	"context"

	"github.com/radni/soapbox/internal/core"
	"github.com/radni/soapbox/internal/modules/users/migrations"
)

func Load(ctx context.Context, deps core.ModuleDeps) error {
	if err := deps.DB.Migrate(ctx, "users", migrations.Files); err != nil {
		return err
	}

	store := NewStore(deps.DB)
	tokens := NewTokenService(deps.Config.JWT)
	service := NewService(deps.DB, store, tokens, deps.Bus, deps.Logger, deps.Config.JWT, deps.Config.Server.IsProd())
	handler := NewHandler(service, deps.Logger)

	if err := RegisterQueries(deps.Bus, store); err != nil {
		return err
	}

	handler.Routes(deps.Router, deps.AuthRequired, deps.AuthOptional)

	return nil
}
