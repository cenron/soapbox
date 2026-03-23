package feed

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

const backfillLimit = 50

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

// GetTimeline returns a paginated timeline for the given user.
func (s *Service) GetTimeline(ctx context.Context, userID types.ID, viewerID *types.ID, params types.CursorParams) (*types.CursorPage[postResponse], error) {
	entries, hasMore, err := s.store.GetTimeline(ctx, userID, params)
	if err != nil {
		return nil, fmt.Errorf("service: get timeline: %w", err)
	}

	if len(entries) == 0 {
		return &types.CursorPage[postResponse]{
			Items:   []postResponse{},
			HasMore: false,
		}, nil
	}

	posts, err := s.fetchPosts(entries, viewerID)
	if err != nil {
		return nil, fmt.Errorf("service: get timeline: %w", err)
	}

	var nextCursor string
	if hasMore && len(entries) > 0 {
		nextCursor = entries[len(entries)-1].CreatedAt.UTC().Format(time.RFC3339Nano)
	}

	return &types.CursorPage[postResponse]{
		Items:      posts,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// HandlePostCreated fans out a new post to all followers' timelines.
func (s *Service) HandlePostCreated(event postCreatedEvent) {
	ctx := context.Background()

	followerIDs, err := s.getFollowerIDs(event.AuthorID)
	if err != nil {
		s.logger.Error("feed: post created: get followers",
			"author_id", event.AuthorID,
			"error", err,
		)
		return
	}

	// Include the author's own timeline.
	recipients := make([]types.ID, 0, len(followerIDs)+1)
	recipients = append(recipients, followerIDs...)
	recipients = append(recipients, event.AuthorID)

	entries := make([]TimelineEntry, len(recipients))
	for i, uid := range recipients {
		entries[i] = TimelineEntry{
			UserID:    uid,
			PostID:    event.PostID,
			CreatedAt: event.CreatedAt,
		}
	}

	if err := s.store.AddToTimelines(ctx, entries); err != nil {
		s.logger.Error("feed: post created: add to timelines",
			"post_id", event.PostID,
			"error", err,
		)
		return
	}

	// Push WebSocket notification to connected followers (not the author).
	for _, uid := range followerIDs {
		s.wsHub.Send(uid, ws.Message{
			Type: "new_posts",
			Data: map[string]int{"count": 1},
		})
	}
}

// HandlePostDeleted removes a deleted post from all timelines.
func (s *Service) HandlePostDeleted(event postDeletedEvent) {
	ctx := context.Background()

	if err := s.store.RemoveFromAllTimelines(ctx, event.PostID); err != nil {
		s.logger.Error("feed: post deleted: remove from timelines",
			"post_id", event.PostID,
			"error", err,
		)
	}
}

// HandleUserFollowed backfills recent posts from the followed user into the follower's timeline.
func (s *Service) HandleUserFollowed(event userFollowedEvent) {
	ctx := context.Background()

	posts, err := s.getRecentPostsByAuthor(event.FollowingID)
	if err != nil {
		s.logger.Error("feed: user followed: get recent posts",
			"follower_id", event.FollowerID,
			"following_id", event.FollowingID,
			"error", err,
		)
		return
	}

	if len(posts) == 0 {
		return
	}

	entries := make([]TimelineEntry, len(posts))
	for i, p := range posts {
		entries[i] = TimelineEntry{
			UserID:    event.FollowerID,
			PostID:    p.ID,
			CreatedAt: p.CreatedAt,
		}
	}

	if err := s.store.AddToTimelines(ctx, entries); err != nil {
		s.logger.Error("feed: user followed: backfill timelines",
			"follower_id", event.FollowerID,
			"following_id", event.FollowingID,
			"error", err,
		)
	}
}

// HandleUserUnfollowed removes the unfollowed user's posts from the unfollower's timeline.
func (s *Service) HandleUserUnfollowed(event userUnfollowedEvent) {
	ctx := context.Background()

	posts, err := s.getRecentPostsByAuthor(event.FollowingID)
	if err != nil {
		s.logger.Error("feed: user unfollowed: get posts",
			"follower_id", event.FollowerID,
			"following_id", event.FollowingID,
			"error", err,
		)
		return
	}

	if len(posts) == 0 {
		return
	}

	postIDs := make([]types.ID, len(posts))
	for i, p := range posts {
		postIDs[i] = p.ID
	}

	if err := s.store.RemovePostsFromTimeline(ctx, event.FollowerID, postIDs); err != nil {
		s.logger.Error("feed: user unfollowed: remove posts",
			"follower_id", event.FollowerID,
			"following_id", event.FollowingID,
			"error", err,
		)
	}
}

// --- helpers ---

// fetchPosts queries the posts module for full post data, preserving timeline order.
func (s *Service) fetchPosts(entries []TimelineEntry, viewerID *types.ID) ([]postResponse, error) {
	ids := make([]types.ID, len(entries))
	for i := range entries {
		ids[i] = entries[i].PostID
	}

	result, err := s.bus.Query(postsQueryGetByIDs, postsGetByIDsQuery{
		PostIDs:  ids,
		ViewerID: viewerID,
	})
	if err != nil {
		return nil, fmt.Errorf("fetch posts: %w", err)
	}

	posts, err := bus.Convert[[]postResponse](result)
	if err != nil {
		return nil, fmt.Errorf("fetch posts: convert response: %w", err)
	}

	// Re-order to match timeline entry order (newest first).
	postMap := make(map[types.ID]postResponse, len(posts))
	for _, p := range posts {
		postMap[p.ID] = p
	}

	ordered := make([]postResponse, 0, len(entries))
	for _, entry := range entries {
		if p, ok := postMap[entry.PostID]; ok {
			ordered = append(ordered, p)
		}
	}

	return ordered, nil
}

// getFollowerIDs queries the users module for IDs of users who follow the given user.
func (s *Service) getFollowerIDs(userID types.ID) ([]types.ID, error) {
	result, err := s.bus.Query(usersQueryGetFollowerIDs, usersGetFollowerIDsQuery{
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	ids, err := bus.Convert[[]types.ID](result)
	if err != nil {
		return nil, fmt.Errorf("get follower ids: convert response: %w", err)
	}

	return ids, nil
}

// getRecentPostsByAuthor queries the posts module for recent posts by the given author.
func (s *Service) getRecentPostsByAuthor(authorID types.ID) ([]postResponse, error) {
	result, err := s.bus.Query(postsQueryGetByAuthor, postsGetByAuthorQuery{
		AuthorID: authorID,
		Limit:    backfillLimit,
	})
	if err != nil {
		return nil, err
	}

	page, err := bus.Convert[postsCursorPage](result)
	if err != nil {
		return nil, fmt.Errorf("get recent posts: convert response: %w", err)
	}

	return page.Items, nil
}
