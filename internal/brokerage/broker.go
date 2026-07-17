// Package brokerage defines the common contract implemented by every
// upstream brokerage driver (Alpaca, OANDA, Questrade), so the gateway and
// engine can route across brokers without asset-class-specific branching.
package brokerage

import (
	"context"

	"vot-tradings/internal/models"
)

// Quote is a normalized top-of-book price snapshot, regardless of which
// broker or asset class it was sourced from.
type Quote struct {
	Symbol    string
	Bid       float64
	Ask       float64
	Timestamp int64 // unix millis
}

// PlaceOrderRequest is the broker-agnostic order submission payload.
type PlaceOrderRequest struct {
	Symbol     string
	Side       models.OrderSide
	Type       models.OrderType
	Quantity   float64
	LimitPrice *float64
}

// Broker is implemented by every brokerage driver. Callers depend on this
// interface, never on a concrete client, so a new venue only requires a new
// implementation of this contract.
type Broker interface {
	// Name identifies which broker this driver talks to.
	Name() models.BrokerName

	// GetAccount fetches the current account snapshot (equity, buying
	// power, cash, PDT flag).
	GetAccount(ctx context.Context) (models.Account, error)

	// GetQuote fetches the latest top-of-book quote for a symbol.
	GetQuote(ctx context.Context, symbol string) (Quote, error)

	// PlaceOrder submits an order and returns the broker's order record.
	PlaceOrder(ctx context.Context, req PlaceOrderRequest) (models.Order, error)
}
