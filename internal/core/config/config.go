package config

import (
	"strconv"
	"time"

	env "github.com/caarlos0/env/v11"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	S3       S3Config
	Mail     MailConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Host string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port int    `env:"SERVER_PORT" envDefault:"8080"`
	Env  string `env:"APP_ENV"     envDefault:"development"`
}

func (c ServerConfig) IsProd() bool {
	return c.Env == "production"
}

type DatabaseConfig struct {
	Host     string `env:"DB_HOST"     envDefault:"localhost"`
	Port     int    `env:"DB_PORT"     envDefault:"5432"`
	Name     string `env:"DB_NAME"     envDefault:"soapbox"`
	User     string `env:"DB_USER"     envDefault:"soapbox"`
	Password string `env:"DB_PASSWORD" envDefault:"soapbox"`
	SSLMode  string `env:"DB_SSLMODE"  envDefault:"disable"`
}

func (c DatabaseConfig) DSN() string {
	return "postgres://" + c.User + ":" + c.Password +
		"@" + c.Host + ":" + strconv.Itoa(c.Port) +
		"/" + c.Name + "?sslmode=" + c.SSLMode
}

type S3Config struct {
	Endpoint  string `env:"S3_ENDPOINT"   envDefault:"http://localhost:9000"`
	Bucket    string `env:"S3_BUCKET"     envDefault:"soapbox"`
	AccessKey string `env:"S3_ACCESS_KEY" envDefault:"minioadmin"`
	SecretKey string `env:"S3_SECRET_KEY" envDefault:"minioadmin"`
	Region    string `env:"S3_REGION"     envDefault:"us-east-1"`
}

type MailConfig struct {
	Host string `env:"MAIL_HOST" envDefault:"localhost"`
	Port int    `env:"MAIL_PORT" envDefault:"1025"`
	From string `env:"MAIL_FROM" envDefault:"noreply@soapbox.dev"`
}

type JWTConfig struct {
	Secret     string        `env:"JWT_SECRET"      envDefault:"dev-secret-change-in-production"`
	AccessTTL  time.Duration `env:"JWT_ACCESS_TTL"  envDefault:"15m"`
	RefreshTTL time.Duration `env:"JWT_REFRESH_TTL" envDefault:"168h"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
