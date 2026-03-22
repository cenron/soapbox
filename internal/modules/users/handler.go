package users

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/radni/soapbox/internal/core/httpkit"
	"github.com/radni/soapbox/internal/core/types"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Routes mounts all users and auth endpoints onto the provided router.
func (h *Handler) Routes(r chi.Router, authRequired, authOptional func(http.Handler) http.Handler) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.handleRegister)
			r.Post("/login", h.handleLogin)
			r.Post("/refresh", h.handleRefresh)
			r.With(authRequired).Post("/logout", h.handleLogout)
		})

		r.Route("/users", func(r chi.Router) {
			// /search must precede /{username} to avoid being captured as a username param.
			r.With(authOptional).Get("/search", h.handleSearchUsers)
			r.With(authOptional).Get("/{username}", h.handleGetProfile)
			r.With(authOptional).Get("/{username}/followers", h.handleGetFollowers)
			r.With(authOptional).Get("/{username}/following", h.handleGetFollowing)

			r.Group(func(r chi.Router) {
				r.Use(authRequired)
				r.Put("/me", h.handleUpdateProfile)
				r.Post("/{username}/follow", h.handleFollow)
				r.Delete("/{username}/follow", h.handleUnfollow)
			})
		})
	})
}

// handleRegister creates a new user account.
//
// @Summary      Register a new user
// @Description  Create an account with email, password, and username. Returns an access token and sets a refresh token cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RegisterRequest true "Registration details"
// @Success      201 {object} AuthResponse
// @Failure      409 {object} types.AppError "Email or username already taken"
// @Failure      422 {object} types.AppError "Validation error"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /auth/register [post]
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpkit.Error(w, types.NewValidation("invalid request body"))
		return
	}

	resp, cookie, err := h.service.Register(r.Context(), req)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	http.SetCookie(w, cookie)
	httpkit.JSON(w, http.StatusCreated, resp)
}

// handleLogin authenticates an existing user.
//
// @Summary      Log in
// @Description  Authenticate with email and password. Returns an access token and sets a refresh token cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body LoginRequest true "Login credentials"
// @Success      200 {object} AuthResponse
// @Failure      401 {object} types.AppError "Invalid credentials"
// @Failure      422 {object} types.AppError "Validation error"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /auth/login [post]
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpkit.Error(w, types.NewValidation("invalid request body"))
		return
	}

	resp, cookie, err := h.service.Login(r.Context(), req)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	http.SetCookie(w, cookie)
	httpkit.JSON(w, http.StatusOK, resp)
}

// handleRefresh rotates the refresh token and issues a new access token.
//
// @Summary      Refresh access token
// @Description  Exchange a valid refresh token (from cookie) for a new access token and rotated refresh token cookie.
// @Tags         auth
// @Produce      json
// @Success      200 {object} RefreshResponse
// @Failure      401 {object} types.AppError "Missing or invalid refresh token"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /auth/refresh [post]
func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	resp, newCookie, err := h.service.Refresh(r.Context(), cookie.Value)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	http.SetCookie(w, newCookie)
	httpkit.JSON(w, http.StatusOK, resp)
}

// handleLogout invalidates the current session.
//
// @Summary      Log out
// @Description  Invalidate the refresh token session. Requires a valid access token.
// @Tags         auth
// @Security     BearerAuth
// @Success      204
// @Failure      401 {object} types.AppError "Not authenticated"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /auth/logout [post]
func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	if err := h.service.Logout(r.Context(), cookie.Value); err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.NoContent(w)
}

// handleGetProfile returns a user's public profile.
//
// @Summary      Get user profile
// @Description  Retrieve a user's profile by username. If authenticated, includes whether the viewer follows this user.
// @Tags         users
// @Produce      json
// @Param        username path string true "Username"
// @Success      200 {object} ProfileResponse
// @Failure      404 {object} types.AppError "User not found"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /users/{username} [get]
func (h *Handler) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	var viewerID *types.ID
	if id, ok := httpkit.UserIDFrom(r.Context()); ok {
		viewerID = &id
	}

	resp, err := h.service.GetProfile(r.Context(), username, viewerID)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.JSON(w, http.StatusOK, resp)
}

