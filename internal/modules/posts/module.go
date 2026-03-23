package posts

import (
	"context"

	"github.com/radni/soapbox/internal/core"
	"github.com/radni/soapbox/internal/modules/posts/migrations"
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

	if err := deps.Bus.Subscribe(usersTopicProfileUpdated, func(event any) {
		e, ok := event.(userProfileUpdatedEvent)
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
