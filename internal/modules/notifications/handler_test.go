package notifications

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
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
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := &Service{logger: logger}
	return NewHandler(svc, logger)
}

func TestHandleList_NoAuth(t *testing.T) {
	h := newTestHandler()
	r := setupRouter(h, noAuth)

	rec := testutil.DoRequest(t, r, "GET", "/api/v1/notifications", "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleMarkRead_NoAuth(t *testing.T) {
	h := newTestHandler()
	r := setupRouter(h, noAuth)

	rec := testutil.DoRequest(t, r, "PUT", "/api/v1/notifications/some-id/read", "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleMarkAllRead_NoAuth(t *testing.T) {
	h := newTestHandler()
	r := setupRouter(h, noAuth)

	rec := testutil.DoRequest(t, r, "PUT", "/api/v1/notifications/read-all", "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleMarkRead_InvalidID(t *testing.T) {
	h := newTestHandler()
	r := setupRouter(h, fakeAuthRequired)

	rec := testutil.DoRequest(t, r, "PUT", "/api/v1/notifications/not-a-uuid/read", "")
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}
