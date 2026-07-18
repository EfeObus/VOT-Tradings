// Package oanda implements the brokerage.Broker contract against OANDA's
// v20 REST API for spot forex.
//
// Reference: https://developer.oanda.com/rest-live-v20/introduction/
package oanda

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"vot-tradings/internal/brokerage"
	"vot-tradings/internal/config"
	"vot-tradings/internal/models"
)

type Client struct {
	cfg        config.OANDAConfig
	httpClient *http.Client
}

func New(cfg config.OANDAConfig) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Name() models.BrokerName {
	return models.BrokerOANDA
}

func (c *Client) authHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.cfg.AccessToken)
	req.Header.Set("Content-Type", "application/json")
}

type accountSummaryResponse struct {
	Account struct {
		Currency          string `json:"currency"`
		Balance           string `json:"balance"`
		MarginAvailable   string `json:"marginAvailable"`
		NAV               string `json:"NAV"`
	} `json:"account"`
}

func (c *Client) GetAccount(ctx context.Context) (models.Account, error) {
	url := fmt.Sprintf("%s/v3/accounts/%s/summary", c.cfg.BaseURL, c.cfg.AccountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return models.Account{}, err
	}
	c.authHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return models.Account{}, fmt.Errorf("oanda: get account: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.Account{}, fmt.Errorf("oanda: get account: status %d: %s", resp.StatusCode, body)
	}

	var out accountSummaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return models.Account{}, fmt.Errorf("oanda: decode account: %w", err)
	}

	nav, _ := strconv.ParseFloat(out.Account.NAV, 64)
	balance, _ := strconv.ParseFloat(out.Account.Balance, 64)
	marginAvail, _ := strconv.ParseFloat(out.Account.MarginAvailable, 64)

	return models.Account{
		ID:          c.cfg.AccountID,
		Broker:      models.BrokerOANDA,
		Currency:    out.Account.Currency,
		Equity:      nav,
		BuyingPower: marginAvail,
		Cash:        balance,
		UpdatedAt:   time.Now().UTC(),
	}, nil
}

type pricingResponse struct {
	Prices []struct {
		Instrument string `json:"instrument"`
		Time       string `json:"time"`
		Bids       []struct {
			Price string `json:"price"`
		} `json:"bids"`
		Asks []struct {
			Price string `json:"price"`
		} `json:"asks"`
	} `json:"prices"`
}

func (c *Client) GetQuote(ctx context.Context, symbol string) (brokerage.Quote, error) {
	url := fmt.Sprintf("%s/v3/accounts/%s/pricing?instruments=%s", c.cfg.BaseURL, c.cfg.AccountID, symbol)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return brokerage.Quote{}, err
	}
	c.authHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return brokerage.Quote{}, fmt.Errorf("oanda: get quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return brokerage.Quote{}, fmt.Errorf("oanda: get quote: status %d: %s", resp.StatusCode, body)
	}

	var out pricingResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return brokerage.Quote{}, fmt.Errorf("oanda: decode quote: %w", err)
	}
	if len(out.Prices) == 0 || len(out.Prices[0].Bids) == 0 || len(out.Prices[0].Asks) == 0 {
		return brokerage.Quote{}, fmt.Errorf("oanda: no pricing data for %s", symbol)
	}

	price := out.Prices[0]
	bid, _ := strconv.ParseFloat(price.Bids[0].Price, 64)
	ask, _ := strconv.ParseFloat(price.Asks[0].Price, 64)
	ts, _ := time.Parse(time.RFC3339, price.Time)

	return brokerage.Quote{
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Timestamp: ts.UnixMilli(),
	}, nil
}

type orderCreatePayload struct {
	Order struct {
		Units       string `json:"units"`
		Instrument  string `json:"instrument"`
		Type        string `json:"type"`
		Price       string `json:"price,omitempty"`
		TimeInForce string `json:"timeInForce"`
	} `json:"order"`
}

type orderCreateResponse struct {
	OrderCreateTransaction struct {
		ID string `json:"id"`
	} `json:"orderCreateTransaction"`
	OrderFillTransaction struct {
		ID string `json:"id"`
	} `json:"orderFillTransaction"`
}

