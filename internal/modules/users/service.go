package users

import (
	"log/slog"

	"github.com/radni/soapbox/internal/core/bus"
	"github.com/radni/soapbox/internal/core/config"
)

type Service struct {
	store  *Store
	tokens *TokenService
	bus    bus.Bus
	logger *slog.Logger
	config *config.Config
}

func NewService(store *Store, tokens *TokenService, b bus.Bus, logger *slog.Logger, cfg *config.Config) *Service {
	return &Service{
		store:  store,
		tokens: tokens,
		bus:    b,
		logger: logger,
		config: cfg,
	}
}
