// Package userbrokers builds each user's own brokerage.Broker clients from
// their stored, decrypted credentials — the multi-tenant replacement for
// the old single shared set of broker credentials read once from .env.
package userbrokers

import (
	"context"
	"errors"
	"fmt"

	"vot-tradings/internal/brokerage"
	"vot-tradings/internal/brokerage/alpaca"
	"vot-tradings/internal/brokerage/oanda"
	"vot-tradings/internal/brokerage/questrade"
	"vot-tradings/internal/config"
	"vot-tradings/internal/db"
)

// CredentialLoader is the subset of *db.CredentialStore this package
// depends on, so it doesn't require a live Postgres pool to unit test.
type CredentialLoader interface {
	Load(ctx context.Context, userID, broker string) (map[string]string, error)
	ListConnectedBrokers(ctx context.Context, userID string) ([]string, error)
}

// Factory builds brokerage.Broker clients on demand for a specific user.
type Factory struct {
	Credentials CredentialLoader
	Tokens      *db.OAuthTokenStore
}

func NewFactory(credentials CredentialLoader, tokens *db.OAuthTokenStore) *Factory {
	return &Factory{Credentials: credentials, Tokens: tokens}
}

// BuildAll constructs a broker client for every broker userID has
// connected credentials for.
func (f *Factory) BuildAll(ctx context.Context, userID string) ([]brokerage.Broker, error) {
	connected, err := f.Credentials.ListConnectedBrokers(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("userbrokers: list connected brokers: %w", err)
	}

	brokers := make([]brokerage.Broker, 0, len(connected))
	for _, name := range connected {
		b, err := f.Build(ctx, userID, name)
		if err != nil {
			return nil, err
		}
		if b != nil {
			brokers = append(brokers, b)
		}
	}
	return brokers, nil
}

// Build constructs a single named broker's client for userID. Returns
// (nil, nil) — not an error — if the user hasn't connected that broker.
func (f *Factory) Build(ctx context.Context, userID, broker string) (brokerage.Broker, error) {
	creds, err := f.Credentials.Load(ctx, userID, broker)
	if errors.Is(err, db.ErrCredentialNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("userbrokers: load %s credentials: %w", broker, err)
	}

	switch broker {
	case "alpaca":
		return alpaca.New(config.AlpacaConfig{
			APIKeyID:  creds["api_key_id"],
			SecretKey: creds["secret_key"],
			BaseURL:   orDefault(creds["base_url"], "https://paper-api.alpaca.markets"),
			DataURL:   orDefault(creds["data_url"], "https://data.alpaca.markets"),
		}), nil
	case "oanda":
		return oanda.New(config.OANDAConfig{
			AccountID:   creds["account_id"],
			AccessToken: creds["access_token"],
			BaseURL:     orDefault(creds["base_url"], "https://api-fxpractice.oanda.com"),
		}), nil
	case "questrade":
		return questrade.New(config.QuestradeConfig{
			RefreshToken: creds["refresh_token"],
			AuthURL:      orDefault(creds["auth_url"], "https://login.questrade.com/oauth2/token"),
		}, f.Tokens.ForUser(userID)), nil
	default:
		return nil, fmt.Errorf("userbrokers: unknown broker %q", broker)
	}
}

func orDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
