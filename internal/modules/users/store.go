package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/radni/soapbox/internal/core/db"
	"github.com/radni/soapbox/internal/core/types"
)

// Model structs

type Profile struct {
	ID          types.ID  `db:"id"`
	Username    string    `db:"username"`
	DisplayName string    `db:"display_name"`
	Bio         string    `db:"bio"`
	AvatarURL   string    `db:"avatar_url"`
	Verified    bool      `db:"verified"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type Credential struct {
	ID           types.ID  `db:"id"`
	UserID       types.ID  `db:"user_id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

type Session struct {
	ID               types.ID  `db:"id"`
	UserID           types.ID  `db:"user_id"`
	RefreshTokenHash string    `db:"refresh_token_hash"`
	ExpiresAt        time.Time `db:"expires_at"`
	CreatedAt        time.Time `db:"created_at"`
}

type Role struct {
	ID        types.ID  `db:"id"`
	UserID    types.ID  `db:"user_id"`
	Role      string    `db:"role"`
	CreatedAt time.Time `db:"created_at"`
}

// Store

type Store struct {
	db *db.DB
}

func NewStore(database *db.DB) *Store {
	return &Store{db: database}
}

// isUniqueViolation checks whether err is a Postgres unique constraint violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// Auth operations

func (s *Store) CreateProfile(ctx context.Context, tx *sqlx.Tx, p *Profile) error {
	const q = `
		INSERT INTO users.profiles (id, username, display_name, bio, avatar_url, verified, created_at, updated_at)
		VALUES (:id, :username, :display_name, :bio, :avatar_url, :verified, :created_at, :updated_at)`

	_, err := tx.NamedExecContext(ctx, q, p)
	if err != nil {
		if isUniqueViolation(err) {
			return types.NewConflict("username already taken")
		}
		return fmt.Errorf("store: create profile: %w", err)
	}
	return nil
}

func (s *Store) CreateCredential(ctx context.Context, tx *sqlx.Tx, c *Credential) error {
	const q = `
		INSERT INTO users.credentials (id, user_id, email, password_hash, created_at)
		VALUES (:id, :user_id, :email, :password_hash, :created_at)`

	_, err := tx.NamedExecContext(ctx, q, c)
	if err != nil {
		if isUniqueViolation(err) {
			return types.NewConflict("email already registered")
		}
		return fmt.Errorf("store: create credential: %w", err)
	}
	return nil
}

func (s *Store) GetCredentialByEmail(ctx context.Context, email string) (*Credential, error) {
	const q = `SELECT id, user_id, email, password_hash, created_at FROM users.credentials WHERE lower(email) = lower($1)`

	var c Credential
	if err := s.db.Conn.GetContext(ctx, &c, q, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.NewNotFound("user")
		}
		return nil, fmt.Errorf("store: get credential by email: %w", err)
	}
	return &c, nil
}

func (s *Store) CreateSession(ctx context.Context, sess *Session) error {
	const q = `
		INSERT INTO users.sessions (id, user_id, refresh_token_hash, expires_at, created_at)
		VALUES (:id, :user_id, :refresh_token_hash, :expires_at, :created_at)`

	_, err := s.db.Conn.NamedExecContext(ctx, q, sess)
	if err != nil {
		return fmt.Errorf("store: create session: %w", err)
	}
	return nil
}

func (s *Store) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	const q = `
		SELECT id, user_id, refresh_token_hash, expires_at, created_at
		FROM users.sessions
		WHERE refresh_token_hash = $1
		  AND expires_at > now()`

	var sess Session
	if err := s.db.Conn.GetContext(ctx, &sess, q, tokenHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.NewNotFound("session")
		}
		return nil, fmt.Errorf("store: get session by token hash: %w", err)
	}
	return &sess, nil
}

func (s *Store) DeleteSession(ctx context.Context, id types.ID) error {
	const q = `DELETE FROM users.sessions WHERE id = $1`

	if _, err := s.db.Conn.ExecContext(ctx, q, id); err != nil {
		return fmt.Errorf("store: delete session: %w", err)
	}
	return nil
}

