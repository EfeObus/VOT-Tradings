// Package db wraps the PostgreSQL connection pool used as VOT Tradings'
// primary relational ledger (accounts, orders, positions, predictions).
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"vot-tradings/internal/config"
)

// Connect opens a pooled connection to Postgres and applies schema.sql so a
// fresh local database is ready to use without a separate migration step.
func Connect(ctx context.Context, cfg config.PostgresConfig, schema string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("db: connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	if schema != "" {
		if _, err := pool.Exec(ctx, schema); err != nil {
			pool.Close()
			return nil, fmt.Errorf("db: apply schema: %w", err)
		}
	}

	return pool, nil
}
