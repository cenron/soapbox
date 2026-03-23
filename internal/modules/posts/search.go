package posts

import (
	"context"
	"fmt"
	"strings"

	"github.com/radni/soapbox/internal/core/types"
)

// SearchPosts performs a full-text search over post body content.
func (s *Service) SearchPosts(ctx context.Context, query string, viewerID *types.ID, params types.CursorParams) (*types.CursorPage[PostResponse], error) {
	posts, hasMore, err := s.store.SearchPosts(ctx, query, params)
	if err != nil {
		return nil, err
	}

	return s.buildPostCursorPage(ctx, posts, hasMore, viewerID)
}

// SearchByHashtag returns posts containing a specific hashtag.
func (s *Service) SearchByHashtag(ctx context.Context, tag string, viewerID *types.ID, params types.CursorParams) (*types.CursorPage[PostResponse], error) {
	posts, hasMore, err := s.store.SearchByHashtag(ctx, tag, params)
	if err != nil {
		return nil, err
	}

	return s.buildPostCursorPage(ctx, posts, hasMore, viewerID)
}

// Store search operations

func (s *Store) SearchPosts(ctx context.Context, query string, params types.CursorParams) ([]Post, bool, error) {
	if err := validateTimestampCursor(params.Cursor); err != nil {
		return nil, false, err
	}

	const q = `
		SELECT id, author_id, author_username, author_display_name,
		       author_avatar_url, author_verified, body, parent_id, repost_of_id,
		       like_count, repost_count, reply_count, created_at, updated_at
		FROM posts.posts
		WHERE to_tsvector('english', body) @@ plainto_tsquery('english', $1)
		  AND ($2 = '' OR created_at < $2::timestamptz)
		ORDER BY created_at DESC
		LIMIT $3`

	limit := params.Limit + 1
	var posts []Post
	if err := s.db.Conn.SelectContext(ctx, &posts, q, query, params.Cursor, limit); err != nil {
		return nil, false, fmt.Errorf("store: search posts: %w", err)
	}

	hasMore := len(posts) > params.Limit
	if hasMore {
		posts = posts[:params.Limit]
	}
	return posts, hasMore, nil
}

func (s *Store) SearchByHashtag(ctx context.Context, tag string, params types.CursorParams) ([]Post, bool, error) {
	if err := validateTimestampCursor(params.Cursor); err != nil {
		return nil, false, err
	}

	const q = `
		SELECT p.id, p.author_id, p.author_username, p.author_display_name,
		       p.author_avatar_url, p.author_verified, p.body, p.parent_id, p.repost_of_id,
		       p.like_count, p.repost_count, p.reply_count, p.created_at, p.updated_at
		FROM posts.posts p
		JOIN posts.hashtags h ON h.post_id = p.id
		WHERE lower(h.tag) = $1
		  AND ($2 = '' OR p.created_at < $2::timestamptz)
		ORDER BY p.created_at DESC
		LIMIT $3`

	normalizedTag := strings.ToLower(strings.TrimPrefix(tag, "#"))

	limit := params.Limit + 1
	var posts []Post
	if err := s.db.Conn.SelectContext(ctx, &posts, q, normalizedTag, params.Cursor, limit); err != nil {
		return nil, false, fmt.Errorf("store: search by hashtag: %w", err)
	}

	hasMore := len(posts) > params.Limit
	if hasMore {
		posts = posts[:params.Limit]
	}
	return posts, hasMore, nil
}
