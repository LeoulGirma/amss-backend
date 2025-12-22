package handlers

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/aeromaintain/amss/internal/infra/redis"
	"github.com/google/uuid"
)

type AuthHandler struct {
	Service         *services.AuthService
	RateLimiter     *redis.RateLimiter
	LoginIPLimit    int
	LoginEmailLimit int
}

var errRateLimited = errors.New("rate limit exceeded")

type loginRequest struct {
	OrgID    string `json:"org_id" validate:"required,uuid"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	orgID, err := uuid.Parse(strings.TrimSpace(req.OrgID))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid org_id")
		return
	}
	email := strings.TrimSpace(req.Email)
	if err := h.checkLoginRateLimits(r, email); err != nil {
		if errors.Is(err, errRateLimited) {
			w.Header().Set("Retry-After", "60")
			writeError(w, r, http.StatusTooManyRequests, "RATE_LIMITED", "too many login attempts")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "RATE_LIMIT_ERROR", "rate limit unavailable")
		return
	}

	pair, err := h.Service.Login(r.Context(), orgID, email, req.Password)
	if err != nil {
		if err == domain.ErrUnauthorized {
			writeError(w, r, http.StatusUnauthorized, "AUTH_INVALID", "invalid credentials")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "AUTH_ERROR", "login failed")
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    int64(pair.ExpiresIn.Seconds()),
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	pair, err := h.Service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if err == domain.ErrUnauthorized {
			writeError(w, r, http.StatusUnauthorized, "AUTH_INVALID", "invalid refresh token")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "AUTH_ERROR", "refresh failed")
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    int64(pair.ExpiresIn.Seconds()),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if err := h.Service.Logout(r.Context(), req.RefreshToken); err != nil {
		if err == domain.ErrUnauthorized {
			writeError(w, r, http.StatusUnauthorized, "AUTH_INVALID", "invalid refresh token")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "AUTH_ERROR", "logout failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) checkLoginRateLimits(r *http.Request, email string) error {
	if h.RateLimiter == nil {
		return nil
	}
	ipLimit := h.LoginIPLimit
	if ipLimit == 0 {
		ipLimit = 5
	}
	emailLimit := h.LoginEmailLimit
	if emailLimit == 0 {
		emailLimit = 10
	}
	ip := clientIP(r)
	allowed, _, _, err := h.RateLimiter.Allow(r.Context(), "login:ip:"+ip, ipLimit, time.Minute)
	if err != nil {
		return err
	}
	if !allowed {
		return errRateLimited
	}
	allowed, _, _, err = h.RateLimiter.Allow(r.Context(), "login:email:"+strings.ToLower(email), emailLimit, time.Minute)
	if err != nil {
		return err
	}
	if !allowed {
		return errRateLimited
	}
	return nil
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
