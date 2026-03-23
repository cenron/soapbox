package posts

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

type Post struct {
	ID                types.ID  `db:"id"`
	AuthorID          types.ID  `db:"author_id"`
	AuthorUsername    string    `db:"author_username"`
	AuthorDisplayName string    `db:"author_display_name"`
	AuthorAvatarURL   string    `db:"author_avatar_url"`
	AuthorVerified    bool      `db:"author_verified"`
	Body              string    `db:"body"`
	ParentID          *types.ID `db:"parent_id"`
	RepostOfID        *types.ID `db:"repost_of_id"`
	LikeCount         int       `db:"like_count"`
	RepostCount       int       `db:"repost_count"`
	ReplyCount        int       `db:"reply_count"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

type PostMedia struct {
	ID        types.ID `db:"id"`
	PostID    types.ID `db:"post_id"`
	MediaURL  string   `db:"media_url"`
	MediaType string   `db:"media_type"`
	Position  int      `db:"position"`
}

type LinkPreview struct {
	ID          types.ID `db:"id"`
	PostID      types.ID `db:"post_id"`
	URL         string   `db:"url"`
	Title       string   `db:"title"`
	Description string   `db:"description"`
	ImageURL    string   `db:"image_url"`
}

type Hashtag struct {
	PostID types.ID `db:"post_id"`
	Tag    string   `db:"tag"`
}

type Like struct {
	PostID    types.ID  `db:"post_id"`
	UserID    types.ID  `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

// Store

type Store struct {
	db *db.DB
}

func NewStore(database *db.DB) *Store {
	return &Store{db: database}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// Post operations

func (s *Store) CreatePost(ctx context.Context, tx *sqlx.Tx, p *Post) error {
	const q = `
		INSERT INTO posts.posts (
			id, author_id, author_username, author_display_name,
			author_avatar_url, author_verified, body, parent_id, repost_of_id,
			like_count, repost_count, reply_count, created_at, updated_at
		) VALUES (
			:id, :author_id, :author_username, :author_display_name,
			:author_avatar_url, :author_verified, :body, :parent_id, :repost_of_id,
			:like_count, :repost_count, :reply_count, :created_at, :updated_at
		)`

	_, err := tx.NamedExecContext(ctx, q, p)
	if err != nil {
		return fmt.Errorf("store: create post: %w", err)
	}
	return nil
}

func (s *Store) GetPostByID(ctx context.Context, id types.ID) (*Post, error) {
	const q = `
		SELECT id, author_id, author_username, author_display_name,
		       author_avatar_url, author_verified, body, parent_id, repost_of_id,
		       like_count, repost_count, reply_count, created_at, updated_at
		FROM posts.posts
		WHERE id = $1`

	var p Post
	if err := s.db.Conn.GetContext(ctx, &p, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.NewNotFound("post not found")
		}
		return nil, fmt.Errorf("store: get post by id: %w", err)
	}
	return &p, nil
}

func (s *Store) GetPostsByIDs(ctx context.Context, ids []types.ID) ([]Post, error) {
	if len(ids) == 0 {
		return []Post{}, nil
	}

	query, args, err := sqlx.In(
		`SELECT id, author_id, author_username, author_display_name,
		        author_avatar_url, author_verified, body, parent_id, repost_of_id,
		        like_count, repost_count, reply_count, created_at, updated_at
		 FROM posts.posts WHERE id IN (?)`,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("store: get posts by ids: build query: %w", err)
	}

	query = s.db.Conn.Rebind(query)

	var posts []Post
	if err := s.db.Conn.SelectContext(ctx, &posts, query, args...); err != nil {
		return nil, fmt.Errorf("store: get posts by ids: %w", err)
	}
	return posts, nil
}

func (s *Store) DeletePost(ctx context.Context, id types.ID) error {
	const q = `DELETE FROM posts.posts WHERE id = $1`

	res, err := s.db.Conn.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("store: delete post: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: delete post: rows affected: %w", err)
	}
	if rows == 0 {
		return types.NewNotFound("post not found")
	}
	return nil
}

