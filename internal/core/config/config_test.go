package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "development", cfg.Server.Env)
	assert.False(t, cfg.Server.IsProd())

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "soapbox", cfg.Database.Name)
	assert.Contains(t, cfg.Database.DSN(), "postgres://soapbox:soapbox@localhost:5432/soapbox")

	assert.Equal(t, "http://localhost:9000", cfg.S3.Endpoint)
	assert.Equal(t, "soapbox", cfg.S3.Bucket)

	assert.Equal(t, "localhost", cfg.Mail.Host)
	assert.Equal(t, 1025, cfg.Mail.Port)

	assert.Equal(t, "dev-secret-change-in-production", cfg.JWT.Secret)
	assert.Equal(t, 15*time.Minute, cfg.JWT.AccessTTL)
	assert.Equal(t, 168*time.Hour, cfg.JWT.RefreshTTL)
}

func TestServerConfig_IsProd(t *testing.T) {
	cfg := ServerConfig{Env: "production"}
	assert.True(t, cfg.IsProd())

	cfg.Env = "development"
	assert.False(t, cfg.IsProd())
}

func TestConfig_Validate_RejectsDefaultSecretInProd(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Env: "production"},
		JWT:    JWTConfig{Secret: "dev-secret-change-in-production"},
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET must be changed")
}

func TestConfig_Validate_AllowsDefaultSecretInDev(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Env: "development"},
		JWT:    JWTConfig{Secret: "dev-secret-change-in-production"},
	}
	err := cfg.Validate()
	require.NoError(t, err)
}

func TestConfig_Validate_AllowsCustomSecretInProd(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Env: "production"},
		JWT:    JWTConfig{Secret: "my-real-production-secret"},
	}
	err := cfg.Validate()
	require.NoError(t, err)
}
