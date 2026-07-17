// Package httpapi exposes the VOT Tradings gateway's HTTP surface: health
// checks for the orchestrator/load balancer, and the unified cross-broker
// balance view described in the middleware core's "Buying Power Compute
// Engine".
package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"vot-tradings/internal/brokerage"
	"vot-tradings/internal/engine"
	"vot-tradings/internal/models"
)

// Server holds the dependencies the HTTP handlers need to serve requests.
type Server struct {
	Brokers    []brokerage.Broker
	DB         *pgxpool.Pool
	Cache      *redis.Client
	Logger     *slog.Logger
	USDCADRate float64

	// AssetsDir is the directory brand assets (e.g. logo.png) are served
	// from. See assets/logo.png — the single canonical app logo; the web
	// client should reference it via GET /logo.png rather than bundling its
	// own copy.
	AssetsDir string

	// AllowedOrigins lists the exact Origin values (e.g.
	// "http://localhost:5173") the browser-based web client is served from.
	// The gateway and the SPA run on different ports/hosts in every
	// environment, so without this, every browser fetch from the web client
	// is blocked by CORS before it reaches these handlers.
	AllowedOrigins []string
}

// Routes builds the gateway's HTTP handler tree.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /api/v1/balance", s.handleBalance)
	mux.HandleFunc("GET /logo.png", s.handleLogo)
	return s.withCORS(mux)
}

// withCORS allows exactly the configured origins to read gateway responses
// from a browser. Origins not on the list get no CORS headers at all, so
// the browser's same-origin policy keeps blocking them by default.
func (s *Server) withCORS(next http.Handler) http.Handler {
	allowed := make(map[string]bool, len(s.AllowedOrigins))
	for _, o := range s.AllowedOrigins {
		allowed[o] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleLogo serves the canonical app logo so every client can point at one
// URL instead of bundling its own copy of the asset.
func (s *Server) handleLogo(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.AssetsDir, "logo.png"))
}

type healthStatus struct {
	Status   string `json:"status"`
	Postgres string `json:"postgres"`
	Redis    string `json:"redis"`
}

// handleHealthz pings Postgres and Redis so an orchestrator can distinguish
// "process is up" from "process can actually serve traffic".
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	out := healthStatus{Status: "ok", Postgres: "ok", Redis: "ok"}
	code := http.StatusOK

	if err := s.DB.Ping(ctx); err != nil {
		out.Postgres = err.Error()
		out.Status = "degraded"
		code = http.StatusServiceUnavailable
	}
	if err := s.Cache.Ping(ctx).Err(); err != nil {
		out.Redis = err.Error()
		out.Status = "degraded"
		code = http.StatusServiceUnavailable
	}

	writeJSON(w, code, out)
}

type brokerAccountResult struct {
	broker  models.BrokerName
	account models.Account
	err     error
}

// brokerStatus reports one broker's fetch outcome so clients can render
// every configured broker (connected or not) without parsing error strings.
type brokerStatus struct {
	Broker  models.BrokerName `json:"broker"`
	Account *models.Account   `json:"account,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type balanceResponse struct {
	Unified engine.UnifiedBalance `json:"unified"`
	Brokers []brokerStatus        `json:"brokers"`
}

// handleBalance fans out to every configured broker concurrently and rolls
// the results up into a single USD-denominated view via
// engine.AggregateBalances.
func (s *Server) handleBalance(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	results := make([]brokerAccountResult, len(s.Brokers))
	var wg sync.WaitGroup
	for i, b := range s.Brokers {
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
		statuses[i] = brokerStatus{Broker: res.broker, Account: &res.account}
		accounts = append(accounts, res.account)
	}

	unified := engine.AggregateBalances(accounts, s.USDCADRate)

	writeJSON(w, http.StatusOK, balanceResponse{
		Unified: unified,
		Brokers: statuses,
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
