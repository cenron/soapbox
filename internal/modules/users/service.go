package users

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/radni/soapbox/internal/core/bus"
	"github.com/radni/soapbox/internal/core/config"
	"github.com/radni/soapbox/internal/core/db"
	"github.com/radni/soapbox/internal/core/types"
)

type Service struct {
	db     *db.DB
	store  *Store
	tokens *TokenService
	bus    bus.Bus
	logger *slog.Logger
	jwt    config.JWTConfig
	isProd bool
}

func NewService(database *db.DB, store *Store, tokens *TokenService, b bus.Bus, logger *slog.Logger, jwt config.JWTConfig, isProd bool) *Service {
	return &Service{
		db:     database,
		store:  store,
		tokens: tokens,
		bus:    b,
		logger: logger,
		jwt:    jwt,
		isProd: isProd,
	}
}

// Register creates a new user account and returns auth tokens.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, *http.Cookie, error) {
	if err := validateRegisterRequest(req); err != nil {
		return nil, nil, err
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("service: register: hash password: %w", err)
	}

	now := time.Now().UTC()

	userID, err := types.NewID()
	if err != nil {
		return nil, nil, fmt.Errorf("service: register: generate user id: %w", err)
	}

	credentialID, err := types.NewID()
	if err != nil {
		return nil, nil, fmt.Errorf("service: register: generate credential id: %w", err)
	}

	profile := &Profile{
		ID:          userID,
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Bio:         "",
		AvatarURL:   "",
		Verified:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	credential := &Credential{
		ID:           credentialID,
		UserID:       userID,
		Email:        strings.ToLower(req.Email),
		PasswordHash: passwordHash,
		CreatedAt:    now,
	}

	err = s.db.WithTx(ctx, func(tx *sqlx.Tx) error {
		if err := s.store.CreateProfile(ctx, tx, profile); err != nil {
			return err
		}
		return s.store.CreateCredential(ctx, tx, credential)
	})
	if err != nil {
		return nil, nil, err
	}

	_, cookie, session, err := s.createSession(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	if err := s.store.CreateSession(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("service: register: create session: %w", err)
	}

	accessToken, err := s.tokens.GenerateAccessToken(userID, req.Username, RoleUser, false)
	if err != nil {
		return nil, nil, fmt.Errorf("service: register: generate access token: %w", err)
	}

	if err := s.bus.Publish(TopicRegistered, UserRegisteredEvent{
		UserID:   userID,
		Username: req.Username,
		Email:    credential.Email,
	}); err != nil {
		s.logger.Warn("service: register: publish event failed", "error", err)
	}

	resp := &AuthResponse{
		AccessToken: accessToken,
		User: ProfileResponse{
			ID:             userID,
			Username:       profile.Username,
			DisplayName:    profile.DisplayName,
			Bio:            profile.Bio,
			AvatarURL:      profile.AvatarURL,
			Verified:       profile.Verified,
			FollowerCount:  0,
			FollowingCount: 0,
			IsFollowing:    false,
			CreatedAt:      profile.CreatedAt,
		},
	}

	return resp, cookie, nil
}

// Login authenticates a user and returns auth tokens.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, *http.Cookie, error) {
	credential, err := s.store.GetCredentialByEmail(ctx, req.Email)
	if err != nil {
		if _, ok := types.IsAppError(err); ok {
			return nil, nil, types.ErrUnauthorized()
		}
		return nil, nil, fmt.Errorf("service: login: get credential: %w", err)
	}

	if err := CheckPassword(credential.PasswordHash, req.Password); err != nil {
		if errors.Is(err, ErrInvalidPassword) {
			return nil, nil, types.ErrUnauthorized()
		}
		return nil, nil, fmt.Errorf("service: login: check password: %w", err)
	}

	profile, err := s.store.GetProfileByID(ctx, credential.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("service: login: get profile: %w", err)
	}

	role, err := s.store.GetRoleByUserID(ctx, credential.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("service: login: get role: %w", err)
	}

	_, cookie, session, err := s.createSession(ctx, credential.UserID)
	if err != nil {
		return nil, nil, err
	}

	if err := s.store.CreateSession(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("service: login: create session: %w", err)
	}

	accessToken, err := s.tokens.GenerateAccessToken(credential.UserID, profile.Username, role, profile.Verified)
	if err != nil {
		return nil, nil, fmt.Errorf("service: login: generate access token: %w", err)
	}

	followerCount, followingCount, err := s.getCounts(ctx, profile.ID)
	if err != nil {
		return nil, nil, err
	}

	resp := &AuthResponse{
		AccessToken: accessToken,
		User:        profileToResponse(profile, followerCount, followingCount, false),
	}

	return resp, cookie, nil
}

