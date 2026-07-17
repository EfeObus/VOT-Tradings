// Package questrade implements the brokerage.Broker contract against
// Questrade's REST API for Canadian (TSX/TSX-V/NEO) equities.
//
// Questrade uses OAuth 2.0 refresh-token rotation: every token exchange
// returns both a short-lived access token AND a new refresh token, which
// must replace the previous one (the old refresh token is invalidated).
// This client keeps the current refresh token and the account-specific
// api_server in memory and re-authenticates lazily when the access token
// expires.
//
// Reference: https://www.questrade.com/api/documentation/authentication
package questrade

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"vot-tradings/internal/brokerage"
	"vot-tradings/internal/config"
	"vot-tradings/internal/models"
)

// TokenStore persists the rotating refresh token across process restarts.
// Questrade invalidates the previous refresh token the instant a new one is
// issued, so without this, every restart burns access and forces the user
// back through Questrade's website to re-authorize.
type TokenStore interface {
	LoadRefreshToken(ctx context.Context, broker string) (string, error)
	SaveRefreshToken(ctx context.Context, broker, token string) error
}

type Client struct {
	cfg        config.QuestradeConfig
	httpClient *http.Client
	store      TokenStore
	Logger     *slog.Logger

	mu              sync.Mutex
	refreshToken    string
	loadedFromStore bool
	accessToken     string
	apiServer       string
	expiresAt       time.Time
	accountID       string
}

// New constructs a Questrade client. store may be nil, in which case the
// client behaves as before: cfg.RefreshToken is the only source of truth
// and rotated tokens live only in process memory.
func New(cfg config.QuestradeConfig, store TokenStore) *Client {
	return &Client{
		cfg:          cfg,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		refreshToken: cfg.RefreshToken,
		store:        store,
	}
}

func (c *Client) logger() *slog.Logger {
	if c.Logger != nil {
		return c.Logger
	}
	return slog.Default()
}

