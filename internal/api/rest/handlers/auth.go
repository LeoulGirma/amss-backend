package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/aeromaintain/amss/internal/api/rest/middleware"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/aeromaintain/amss/internal/infra/redis"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type AuthHandler struct {
	Service         *services.AuthService
	Repository      ports.AuthRepository
	RateLimiter     *redis.RateLimiter
	LoginIPLimit    int
	LoginEmailLimit int
	Logger          zerolog.Logger
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

type lookupRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type orgInfoResponse struct {
	OrgID   string `json:"org_id"`
	OrgName string `json:"org_name"`
}

type lookupResponse struct {
	Organizations []orgInfoResponse `json:"organizations"`
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

func (h *AuthHandler) Lookup(w http.ResponseWriter, r *http.Request) {
	var req lookupRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	orgs, err := h.Repository.LookupOrgsByEmail(r.Context(), email)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "LOOKUP_ERROR", "lookup failed")
		return
	}

	resp := lookupResponse{Organizations: make([]orgInfoResponse, 0, len(orgs))}
	for _, org := range orgs {
		resp.Organizations = append(resp.Organizations, orgInfoResponse{
			OrgID:   org.OrgID.String(),
			OrgName: org.OrgName,
		})
	}

	writeJSON(w, http.StatusOK, resp)
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
		h.logRateLimitError(r, "ip", ip, err)
		return err
	}
	if !allowed {
		return errRateLimited
	}
	allowed, _, _, err = h.RateLimiter.Allow(r.Context(), "login:email:"+strings.ToLower(email), emailLimit, time.Minute)
	if err != nil {
		h.logRateLimitError(r, "email", strings.ToLower(email), err)
		return err
	}
	if !allowed {
		return errRateLimited
	}
	return nil
}

func (h *AuthHandler) logRateLimitError(r *http.Request, dimension, identifier string, err error) {
	h.Logger.Error().
		Err(err).
		Str("request_id", middleware.RequestIDFromContext(r.Context())).
		Str("dimension", dimension).
		Str("identifier_hash", hashIdentifier(identifier)).
		Str("client_ip", clientIP(r)).
		Msg("login_rate_limit_error")
}

func hashIdentifier(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
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

// meResponse represents the current user's profile
type meResponse struct {
	ID        string  `json:"id"`
	OrgID     string  `json:"org_id"`
	Email     string  `json:"email"`
	Role      string  `json:"role"`
	LastLogin *string `json:"last_login"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// Me returns the current authenticated user's profile
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	principal, ok := getPrincipal(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}

	user, err := h.Repository.GetUserByID(r.Context(), principal.OrgID, principal.UserID)
	if err != nil {
		if err == domain.ErrNotFound {
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "failed to get user")
		return
	}

	var lastLogin *string
	if user.LastLogin != nil {
		formatted := user.LastLogin.Format(time.RFC3339)
		lastLogin = &formatted
	}

	writeJSON(w, http.StatusOK, meResponse{
		ID:        user.ID.String(),
		OrgID:     user.OrgID.String(),
		Email:     user.Email,
		Role:      string(user.Role),
		LastLogin: lastLogin,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	})
}
