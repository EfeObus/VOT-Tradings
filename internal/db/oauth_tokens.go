package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OAuthTokenStore persists rotating broker OAuth refresh tokens (see
// broker_oauth_tokens in schema.sql) so a process restart doesn't strand a
// broker integration whose refresh tokens are single-use, such as
// Questrade's.
type OAuthTokenStore struct {
	pool *pgxpool.Pool
}

func NewOAuthTokenStore(pool *pgxpool.Pool) *OAuthTokenStore {
	return &OAuthTokenStore{pool: pool}
}

// LoadRefreshToken returns the last-persisted refresh token for broker, or
// "" if none has been stored yet.
func (s *OAuthTokenStore) LoadRefreshToken(ctx context.Context, broker string) (string, error) {
	var token string
	err := s.pool.QueryRow(ctx,
		`SELECT refresh_token FROM broker_oauth_tokens WHERE broker = $1`, broker,
	).Scan(&token)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return token, err
}

// SaveRefreshToken upserts the current refresh token for broker.
func (s *OAuthTokenStore) SaveRefreshToken(ctx context.Context, broker, token string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO broker_oauth_tokens (broker, refresh_token, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (broker) DO UPDATE
			SET refresh_token = EXCLUDED.refresh_token, updated_at = now()
	`, broker, token)
	return err
}