func (s *Store) GetReplies(ctx context.Context, parentID types.ID, params types.CursorParams) ([]Post, bool, error) {
	if err := validateTimestampCursor(params.Cursor); err != nil {
		return nil, false, err
	}

	const q = `
		SELECT id, author_id, author_username, author_display_name,
		       author_avatar_url, author_verified, body, parent_id, repost_of_id,
		       like_count, repost_count, reply_count, created_at, updated_at
		FROM posts.posts
		WHERE parent_id = $1
		  AND ($2 = '' OR created_at > $2::timestamptz)
		ORDER BY created_at ASC
		LIMIT $3`

	limit := params.Limit + 1
	var posts []Post
	if err := s.db.Conn.SelectContext(ctx, &posts, q, parentID, params.Cursor, limit); err != nil {
		return nil, false, fmt.Errorf("store: get replies: %w", err)
	}

	hasMore := len(posts) > params.Limit
	if hasMore {
		posts = posts[:params.Limit]
	}
	return posts, hasMore, nil
}

func (s *Store) GetPostsByAuthor(ctx context.Context, authorID types.ID, params types.CursorParams) ([]Post, bool, error) {
	if err := validateTimestampCursor(params.Cursor); err != nil {
		return nil, false, err
	}

	const q = `
		SELECT id, author_id, author_username, author_display_name,
		       author_avatar_url, author_verified, body, parent_id, repost_of_id,
		       like_count, repost_count, reply_count, created_at, updated_at
		FROM posts.posts
		WHERE author_id = $1
		  AND parent_id IS NULL
		  AND ($2 = '' OR created_at < $2::timestamptz)
		ORDER BY created_at DESC
		LIMIT $3`

	limit := params.Limit + 1
	var posts []Post
	if err := s.db.Conn.SelectContext(ctx, &posts, q, authorID, params.Cursor, limit); err != nil {
		return nil, false, fmt.Errorf("store: get posts by author: %w", err)
	}

	hasMore := len(posts) > params.Limit
	if hasMore {
		posts = posts[:params.Limit]
	}
	return posts, hasMore, nil
}

func (s *Store) IncrementReplyCount(ctx context.Context, tx *sqlx.Tx, postID types.ID) error {
	const q = `UPDATE posts.posts SET reply_count = reply_count + 1 WHERE id = $1`

	_, err := tx.ExecContext(ctx, q, postID)
	if err != nil {
		return fmt.Errorf("store: increment reply count: %w", err)
	}
	return nil
}

func (s *Store) DecrementReplyCount(ctx context.Context, parentID types.ID) error {
	const q = `UPDATE posts.posts SET reply_count = GREATEST(reply_count - 1, 0) WHERE id = $1`

	_, err := s.db.Conn.ExecContext(ctx, q, parentID)
	if err != nil {
		return fmt.Errorf("store: decrement reply count: %w", err)
	}
	return nil
}

// Media operations

func (s *Store) CreatePostMedia(ctx context.Context, tx *sqlx.Tx, m *PostMedia) error {
	const q = `
		INSERT INTO posts.media (id, post_id, media_url, media_type, position)
		VALUES (:id, :post_id, :media_url, :media_type, :position)`

	_, err := tx.NamedExecContext(ctx, q, m)
	if err != nil {
		return fmt.Errorf("store: create post media: %w", err)
	}
	return nil
}

func (s *Store) GetMediaByPostID(ctx context.Context, postID types.ID) ([]PostMedia, error) {
	const q = `
		SELECT id, post_id, media_url, media_type, position
		FROM posts.media
		WHERE post_id = $1
		ORDER BY position`

	var media []PostMedia
	if err := s.db.Conn.SelectContext(ctx, &media, q, postID); err != nil {
		return nil, fmt.Errorf("store: get media by post id: %w", err)
	}
	return media, nil
}

