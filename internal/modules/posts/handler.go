package posts

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

// Routes mounts all post endpoints onto the provided router.
func (h *Handler) Routes(r chi.Router, authRequired, authOptional func(http.Handler) http.Handler) {
	r.Route("/api/v1/posts", func(r chi.Router) {
		r.With(authRequired).Post("/", h.handleCreatePost)
		r.With(authOptional).Get("/search", h.handleSearchPosts)
		r.With(authOptional).Get("/{id}", h.handleGetPost)
		r.With(authOptional).Get("/{id}/replies", h.handleGetReplies)
		r.With(authRequired).Delete("/{id}", h.handleDeletePost)
		r.With(authRequired).Post("/{id}/like", h.handleLikePost)
		r.With(authRequired).Delete("/{id}/like", h.handleUnlikePost)
		r.With(authRequired).Post("/{id}/repost", h.handleRepostPost)
		r.With(authRequired).Delete("/{id}/repost", h.handleUndoRepost)
	})

	r.With(authOptional).Get("/api/v1/users/{username}/posts", h.handleGetUserPosts)
}

// handleCreatePost creates a new post.
//
// @Summary      Create a post
// @Description  Create a new post with text and optional image attachments. Supports replies (parent_id) and reposts (repost_of_id). Automatically extracts hashtags and fetches link previews.
// @Tags         posts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body CreatePostRequest true "Post content"
// @Success      201 {object} PostResponse
// @Failure      401 {object} types.AppError "Not authenticated"
// @Failure      422 {object} types.AppError "Validation error"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /posts [post]
func (h *Handler) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpkit.Error(w, types.NewValidation("invalid request body"))
		return
	}

	resp, err := h.service.CreatePost(r.Context(), userID, req)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.JSON(w, http.StatusCreated, resp)
}

// handleGetPost retrieves a single post by ID.
//
// @Summary      Get a post
// @Description  Retrieve a post by ID. If authenticated, includes whether the viewer has liked or reposted this post.
// @Tags         posts
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Post ID"
// @Success      200 {object} PostResponse
// @Failure      404 {object} types.AppError "Post not found"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /posts/{id} [get]
func (h *Handler) handleGetPost(w http.ResponseWriter, r *http.Request) {
	postID, err := types.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httpkit.Error(w, types.NewValidation("invalid post id"))
		return
	}

	var viewerID *types.ID
	if id, ok := httpkit.UserIDFrom(r.Context()); ok {
		viewerID = &id
	}

	resp, err := h.service.GetPost(r.Context(), postID, viewerID)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.JSON(w, http.StatusOK, resp)
}

// handleDeletePost deletes a post owned by the authenticated user.
//
// @Summary      Delete a post
// @Description  Delete a post. Only the post author can delete their own post.
// @Tags         posts
// @Security     BearerAuth
// @Param        id path string true "Post ID"
// @Success      204
// @Failure      401 {object} types.AppError "Not authenticated"
// @Failure      403 {object} types.AppError "Not the post author"
// @Failure      404 {object} types.AppError "Post not found"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /posts/{id} [delete]
func (h *Handler) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	postID, err := types.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httpkit.Error(w, types.NewValidation("invalid post id"))
		return
	}

	if err := h.service.DeletePost(r.Context(), postID, userID); err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.NoContent(w)
}

// handleGetReplies returns a paginated list of replies to a post.
//
// @Summary      Get replies
// @Description  Retrieve replies to a post, cursor-paginated in chronological order.
// @Tags         posts
// @Security     BearerAuth
// @Produce      json
// @Param        id     path  string true  "Parent post ID"
// @Param        cursor query string false "Pagination cursor"
// @Param        limit  query int    false "Page size (default 20, max 100)"
// @Success      200 {object} PostCursorPage
// @Failure      404 {object} types.AppError "Post not found"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /posts/{id}/replies [get]
func (h *Handler) handleGetReplies(w http.ResponseWriter, r *http.Request) {
	parentID, err := types.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httpkit.Error(w, types.NewValidation("invalid post id"))
		return
	}

	var viewerID *types.ID
	if id, ok := httpkit.UserIDFrom(r.Context()); ok {
		viewerID = &id
	}

	params := types.ParseCursorParams(r)

	page, err := h.service.GetReplies(r.Context(), parentID, viewerID, params)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.CursorResponse(w, *page)
}

