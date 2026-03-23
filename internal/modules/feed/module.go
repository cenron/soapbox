package feed

import (
	"context"

	"github.com/radni/soapbox/internal/core"
	"github.com/radni/soapbox/internal/core/bus"
	"github.com/radni/soapbox/internal/modules/feed/migrations"
)

func Load(ctx context.Context, deps core.ModuleDeps) error {
	if err := deps.DB.Migrate(ctx, "feed", migrations.Files); err != nil {
		return err
	}

	store := NewStore(deps.DB)
	service := NewService(deps.DB, store, deps.Bus, deps.WSHub, deps.Logger)
	handler := NewHandler(service, deps.Logger)

	if err := RegisterQueries(deps.Bus, service); err != nil {
		return err
	}

	// Subscribe to posts.created — fan out to followers' timelines.
	if err := deps.Bus.Subscribe(postsTopicCreated, func(event any) {
		e, err := bus.Convert[postCreatedEvent](event)
		if err != nil {
			deps.Logger.Error("feed: posts.created: invalid event type", "error", err)
			return
		}
		service.HandlePostCreated(e)
	}); err != nil {
		return err
	}

	// Subscribe to posts.deleted — remove from all timelines.
	if err := deps.Bus.Subscribe(postsTopicDeleted, func(event any) {
		e, err := bus.Convert[postDeletedEvent](event)
		if err != nil {
			deps.Logger.Error("feed: posts.deleted: invalid event type", "error", err)
			return
		}
		service.HandlePostDeleted(e)
	}); err != nil {
		return err
	}

	// Subscribe to users.followed — backfill recent posts.
	if err := deps.Bus.Subscribe(usersTopicFollowed, func(event any) {
		e, err := bus.Convert[userFollowedEvent](event)
		if err != nil {
			deps.Logger.Error("feed: users.followed: invalid event type", "error", err)
			return
		}
		service.HandleUserFollowed(e)
	}); err != nil {
		return err
	}

	// Subscribe to users.unfollowed — remove unfollowed user's posts.
	if err := deps.Bus.Subscribe(usersTopicUnfollowed, func(event any) {
		e, err := bus.Convert[userUnfollowedEvent](event)
		if err != nil {
			deps.Logger.Error("feed: users.unfollowed: invalid event type", "error", err)
			return
		}
		service.HandleUserUnfollowed(e)
	}); err != nil {
		return err
	}

	handler.Routes(deps.Router, deps.AuthRequired)

	return nil
}
