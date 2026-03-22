package db

import (
	"context"
	"fmt"
	"io/fs"
	"regexp"

	"github.com/pressly/goose/v3"
)

var validSchema = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

func (db *DB) Migrate(ctx context.Context, schema string, migrations fs.FS) error {
	if !validSchema.MatchString(schema) {
		return fmt.Errorf("db: invalid schema name %q", schema)
	}

	sqlDB := db.Conn.DB

	if _, err := sqlDB.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)); err != nil {
		return fmt.Errorf("db: create schema %s: %w", schema, err)
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		sqlDB,
		migrations,
		goose.WithTableName(schema+".goose_db_version"),
	)
	if err != nil {
		return fmt.Errorf("db: migrate provider: %w", err)
	}
	defer provider.Close()

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("db: migrate up %s: %w", schema, err)
	}

	return nil
}
