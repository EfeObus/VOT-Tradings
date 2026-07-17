package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"vot-tradings/pkg/crypto"
)

// ErrCredentialNotFound is returned when a user hasn't connected the
// requested broker.
var ErrCredentialNotFound = errors.New("db: broker credential not found")

// CredentialStore persists each user's brokerage credentials, encrypted at
// rest with the gateway's AES-GCM key (see pkg/crypto). Credentials are
// stored as an opaque JSON object whose shape is broker-specific — e.g.
// {"api_key_id":"...","secret_key":"..."} for Alpaca — and interpreted by
// the broker factory that builds each user's brokerage.Broker clients.
type CredentialStore struct {
	pool *pgxpool.Pool
	box  *crypto.Box
}

func NewCredentialStore(pool *pgxpool.Pool, box *crypto.Box) *CredentialStore {
	return &CredentialStore{pool: pool, box: box}
}

// Save encrypts and upserts a user's credentials for one broker.
func (s *CredentialStore) Save(ctx context.Context, userID, broker string, credentials map[string]string) error {
	plaintext, err := json.Marshal(credentials)
	if err != nil {
		return fmt.Errorf("db: marshal credentials: %w", err)
	}
	encrypted, err := s.box.Encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("db: encrypt credentials: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO user_broker_credentials (user_id, broker, encrypted_credentials, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id, broker) DO UPDATE
			SET encrypted_credentials = EXCLUDED.encrypted_credentials, updated_at = now()
	`, userID, broker, encrypted)
	if err != nil {
		return fmt.Errorf("db: save credentials: %w", err)
	}
	return nil
}

// Load decrypts and returns a user's stored credentials for one broker.
func (s *CredentialStore) Load(ctx context.Context, userID, broker string) (map[string]string, error) {
	var encrypted []byte
	err := s.pool.QueryRow(ctx, `
		SELECT encrypted_credentials FROM user_broker_credentials WHERE user_id = $1 AND broker = $2
	`, userID, broker).Scan(&encrypted)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCredentialNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("db: load credentials: %w", err)
	}

	plaintext, err := s.box.Decrypt(encrypted)
	if err != nil {
		return nil, fmt.Errorf("db: decrypt credentials: %w", err)
	}

	var out map[string]string
	if err := json.Unmarshal(plaintext, &out); err != nil {
		return nil, fmt.Errorf("db: unmarshal credentials: %w", err)
	}
	return out, nil
}

// ListConnectedBrokers returns the broker names a user has saved
// credentials for, without decrypting anything.
func (s *CredentialStore) ListConnectedBrokers(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT broker FROM user_broker_credentials WHERE user_id = $1 ORDER BY broker
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("db: list connected brokers: %w", err)
	}
	defer rows.Close()

	var brokers []string
	for rows.Next() {
		var broker string
		if err := rows.Scan(&broker); err != nil {
			return nil, fmt.Errorf("db: scan broker: %w", err)
		}
		brokers = append(brokers, broker)
	}
	return brokers, rows.Err()
}

// Delete removes a user's stored credentials for one broker.
func (s *CredentialStore) Delete(ctx context.Context, userID, broker string) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM user_broker_credentials WHERE user_id = $1 AND broker = $2
	`, userID, broker)
	if err != nil {
		return fmt.Errorf("db: delete credentials: %w", err)
	}
	return nil
}
