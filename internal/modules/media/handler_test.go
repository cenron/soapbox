package media

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/radni/soapbox/internal/core/config"
	"github.com/radni/soapbox/internal/core/httpkit"
	"github.com/radni/soapbox/internal/core/testutil"
	"github.com/radni/soapbox/internal/core/types"
	"github.com/stretchr/testify/assert"
)

func fakeAuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := types.NewID()
		ctx := httpkit.WithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func noAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func setupRouter(h *Handler, auth func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	h.Routes(r, auth)
	return r
}

func newTestHandler() *Handler {
	s3Client := &mockS3{presignURL: "http://example.com/presigned"}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := config.S3Config{Endpoint: "http://localhost:9000", Bucket: "soapbox"}
	svc := NewService(nil, s3Client, cfg, logger)

	return NewHandler(svc, logger)
}

func TestHandleUploadURL_NoAuth(t *testing.T) {
	h := newTestHandler()
	router := setupRouter(h, noAuth)

	rec := testutil.DoRequest(t, router, "POST", "/api/v1/media/upload-url",
		`{"content_type":"image/jpeg","filename":"photo.jpg"}`)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleUploadURL_InvalidBody(t *testing.T) {
	h := newTestHandler()
	router := setupRouter(h, fakeAuthRequired)

	rec := testutil.DoRequest(t, router, "POST", "/api/v1/media/upload-url", "not json")

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestHandleUploadURL_InvalidContentType(t *testing.T) {
	h := newTestHandler()
	router := setupRouter(h, fakeAuthRequired)

	rec := testutil.DoRequest(t, router, "POST", "/api/v1/media/upload-url",
		`{"content_type":"application/pdf","filename":"doc.pdf"}`)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestHandleConfirmUpload_InvalidID(t *testing.T) {
	h := newTestHandler()
	router := setupRouter(h, fakeAuthRequired)

	rec := testutil.DoRequest(t, router, "POST", "/api/v1/media/not-a-uuid/confirm", "")

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}
