package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func (db *DB) WithTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := db.Conn.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db: begin tx: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("db: rollback failed: %w (original: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("db: commit tx: %w", err)
	}

	return nil
}