// Refresh rotates a refresh token and returns a new access token.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*RefreshResponse, *http.Cookie, error) {
	tokenHash := s.hashToken(refreshToken)

	oldSession, err := s.store.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if _, ok := types.IsAppError(err); ok {
			return nil, nil, types.ErrUnauthorized()
		}
		return nil, nil, fmt.Errorf("service: refresh: get session: %w", err)
	}

	_, cookie, newSession, err := s.createSession(ctx, oldSession.UserID)
	if err != nil {
		return nil, nil, err
	}

	if err := s.store.RotateSession(ctx, oldSession.ID, newSession); err != nil {
		return nil, nil, fmt.Errorf("service: refresh: rotate session: %w", err)
	}

	profile, err := s.store.GetProfileByID(ctx, oldSession.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("service: refresh: get profile: %w", err)
	}

	role, err := s.store.GetRoleByUserID(ctx, oldSession.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("service: refresh: get role: %w", err)
	}

	accessToken, err := s.tokens.GenerateAccessToken(oldSession.UserID, profile.Username, role, profile.Verified)
	if err != nil {
		return nil, nil, fmt.Errorf("service: refresh: generate access token: %w", err)
	}

	return &RefreshResponse{AccessToken: accessToken}, cookie, nil
}

// Logout invalidates all sessions for the authenticated user.
func (s *Service) Logout(ctx context.Context, userID types.ID) error {
	if err := s.store.DeleteSessionsByUserID(ctx, userID); err != nil {
		return fmt.Errorf("service: logout: delete sessions: %w", err)
	}
	return nil
}

