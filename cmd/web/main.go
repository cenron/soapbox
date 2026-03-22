package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/radni/soapbox/api/swagger"
	"github.com/radni/soapbox/internal/core/bus"
	"github.com/radni/soapbox/internal/core/cache"
	"github.com/radni/soapbox/internal/core/config"
	"github.com/radni/soapbox/internal/core/db"
	"github.com/radni/soapbox/internal/core/httpkit"
	"github.com/radni/soapbox/internal/core/registry"
	"github.com/radni/soapbox/web"
)

// @title           Soapbox API
// @version         1.0
// @description     Chronological microblogging platform — pre-2022 Twitter clone.

// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http

// @accept          json
// @produce         json

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.Server)

	database, err := db.New(ctx, cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database — is Postgres running? Try: make docker-up",
			"host", cfg.Database.Host,
			"port", cfg.Database.Port,
			"error", err,
		)
		os.Exit(1)
	}

	eventBus := bus.NewMemoryBus(logger)
	reg := registry.NewMemoryRegistry()
	appCache := cache.NewMemoryCache()
	server := httpkit.NewServer(cfg.Server, logger)

	server.Router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	_ = database
	_ = eventBus
	_ = reg
	_ = appCache

	// Future: load modules here
	// auth.Load(core.ModuleDeps{...})
	// users.Load(core.ModuleDeps{...})

	// Serve embedded SPA — API routes match first, unmatched routes serve the frontend.
	// In dev, run `make web-build` first or use Vite dev server at :5173 instead.
	staticFS, fsErr := fs.Sub(web.StaticFiles, "dist")
	if fsErr != nil {
		logger.Error("failed to create sub filesystem for SPA", "error", fsErr)
		os.Exit(1)
	}
	server.Router.NotFound(httpkit.SPAHandler(staticFS))

	errCh := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	addr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	logger.Info("soapbox started", "addr", addr)
	logger.Info("swagger UI", "url", "http://localhost:"+strconv.Itoa(cfg.Server.Port)+"/swagger/index.html")
	logger.Info("web app", "url", "http://localhost:"+strconv.Itoa(cfg.Server.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		logger.Error("server error", "error", err)
	case <-quit:
	}

	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}

	if err := database.Close(); err != nil {
		logger.Error("database close error", "error", err)
	}
	logger.Info("shutdown complete")
}

func setupLogger(cfg config.ServerConfig) *slog.Logger {
	var handler slog.Handler

	if cfg.IsProd() {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