// handleUpdateProfile updates the authenticated user's profile.
//
// @Summary      Update own profile
// @Description  Update display name, bio, or avatar URL for the authenticated user.
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body UpdateProfileRequest true "Profile fields to update"
// @Success      200 {object} ProfileResponse
// @Failure      401 {object} types.AppError "Not authenticated"
// @Failure      422 {object} types.AppError "Validation error"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /users/me [put]
func (h *Handler) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpkit.Error(w, types.NewValidation("invalid request body"))
		return
	}

	resp, err := h.service.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.JSON(w, http.StatusOK, resp)
}

// handleFollow follows a user.
//
// @Summary      Follow a user
// @Description  Follow the user identified by username. Requires authentication.
// @Tags         users
// @Security     BearerAuth
// @Param        username path string true "Username to follow"
// @Success      204
// @Failure      401 {object} types.AppError "Not authenticated"
// @Failure      404 {object} types.AppError "User not found"
// @Failure      409 {object} types.AppError "Already following"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /users/{username}/follow [post]
func (h *Handler) handleFollow(w http.ResponseWriter, r *http.Request) {
	followerID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	username := chi.URLParam(r, "username")

	if err := h.service.Follow(r.Context(), followerID, username); err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.NoContent(w)
}

// handleUnfollow unfollows a user.
//
// @Summary      Unfollow a user
// @Description  Remove a follow relationship with the user identified by username. Requires authentication.
// @Tags         users
// @Security     BearerAuth
// @Param        username path string true "Username to unfollow"
// @Success      204
// @Failure      401 {object} types.AppError "Not authenticated"
// @Failure      404 {object} types.AppError "User not found"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /users/{username}/follow [delete]
func (h *Handler) handleUnfollow(w http.ResponseWriter, r *http.Request) {
	followerID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	username := chi.URLParam(r, "username")

	if err := h.service.Unfollow(r.Context(), followerID, username); err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.NoContent(w)
}

// handleGetFollowers returns a paginated list of a user's followers.
//
// @Summary      List followers
// @Description  Retrieve a cursor-paginated list of users who follow the given username.
// @Tags         users
// @Produce      json
// @Param        username path  string true  "Username"
// @Param        cursor   query string false "Pagination cursor"
// @Param        limit    query int    false "Page size (default 20, max 100)"
// @Success      200 {object} ProfileCursorPage
// @Failure      404 {object} types.AppError "User not found"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /users/{username}/followers [get]
func (h *Handler) handleGetFollowers(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	params := types.ParseCursorParams(r)

	page, err := h.service.GetFollowers(r.Context(), username, params)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.CursorResponse(w, *page)
}

// handleGetFollowing returns a paginated list of users that a user follows.
//
// @Summary      List following
// @Description  Retrieve a cursor-paginated list of users that the given username follows.
// @Tags         users
// @Produce      json
// @Param        username path  string true  "Username"
// @Param        cursor   query string false "Pagination cursor"
// @Param        limit    query int    false "Page size (default 20, max 100)"
// @Success      200 {object} ProfileCursorPage
// @Failure      404 {object} types.AppError "User not found"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /users/{username}/following [get]
func (h *Handler) handleGetFollowing(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	params := types.ParseCursorParams(r)

	page, err := h.service.GetFollowing(r.Context(), username, params)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.CursorResponse(w, *page)
}

// handleSearchUsers searches for users by username or display name.
//
// @Summary      Search users
// @Description  Full-text search over usernames and display names, cursor-paginated.
// @Tags         users
// @Produce      json
// @Param        q      query string true  "Search query"
// @Param        cursor query string false "Pagination cursor"
// @Param        limit  query int    false "Page size (default 20, max 100)"
// @Success      200 {object} ProfileCursorPage
// @Failure      422 {object} types.AppError "Missing query parameter"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /users/search [get]
func (h *Handler) handleSearchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		httpkit.Error(w, types.NewValidation("query parameter 'q' is required"))
		return
	}

	params := types.ParseCursorParams(r)

	page, err := h.service.SearchUsers(r.Context(), query, params)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.CursorResponse(w, *page)
}
