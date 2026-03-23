package feed

import (
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
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Routes(r chi.Router, authRequired func(http.Handler) http.Handler) {
	r.With(authRequired).Get("/api/v1/feed", h.handleGetTimeline)
}

// handleGetTimeline godoc
//
//	@Summary      Get timeline
//	@Description  Returns the authenticated user's chronological home timeline. Shows posts from followed users and the user's own posts.
//	@Tags         feed
//	@Security     BearerAuth
//	@Produce      json
//	@Param        cursor query string false "Pagination cursor"
//	@Param        limit  query int    false "Page size (default 20, max 100)"
//	@Success      200 {object} TimelineCursorPage
//	@Failure      401 {object} types.AppError "Not authenticated"
//	@Failure      500 {object} types.AppError "Internal server error"
//	@Router       /feed [get]
func (h *Handler) handleGetTimeline(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	params := types.ParseCursorParams(r)

	page, err := h.service.GetTimeline(r.Context(), userID, &userID, params)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.JSON(w, http.StatusOK, page)
}
