package feed

import (
	"time"

	"github.com/radni/soapbox/internal/core/types"
)

// Event topic constants (owned by other modules).
const (
	postsTopicCreated    = "posts.created"
	postsTopicDeleted    = "posts.deleted"
	usersTopicFollowed   = "users.followed"
	usersTopicUnfollowed = "users.unfollowed"
)

// Bus query constants (owned by other modules).
const (
	postsQueryGetByIDs       = "posts.GetByIDs"
	postsQueryGetByAuthor    = "posts.GetByAuthor"
	usersQueryGetFollowerIDs = "users.GetFollowerIDs"
)

// Bus query exposed by this module.
const (
	QueryGetTimeline = "feed.GetTimeline"
)

// GetTimelineQuery is the bus query request for feed.GetTimeline.
type GetTimelineQuery struct {
	UserID   types.ID
	ViewerID *types.ID
	Cursor   string
	Limit    int
}

// --- Bus contract mirrors ---
// Local copies of types owned by other modules.
// JSON tags must match the publisher's struct layout exactly.

// Mirrors posts.PostCreatedEvent.
type postCreatedEvent struct {
	PostID         types.ID  `json:"post_id"`
	AuthorID       types.ID  `json:"author_id"`
	AuthorUsername string    `json:"author_username"`
	Body           string    `json:"body"`
	ParentID       *types.ID `json:"parent_id,omitempty"`
	RepostOfID     *types.ID `json:"repost_of_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// Mirrors posts.PostDeletedEvent.
type postDeletedEvent struct {
	PostID    types.ID  `json:"post_id"`
	AuthorID  types.ID  `json:"author_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

// Mirrors users.UserFollowedEvent.
type userFollowedEvent struct {
	FollowerID  types.ID `json:"follower_id"`
	FollowingID types.ID `json:"following_id"`
}

// Mirrors users.UserUnfollowedEvent.
type userUnfollowedEvent struct {
	FollowerID  types.ID `json:"follower_id"`
	FollowingID types.ID `json:"following_id"`
}

// Mirrors users.GetFollowerIDsQuery.
type usersGetFollowerIDsQuery struct {
	UserID types.ID
}

// Mirrors posts.GetByIDsQuery.
type postsGetByIDsQuery struct {
	PostIDs  []types.ID
	ViewerID *types.ID
}

// Mirrors posts.GetByAuthorQuery.
type postsGetByAuthorQuery struct {
	AuthorID types.ID
	ViewerID *types.ID
	Cursor   string
	Limit    int
}

// Mirrors posts.PostResponse (subset needed for timeline rendering).
type postResponse struct {
	ID                types.ID            `json:"id"`
	AuthorID          types.ID            `json:"author_id"`
	AuthorUsername    string              `json:"author_username"`
	AuthorDisplayName string              `json:"author_display_name"`
	AuthorAvatarURL   string              `json:"author_avatar_url"`
	AuthorVerified    bool                `json:"author_verified"`
	Body              string              `json:"body"`
	ParentID          *types.ID           `json:"parent_id"`
	RepostOfID        *types.ID           `json:"repost_of_id"`
	Media             []postMediaResponse `json:"media"`
	LinkPreview       *postLinkPreview    `json:"link_preview"`
	Hashtags          []string            `json:"hashtags"`
	LikeCount         int                 `json:"like_count"`
	RepostCount       int                 `json:"repost_count"`
	ReplyCount        int                 `json:"reply_count"`
	LikedByMe         bool                `json:"liked_by_me"`
	RepostedByMe      bool                `json:"reposted_by_me"`
	CreatedAt         time.Time           `json:"created_at"`
}

type postMediaResponse struct {
	ID        types.ID `json:"id"`
	MediaURL  string   `json:"media_url"`
	MediaType string   `json:"media_type"`
	Position  int      `json:"position"`
}

type postLinkPreview struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

// Mirrors types.CursorPage[posts.PostResponse] for bus query results.
type postsCursorPage struct {
	Items      []postResponse `json:"items"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

// Swagger response types (swaggo does not support Go generics).

// TimelineCursorPage is the response for GET /api/v1/feed.
type TimelineCursorPage struct {
	Items      []postResponse `json:"items"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}
