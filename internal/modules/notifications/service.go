package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/radni/soapbox/internal/core/bus"
	"github.com/radni/soapbox/internal/core/db"
	"github.com/radni/soapbox/internal/core/types"
	"github.com/radni/soapbox/internal/core/ws"
)

type Service struct {
	db     *db.DB
	store  *Store
	bus    bus.Bus
	wsHub  *ws.Hub
	logger *slog.Logger
}

func NewService(database *db.DB, store *Store, b bus.Bus, hub *ws.Hub, logger *slog.Logger) *Service {
	return &Service{
		db:     database,
		store:  store,
		bus:    b,
		wsHub:  hub,
		logger: logger,
	}
}

// ListNotifications returns paginated notifications with actor info enriched via bus.
func (s *Service) ListNotifications(ctx context.Context, userID types.ID, params types.CursorParams) (*types.CursorPage[NotificationResponse], error) {
	rows, hasMore, err := s.store.List(ctx, userID, params)
	if err != nil {
		return nil, fmt.Errorf("service: list notifications: %w", err)
	}

	if len(rows) == 0 {
		return &types.CursorPage[NotificationResponse]{
			Items:   []NotificationResponse{},
			HasMore: false,
		}, nil
	}

	responses, err := s.enrichWithActors(rows)
	if err != nil {
		return nil, fmt.Errorf("service: list notifications: %w", err)
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].CreatedAt.UTC().Format(time.RFC3339Nano)
	}

	return &types.CursorPage[NotificationResponse]{
		Items:      responses,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// MarkRead marks a single notification as read.
func (s *Service) MarkRead(ctx context.Context, id, userID types.ID) error {
	found, err := s.store.MarkRead(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("service: mark read: %w", err)
	}

	if !found {
		return types.NewNotFound("notification not found")
	}

	return nil
}

// MarkAllRead marks all of a user's notifications as read.
func (s *Service) MarkAllRead(ctx context.Context, userID types.ID) error {
	if err := s.store.MarkAllRead(ctx, userID); err != nil {
		return fmt.Errorf("service: mark all read: %w", err)
	}
	return nil
}

// HandlePostLiked creates a notification when someone likes a post.
func (s *Service) HandlePostLiked(event postLikedEvent) {
	if event.UserID == event.AuthorID {
		return
	}

	s.createAndPush(event.AuthorID, TypeLike, event.UserID, &event.PostID)
}

// HandlePostReposted creates a notification when someone reposts a post.
func (s *Service) HandlePostReposted(event postRepostedEvent) {
	if event.UserID == event.AuthorID {
		return
	}

	s.createAndPush(event.AuthorID, TypeRepost, event.UserID, &event.PostID)
}

// HandlePostCreated creates a reply notification when someone replies to a post.
func (s *Service) HandlePostCreated(event postCreatedEvent) {
	if event.ParentID == nil {
		return
	}

	parentAuthorID, err := s.getPostAuthor(*event.ParentID)
	if err != nil {
		s.logger.Error("notifications: post created: get parent author",
			"parent_id", event.ParentID,
			"error", err,
		)
		return
	}

	if event.AuthorID == parentAuthorID {
		return
	}

	s.createAndPush(parentAuthorID, TypeReply, event.AuthorID, &event.PostID)
}

// HandleUserFollowed creates a notification when someone follows a user.
func (s *Service) HandleUserFollowed(event userFollowedEvent) {
	s.createAndPush(event.FollowingID, TypeFollow, event.FollowerID, nil)
}

// --- helpers ---

// createAndPush inserts a notification, pushes a WebSocket message, and publishes a bus event.
func (s *Service) createAndPush(recipientID types.ID, notifType string, actorID types.ID, postID *types.ID) {
	ctx := context.Background()

	id, err := types.NewID()
	if err != nil {
		s.logger.Error("notifications: generate id", "error", err)
		return
	}

	n := Notification{
		ID:        id,
		UserID:    recipientID,
		Type:      notifType,
		ActorID:   actorID,
		PostID:    postID,
		Read:      false,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.store.Insert(ctx, n); err != nil {
		s.logger.Error("notifications: insert",
			"type", notifType,
			"recipient", recipientID,
			"actor", actorID,
			"error", err,
		)
		return
	}

	s.wsHub.Send(recipientID, ws.Message{
		Type: "new_notification",
		Data: map[string]any{
			"id":       n.ID,
			"type":     n.Type,
			"actor_id": n.ActorID,
			"post_id":  n.PostID,
		},
	})

	_ = s.bus.Publish(TopicNew, NewNotificationEvent{
		ID:      n.ID,
		UserID:  n.UserID,
		Type:    n.Type,
		ActorID: n.ActorID,
		PostID:  n.PostID,
	})
}

// enrichWithActors fetches actor profiles and maps them onto notification responses.
func (s *Service) enrichWithActors(rows []Notification) ([]NotificationResponse, error) {
	actorIDs := uniqueIDs(rows)

	result, err := s.bus.Query(usersQueryGetProfiles, usersGetProfilesQuery{
		UserIDs: actorIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("enrich actors: %w", err)
	}

	profiles, err := bus.Convert[[]userProfileResponse](result)
	if err != nil {
		return nil, fmt.Errorf("enrich actors: convert response: %w", err)
	}

	profileMap := make(map[types.ID]userProfileResponse, len(profiles))
	for _, p := range profiles {
		profileMap[p.ID] = p
	}

	responses := make([]NotificationResponse, len(rows))
	for i, n := range rows {
		actor := profileMap[n.ActorID]
		responses[i] = NotificationResponse{
			ID:               n.ID,
			Type:             n.Type,
			ActorID:          n.ActorID,
			ActorUsername:    actor.Username,
			ActorDisplayName: actor.DisplayName,
			ActorAvatarURL:   actor.AvatarURL,
			PostID:           n.PostID,
			Read:             n.Read,
			CreatedAt:        n.CreatedAt,
		}
	}

	return responses, nil
}

// getPostAuthor looks up the author of a post via the posts bus query.
func (s *Service) getPostAuthor(postID types.ID) (types.ID, error) {
	result, err := s.bus.Query(postsQueryGetByIDs, postsGetByIDsQuery{
		PostIDs: []types.ID{postID},
	})
	if err != nil {
		return types.ZeroID, fmt.Errorf("get post author: %w", err)
	}

	posts, err := bus.Convert[[]postResponse](result)
	if err != nil {
		return types.ZeroID, fmt.Errorf("get post author: convert response: %w", err)
	}

	if len(posts) == 0 {
		return types.ZeroID, fmt.Errorf("get post author: post %s not found", postID)
	}

	return posts[0].AuthorID, nil
}

// uniqueIDs extracts deduplicated actor IDs from notifications.
func uniqueIDs(rows []Notification) []types.ID {
	seen := make(map[types.ID]struct{}, len(rows))
	ids := make([]types.ID, 0, len(rows))

	for _, n := range rows {
		if _, ok := seen[n.ActorID]; ok {
			continue
		}
		seen[n.ActorID] = struct{}{}
		ids = append(ids, n.ActorID)
	}

	return ids
}
