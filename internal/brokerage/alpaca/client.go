// Package alpaca implements the brokerage.Broker contract against Alpaca's
// paper/live trading REST API for US equities.
//
// Reference: https://docs.alpaca.markets/reference
package alpaca

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"vot-tradings/internal/brokerage"
	"vot-tradings/internal/config"
	"vot-tradings/internal/models"
)

type Client struct {
	cfg        config.AlpacaConfig
	httpClient *http.Client
}

func New(cfg config.AlpacaConfig) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Name() models.BrokerName {
	return models.BrokerAlpaca
}

func (c *Client) authHeaders(req *http.Request) {
	req.Header.Set("APCA-API-KEY-ID", c.cfg.APIKeyID)
	req.Header.Set("APCA-API-SECRET-KEY", c.cfg.SecretKey)
	req.Header.Set("Content-Type", "application/json")
}

type accountResponse struct {
	ID               string `json:"id"`
	Currency         string `json:"currency"`
	Cash             string `json:"cash"`
	Equity           string `json:"equity"`
	BuyingPower      string `json:"buying_power"`
	PatternDayTrader bool   `json:"pattern_day_trader"`
}

func (c *Client) GetAccount(ctx context.Context) (models.Account, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.BaseURL+"/v2/account", nil)
	if err != nil {
		return models.Account{}, err
	}
	c.authHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return models.Account{}, fmt.Errorf("alpaca: get account: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.Account{}, fmt.Errorf("alpaca: get account: status %d: %s", resp.StatusCode, body)
	}

	var out accountResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return models.Account{}, fmt.Errorf("alpaca: decode account: %w", err)
	}

	return models.Account{
		ID:               out.ID,
		Broker:           models.BrokerAlpaca,
		Currency:         out.Currency,
		Equity:           parseFloat(out.Equity),
		BuyingPower:      parseFloat(out.BuyingPower),
		Cash:             parseFloat(out.Cash),
		PatternDayTrader: out.PatternDayTrader,
		UpdatedAt:        time.Now().UTC(),
	}, nil
}

type latestQuoteResponse struct {
	Symbol string `json:"symbol"`
	Quote  struct {
		BidPrice  float64   `json:"bp"`
		AskPrice  float64   `json:"ap"`
		Timestamp time.Time `json:"t"`
	} `json:"quote"`
}

func (c *Client) GetQuote(ctx context.Context, symbol string) (brokerage.Quote, error) {
	url := fmt.Sprintf("%s/v2/stocks/%s/quotes/latest", c.cfg.DataURL, symbol)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return brokerage.Quote{}, err
	}
	c.authHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return brokerage.Quote{}, fmt.Errorf("alpaca: get quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return brokerage.Quote{}, fmt.Errorf("alpaca: get quote: status %d: %s", resp.StatusCode, body)
	}

	var out latestQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return brokerage.Quote{}, fmt.Errorf("alpaca: decode quote: %w", err)
	}

	return brokerage.Quote{
		Symbol:    symbol,
		Bid:       out.Quote.BidPrice,
		Ask:       out.Quote.AskPrice,
		Timestamp: out.Quote.Timestamp.UnixMilli(),
	}, nil
}

type placeOrderPayload struct {
	Symbol      string  `json:"symbol"`
	Qty         float64 `json:"qty"`
	Side        string  `json:"side"`
	Type        string  `json:"type"`
	TimeInForce string  `json:"time_in_force"`
	LimitPrice  *float64 `json:"limit_price,omitempty"`
}

type orderResponse struct {
	ID        string `json:"id"`
	Symbol    string `json:"symbol"`
	Side      string `json:"side"`
	Type      string `json:"type"`
	Qty       string `json:"qty"`
	Status    string `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *Client) PlaceOrder(ctx context.Context, req brokerage.PlaceOrderRequest) (models.Order, error) {
	payload := placeOrderPayload{
		Symbol:      req.Symbol,
		Qty:         req.Quantity,
		Side:        string(req.Side),
		Type:        string(req.Type),
		TimeInForce: "day",
		LimitPrice:  req.LimitPrice,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return models.Order{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/v2/orders", bytes.NewReader(body))
	if err != nil {
		return models.Order{}, err
	}
	c.authHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return models.Order{}, fmt.Errorf("alpaca: place order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return models.Order{}, fmt.Errorf("alpaca: place order: status %d: %s", resp.StatusCode, respBody)
	}

	var out orderResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return models.Order{}, fmt.Errorf("alpaca: decode order: %w", err)
	}

	return models.Order{
		Broker:        models.BrokerAlpaca,
		Symbol:        out.Symbol,
		Side:          req.Side,
		Type:          req.Type,
		Quantity:      req.Quantity,
		LimitPrice:    req.LimitPrice,
		Status:        models.OrderStatus(out.Status),
		BrokerOrderID: out.ID,
		CreatedAt:     out.CreatedAt,
		UpdatedAt:     out.UpdatedAt,
	}, nil
}

func parseFloat(s string) float64 {
	var f float64
	_, _ = fmt.Sscanf(s, "%f", &f)
	return f
}
