// Package httpapi exposes the VOT Tradings gateway's HTTP surface: auth,
// health checks, per-user broker credential management, and the unified
// cross-broker balance/quote views built from each user's own connected
// brokers.
package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"vot-tradings/internal/auth"
	"vot-tradings/internal/config"
	"vot-tradings/internal/db"
	"vot-tradings/internal/userbrokers"
)

// Server holds the dependencies the HTTP handlers need to serve requests.
type Server struct {
	DB      *pgxpool.Pool
	Cache   *redis.Client
	Logger  *slog.Logger
	Config  config.Config

	Users       *db.UserStore
	Sessions    *auth.SessionStore
	Credentials *db.CredentialStore
	Brokers     *userbrokers.Factory

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
	mux.HandleFunc("GET /logo.png", s.handleLogo)

	mux.HandleFunc("POST /api/v1/auth/register", s.handleRegister)
	mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
	mux.HandleFunc("POST /api/v1/auth/logout", s.handleLogout)
	mux.HandleFunc("GET /api/v1/auth/me", s.requireAuth(s.handleMe))

	mux.HandleFunc("GET /api/v1/broker-credentials", s.requireAuth(s.handleListBrokerCredentials))
	mux.HandleFunc("POST /api/v1/broker-credentials", s.requireAuth(s.handleSaveBrokerCredential))
	mux.HandleFunc("DELETE /api/v1/broker-credentials", s.requireAuth(s.handleDeleteBrokerCredential))
	mux.HandleFunc("POST /api/v1/broker-credentials/import-env", s.requireAuth(s.handleImportEnvCredentials))

	mux.HandleFunc("GET /api/v1/balance", s.requireAuth(s.handleBalance))
	mux.HandleFunc("GET /api/v1/quote", s.requireAuth(s.handleQuote))

	return s.withCORS(mux)
}

// withCORS allows exactly the configured origins to read gateway responses
// from a browser, and to send the session cookie cross-origin. Origins not
// on the list get no CORS headers at all, so the browser's same-origin
// policy keeps blocking them by default.
func (s *Server) withCORS(next http.Handler) http.Handler {
	allowed := make(map[string]bool, len(s.AllowedOrigins))
	for _, o := range s.AllowedOrigins {
		allowed[o] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
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
// "process is up" from "process can actually serve traffic". Deliberately
// unauthenticated — load balancers hitting this shouldn't need a session.
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

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]string{"error": message})
}