// handleLikePost adds a like to a post.
//
// @Summary      Like a post
// @Description  Add a like to a post. Returns the updated like count.
// @Tags         posts
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Post ID"
// @Success      200 {object} LikeResponse
// @Failure      401 {object} types.AppError "Not authenticated"
// @Failure      404 {object} types.AppError "Post not found"
// @Failure      409 {object} types.AppError "Already liked"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /posts/{id}/like [post]
func (h *Handler) handleLikePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	postID, err := types.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httpkit.Error(w, types.NewValidation("invalid post id"))
		return
	}

	resp, err := h.service.LikePost(r.Context(), postID, userID)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.JSON(w, http.StatusOK, resp)
}

// handleUnlikePost removes a like from a post.
//
// @Summary      Unlike a post
// @Description  Remove a like from a post. Returns the updated like count.
// @Tags         posts
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Post ID"
// @Success      200 {object} LikeResponse
// @Failure      401 {object} types.AppError "Not authenticated"
// @Failure      404 {object} types.AppError "Like not found"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /posts/{id}/like [delete]
func (h *Handler) handleUnlikePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	postID, err := types.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httpkit.Error(w, types.NewValidation("invalid post id"))
		return
	}

	resp, err := h.service.UnlikePost(r.Context(), postID, userID)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.JSON(w, http.StatusOK, resp)
}

// handleRepostPost creates a repost of a post.
//
// @Summary      Repost a post
// @Description  Repost an existing post. Returns the updated repost count.
// @Tags         posts
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Post ID to repost"
// @Success      200 {object} RepostResponse
// @Failure      401 {object} types.AppError "Not authenticated"
// @Failure      404 {object} types.AppError "Post not found"
// @Failure      409 {object} types.AppError "Already reposted"
// @Failure      422 {object} types.AppError "Cannot repost a repost"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /posts/{id}/repost [post]
func (h *Handler) handleRepostPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	postID, err := types.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httpkit.Error(w, types.NewValidation("invalid post id"))
		return
	}

	resp, err := h.service.RepostPost(r.Context(), postID, userID)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.JSON(w, http.StatusOK, resp)
}

// handleUndoRepost removes a repost.
//
// @Summary      Undo repost
// @Description  Remove a repost of a post. Returns the updated repost count.
// @Tags         posts
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Original post ID"
// @Success      200 {object} RepostResponse
// @Failure      401 {object} types.AppError "Not authenticated"
// @Failure      404 {object} types.AppError "Repost not found"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /posts/{id}/repost [delete]
func (h *Handler) handleUndoRepost(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	postID, err := types.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httpkit.Error(w, types.NewValidation("invalid post id"))
		return
	}

	resp, err := h.service.UndoRepost(r.Context(), postID, userID)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.JSON(w, http.StatusOK, resp)
}

// handleSearchPosts searches posts by body text or hashtag.
//
// @Summary      Search posts
// @Description  Full-text search over post body content, cursor-paginated.
// @Tags         posts
// @Produce      json
// @Param        q      query string true  "Search query"
// @Param        cursor query string false "Pagination cursor"
// @Param        limit  query int    false "Page size (default 20, max 100)"
// @Success      200 {object} PostCursorPage
// @Failure      422 {object} types.AppError "Missing query parameter"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /posts/search [get]
func (h *Handler) handleSearchPosts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		httpkit.Error(w, types.NewValidation("query parameter 'q' is required"))
		return
	}

	var viewerID *types.ID
	if id, ok := httpkit.UserIDFrom(r.Context()); ok {
		viewerID = &id
	}

	params := types.ParseCursorParams(r)

	page, err := h.service.SearchPosts(r.Context(), query, viewerID, params)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.CursorResponse(w, *page)
}

// handleGetUserPosts returns a paginated list of posts by a user.
//
// @Summary      Get user posts
// @Description  Retrieve root posts by a user, cursor-paginated in reverse chronological order. If authenticated, includes liked/reposted status.
// @Tags         posts
// @Security     BearerAuth
// @Produce      json
// @Param        username path  string true  "Username"
// @Param        cursor   query string false "Pagination cursor"
// @Param        limit    query int    false "Page size (default 20, max 100)"
// @Success      200 {object} PostCursorPage
// @Failure      404 {object} types.AppError "User not found"
// @Failure      500 {object} types.AppError "Internal server error"
// @Router       /users/{username}/posts [get]
func (h *Handler) handleGetUserPosts(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	var viewerID *types.ID
	if id, ok := httpkit.UserIDFrom(r.Context()); ok {
		viewerID = &id
	}

	params := types.ParseCursorParams(r)

	page, err := h.service.GetUserPosts(r.Context(), username, viewerID, params)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.CursorResponse(w, *page)
}
