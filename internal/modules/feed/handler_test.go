package feed

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/radni/soapbox/internal/core/testutil"
	"github.com/stretchr/testify/assert"
)

func noAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func TestHandleGetTimeline_NoAuth(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := &Service{logger: logger}
	h := NewHandler(svc, logger)

	r := chi.NewRouter()
	h.Routes(r, noAuth)

	rec := testutil.DoRequest(t, r, "GET", "/api/v1/feed", "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
