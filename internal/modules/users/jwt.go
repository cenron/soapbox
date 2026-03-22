package users

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/radni/soapbox/internal/core/config"
	"github.com/radni/soapbox/internal/core/types"
)

type Claims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	Role     string `json:"role"`
	Verified bool   `json:"verified"`
}

type TokenService struct {
	config config.JWTConfig
}

func NewTokenService(cfg config.JWTConfig) *TokenService {
	return &TokenService{config: cfg}
}

func (t *TokenService) GenerateAccessToken(userID types.ID, username, role string, verified bool) (string, error) {
	now := time.Now()

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.config.AccessTTL)),
		},
		Username: username,
		Role:     role,
		Verified: verified,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(t.config.Secret))
	if err != nil {
		return "", fmt.Errorf("jwt: sign token: %w", err)
	}

	return signed, nil
}

func (t *TokenService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("jwt: unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.config.Secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt: parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("jwt: invalid token claims")
	}

	return claims, nil
}

func (t *TokenService) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("jwt: generate refresh token: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}
