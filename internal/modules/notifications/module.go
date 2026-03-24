package notifications

import (
	"context"

	"github.com/radni/soapbox/internal/core"
	"github.com/radni/soapbox/internal/core/bus"
	"github.com/radni/soapbox/internal/modules/notifications/migrations"
)

func Load(ctx context.Context, deps core.ModuleDeps) error {
	if err := deps.DB.Migrate(ctx, "notifications", migrations.Files); err != nil {
		return err
	}

	store := NewStore(deps.DB)
	service := NewService(deps.DB, store, deps.Bus, deps.WSHub, deps.Logger)
	handler := NewHandler(service, deps.Logger)

	if err := RegisterQueries(deps.Bus, service); err != nil {
		return err
	}

	// Subscribe to posts.liked — notify post author.
	if err := deps.Bus.Subscribe(postsTopicLiked, func(event any) {
		e, err := bus.Convert[postLikedEvent](event)
		if err != nil {
			deps.Logger.Error("notifications: posts.liked: invalid event type", "error", err)
			return
		}
		service.HandlePostLiked(e)
	}); err != nil {
		return err
	}

	// Subscribe to posts.reposted — notify post author.
	if err := deps.Bus.Subscribe(postsTopicReposted, func(event any) {
		e, err := bus.Convert[postRepostedEvent](event)
		if err != nil {
			deps.Logger.Error("notifications: posts.reposted: invalid event type", "error", err)
			return
		}
		service.HandlePostReposted(e)
	}); err != nil {
		return err
	}

	// Subscribe to posts.created — notify parent post author on replies.
	if err := deps.Bus.Subscribe(postsTopicCreated, func(event any) {
		e, err := bus.Convert[postCreatedEvent](event)
		if err != nil {
			deps.Logger.Error("notifications: posts.created: invalid event type", "error", err)
			return
		}
		service.HandlePostCreated(e)
	}); err != nil {
		return err
	}

	// Subscribe to users.followed — notify followed user.
	if err := deps.Bus.Subscribe(usersTopicFollowed, func(event any) {
		e, err := bus.Convert[userFollowedEvent](event)
		if err != nil {
			deps.Logger.Error("notifications: users.followed: invalid event type", "error", err)
			return
		}
		service.HandleUserFollowed(e)
	}); err != nil {
		return err
	}

	handler.Routes(deps.Router, deps.AuthRequired)

	return nil
}
