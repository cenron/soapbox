package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/radni/soapbox/internal/core/db"
	"github.com/radni/soapbox/internal/core/types"
)

// Notification represents a row in notifications.notifications.
type Notification struct {
	ID        types.ID  `db:"id"`
	UserID    types.ID  `db:"user_id"`
	Type      string    `db:"type"`
	ActorID   types.ID  `db:"actor_id"`
	PostID    *types.ID `db:"post_id"`
	Read      bool      `db:"read"`
	CreatedAt time.Time `db:"created_at"`
}

type Store struct {
	db *db.DB
}

func NewStore(database *db.DB) *Store {
	return &Store{db: database}
}

// Insert creates a new notification.
func (s *Store) Insert(ctx context.Context, n Notification) error {
	const q = `
		INSERT INTO notifications.notifications (id, user_id, type, actor_id, post_id, read, created_at)
		VALUES (:id, :user_id, :type, :actor_id, :post_id, :read, :created_at)`

	_, err := s.db.Conn.NamedExecContext(ctx, q, n)
	if err != nil {
		return fmt.Errorf("store: insert notification: %w", err)
	}
	return nil
}

// List returns paginated notifications for a user, newest first.
func (s *Store) List(ctx context.Context, userID types.ID, params types.CursorParams) ([]Notification, bool, error) {
	if err := validateTimestampCursor(params.Cursor); err != nil {
		return nil, false, err
	}

	const q = `
		SELECT id, user_id, type, actor_id, post_id, read, created_at
		FROM notifications.notifications
		WHERE user_id = $1
		  AND ($2 = '' OR created_at < $2::timestamptz)
		ORDER BY created_at DESC
		LIMIT $3`

	limit := params.Limit + 1
	var rows []Notification

	if err := s.db.Conn.SelectContext(ctx, &rows, q, userID, params.Cursor, limit); err != nil {
		return nil, false, fmt.Errorf("store: list notifications: %w", err)
	}

	hasMore := len(rows) > params.Limit
	if hasMore {
		rows = rows[:params.Limit]
	}

	return rows, hasMore, nil
}

// MarkRead marks a single notification as read. Returns false if not found or not owned.
func (s *Store) MarkRead(ctx context.Context, id, userID types.ID) (bool, error) {
	const q = `
		UPDATE notifications.notifications
		SET read = TRUE
		WHERE id = $1 AND user_id = $2`

	result, err := s.db.Conn.ExecContext(ctx, q, id, userID)
	if err != nil {
		return false, fmt.Errorf("store: mark read: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("store: mark read: rows affected: %w", err)
	}

	return rows > 0, nil
}

// MarkAllRead marks all of a user's notifications as read.
func (s *Store) MarkAllRead(ctx context.Context, userID types.ID) error {
	const q = `
		UPDATE notifications.notifications
		SET read = TRUE
		WHERE user_id = $1 AND read = FALSE`

	_, err := s.db.Conn.ExecContext(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("store: mark all read: %w", err)
	}
	return nil
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