// GetProfile returns a profile by username with follower/following counts and follow status.
func (s *Service) GetProfile(ctx context.Context, username string, viewerID *types.ID) (*ProfileResponse, error) {
	profile, err := s.store.GetProfileByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	followerCount, followingCount, err := s.getCounts(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	isFollowing := false
	if viewerID != nil {
		isFollowing, err = s.store.IsFollowing(ctx, *viewerID, profile.ID)
		if err != nil {
			return nil, fmt.Errorf("service: get profile: check following: %w", err)
		}
	}

	resp := profileToResponse(profile, followerCount, followingCount, isFollowing)
	return &resp, nil
}

// UpdateProfile applies partial updates to a user's profile.
func (s *Service) UpdateProfile(ctx context.Context, userID types.ID, req UpdateProfileRequest) (*ProfileResponse, error) {
	profile, err := s.store.UpdateProfile(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	followerCount, followingCount, err := s.getCounts(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	if err := s.bus.Publish(TopicProfileUpdated, ProfileUpdatedEvent{
		UserID:      profile.ID,
		Username:    profile.Username,
		DisplayName: profile.DisplayName,
		AvatarURL:   profile.AvatarURL,
		Verified:    profile.Verified,
	}); err != nil {
		s.logger.Warn("service: update profile: publish event failed", "error", err)
	}

	resp := profileToResponse(profile, followerCount, followingCount, false)
	return &resp, nil
}

// Follow creates a follow relationship from followerID to the user identified by username.
func (s *Service) Follow(ctx context.Context, followerID types.ID, username string) error {
	target, err := s.store.GetProfileByUsername(ctx, username)
	if err != nil {
		return err
	}

	if followerID == target.ID {
		return types.NewValidation("cannot follow yourself")
	}

	if err := s.store.CreateFollow(ctx, followerID, target.ID); err != nil {
		return err
	}

	if err := s.bus.Publish(TopicFollowed, UserFollowedEvent{
		FollowerID:  followerID,
		FollowingID: target.ID,
	}); err != nil {
		s.logger.Warn("service: follow: publish event failed", "error", err)
	}

	return nil
}

// Unfollow removes the follow relationship from followerID to the user identified by username.
func (s *Service) Unfollow(ctx context.Context, followerID types.ID, username string) error {
	target, err := s.store.GetProfileByUsername(ctx, username)
	if err != nil {
		return err
	}

	if err := s.store.DeleteFollow(ctx, followerID, target.ID); err != nil {
		return err
	}

	if err := s.bus.Publish(TopicUnfollowed, UserUnfollowedEvent{
		FollowerID:  followerID,
		FollowingID: target.ID,
	}); err != nil {
		s.logger.Warn("service: unfollow: publish event failed", "error", err)
	}

	return nil
}

// GetFollowers returns a cursor page of profiles that follow the given username.
func (s *Service) GetFollowers(ctx context.Context, username string, params types.CursorParams) (*types.CursorPage[ProfileResponse], error) {
	profile, err := s.store.GetProfileByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	page, err := s.store.GetFollowers(ctx, profile.ID, params)
	if err != nil {
		return nil, err
	}

	resp := profileResponsePage(page)
	return &resp, nil
}

// GetFollowing returns a cursor page of profiles that the given username follows.
func (s *Service) GetFollowing(ctx context.Context, username string, params types.CursorParams) (*types.CursorPage[ProfileResponse], error) {
	profile, err := s.store.GetProfileByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	page, err := s.store.GetFollowing(ctx, profile.ID, params)
	if err != nil {
		return nil, err
	}

	resp := profileResponsePage(page)
	return &resp, nil
}

// SearchUsers returns a cursor page of profiles matching the query string.
func (s *Service) SearchUsers(ctx context.Context, query string, params types.CursorParams) (*types.CursorPage[ProfileResponse], error) {
	page, err := s.store.SearchUsers(ctx, query, params)
	if err != nil {
		return nil, err
	}

	resp := profileResponsePage(page)
	return &resp, nil
}

// --- helpers ---

func (s *Service) createSession(ctx context.Context, userID types.ID) (string, *http.Cookie, *Session, error) {
	refreshToken, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return "", nil, nil, fmt.Errorf("service: create session: generate refresh token: %w", err)
	}

	now := time.Now().UTC()

	sessionID, err := types.NewID()
	if err != nil {
		return "", nil, nil, fmt.Errorf("service: create session: generate session id: %w", err)
	}

	session := &Session{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: s.hashToken(refreshToken),
		ExpiresAt:        now.Add(s.jwt.RefreshTTL),
		CreatedAt:        now,
	}

	cookie := s.makeRefreshCookie(refreshToken, int(s.jwt.RefreshTTL.Seconds()))
	return refreshToken, cookie, session, nil
}

func (s *Service) getCounts(ctx context.Context, userID types.ID) (followerCount, followingCount int, err error) {
	followerCount, err = s.store.GetFollowerCount(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("service: get follower count: %w", err)
	}

	followingCount, err = s.store.GetFollowingCount(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("service: get following count: %w", err)
	}

	return followerCount, followingCount, nil
}

func profileResponsePage(page types.CursorPage[Profile]) types.CursorPage[ProfileResponse] {
	items := make([]ProfileResponse, len(page.Items))
	for i := range page.Items {
		items[i] = profileToResponse(&page.Items[i], 0, 0, false)
	}

	return types.CursorPage[ProfileResponse]{
		Items:      items,
		NextCursor: page.NextCursor,
		HasMore:    page.HasMore,
	}
}

func (s *Service) hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum)
}

func (s *Service) makeRefreshCookie(token string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   s.isProd,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	}
}

// validateRegisterRequest checks required fields before hitting the database.
func validateRegisterRequest(req RegisterRequest) error {
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		return types.NewValidation("email is required and must be a valid address")
	}
	if len(req.Password) < 8 {
		return types.NewValidation("password must be at least 8 characters")
	}
	if req.Username == "" {
		return types.NewValidation("username is required")
	}
	if req.DisplayName == "" {
		return types.NewValidation("display_name is required")
	}
	return nil
}