// RotateSession atomically deletes the old session and creates a new one.
func (s *Store) RotateSession(ctx context.Context, oldID types.ID, newSession *Session) error {
	return s.db.WithTx(ctx, func(tx *sqlx.Tx) error {
		if _, err := tx.ExecContext(ctx, "DELETE FROM users.sessions WHERE id = $1", oldID); err != nil {
			return fmt.Errorf("store: rotate session: delete old: %w", err)
		}

		const q = `INSERT INTO users.sessions (id, user_id, refresh_token_hash, expires_at, created_at)
		           VALUES (:id, :user_id, :refresh_token_hash, :expires_at, :created_at)`
		if _, err := tx.NamedExecContext(ctx, q, newSession); err != nil {
			return fmt.Errorf("store: rotate session: create new: %w", err)
		}

		return nil
	})
}

func (s *Store) DeleteSessionsByUserID(ctx context.Context, userID types.ID) error {
	const q = `DELETE FROM users.sessions WHERE user_id = $1`

	if _, err := s.db.Conn.ExecContext(ctx, q, userID); err != nil {
		return fmt.Errorf("store: delete sessions by user id: %w", err)
	}
	return nil
}

func (s *Store) GetRoleByUserID(ctx context.Context, userID types.ID) (string, error) {
	const q = `SELECT role FROM users.roles WHERE user_id = $1`

	var role string
	if err := s.db.Conn.GetContext(ctx, &role, q, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RoleUser, nil
		}
		return "", fmt.Errorf("store: get role by user id: %w", err)
	}
	return role, nil
}

func (s *Store) CreateRole(ctx context.Context, tx *sqlx.Tx, r *Role) error {
	const q = `
		INSERT INTO users.roles (id, user_id, role, created_at)
		VALUES (:id, :user_id, :role, :created_at)`

	_, err := tx.NamedExecContext(ctx, q, r)
	if err != nil {
		return fmt.Errorf("store: create role: %w", err)
	}
	return nil
}

func (s *Store) UpdateRole(ctx context.Context, userID types.ID, role string) error {
	const q = `UPDATE users.roles SET role = $1 WHERE user_id = $2`

	if _, err := s.db.Conn.ExecContext(ctx, q, role, userID); err != nil {
		return fmt.Errorf("store: update role: %w", err)
	}
	return nil
}

func (s *Store) DeleteRole(ctx context.Context, userID types.ID) error {
	const q = `DELETE FROM users.roles WHERE user_id = $1`

	if _, err := s.db.Conn.ExecContext(ctx, q, userID); err != nil {
		return fmt.Errorf("store: delete role: %w", err)
	}
	return nil
}

// Profile operations

func (s *Store) GetProfileByID(ctx context.Context, id types.ID) (*Profile, error) {
	const q = `SELECT id, username, display_name, bio, avatar_url, verified, created_at, updated_at FROM users.profiles WHERE id = $1`

	var p Profile
	if err := s.db.Conn.GetContext(ctx, &p, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.NewNotFound("user")
		}
		return nil, fmt.Errorf("store: get profile by id: %w", err)
	}
	return &p, nil
}

func (s *Store) GetProfileByUsername(ctx context.Context, username string) (*Profile, error) {
	const q = `SELECT id, username, display_name, bio, avatar_url, verified, created_at, updated_at FROM users.profiles WHERE lower(username) = lower($1)`

	var p Profile
	if err := s.db.Conn.GetContext(ctx, &p, q, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.NewNotFound("user")
		}
		return nil, fmt.Errorf("store: get profile by username: %w", err)
	}
	return &p, nil
}

func (s *Store) GetProfilesByIDs(ctx context.Context, ids []types.ID) ([]Profile, error) {
	if len(ids) == 0 {
		return []Profile{}, nil
	}

	query, args, err := sqlx.In(
		`SELECT id, username, display_name, bio, avatar_url, verified, created_at, updated_at FROM users.profiles WHERE id IN (?)`,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("store: get profiles by ids: build query: %w", err)
	}

	query = s.db.Conn.Rebind(query)

	var profiles []Profile
	if err := s.db.Conn.SelectContext(ctx, &profiles, query, args...); err != nil {
		return nil, fmt.Errorf("store: get profiles by ids: %w", err)
	}
	return profiles, nil
}

