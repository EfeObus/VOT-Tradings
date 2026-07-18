package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"vot-tradings/internal/brokerage"
	"vot-tradings/internal/engine"
	"vot-tradings/internal/models"
)

type createOrderRequest struct {
	Broker     string   `json:"broker"`
	Symbol     string   `json:"symbol"`
	Side       string   `json:"side"`
	Type       string   `json:"type"`
	Quantity   float64  `json:"quantity"`
	LimitPrice *float64 `json:"limit_price,omitempty"`
}

// handleCreateOrder places a real order against one of the authenticated
// user's connected brokers. This is the one endpoint in the gateway that
// moves real money if the underlying broker credentials point at a live
// account — see the root README's operational notes on OANDA.
func (s *Server) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	ctx := r.Context()

	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !isKnownBroker(req.Broker) {
		writeError(w, http.StatusBadRequest, "unknown broker: "+req.Broker)
		return
	}
	side := models.OrderSide(req.Side)
	if side != models.OrderSideBuy && side != models.OrderSideSell {
		writeError(w, http.StatusBadRequest, "side must be \"buy\" or \"sell\"")
		return
	}
	orderType := models.OrderType(req.Type)
	if orderType != models.OrderTypeMarket && orderType != models.OrderTypeLimit {
		writeError(w, http.StatusBadRequest, "type must be \"market\" or \"limit\"")
		return
	}
	if req.Symbol == "" {
		writeError(w, http.StatusBadRequest, "symbol is required")
		return
	}
	if req.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "quantity must be positive")
		return
	}
	if orderType == models.OrderTypeLimit && req.LimitPrice == nil {
		writeError(w, http.StatusBadRequest, "limit_price is required for limit orders")
		return
	}

	target, err := s.Brokers.Build(ctx, userID, req.Broker)
	if err != nil {
		s.Logger.Error("create order: build broker", "broker", req.Broker, "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if target == nil {
		writeError(w, http.StatusNotFound, "broker not connected: "+req.Broker)
		return
	}

	// FINRA's Pattern Day Trader rule only applies to US margin equity
	// accounts — i.e. Alpaca here, not OANDA (forex) or Questrade
	// (Canadian equities, under IIROC rules instead). Applying it broker-
	// wide would incorrectly block legitimate non-US trades.
	if models.BrokerName(req.Broker) == models.BrokerAlpaca {
		account, err := target.GetAccount(ctx)
		if err != nil {
			writeError(w, http.StatusBadGateway, "fetch account for PDT check: "+err.Error())
			return
		}

		trades, err := s.DayTrades.ListRecent(ctx, userID, models.BrokerAlpaca)
		if err != nil {
			s.Logger.Error("create order: list day trades", "error", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		pdt := engine.CheckPDT(account, userID, models.BrokerAlpaca, trades, time.Now())
		if pdt.Restricted {
			writeError(w, http.StatusForbidden, pdt.Reason)
			return
		}
	}

	order, err := target.PlaceOrder(ctx, brokerage.PlaceOrderRequest{
		Symbol:     req.Symbol,
		Side:       side,
		Type:       orderType,
		Quantity:   req.Quantity,
		LimitPrice: req.LimitPrice,
	})
	if err != nil {
		s.Logger.Warn("create order: broker rejected", "broker", req.Broker, "symbol", req.Symbol, "error", err)
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	order.UserID = userID

	if err := s.Orders.Create(ctx, uuid.NewString(), order); err != nil {
		// The broker already accepted the order at this point — a
		// persistence failure here must not be reported as an order
		// failure, or a caller retrying could double-submit a real trade.
		s.Logger.Error("create order: persist", "error", err)
	}

	writeJSON(w, http.StatusCreated, order)
}

// handleListOrders returns the authenticated user's own order history.
func (s *Server) handleListOrders(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	orders, err := s.Orders.ListByUser(r.Context(), userID)
	if err != nil {
		s.Logger.Error("list orders", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if orders == nil {
		orders = []models.Order{}
	}
	writeJSON(w, http.StatusOK, orders)
}
