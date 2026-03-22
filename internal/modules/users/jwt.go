package users

import "github.com/radni/soapbox/internal/core/config"

type TokenService struct {
	config config.JWTConfig
}

func NewTokenService(cfg config.JWTConfig) *TokenService {
	return &TokenService{config: cfg}
}