func (c *Client) PlaceOrder(ctx context.Context, req brokerage.PlaceOrderRequest) (models.Order, error) {
	units := req.Quantity
	if req.Side == models.OrderSideSell {
		units = -units
	}

	var payload orderCreatePayload
	payload.Order.Units = strconv.FormatFloat(units, 'f', -1, 64)
	payload.Order.Instrument = req.Symbol
	payload.Order.TimeInForce = "FOK"
	if req.Type == models.OrderTypeLimit && req.LimitPrice != nil {
		payload.Order.Type = "LIMIT"
		payload.Order.Price = strconv.FormatFloat(*req.LimitPrice, 'f', -1, 64)
		payload.Order.TimeInForce = "GTC"
	} else {
		payload.Order.Type = "MARKET"
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return models.Order{}, err
	}

	url := fmt.Sprintf("%s/v3/accounts/%s/orders", c.cfg.BaseURL, c.cfg.AccountID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return models.Order{}, err
	}
	c.authHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return models.Order{}, fmt.Errorf("oanda: place order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return models.Order{}, fmt.Errorf("oanda: place order: status %d: %s", resp.StatusCode, respBody)
	}

	var out orderCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return models.Order{}, fmt.Errorf("oanda: decode order: %w", err)
	}

	status := models.OrderStatusAccepted
	brokerOrderID := out.OrderCreateTransaction.ID
	if out.OrderFillTransaction.ID != "" {
		status = models.OrderStatusFilled
		brokerOrderID = out.OrderFillTransaction.ID
	}

	return models.Order{
		Broker:        models.BrokerOANDA,
		Symbol:        req.Symbol,
		Side:          req.Side,
		Type:          req.Type,
		Quantity:      req.Quantity,
		LimitPrice:    req.LimitPrice,
		Status:        status,
		BrokerOrderID: brokerOrderID,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}, nil
}

// PriceUpdate is a single normalized tick from OANDA's streaming pricing
// endpoint.
type PriceUpdate struct {
	Instrument string
	Bid        float64
	Ask        float64
	Time       time.Time
}

// streamBaseURL derives OANDA's streaming hostname from the configured REST
// base URL. OANDA serves streaming pricing on a *different host* than its
// REST API (stream-fxpractice/stream-fxtrade vs api-fxpractice/api-fxtrade)
// — reusing cfg.BaseURL directly would 404.
func (c *Client) streamBaseURL() (string, error) {
	switch c.cfg.BaseURL {
	case "https://api-fxpractice.oanda.com":
		return "https://stream-fxpractice.oanda.com", nil
	case "https://api-fxtrade.oanda.com":
		return "https://stream-fxtrade.oanda.com", nil
	default:
		return "", fmt.Errorf("oanda: cannot derive streaming host from base_url %q", c.cfg.BaseURL)
	}
}

// StreamPricing opens OANDA's chunked-HTTP pricing stream for the given
// instruments and pushes normalized ticks to the returned channel until ctx
// is canceled, the caller stops reading, or the connection drops. The
// channel is closed when streaming ends for any reason.
//
// Uses a dedicated http.Client with no request timeout — c.httpClient's 10s
// timeout covers the whole request including the body read, which would
// kill a long-lived stream after 10 seconds regardless of activity.
func (c *Client) StreamPricing(ctx context.Context, instruments []string) (<-chan PriceUpdate, error) {
	streamHost, err := c.streamBaseURL()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v3/accounts/%s/pricing/stream?instruments=%s",
		streamHost, c.cfg.AccountID, strings.Join(instruments, ","))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.authHeaders(req)

	streamClient := &http.Client{}
	resp, err := streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oanda: stream pricing: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("oanda: stream pricing: status %d: %s", resp.StatusCode, body)
	}

	out := make(chan PriceUpdate)
	go func() {
		defer close(out)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var raw struct {
				Type       string `json:"type"`
				Instrument string `json:"instrument"`
				Time       string `json:"time"`
				Bids       []struct {
					Price string `json:"price"`
				} `json:"bids"`
				Asks []struct {
					Price string `json:"price"`
				} `json:"asks"`
			}
			// Heartbeats (type "HEARTBEAT") and any unparseable line are
			// skipped rather than treated as errors — OANDA sends
			// heartbeats every ~5s to keep the connection alive.
			if err := json.Unmarshal(line, &raw); err != nil || raw.Type != "PRICE" {
				continue
			}
			if len(raw.Bids) == 0 || len(raw.Asks) == 0 {
				continue
			}

			bid, _ := strconv.ParseFloat(raw.Bids[0].Price, 64)
			ask, _ := strconv.ParseFloat(raw.Asks[0].Price, 64)
			t, _ := time.Parse(time.RFC3339, raw.Time)

			select {
			case out <- PriceUpdate{Instrument: raw.Instrument, Bid: bid, Ask: ask, Time: t}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}
