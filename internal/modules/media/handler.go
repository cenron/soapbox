package media

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

// Routes mounts all media endpoints onto the provided router.
func (h *Handler) Routes(r chi.Router, authRequired func(http.Handler) http.Handler) {
	r.Route("/api/v1/media", func(r chi.Router) {
		r.Use(authRequired)
		r.Post("/upload-url", h.handleUploadURL)
		r.Post("/{id}/confirm", h.handleConfirmUpload)
	})
}

// handleUploadURL generates a presigned S3 upload URL.
//
// @Summary      Request a presigned upload URL
// @Description  Returns a presigned S3 URL for direct file upload. The client uploads the file directly to S3 using this URL, then confirms the upload.
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        body body UploadURLRequest true "Upload details"
// @Success      201 {object} UploadURLResponse
// @Failure      401 {object} types.AppError "Unauthorized"
// @Failure      422 {object} types.AppError "Unsupported content type"
// @Failure      500 {object} types.AppError "Internal server error"
// @Security     BearerAuth
// @Router       /media/upload-url [post]
func (h *Handler) handleUploadURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	var req UploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpkit.Error(w, types.NewValidation("invalid request body"))
		return
	}

	resp, err := h.service.RequestUpload(r.Context(), userID, req)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.JSON(w, http.StatusCreated, resp)
}

// handleConfirmUpload marks an upload as ready after the client finishes uploading to S3.
//
// @Summary      Confirm an upload
// @Description  Marks a pending upload as ready. Call this after successfully uploading the file to the presigned URL.
// @Tags         media
// @Produce      json
// @Param        id path string true "Upload ID"
// @Success      200 {object} UploadResponse
// @Failure      401 {object} types.AppError "Unauthorized"
// @Failure      403 {object} types.AppError "Not the upload owner"
// @Failure      404 {object} types.AppError "Upload not found"
// @Failure      409 {object} types.AppError "Already confirmed"
// @Failure      500 {object} types.AppError "Internal server error"
// @Security     BearerAuth
// @Router       /media/{id}/confirm [post]
func (h *Handler) handleConfirmUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpkit.UserIDFrom(r.Context())
	if !ok {
		httpkit.Error(w, types.ErrUnauthorized())
		return
	}

	uploadID, err := types.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httpkit.Error(w, types.NewValidation("invalid upload id"))
		return
	}

	resp, err := h.service.ConfirmUpload(r.Context(), userID, uploadID)
	if err != nil {
		httpkit.Error(w, err)
		return
	}

	httpkit.JSON(w, http.StatusOK, resp)
}
