package db

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/radni/soapbox/internal/core/config"
)

type DB struct {
	Conn *sqlx.DB
}

func New(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	conn, err := sqlx.ConnectContext(ctx, "pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("db: connect: %w", err)
	}

	return &DB{Conn: conn}, nil
}

func (db *DB) Close() error {
	return db.Conn.Close()
}

func (db *DB) Health(ctx context.Context) error {
	return db.Conn.PingContext(ctx)
}