func (s *Store) GetMediaByPostIDs(ctx context.Context, postIDs []types.ID) (map[types.ID][]PostMedia, error) {
	if len(postIDs) == 0 {
		return map[types.ID][]PostMedia{}, nil
	}

	query, args, err := sqlx.In(
		`SELECT id, post_id, media_url, media_type, position
		 FROM posts.media WHERE post_id IN (?) ORDER BY position`,
		postIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("store: get media by post ids: build query: %w", err)
	}

	query = s.db.Conn.Rebind(query)

	var media []PostMedia
	if err := s.db.Conn.SelectContext(ctx, &media, query, args...); err != nil {
		return nil, fmt.Errorf("store: get media by post ids: %w", err)
	}

	result := make(map[types.ID][]PostMedia, len(postIDs))
	for i := range media {
		result[media[i].PostID] = append(result[media[i].PostID], media[i])
	}
	return result, nil
}

// Link preview operations

func (s *Store) CreateLinkPreview(ctx context.Context, tx *sqlx.Tx, lp *LinkPreview) error {
	const q = `
		INSERT INTO posts.link_previews (id, post_id, url, title, description, image_url)
		VALUES (:id, :post_id, :url, :title, :description, :image_url)`

	_, err := tx.NamedExecContext(ctx, q, lp)
	if err != nil {
		return fmt.Errorf("store: create link preview: %w", err)
	}
	return nil
}

func (s *Store) GetLinkPreviewByPostID(ctx context.Context, postID types.ID) (*LinkPreview, error) {
	const q = `
		SELECT id, post_id, url, title, description, image_url
		FROM posts.link_previews
		WHERE post_id = $1`

	var lp LinkPreview
	if err := s.db.Conn.GetContext(ctx, &lp, q, postID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("store: get link preview by post id: %w", err)
	}
	return &lp, nil
}

func (s *Store) GetLinkPreviewsByPostIDs(ctx context.Context, postIDs []types.ID) (map[types.ID]*LinkPreview, error) {
	if len(postIDs) == 0 {
		return map[types.ID]*LinkPreview{}, nil
	}

	query, args, err := sqlx.In(
		`SELECT id, post_id, url, title, description, image_url
		 FROM posts.link_previews WHERE post_id IN (?)`,
		postIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("store: get link previews by post ids: build query: %w", err)
	}

	query = s.db.Conn.Rebind(query)

	var previews []LinkPreview
	if err := s.db.Conn.SelectContext(ctx, &previews, query, args...); err != nil {
		return nil, fmt.Errorf("store: get link previews by post ids: %w", err)
	}

	result := make(map[types.ID]*LinkPreview, len(previews))
	for i := range previews {
		result[previews[i].PostID] = &previews[i]
	}
	return result, nil
}

// Hashtag operations

func (s *Store) CreateHashtags(ctx context.Context, tx *sqlx.Tx, postID types.ID, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	const q = `INSERT INTO posts.hashtags (post_id, tag) VALUES ($1, $2)`

	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, q, postID, strings.ToLower(tag)); err != nil {
			return fmt.Errorf("store: create hashtag %q: %w", tag, err)
		}
	}
	return nil
}

func (s *Store) GetHashtagsByPostID(ctx context.Context, postID types.ID) ([]string, error) {
	const q = `SELECT tag FROM posts.hashtags WHERE post_id = $1 ORDER BY tag`

	var tags []string
	if err := s.db.Conn.SelectContext(ctx, &tags, q, postID); err != nil {
		return nil, fmt.Errorf("store: get hashtags by post id: %w", err)
	}
	return tags, nil
}

func (s *Store) GetHashtagsByPostIDs(ctx context.Context, postIDs []types.ID) (map[types.ID][]string, error) {
	if len(postIDs) == 0 {
		return map[types.ID][]string{}, nil
	}

	query, args, err := sqlx.In(
		`SELECT post_id, tag FROM posts.hashtags WHERE post_id IN (?) ORDER BY tag`,
		postIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("store: get hashtags by post ids: build query: %w", err)
	}

	query = s.db.Conn.Rebind(query)

	var rows []Hashtag
	if err := s.db.Conn.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("store: get hashtags by post ids: %w", err)
	}

	result := make(map[types.ID][]string, len(postIDs))
	for _, h := range rows {
		result[h.PostID] = append(result[h.PostID], h.Tag)
	}
	return result, nil
}

