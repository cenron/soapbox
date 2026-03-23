package posts

import (
	"time"

	"github.com/radni/soapbox/internal/core/types"
)

// Request types

type CreatePostRequest struct {
	Body       string   `json:"body"`
	MediaIDs   []string `json:"media_ids,omitempty"`
	ParentID   *string  `json:"parent_id,omitempty"`
	RepostOfID *string  `json:"repost_of_id,omitempty"`
}

// Response types

type PostResponse struct {
	ID                types.ID             `json:"id"`
	AuthorID          types.ID             `json:"author_id"`
	AuthorUsername    string               `json:"author_username"`
	AuthorDisplayName string               `json:"author_display_name"`
	AuthorAvatarURL   string               `json:"author_avatar_url"`
	AuthorVerified    bool                 `json:"author_verified"`
	Body              string               `json:"body"`
	ParentID          *types.ID            `json:"parent_id"`
	RepostOfID        *types.ID            `json:"repost_of_id"`
	Media             []MediaResponse      `json:"media"`
	LinkPreview       *LinkPreviewResponse `json:"link_preview"`
	Hashtags          []string             `json:"hashtags"`
	LikeCount         int                  `json:"like_count"`
	RepostCount       int                  `json:"repost_count"`
	ReplyCount        int                  `json:"reply_count"`
	LikedByMe         bool                 `json:"liked_by_me"`
	RepostedByMe      bool                 `json:"reposted_by_me"`
	CreatedAt         time.Time            `json:"created_at"`
}

type MediaResponse struct {
	ID        types.ID `json:"id"`
	MediaURL  string   `json:"media_url"`
	MediaType string   `json:"media_type"`
	Position  int      `json:"position"`
}

type LinkPreviewResponse struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

type LikeResponse struct {
	PostID    types.ID `json:"post_id"`
	LikeCount int      `json:"like_count"`
	LikedByMe bool     `json:"liked_by_me"`
}

type RepostResponse struct {
	PostID       types.ID `json:"post_id"`
	RepostCount  int      `json:"repost_count"`
	RepostedByMe bool     `json:"reposted_by_me"`
}

// Event types

type PostCreatedEvent struct {
	PostID         types.ID  `json:"post_id"`
	AuthorID       types.ID  `json:"author_id"`
	AuthorUsername string    `json:"author_username"`
	Body           string    `json:"body"`
	ParentID       *types.ID `json:"parent_id,omitempty"`
	RepostOfID     *types.ID `json:"repost_of_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type PostLikedEvent struct {
	PostID    types.ID  `json:"post_id"`
	UserID    types.ID  `json:"user_id"`
	AuthorID  types.ID  `json:"author_id"`
	LikeCount int       `json:"like_count"`
	CreatedAt time.Time `json:"created_at"`
}

type PostRepostedEvent struct {
	PostID      types.ID  `json:"post_id"`
	UserID      types.ID  `json:"user_id"`
	AuthorID    types.ID  `json:"author_id"`
	RepostCount int       `json:"repost_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type PostDeletedEvent struct {
	PostID    types.ID  `json:"post_id"`
	AuthorID  types.ID  `json:"author_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

// Topic constants

const (
	TopicCreated  = "posts.created"
	TopicLiked    = "posts.liked"
	TopicReposted = "posts.reposted"
	TopicDeleted  = "posts.deleted"
)

// Query name constants

const (
	QueryGetByIDs      = "posts.GetByIDs"
	QueryGetByAuthor   = "posts.GetByAuthor"
	QueryGetThread     = "posts.GetThread"
	QuerySearch        = "posts.Search"
	QuerySearchHashtag = "posts.SearchHashtag"
)

// Bus query types

type GetByIDsQuery struct {
	PostIDs  []types.ID
	ViewerID *types.ID
}

type GetByAuthorQuery struct {
	AuthorID types.ID
	ViewerID *types.ID
	Cursor   string
	Limit    int
}

type GetThreadQuery struct {
	RootPostID types.ID
	ViewerID   *types.ID
}

type SearchPostsQuery struct {
	Q        string
	ViewerID *types.ID
	Cursor   string
	Limit    int
}

type SearchHashtagQuery struct {
	Tag      string
	ViewerID *types.ID
	Cursor   string
	Limit    int
}

// Swagger response types (swaggo does not support Go generics)

type PostCursorPage struct {
	Items      []PostResponse `json:"items"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}
