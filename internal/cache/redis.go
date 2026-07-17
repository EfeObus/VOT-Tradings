// Package cache wraps the Redis Stack client used as VOT Tradings' low-
// latency market-state cache and Pub/Sub event distribution layer.
package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"vot-tradings/internal/config"
)

// Connect opens a Redis client and verifies connectivity with a PING.
func Connect(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cache: ping: %w", err)
	}

	return client, nil
}

// QuoteChannel returns the Pub/Sub channel name streamed to for a given
// symbol's normalized quote updates.
func QuoteChannel(symbol string) string {
	return fmt.Sprintf("quotes:%s", symbol)
}

// QuoteKey returns the cache key holding the latest known quote for a symbol.
func QuoteKey(symbol string) string {
	return fmt.Sprintf("quote:%s", symbol)
}
