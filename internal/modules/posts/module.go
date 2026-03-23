package posts

import (
	"context"

	"github.com/radni/soapbox/internal/core"
	"github.com/radni/soapbox/internal/modules/posts/migrations"
	"github.com/radni/soapbox/internal/modules/users"
)

func Load(ctx context.Context, deps core.ModuleDeps) error {
	if err := deps.DB.Migrate(ctx, "posts", migrations.Files); err != nil {
		return err
	}

	store := NewStore(deps.DB)
	service := NewService(deps.DB, store, deps.Bus, deps.Logger)
	handler := NewHandler(service, deps.Logger)

	if err := RegisterQueries(deps.Bus, service); err != nil {
		return err
	}

	if err := deps.Bus.Subscribe(users.TopicProfileUpdated, func(event any) {
		e, ok := event.(users.ProfileUpdatedEvent)
		if !ok {
			deps.Logger.Error("posts: profile_updated: invalid event type")
			return
		}
		service.HandleProfileUpdated(e)
	}); err != nil {
		return err
	}

	handler.Routes(deps.Router, deps.AuthRequired, deps.AuthOptional)

	return nil
}
