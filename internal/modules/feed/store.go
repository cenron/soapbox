package feed

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/radni/soapbox/internal/core/db"
	"github.com/radni/soapbox/internal/core/types"
)

// TimelineEntry represents a row in feed.timelines.
type TimelineEntry struct {
	UserID    types.ID  `db:"user_id"`
	PostID    types.ID  `db:"post_id"`
	CreatedAt time.Time `db:"created_at"`
}

type Store struct {
	db *db.DB
}

func NewStore(database *db.DB) *Store {
	return &Store{db: database}
}

// AddToTimelines batch-inserts timeline entries. Duplicates are silently ignored.
func (s *Store) AddToTimelines(ctx context.Context, entries []TimelineEntry) error {
	if len(entries) == 0 {
		return nil
	}

	const q = `
		INSERT INTO feed.timelines (user_id, post_id, created_at)
		VALUES (:user_id, :post_id, :created_at)
		ON CONFLICT DO NOTHING`

	_, err := s.db.Conn.NamedExecContext(ctx, q, entries)
	if err != nil {
		return fmt.Errorf("store: add to timelines: %w", err)
	}
	return nil
}

// RemoveFromAllTimelines removes a post from every user's timeline.
func (s *Store) RemoveFromAllTimelines(ctx context.Context, postID types.ID) error {
	const q = `DELETE FROM feed.timelines WHERE post_id = $1`

	_, err := s.db.Conn.ExecContext(ctx, q, postID)
	if err != nil {
		return fmt.Errorf("store: remove from all timelines: %w", err)
	}
	return nil
}

// RemovePostsFromTimeline removes specific posts from a single user's timeline.
func (s *Store) RemovePostsFromTimeline(ctx context.Context, userID types.ID, postIDs []types.ID) error {
	if len(postIDs) == 0 {
		return nil
	}

	query, args, err := sqlx.In(
		`DELETE FROM feed.timelines WHERE user_id = ? AND post_id IN (?)`,
		userID, postIDs,
	)
	if err != nil {
		return fmt.Errorf("store: remove posts from timeline: build query: %w", err)
	}

	query = s.db.Conn.Rebind(query)

	_, err = s.db.Conn.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("store: remove posts from timeline: %w", err)
	}
	return nil
}

// GetTimeline returns paginated timeline entries for a user, newest first.
func (s *Store) GetTimeline(ctx context.Context, userID types.ID, params types.CursorParams) ([]TimelineEntry, bool, error) {
	if err := validateTimestampCursor(params.Cursor); err != nil {
		return nil, false, err
	}

	const q = `
		SELECT user_id, post_id, created_at
		FROM feed.timelines
		WHERE user_id = $1
		  AND ($2 = '' OR created_at < $2::timestamptz)
		ORDER BY created_at DESC
		LIMIT $3`

	limit := params.Limit + 1
	var entries []TimelineEntry

	if err := s.db.Conn.SelectContext(ctx, &entries, q, userID, params.Cursor, limit); err != nil {
		return nil, false, fmt.Errorf("store: get timeline: %w", err)
	}

	hasMore := len(entries) > params.Limit
	if hasMore {
		entries = entries[:params.Limit]
	}

	return entries, hasMore, nil
}

// validateTimestampCursor checks that a cursor is empty or a valid RFC3339 timestamp.
func validateTimestampCursor(cursor string) error {
	if cursor == "" {
		return nil
	}

	_, err := time.Parse(time.RFC3339Nano, cursor)
	if err != nil {
		return types.NewValidation("invalid cursor format")
	}
	return nil
}
