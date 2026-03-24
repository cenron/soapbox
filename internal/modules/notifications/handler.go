package notifications

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
	r.With(authRequired).Get("/api/v1/notifications", h.handleList)
	r.With(authRequired).Put("/api/v1/notifications/{id}/read", h.handleMarkRead)
	r.With(authRequired).Put("/api/v1/notifications/read-all", h.handleMarkAllRead)
}

// handleList godoc
//
//	@Summary      List notifications
//	@Description  Returns the authenticated user's notifications, newest first, with cursor-based pagination.
//	@Tags         notifications
//	@Security     BearerAuth
//	@Produce      json
//	@Param        cursor query string false "Pagination cursor"
//	@Param        limit  query int    false "Page size (default 20, max 100)"
//	@Success      200 {object} NotificationCursorPage
//	@Failure      401 {object} types.AppError "Not authenticated"
//	@Failure      500 {object} types.AppError "Internal server error"
//	@Router       /notifications [get]
func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	params := types.ParseCursorParams(r)

	page, err := h.service.ListNotifications(r.Context(), userID, params)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.JSON(w, http.StatusOK, page)
}

// handleMarkRead godoc
//
//	@Summary      Mark notification as read
//	@Description  Marks a single notification as read. The notification must belong to the authenticated user.
//	@Tags         notifications
//	@Security     BearerAuth
//	@Param        id path string true "Notification ID"
//	@Success      204 "Notification marked as read"
//	@Failure      401 {object} types.AppError "Not authenticated"
//	@Failure      404 {object} types.AppError "Notification not found"
//	@Failure      500 {object} types.AppError "Internal server error"
//	@Router       /notifications/{id}/read [put]
func (h *Handler) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := types.ParseID(idStr)
	if err != nil {
		httpkit.Error(w, types.NewValidation("invalid notification id"))
		return
	}

	if err := h.service.MarkRead(r.Context(), id, userID); err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.NoContent(w)
}

// handleMarkAllRead godoc
//
//	@Summary      Mark all notifications as read
//	@Description  Marks all of the authenticated user's unread notifications as read.
//	@Tags         notifications
//	@Security     BearerAuth
//	@Success      204 "All notifications marked as read"
//	@Failure      401 {object} types.AppError "Not authenticated"
//	@Failure      500 {object} types.AppError "Internal server error"
//	@Router       /notifications/read-all [put]
func (h *Handler) handleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	if err := h.service.MarkAllRead(r.Context(), userID); err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.NoContent(w)
}