// Like operations

func (s *Store) CreateLike(ctx context.Context, postID, userID types.ID) error {
	const q = `INSERT INTO posts.likes (post_id, user_id, created_at) VALUES ($1, $2, $3)`

	_, err := s.db.Conn.ExecContext(ctx, q, postID, userID, time.Now().UTC())
	if err != nil {
		if isUniqueViolation(err) {
			return types.NewConflict("already liked")
		}
		return fmt.Errorf("store: create like: %w", err)
	}
	return nil
}

func (s *Store) DeleteLike(ctx context.Context, postID, userID types.ID) error {
	const q = `DELETE FROM posts.likes WHERE post_id = $1 AND user_id = $2`

	res, err := s.db.Conn.ExecContext(ctx, q, postID, userID)
	if err != nil {
		return fmt.Errorf("store: delete like: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: delete like: rows affected: %w", err)
	}
	if rows == 0 {
		return types.NewNotFound("like not found")
	}
	return nil
}

func (s *Store) IncrementLikeCount(ctx context.Context, postID types.ID) (int, error) {
	const q = `UPDATE posts.posts SET like_count = like_count + 1 WHERE id = $1 RETURNING like_count`

	var count int
	if err := s.db.Conn.GetContext(ctx, &count, q, postID); err != nil {
		return 0, fmt.Errorf("store: increment like count: %w", err)
	}
	return count, nil
}

func (s *Store) DecrementLikeCount(ctx context.Context, postID types.ID) (int, error) {
	const q = `UPDATE posts.posts SET like_count = GREATEST(like_count - 1, 0) WHERE id = $1 RETURNING like_count`

	var count int
	if err := s.db.Conn.GetContext(ctx, &count, q, postID); err != nil {
		return 0, fmt.Errorf("store: decrement like count: %w", err)
	}
	return count, nil
}

func (s *Store) IsLikedByUser(ctx context.Context, postID, userID types.ID) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM posts.likes WHERE post_id = $1 AND user_id = $2)`

	var exists bool
	if err := s.db.Conn.GetContext(ctx, &exists, q, postID, userID); err != nil {
		return false, fmt.Errorf("store: is liked by user: %w", err)
	}
	return exists, nil
}

func (s *Store) IsLikedByUserBatch(ctx context.Context, postIDs []types.ID, userID types.ID) (map[types.ID]bool, error) {
	if len(postIDs) == 0 {
		return map[types.ID]bool{}, nil
	}

	query, args, err := sqlx.In(
		`SELECT post_id FROM posts.likes WHERE post_id IN (?) AND user_id = ?`,
		postIDs, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: is liked by user batch: build query: %w", err)
	}

	query = s.db.Conn.Rebind(query)

	var likedIDs []types.ID
	if err := s.db.Conn.SelectContext(ctx, &likedIDs, query, args...); err != nil {
		return nil, fmt.Errorf("store: is liked by user batch: %w", err)
	}

	result := make(map[types.ID]bool, len(likedIDs))
	for _, id := range likedIDs {
		result[id] = true
	}
	return result, nil
}

// Repost operations

func (s *Store) IsRepostedByUser(ctx context.Context, postID, userID types.ID) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM posts.posts WHERE repost_of_id = $1 AND author_id = $2)`

	var exists bool
	if err := s.db.Conn.GetContext(ctx, &exists, q, postID, userID); err != nil {
		return false, fmt.Errorf("store: is reposted by user: %w", err)
	}
	return exists, nil
}