func (c *Client) Name() models.BrokerName {
	return models.BrokerQuestrade
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	APIServer    string `json:"api_server"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

// ensureAuth refreshes the access token if it's missing or about to expire,
// rotating the stored refresh token as Questrade requires.
//
// On the first call, if a TokenStore is configured, its value takes
// precedence over cfg.RefreshToken from .env: it reflects the most recent
// successful rotation, whereas the .env value may already be stale in a
// long-running deployment. If you manually paste a freshly-generated token
// into .env, clear its broker_oauth_tokens row first (or point at a fresh
// database) so the manual value isn't shadowed by a stale persisted one.
func (c *Client) ensureAuth(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		return nil
	}

	if !c.loadedFromStore {
		c.loadedFromStore = true
		if c.store != nil {
			if stored, err := c.store.LoadRefreshToken(ctx, string(models.BrokerQuestrade)); err != nil {
				c.logger().Warn("questrade: failed to load persisted refresh token, falling back to configured token", "error", err)
			} else if stored != "" {
				c.refreshToken = stored
			}
		}
	}

	url := fmt.Sprintf("%s?grant_type=refresh_token&refresh_token=%s", c.cfg.AuthURL, c.refreshToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("questrade: refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("questrade: refresh token: status %d: %s", resp.StatusCode, body)
	}

	var out tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fmt.Errorf("questrade: decode token: %w", err)
	}

	c.accessToken = out.AccessToken
	c.apiServer = strings.TrimRight(out.APIServer, "/")
	c.refreshToken = out.RefreshToken // must persist: old token is now invalid
	c.expiresAt = time.Now().Add(time.Duration(out.ExpiresIn-30) * time.Second)

	if c.store != nil {
		if err := c.store.SaveRefreshToken(ctx, string(models.BrokerQuestrade), c.refreshToken); err != nil {
			c.logger().Warn("questrade: failed to persist rotated refresh token; a restart before the next successful refresh will burn access and require re-authorizing via Questrade's website", "error", err)
		}
	}

	return nil
}

func (c *Client) apiGet(ctx context.Context, path string, out any) error {
	if err := c.ensureAuth(ctx); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiServer+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("questrade: get %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("questrade: get %s: status %d: %s", path, resp.StatusCode, body)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

type accountsResponse struct {
	Accounts []struct {
		Number string `json:"number"`
		Type   string `json:"type"`
	} `json:"accounts"`
}

type balancesResponse struct {
	PerCurrencyBalances []struct {
		Currency    string  `json:"currency"`
		Cash        float64 `json:"cash"`
		MarketValue float64 `json:"marketValue"`
		TotalEquity float64 `json:"totalEquity"`
		BuyingPower float64 `json:"buyingPower"`
	} `json:"perCurrencyBalances"`
}

func (c *Client) GetAccount(ctx context.Context) (models.Account, error) {
	if c.accountID == "" {
		var accts accountsResponse
		if err := c.apiGet(ctx, "/v1/accounts", &accts); err != nil {
			return models.Account{}, err
		}
		if len(accts.Accounts) == 0 {
			return models.Account{}, fmt.Errorf("questrade: no accounts found for this token")
		}
		c.accountID = accts.Accounts[0].Number
	}

	var balances balancesResponse
	if err := c.apiGet(ctx, fmt.Sprintf("/v1/accounts/%s/balances", c.accountID), &balances); err != nil {
		return models.Account{}, err
	}
	if len(balances.PerCurrencyBalances) == 0 {
		return models.Account{}, fmt.Errorf("questrade: no balances returned for account %s", c.accountID)
	}

	bal := balances.PerCurrencyBalances[0]
	return models.Account{
		ID:          c.accountID,
		Broker:      models.BrokerQuestrade,
		Currency:    bal.Currency,
		Equity:      bal.TotalEquity,
		BuyingPower: bal.BuyingPower,
		Cash:        bal.Cash,
		UpdatedAt:   time.Now().UTC(),
	}, nil
}

type symbolSearchResponse struct {
	Symbols []struct {
		SymbolID int    `json:"symbolId"`
		Symbol   string `json:"symbol"`
	} `json:"symbols"`
}

func (c *Client) resolveSymbolID(ctx context.Context, symbol string) (int, error) {
	var out symbolSearchResponse
	if err := c.apiGet(ctx, "/v1/symbols?names="+symbol, &out); err != nil {
		return 0, err
	}
	if len(out.Symbols) == 0 {
		return 0, fmt.Errorf("questrade: symbol not found: %s", symbol)
	}
	return out.Symbols[0].SymbolID, nil
}

type quotesResponse struct {
	Quotes []struct {
		Symbol    string  `json:"symbol"`
		BidPrice  float64 `json:"bidPrice"`
		AskPrice  float64 `json:"askPrice"`
	} `json:"quotes"`
}

func (c *Client) GetQuote(ctx context.Context, symbol string) (brokerage.Quote, error) {
	symbolID, err := c.resolveSymbolID(ctx, symbol)
	if err != nil {
		return brokerage.Quote{}, err
	}

	var out quotesResponse
	if err := c.apiGet(ctx, "/v1/markets/quotes/"+strconv.Itoa(symbolID), &out); err != nil {
		return brokerage.Quote{}, err
	}
	if len(out.Quotes) == 0 {
		return brokerage.Quote{}, fmt.Errorf("questrade: no quote data for %s", symbol)
	}

	q := out.Quotes[0]
	return brokerage.Quote{
		Symbol:    symbol,
		Bid:       q.BidPrice,
		Ask:       q.AskPrice,
		Timestamp: time.Now().UTC().UnixMilli(),
	}, nil
}

// PlaceOrder is not yet implemented: Questrade order submission requires an
// account-specific order impact preview call before the final commit call,
// which needs a real practice account to exercise and verify against. The
// read path (account + quotes) above is fully wired.
func (c *Client) PlaceOrder(ctx context.Context, req brokerage.PlaceOrderRequest) (models.Order, error) {
	return models.Order{}, fmt.Errorf("questrade: order placement not yet implemented")
}
