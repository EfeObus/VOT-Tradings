package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"vot-tradings/internal/auth"
	"vot-tradings/internal/db"
	"vot-tradings/internal/models"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func userIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(userIDContextKey).(string)
	return id
}

// requireAuth gates a handler behind a valid session cookie, injecting the
// authenticated user's ID into the request context.
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.SessionCookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		userID, err := s.Sessions.UserID(r.Context(), cookie.Value)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "session expired or invalid")
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func toUserResponse(u models.User) userResponse {
	return userResponse{ID: u.ID, Email: u.Email, CreatedAt: u.CreatedAt}
}

// startSession mints a session for userID and attaches it as an HttpOnly
// cookie. Secure is only set when the request itself arrived over TLS —
// production deployments must terminate TLS in front of the gateway (or set
// this unconditionally) for the cookie to survive; plain-HTTP local dev
// needs Secure unset or browsers silently drop it.
func (s *Server) startSession(w http.ResponseWriter, r *http.Request, userID string) bool {
	token, err := s.Sessions.Create(r.Context(), userID)
	if err != nil {
		s.Logger.Error("start session", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return false
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})
	return true
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "email is required and password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		s.Logger.Error("register: hash password", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	user, err := s.Users.CreateUser(r.Context(), uuid.NewString(), req.Email, hash)
	if errors.Is(err, db.ErrEmailTaken) {
		writeError(w, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		s.Logger.Error("register: create user", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !s.startSession(w, r, user.ID) {
		return
	}
	writeJSON(w, http.StatusCreated, toUserResponse(user))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := s.Users.GetUserByEmail(r.Context(), req.Email)
	if err != nil || !auth.VerifyPassword(user.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if !s.startSession(w, r, user.ID) {
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.SessionCookieName); err == nil {
		_ = s.Sessions.Delete(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, err := s.Users.GetUserByID(r.Context(), userIDFromContext(r.Context()))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}
