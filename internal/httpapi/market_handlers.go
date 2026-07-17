package httpapi

import (
	"context"
	"net/http"
	"sync"
	"time"

	"vot-tradings/internal/brokerage"
	"vot-tradings/internal/engine"
	"vot-tradings/internal/models"
)

type brokerAccountResult struct {
	broker  models.BrokerName
	account models.Account
	err     error
}

// brokerStatus reports one broker's fetch outcome so clients can render
// every connected broker (working or not) without parsing error strings.
type brokerStatus struct {
	Broker  models.BrokerName `json:"broker"`
	Account *models.Account   `json:"account,omitempty"`
	// EquityUSD is Account.Equity converted to USD via the same rate
	// engine.AggregateBalances uses, so clients can build a cross-broker
	// allocation breakdown without re-implementing (or guessing at) the
	// CAD/USD conversion themselves.
	EquityUSD *float64 `json:"equity_usd,omitempty"`
	Error     string   `json:"error,omitempty"`
}

type balanceResponse struct {
	Unified engine.UnifiedBalance `json:"unified"`
	Brokers []brokerStatus        `json:"brokers"`
}

// handleBalance builds the authenticated user's own brokers from their
// stored credentials, fans out to each concurrently, and rolls the results
// up into a single USD-denominated view via engine.AggregateBalances.
func (s *Server) handleBalance(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userBrokers, err := s.Brokers.BuildAll(ctx, userID)
	if err != nil {
		s.Logger.Error("balance: build user brokers", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	results := make([]brokerAccountResult, len(userBrokers))
	var wg sync.WaitGroup
	for i, b := range userBrokers {
		wg.Add(1)
		go func(i int, b brokerage.Broker) {
			defer wg.Done()
			acct, err := b.GetAccount(ctx)
			results[i] = brokerAccountResult{broker: b.Name(), account: acct, err: err}
		}(i, b)
	}
	wg.Wait()

	statuses := make([]brokerStatus, len(results))
	var accounts []models.Account
	for i, res := range results {
		if res.err != nil {
			s.Logger.Warn("balance: broker account fetch failed", "broker", res.broker, "error", res.err)
			statuses[i] = brokerStatus{Broker: res.broker, Error: res.err.Error()}
			continue
		}
		equityUSD := res.account.Equity * engine.USDRate(res.account.Currency, s.Config.USDCADRate)
		statuses[i] = brokerStatus{Broker: res.broker, Account: &res.account, EquityUSD: &equityUSD}
		accounts = append(accounts, res.account)
	}

	unified := engine.AggregateBalances(accounts, s.Config.USDCADRate)

	writeJSON(w, http.StatusOK, balanceResponse{
		Unified: unified,
		Brokers: statuses,
	})
}

type quoteResponse struct {
	Broker    models.BrokerName `json:"broker"`
	Symbol    string            `json:"symbol"`
	Bid       float64           `json:"bid"`
	Ask       float64           `json:"ask"`
	Timestamp int64             `json:"timestamp"`
}

// handleQuote looks up a single on-demand quote from one of the
// authenticated user's connected brokers. This is a synchronous REST call
// to the broker, not a stream — see the web client's Market page for why
// that distinction matters (no live tick feed exists).
func (s *Server) handleQuote(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	brokerName := r.URL.Query().Get("broker")
	symbol := r.URL.Query().Get("symbol")
	if brokerName == "" || symbol == "" {
		writeError(w, http.StatusBadRequest, "both broker and symbol query params are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	target, err := s.Brokers.Build(ctx, userID, brokerName)
	if err != nil {
		s.Logger.Error("quote: build broker", "broker", brokerName, "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if target == nil {
		writeError(w, http.StatusNotFound, "broker not connected: "+brokerName)
		return
	}

	quote, err := target.GetQuote(ctx, symbol)
	if err != nil {
		s.Logger.Warn("quote: fetch failed", "broker", brokerName, "symbol", symbol, "error", err)
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, quoteResponse{
		Broker:    target.Name(),
		Symbol:    quote.Symbol,
		Bid:       quote.Bid,
		Ask:       quote.Ask,
		Timestamp: quote.Timestamp,
	})
}
