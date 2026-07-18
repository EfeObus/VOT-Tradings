package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"vot-tradings/internal/models"
)

type DayTradeStore struct {
	pool *pgxpool.Pool
}

func NewDayTradeStore(pool *pgxpool.Pool) *DayTradeStore {
	return &DayTradeStore{pool: pool}
}

// ListRecent returns a user's day_trades rows at broker from the last 14
// calendar days — comfortably more than the rolling 5-*business*-day window
// engine.CheckPDT actually applies, so it can do that filtering itself.
func (s *DayTradeStore) ListRecent(ctx context.Context, userID string, broker models.BrokerName) ([]models.DayTrade, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, broker, symbol, trade_date
		FROM day_trades
		WHERE user_id = $1 AND broker = $2 AND trade_date >= $3
	`, userID, broker, time.Now().AddDate(0, 0, -14))
	if err != nil {
		return nil, fmt.Errorf("db: list day trades: %w", err)
	}
	defer rows.Close()

	var trades []models.DayTrade
	for rows.Next() {
		var t models.DayTrade
		if err := rows.Scan(&t.ID, &t.UserID, &t.Broker, &t.Symbol, &t.TradeDate); err != nil {
			return nil, fmt.Errorf("db: scan day trade: %w", err)
		}
		trades = append(trades, t)
	}
	return trades, rows.Err()
}
