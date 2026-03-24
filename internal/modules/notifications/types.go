package notifications

import (
	"time"

	"github.com/radni/soapbox/internal/core/types"
)

// Notification type constants.
const (
	TypeLike   = "like"
	TypeRepost = "repost"
	TypeReply  = "reply"
	TypeFollow = "follow"
)

// Event topic constants (owned by other modules).
const (
	postsTopicCreated  = "posts.created"
	postsTopicLiked    = "posts.liked"
	postsTopicReposted = "posts.reposted"
	usersTopicFollowed = "users.followed"
)

// Event published by this module.
const (
	TopicNew = "notifications.new"
)

// Bus query constants (owned by other modules).
const (
	usersQueryGetProfiles = "users.GetProfiles"
	postsQueryGetByIDs    = "posts.GetByIDs"
)

// Bus query exposed by this module.
const (
	QueryGetForUser = "notifications.GetForUser"
)

// GetForUserQuery is the bus query request for notifications.GetForUser.
type GetForUserQuery struct {
	UserID types.ID
	Cursor string
	Limit  int
}

// NotificationResponse is the API/bus response for a single notification.
type NotificationResponse struct {
	ID               types.ID  `json:"id"`
	Type             string    `json:"type"`
	ActorID          types.ID  `json:"actor_id"`
	ActorUsername    string    `json:"actor_username"`
	ActorDisplayName string    `json:"actor_display_name"`
	ActorAvatarURL   string    `json:"actor_avatar_url"`
	PostID           *types.ID `json:"post_id,omitempty"`
	Read             bool      `json:"read"`
	CreatedAt        time.Time `json:"created_at"`
}

// NewNotificationEvent is published on TopicNew after creating a notification.
type NewNotificationEvent struct {
	ID      types.ID  `json:"id"`
	UserID  types.ID  `json:"user_id"`
	Type    string    `json:"type"`
	ActorID types.ID  `json:"actor_id"`
	PostID  *types.ID `json:"post_id,omitempty"`
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

// Mirrors posts.PostLikedEvent.
type postLikedEvent struct {
	PostID    types.ID  `json:"post_id"`
	UserID    types.ID  `json:"user_id"`
	AuthorID  types.ID  `json:"author_id"`
	LikeCount int       `json:"like_count"`
	CreatedAt time.Time `json:"created_at"`
}

// Mirrors posts.PostRepostedEvent.
type postRepostedEvent struct {
	PostID      types.ID  `json:"post_id"`
	UserID      types.ID  `json:"user_id"`
	AuthorID    types.ID  `json:"author_id"`
	RepostCount int       `json:"repost_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// Mirrors users.UserFollowedEvent.
type userFollowedEvent struct {
	FollowerID  types.ID `json:"follower_id"`
	FollowingID types.ID `json:"following_id"`
}

// Mirrors users.GetProfilesQuery.
type usersGetProfilesQuery struct {
	UserIDs  []types.ID
	ViewerID *types.ID
}

// Mirrors users.ProfileResponse (subset needed for actor enrichment).
type userProfileResponse struct {
	ID          types.ID `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	AvatarURL   string   `json:"avatar_url"`
}

// Mirrors posts.GetByIDsQuery.
type postsGetByIDsQuery struct {
	PostIDs  []types.ID
	ViewerID *types.ID
}

// Mirrors posts.PostResponse (subset needed for reply parent lookup).
type postResponse struct {
	ID       types.ID `json:"id"`
	AuthorID types.ID `json:"author_id"`
}

// Swagger response types (swaggo does not support Go generics).

// NotificationCursorPage is the response for GET /api/v1/notifications.
type NotificationCursorPage struct {
	Items      []NotificationResponse `json:"items"`
	NextCursor string                 `json:"next_cursor,omitempty"`
	HasMore    bool                   `json:"has_more"`
}
