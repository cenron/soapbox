package testutil

import (
	"context"
	"testing"

	"github.com/radni/soapbox/internal/core/config"
	"github.com/radni/soapbox/internal/core/db"
	"github.com/stretchr/testify/require"
)

func NewTestDB(t *testing.T) *db.DB {
	t.Helper()

	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "soapbox",
		User:     "soapbox",
		Password: "soapbox",
		SSLMode:  "disable",
	}

	database, err := db.New(context.Background(), cfg)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = database.Close()
	})

	return database
}

func CleanSchema(t *testing.T, database *db.DB, schema string) {
	t.Helper()

	_, err := database.Conn.ExecContext(context.Background(),
		"DROP SCHEMA IF EXISTS "+schema+" CASCADE")
	require.NoError(t, err)
}
