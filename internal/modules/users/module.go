package users

import (
	"context"

	"github.com/radni/soapbox/internal/core"
	"github.com/radni/soapbox/internal/modules/users/migrations"
)

func Load(deps core.ModuleDeps) error {
	ctx := context.Background()

	if err := deps.DB.Migrate(ctx, "users", migrations.Files); err != nil {
		return err
	}

	return nil
}
