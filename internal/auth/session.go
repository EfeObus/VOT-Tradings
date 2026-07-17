package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SessionCookieName is the HttpOnly cookie carrying the opaque session
// token. See withCORS in internal/httpapi — cookies only survive a
// cross-origin fetch when both CORS credentials and SameSite are set
// correctly, which is why the frontend must call the API with
// `credentials: 'include'`.
const SessionCookieName = "vot_session"

const sessionTTL = 7 * 24 * time.Hour

// SessionStore holds session_token -> userID in Redis with a rolling TTL.
// Sessions are ephemeral by design: losing them just signs everyone out,
// unlike the durable OAuth-token persistence in internal/db.
type SessionStore struct {
	redis *redis.Client
}

func NewSessionStore(redis *redis.Client) *SessionStore {
	return &SessionStore{redis: redis}
}

// Create mints a new session for userID and returns its opaque token.
func (s *SessionStore) Create(ctx context.Context, userID string) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("auth: generate token: %w", err)
	}
	if err := s.redis.Set(ctx, sessionKey(token), userID, sessionTTL).Err(); err != nil {
		return "", fmt.Errorf("auth: create session: %w", err)
	}
	return token, nil
}

// UserID resolves a session token to the user it belongs to. Returns
// redis.Nil (via the standard errors.Is check) if the token is unknown or
// expired.
func (s *SessionStore) UserID(ctx context.Context, token string) (string, error) {
	return s.redis.Get(ctx, sessionKey(token)).Result()
}

// Delete invalidates a session (logout).
func (s *SessionStore) Delete(ctx context.Context, token string) error {
	return s.redis.Del(ctx, sessionKey(token)).Err()
}

func sessionKey(token string) string {
	return "session:" + token
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
