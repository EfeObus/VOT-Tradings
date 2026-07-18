package httpapi

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"

	"vot-tradings/internal/auth"
	"vot-tradings/internal/brokerage/oanda"
)

// handleStreamQuotes upgrades to a WebSocket and forwards a live OANDA
// pricing stream to the client. Streaming is currently OANDA-only — Alpaca
// and Questrade don't have a streamer implementation yet (see
// internal/brokerage/oanda's StreamPricing and the root README).
//
// Auth happens before the upgrade using the same session cookie as every
// other endpoint; failures are plain HTTP errors (not JSON) since a
// WebSocket handshake failure isn't a JSON API response.
func (s *Server) handleStreamQuotes(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(auth.SessionCookieName)
	if err != nil {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	userID, err := s.Sessions.UserID(r.Context(), cookie.Value)
	if err != nil {
		http.Error(w, "session expired or invalid", http.StatusUnauthorized)
		return
	}

	brokerName := r.URL.Query().Get("broker")
	symbol := r.URL.Query().Get("symbol")
	if brokerName == "" || symbol == "" {
		http.Error(w, "broker and symbol query params are required", http.StatusBadRequest)
		return
	}

	target, err := s.Brokers.Build(r.Context(), userID, brokerName)
	if err != nil {
		s.Logger.Error("stream quotes: build broker", "broker", brokerName, "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if target == nil {
		http.Error(w, "broker not connected: "+brokerName, http.StatusNotFound)
		return
	}
	oandaClient, ok := target.(*oanda.Client)
	if !ok {
		http.Error(w, "streaming not implemented for broker: "+brokerName, http.StatusNotImplemented)
		return
	}

	allowedOrigins := make(map[string]bool, len(s.AllowedOrigins))
	for _, o := range s.AllowedOrigins {
		allowedOrigins[o] = true
	}
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return allowedOrigins[r.Header.Get("Origin")] },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Logger.Warn("stream quotes: upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	ticks, err := oandaClient.StreamPricing(ctx, []string{symbol})
	if err != nil {
		_ = conn.WriteJSON(map[string]string{"error": err.Error()})
		return
	}

	for tick := range ticks {
		msg := map[string]any{
			"broker":    "oanda",
			"symbol":    tick.Instrument,
			"bid":       tick.Bid,
			"ask":       tick.Ask,
			"timestamp": tick.Time.UnixMilli(),
		}
		if err := conn.WriteJSON(msg); err != nil {
			// Client disconnected or the connection died — stop reading
			// the upstream OANDA stream too.
			return
		}
	}
}