func (s *Store) UpdateProfile(ctx context.Context, id types.ID, req UpdateProfileRequest) (*Profile, error) {
	const q = `
		UPDATE users.profiles
		SET
			display_name = COALESCE($1, display_name),
			bio          = COALESCE($2, bio),
			avatar_url   = COALESCE($3, avatar_url),
			updated_at   = now()
		WHERE id = $4
		RETURNING id, username, display_name, bio, avatar_url, verified, created_at, updated_at`

	var p Profile
	err := s.db.Conn.GetContext(ctx, &p, q, req.DisplayName, req.Bio, req.AvatarURL, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.NewNotFound("user")
		}
		return nil, fmt.Errorf("store: update profile: %w", err)
	}
	return &p, nil
}

// Follow operations

func (s *Store) CreateFollow(ctx context.Context, followerID, followingID types.ID) error {
	const q = `INSERT INTO users.follows (follower_id, following_id, created_at) VALUES ($1, $2, now())`

	if _, err := s.db.Conn.ExecContext(ctx, q, followerID, followingID); err != nil {
		if isUniqueViolation(err) {
			return types.NewConflict("already following")
		}
		return fmt.Errorf("store: create follow: %w", err)
	}
	return nil
}

func (s *Store) DeleteFollow(ctx context.Context, followerID, followingID types.ID) error {
	const q = `DELETE FROM users.follows WHERE follower_id = $1 AND following_id = $2`

	_, err := s.db.Conn.ExecContext(ctx, q, followerID, followingID)
	if err != nil {
		return fmt.Errorf("store: delete follow: %w", err)
	}

	return nil
}

