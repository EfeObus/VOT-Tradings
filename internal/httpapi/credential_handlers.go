package httpapi

import (
	"encoding/json"
	"net/http"
)

var knownBrokers = []string{"alpaca", "oanda", "questrade"}

func isKnownBroker(name string) bool {
	for _, b := range knownBrokers {
		if b == name {
			return true
		}
	}
	return false
}

type brokerCredentialStatus struct {
	Broker    string `json:"broker"`
	Connected bool   `json:"connected"`
}

// handleListBrokerCredentials reports, for every known broker, whether the
// authenticated user has connected it — never the credentials themselves.
func (s *Server) handleListBrokerCredentials(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	connected, err := s.Credentials.ListConnectedBrokers(r.Context(), userID)
	if err != nil {
		s.Logger.Error("list broker credentials", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	connectedSet := make(map[string]bool, len(connected))
	for _, b := range connected {
		connectedSet[b] = true
	}

	out := make([]brokerCredentialStatus, len(knownBrokers))
	for i, b := range knownBrokers {
		out[i] = brokerCredentialStatus{Broker: b, Connected: connectedSet[b]}
	}
	writeJSON(w, http.StatusOK, out)
}

type saveCredentialRequest struct {
	Broker      string            `json:"broker"`
	Credentials map[string]string `json:"credentials"`
}

// handleSaveBrokerCredential stores (encrypted) the fields a specific
// broker's client needs to authenticate on this user's behalf — e.g. for
// "alpaca": api_key_id, secret_key. See internal/userbrokers.Factory.Build
// for the exact fields each broker reads.
func (s *Server) handleSaveBrokerCredential(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	var req saveCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !isKnownBroker(req.Broker) {
		writeError(w, http.StatusBadRequest, "unknown broker: "+req.Broker)
		return
	}

	if err := s.Credentials.Save(r.Context(), userID, req.Broker, req.Credentials); err != nil {
		s.Logger.Error("save broker credential", "broker", req.Broker, "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteBrokerCredential(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	broker := r.URL.Query().Get("broker")
	if broker == "" {
		writeError(w, http.StatusBadRequest, "broker query param required")
		return
	}

	if err := s.Credentials.Delete(r.Context(), userID, broker); err != nil {
		s.Logger.Error("delete broker credential", "broker", broker, "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleImportEnvCredentials is a local-dev convenience: if the gateway
// process itself was started with legacy single-tenant broker credentials
// in .env (ALPACA_API_KEY_ID etc.), copy whichever ones are non-empty into
// the authenticated user's own encrypted credential rows, skipping any
// broker the user has already connected for real. This exists so testing
// multi-tenant auth doesn't require re-typing secrets that are already
// sitting in .env — it is not how a real second user would connect a
// broker.
func (s *Server) handleImportEnvCredentials(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	existing, err := s.Credentials.ListConnectedBrokers(r.Context(), userID)
	if err != nil {
		s.Logger.Error("import-env: list connected brokers", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	already := make(map[string]bool, len(existing))
	for _, b := range existing {
		already[b] = true
	}

	imported := []string{}

	if !already["alpaca"] && s.Config.Alpaca.APIKeyID != "" && s.Config.Alpaca.SecretKey != "" {
		err := s.Credentials.Save(r.Context(), userID, "alpaca", map[string]string{
			"api_key_id": s.Config.Alpaca.APIKeyID,
			"secret_key": s.Config.Alpaca.SecretKey,
			"base_url":   s.Config.Alpaca.BaseURL,
			"data_url":   s.Config.Alpaca.DataURL,
		})
		if err != nil {
			s.Logger.Error("import-env: save alpaca", "error", err)
		} else {
			imported = append(imported, "alpaca")
		}
	}

	if !already["oanda"] && s.Config.OANDA.AccountID != "" && s.Config.OANDA.AccessToken != "" {
		err := s.Credentials.Save(r.Context(), userID, "oanda", map[string]string{
			"account_id":   s.Config.OANDA.AccountID,
			"access_token": s.Config.OANDA.AccessToken,
			"base_url":     s.Config.OANDA.BaseURL,
		})
		if err != nil {
			s.Logger.Error("import-env: save oanda", "error", err)
		} else {
			imported = append(imported, "oanda")
		}
	}

	if !already["questrade"] && s.Config.Questrade.RefreshToken != "" {
		err := s.Credentials.Save(r.Context(), userID, "questrade", map[string]string{
			"refresh_token": s.Config.Questrade.RefreshToken,
			"auth_url":      s.Config.Questrade.AuthURL,
		})
		if err != nil {
			s.Logger.Error("import-env: save questrade", "error", err)
		} else {
			imported = append(imported, "questrade")
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"imported": imported})
}
