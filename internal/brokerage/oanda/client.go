// Package oanda implements the brokerage.Broker contract against OANDA's
// v20 REST API for spot forex.
//
// Reference: https://developer.oanda.com/rest-live-v20/introduction/
package oanda

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