func (s *Store) IsRepostedByUserBatch(ctx context.Context, postIDs []types.ID, userID types.ID) (map[types.ID]bool, error) {
	if len(postIDs) == 0 {
		return map[types.ID]bool{}, nil
	}

	query, args, err := sqlx.In(
		`SELECT repost_of_id FROM posts.posts WHERE repost_of_id IN (?) AND author_id = ?`,
		postIDs, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: is reposted by user batch: build query: %w", err)
	}

	query = s.db.Conn.Rebind(query)

	var repostedIDs []types.ID
	if err := s.db.Conn.SelectContext(ctx, &repostedIDs, query, args...); err != nil {
		return nil, fmt.Errorf("store: is reposted by user batch: %w", err)
	}

	result := make(map[types.ID]bool, len(repostedIDs))
	for _, id := range repostedIDs {
		result[id] = true
	}
	return result, nil
}

func (s *Store) GetRepostByUser(ctx context.Context, postID, userID types.ID) (*Post, error) {
	const q = `
		SELECT id, author_id, author_username, author_display_name,
		       author_avatar_url, author_verified, body, parent_id, repost_of_id,
		       like_count, repost_count, reply_count, created_at, updated_at
		FROM posts.posts
		WHERE repost_of_id = $1 AND author_id = $2`

	var p Post
	if err := s.db.Conn.GetContext(ctx, &p, q, postID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.NewNotFound("repost not found")
		}
		return nil, fmt.Errorf("store: get repost by user: %w", err)
	}
	return &p, nil
}

func (s *Store) IncrementRepostCount(ctx context.Context, postID types.ID) (int, error) {
	const q = `UPDATE posts.posts SET repost_count = repost_count + 1 WHERE id = $1 RETURNING repost_count`

	var count int
	if err := s.db.Conn.GetContext(ctx, &count, q, postID); err != nil {
		return 0, fmt.Errorf("store: increment repost count: %w", err)
	}
	return count, nil
}

func (s *Store) DecrementRepostCount(ctx context.Context, postID types.ID) (int, error) {
	const q = `UPDATE posts.posts SET repost_count = GREATEST(repost_count - 1, 0) WHERE id = $1 RETURNING repost_count`

	var count int
	if err := s.db.Conn.GetContext(ctx, &count, q, postID); err != nil {
		return 0, fmt.Errorf("store: decrement repost count: %w", err)
	}
	return count, nil
}

// Author posts by username

func (s *Store) GetPostsByUsername(ctx context.Context, username string, params types.CursorParams) ([]Post, bool, error) {
	if err := validateTimestampCursor(params.Cursor); err != nil {
		return nil, false, err
	}

	const q = `
		SELECT id, author_id, author_username, author_display_name,
		       author_avatar_url, author_verified, body, parent_id, repost_of_id,
		       like_count, repost_count, reply_count, created_at, updated_at
		FROM posts.posts
		WHERE lower(author_username) = lower($1)
		  AND parent_id IS NULL
		  AND ($2 = '' OR created_at < $2::timestamptz)
		ORDER BY created_at DESC
		LIMIT $3`

	limit := params.Limit + 1
	var posts []Post
	if err := s.db.Conn.SelectContext(ctx, &posts, q, username, params.Cursor, limit); err != nil {
		return nil, false, fmt.Errorf("store: get posts by username: %w", err)
	}

	hasMore := len(posts) > params.Limit
	if hasMore {
		posts = posts[:params.Limit]
	}
	return posts, hasMore, nil
}

// Denormalization sync

func (s *Store) UpdateAuthorDenorm(ctx context.Context, authorID types.ID, username, displayName, avatarURL string, verified bool) error {
	const q = `
		UPDATE posts.posts
		SET author_username = $1,
		    author_display_name = $2,
		    author_avatar_url = $3,
		    author_verified = $4
		WHERE author_id = $5`

	_, err := s.db.Conn.ExecContext(ctx, q, username, displayName, avatarURL, verified, authorID)
	if err != nil {
		return fmt.Errorf("store: update author denorm: %w", err)
	}
	return nil
}

// --- helpers ---

func validateTimestampCursor(cursor string) error {
	if cursor == "" {
		return nil
	}
	if _, err := time.Parse(time.RFC3339Nano, cursor); err != nil {
		return types.NewValidation("invalid cursor")
	}
	return nil
}
