package users

import (
	"time"

	"github.com/google/uuid"
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
	ID             uuid.UUID `json:"id"`
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
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

type UserFollowedEvent struct {
	FollowerID  uuid.UUID `json:"follower_id"`
	FollowingID uuid.UUID `json:"following_id"`
}

type UserUnfollowedEvent struct {
	FollowerID  uuid.UUID `json:"follower_id"`
	FollowingID uuid.UUID `json:"following_id"`
}

type ProfileUpdatedEvent struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Verified    bool      `json:"verified"`
}

// Bus query types

type GetProfileQuery struct {
	UserID uuid.UUID
}

type GetProfilesQuery struct {
	UserIDs []uuid.UUID
}

type GetFollowingQuery struct {
	UserID uuid.UUID
}

// Topic constants

const (
	TopicRegistered     = "users.registered"
	TopicFollowed       = "users.followed"
	TopicUnfollowed     = "users.unfollowed"
	TopicProfileUpdated = "users.profile_updated"
)

// Query name constants

const (
	QueryGetProfile   = "users.GetProfile"
	QueryGetProfiles  = "users.GetProfiles"
	QueryGetFollowing = "users.GetFollowing"
)
