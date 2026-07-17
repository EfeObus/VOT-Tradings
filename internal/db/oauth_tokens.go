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
// Questrade's. Scoped per-user via ForUser, since each user authorizes
// their own Questrade account.
type OAuthTokenStore struct {
	pool *pgxpool.Pool
}

func NewOAuthTokenStore(pool *pgxpool.Pool) *OAuthTokenStore {
	return &OAuthTokenStore{pool: pool}
}

// ForUser returns a store bound to a single user, satisfying
// questrade.TokenStore.
func (s *OAuthTokenStore) ForUser(userID string) *UserOAuthTokenStore {
	return &UserOAuthTokenStore{pool: s.pool, userID: userID}
}

// UserOAuthTokenStore is an OAuthTokenStore scoped to one user.
type UserOAuthTokenStore struct {
	pool   *pgxpool.Pool
	userID string
}

// LoadRefreshToken returns the last-persisted refresh token for broker, or
// "" if this user hasn't stored one yet.
func (s *UserOAuthTokenStore) LoadRefreshToken(ctx context.Context, broker string) (string, error) {
	var token string
	err := s.pool.QueryRow(ctx,
		`SELECT refresh_token FROM broker_oauth_tokens WHERE user_id = $1 AND broker = $2`,
		s.userID, broker,
	).Scan(&token)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return token, err
}

// SaveRefreshToken upserts this user's current refresh token for broker.
func (s *UserOAuthTokenStore) SaveRefreshToken(ctx context.Context, broker, token string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO broker_oauth_tokens (user_id, broker, refresh_token, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id, broker) DO UPDATE
			SET refresh_token = EXCLUDED.refresh_token, updated_at = now()
	`, s.userID, broker, token)
	return err
}
