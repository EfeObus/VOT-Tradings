// Package models defines the domain structures shared across the VOT
// Tradings Go services (brokerage adapters, engine, gateway).
package models

import "time"

// AssetClass distinguishes the venue/instrument family a symbol belongs to.
type AssetClass string

const (
	AssetClassUSEquity AssetClass = "us_equity"
	AssetClassCAEquity AssetClass = "ca_equity"
	AssetClassForex    AssetClass = "forex"
)

// BrokerName identifies which upstream brokerage a record originated from.
type BrokerName string

const (
	BrokerAlpaca    BrokerName = "alpaca"
	BrokerOANDA     BrokerName = "oanda"
	BrokerQuestrade BrokerName = "questrade"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

type OrderType string

const (
	OrderTypeMarket OrderType = "market"
	OrderTypeLimit  OrderType = "limit"
)

type OrderStatus string

const (
	OrderStatusPendingNew OrderStatus = "pending_new"
	OrderStatusAccepted   OrderStatus = "accepted"
	OrderStatusFilled     OrderStatus = "filled"
	OrderStatusPartial    OrderStatus = "partially_filled"
	OrderStatusCanceled   OrderStatus = "canceled"
	OrderStatusRejected   OrderStatus = "rejected"
)

// Account is a single brokerage account normalized into VOT's unified view.
type Account struct {
	ID            string     `json:"id" db:"id"`
	Broker        BrokerName `json:"broker" db:"broker"`
	Currency      string     `json:"currency" db:"currency"`
	Equity        float64    `json:"equity" db:"equity"`
	BuyingPower   float64    `json:"buying_power" db:"buying_power"`
	Cash          float64    `json:"cash" db:"cash"`
	PatternDayTrader bool    `json:"pattern_day_trader" db:"pattern_day_trader"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// Position is a held quantity of a symbol at a given broker.
type Position struct {
	ID           string     `json:"id" db:"id"`
	AccountID    string     `json:"account_id" db:"account_id"`
	Broker       BrokerName `json:"broker" db:"broker"`
	Symbol       string     `json:"symbol" db:"symbol"`
	AssetClass   AssetClass `json:"asset_class" db:"asset_class"`
	Quantity     float64    `json:"quantity" db:"quantity"`
	AvgEntryPrice float64   `json:"avg_entry_price" db:"avg_entry_price"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Order represents a single order lifecycle record routed to a broker.
type Order struct {
	ID          string      `json:"id" db:"id"`
	AccountID   string      `json:"account_id" db:"account_id"`
	Broker      BrokerName  `json:"broker" db:"broker"`
	Symbol      string      `json:"symbol" db:"symbol"`
	Side        OrderSide   `json:"side" db:"side"`
	Type        OrderType   `json:"type" db:"type"`
	Quantity    float64     `json:"quantity" db:"quantity"`
	LimitPrice  *float64    `json:"limit_price,omitempty" db:"limit_price"`
	Status      OrderStatus `json:"status" db:"status"`
	BrokerOrderID string    `json:"broker_order_id,omitempty" db:"broker_order_id"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// Prediction is one row of the self-correcting-loop inference ledger: a
// forecast emitted by the DL engine, later reconciled against real outcomes.
type Prediction struct {
	InferenceID       string     `json:"inference_id" db:"inference_id"`
	Symbol            string     `json:"symbol" db:"symbol"`
	AssetClass        AssetClass `json:"asset_class" db:"asset_class"`
	HorizonMinutes    int        `json:"horizon_minutes" db:"horizon_minutes"`
	PredictedPrice    float64    `json:"predicted_price" db:"predicted_price"`
	PredictedAt       time.Time  `json:"predicted_at" db:"predicted_at"`
	ActualPrice       *float64   `json:"actual_price,omitempty" db:"actual_price"`
	ResolvedAt        *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	AbsError          *float64   `json:"abs_error,omitempty" db:"abs_error"`
	DirectionCorrect  *bool      `json:"direction_correct,omitempty" db:"direction_correct"`
	ModelVersion      string     `json:"model_version" db:"model_version"`
}

// DayTrade tracks one round-trip (buy+sell of the same symbol in the same
// session) for Pattern Day Trader rolling-5-business-day compliance.
type DayTrade struct {
	ID        string    `json:"id" db:"id"`
	AccountID string    `json:"account_id" db:"account_id"`
	Symbol    string    `json:"symbol" db:"symbol"`
	TradeDate time.Time `json:"trade_date" db:"trade_date"`
}

// User is a registered VOT Tradings account holder. PasswordHash is never
// serialized to JSON — callers must not accidentally leak it in an API
// response.
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
