package httpkit

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/radni/soapbox/internal/core/config"
)

type Server struct {
	Router chi.Router
	srv    *http.Server
	logger *slog.Logger
}

func NewServer(cfg config.ServerConfig, logger *slog.Logger) *Server {
	r := chi.NewRouter()

	r.Use(RequestID)
	r.Use(CORS)
	r.Use(Recoverer(logger))
	r.Use(Logger(logger))

	s := &Server{
		Router: r,
		logger: logger,
		srv: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler: r,
		},
	}

	r.Get("/healthz", s.healthz)

	return s
}

func (s *Server) Start() error {
	s.logger.Info("server starting", "addr", s.srv.Addr)
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("server shutting down")
	return s.srv.Shutdown(ctx)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
