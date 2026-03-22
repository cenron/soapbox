package core

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/radni/soapbox/internal/core/bus"
	"github.com/radni/soapbox/internal/core/cache"
	"github.com/radni/soapbox/internal/core/config"
	"github.com/radni/soapbox/internal/core/db"
	"github.com/radni/soapbox/internal/core/registry"
)

type ModuleDeps struct {
	DB       *db.DB
	Bus      bus.Bus
	Registry registry.Registry
	Cache    cache.Cache
	Router   chi.Router
	Logger   *slog.Logger
	Config   *config.Config
}
