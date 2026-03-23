package users

import (
	"time"

	"github.com/radni/soapbox/internal/core/types"
)

// Role constants

const (
	RoleUser      = "user"
	RoleModerator = "moderator"
	RoleAdmin     = "admin"
)

// Request types

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name"`
	Bio         *string `json:"bio"`
	AvatarURL   *string `json:"avatar_url"`
}

// Response types

type ProfileResponse struct {
	ID             types.ID  `json:"id"`
	Username       string    `json:"username"`
	DisplayName    string    `json:"display_name"`
	Bio            string    `json:"bio"`
	AvatarURL      string    `json:"avatar_url"`
	Verified       bool      `json:"verified"`
	FollowerCount  int       `json:"follower_count"`
	FollowingCount int       `json:"following_count"`
	IsFollowing    bool      `json:"is_following"`
	CreatedAt      time.Time `json:"created_at"`
}

type AuthResponse struct {
	AccessToken string          `json:"access_token"`
	User        ProfileResponse `json:"user"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

// Event types

type UserRegisteredEvent struct {
	UserID   types.ID `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
}

type UserFollowedEvent struct {
	FollowerID  types.ID `json:"follower_id"`
	FollowingID types.ID `json:"following_id"`
}

type UserUnfollowedEvent struct {
	FollowerID  types.ID `json:"follower_id"`
	FollowingID types.ID `json:"following_id"`
}

type ProfileUpdatedEvent struct {
	UserID      types.ID `json:"user_id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	AvatarURL   string   `json:"avatar_url"`
	Verified    bool     `json:"verified"`
}

// Bus query types

type GetProfileQuery struct {
	UserID   types.ID
	ViewerID *types.ID
}

type GetProfilesQuery struct {
	UserIDs  []types.ID
	ViewerID *types.ID
}

type GetFollowingQuery struct {
	UserID types.ID
}

type GetFollowerIDsQuery struct {
	UserID types.ID
}

// Topic constants

const (
	TopicRegistered     = "users.registered"
	TopicFollowed       = "users.followed"
	TopicUnfollowed     = "users.unfollowed"
	TopicProfileUpdated = "users.profile_updated"
)

// Swagger response types (swaggo does not support Go generics)

type ProfileCursorPage struct {
	Items      []ProfileResponse `json:"items"`
	NextCursor string            `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more"`
}

// Query name constants

const (
	QueryGetProfile     = "users.GetProfile"
	QueryGetProfiles    = "users.GetProfiles"
	QueryGetFollowing   = "users.GetFollowing"
	QueryGetFollowerIDs = "users.GetFollowerIDs"
)