func (s *Store) IsFollowing(ctx context.Context, followerID, followingID types.ID) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM users.follows WHERE follower_id = $1 AND following_id = $2)`

	var exists bool
	if err := s.db.Conn.GetContext(ctx, &exists, q, followerID, followingID); err != nil {
		return false, fmt.Errorf("store: is following: %w", err)
	}
	return exists, nil
}

func (s *Store) GetFollowerCount(ctx context.Context, userID types.ID) (int, error) {
	const q = `SELECT COUNT(*) FROM users.follows WHERE following_id = $1`

	var count int
	if err := s.db.Conn.GetContext(ctx, &count, q, userID); err != nil {
		return 0, fmt.Errorf("store: get follower count: %w", err)
	}
	return count, nil
}

func (s *Store) GetFollowingCount(ctx context.Context, userID types.ID) (int, error) {
	const q = `SELECT COUNT(*) FROM users.follows WHERE follower_id = $1`

	var count int
	if err := s.db.Conn.GetContext(ctx, &count, q, userID); err != nil {
		return 0, fmt.Errorf("store: get following count: %w", err)
	}
	return count, nil
}

// followRow is a scan target for follow-list queries that includes the relationship timestamp.
type followRow struct {
	Profile
	FollowedAt time.Time `db:"followed_at"`
}

func (s *Store) GetFollowers(ctx context.Context, userID types.ID, params types.CursorParams) (types.CursorPage[Profile], error) {
	if err := validateTimestampCursor(params.Cursor); err != nil {
		return types.CursorPage[Profile]{}, err
	}

	const q = `
		SELECT p.id, p.username, p.display_name, p.bio, p.avatar_url, p.verified, p.created_at, p.updated_at,
		       f.created_at AS followed_at
		FROM users.follows f
		JOIN users.profiles p ON p.id = f.follower_id
		WHERE f.following_id = $1
		  AND ($2 = '' OR f.created_at < $2::timestamptz)
		ORDER BY f.created_at DESC
		LIMIT $3`

	limit := params.Limit + 1
	rows := make([]followRow, 0, limit)

	if err := s.db.Conn.SelectContext(ctx, &rows, q, userID, params.Cursor, limit); err != nil {
		return types.CursorPage[Profile]{}, fmt.Errorf("store: get followers: %w", err)
	}

	return buildFollowCursorPage(rows, params.Limit), nil
}

func (s *Store) GetFollowing(ctx context.Context, userID types.ID, params types.CursorParams) (types.CursorPage[Profile], error) {
	if err := validateTimestampCursor(params.Cursor); err != nil {
		return types.CursorPage[Profile]{}, err
	}

	const q = `
		SELECT p.id, p.username, p.display_name, p.bio, p.avatar_url, p.verified, p.created_at, p.updated_at,
		       f.created_at AS followed_at
		FROM users.follows f
		JOIN users.profiles p ON p.id = f.following_id
		WHERE f.follower_id = $1
		  AND ($2 = '' OR f.created_at < $2::timestamptz)
		ORDER BY f.created_at DESC
		LIMIT $3`

	limit := params.Limit + 1
	rows := make([]followRow, 0, limit)

	if err := s.db.Conn.SelectContext(ctx, &rows, q, userID, params.Cursor, limit); err != nil {
		return types.CursorPage[Profile]{}, fmt.Errorf("store: get following: %w", err)
	}

	return buildFollowCursorPage(rows, params.Limit), nil
}

func (s *Store) GetFollowingIDs(ctx context.Context, userID types.ID) ([]types.ID, error) {
	const q = `SELECT following_id FROM users.follows WHERE follower_id = $1`

	var ids []types.ID
	if err := s.db.Conn.SelectContext(ctx, &ids, q, userID); err != nil {
		return nil, fmt.Errorf("store: get following ids: %w", err)
	}
	return ids, nil
}

func (s *Store) GetFollowerIDs(ctx context.Context, userID types.ID) ([]types.ID, error) {
	const q = `SELECT follower_id FROM users.follows WHERE following_id = $1`

	var ids []types.ID
	if err := s.db.Conn.SelectContext(ctx, &ids, q, userID); err != nil {
		return nil, fmt.Errorf("store: get follower ids: %w", err)
	}
	return ids, nil
}

// escapeILIKE escapes ILIKE wildcard characters in user input.
func escapeILIKE(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// validateTimestampCursor checks that a cursor string is either empty or a valid RFC3339 timestamp.
func validateTimestampCursor(cursor string) error {
	if cursor == "" {
		return nil
	}
	if _, err := time.Parse(time.RFC3339Nano, cursor); err != nil {
		return types.NewValidation("invalid cursor")
	}
	return nil
}

// Search

func (s *Store) SearchUsers(ctx context.Context, query string, params types.CursorParams) (types.CursorPage[Profile], error) {
	const q = `
		SELECT id, username, display_name, bio, avatar_url, verified, created_at, updated_at
		FROM users.profiles
		WHERE (username ILIKE $1 OR display_name ILIKE $1)
		  AND ($2 = '' OR username > $2)
		ORDER BY username ASC
		LIMIT $3`

	pattern := "%" + escapeILIKE(query) + "%"
	limit := params.Limit + 1
	rows := make([]Profile, 0, limit)

	if err := s.db.Conn.SelectContext(ctx, &rows, q, pattern, params.Cursor, limit); err != nil {
		return types.CursorPage[Profile]{}, fmt.Errorf("store: search users: %w", err)
	}

	return buildSearchCursorPage(rows, params.Limit), nil
}

// buildFollowCursorPage assembles a CursorPage from follow-list rows fetched with limit+1.
// The cursor value is the RFC3339Nano timestamp of the follow relationship (followed_at),
// NOT the profile's created_at — the WHERE clause filters on f.created_at.
func buildFollowCursorPage(rows []followRow, limit int) types.CursorPage[Profile] {
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	profiles := make([]Profile, len(rows))
	for i := range rows {
		profiles[i] = rows[i].Profile
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].FollowedAt.UTC().Format(time.RFC3339Nano)
	}

	return types.CursorPage[Profile]{
		Items:      profiles,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}

// buildSearchCursorPage assembles a CursorPage for username-ordered search results.
// The cursor value is the username of the last included item.
func buildSearchCursorPage(rows []Profile, limit int) types.CursorPage[Profile] {
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].Username
	}

	return types.CursorPage[Profile]{
		Items:      rows,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}
