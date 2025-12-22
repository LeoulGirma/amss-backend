package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aeromaintain/amss/pkg/observability"
	"github.com/google/uuid"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time, error)
}

type RateLimitConfig struct {
	Limiter            RateLimiter
	Window             time.Duration
	Category           string
	DefaultLimit       int
	LimitFor           func(ctx context.Context, orgID uuid.UUID) (int, error)
	APIKeyHeader       string
	APIKeyCategory     string
	APIKeyDefaultLimit int
	APIKeyLimitFor     func(ctx context.Context, orgID uuid.UUID) (int, error)
}

func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if cfg.Window == 0 {
		cfg.Window = time.Minute
	}
	category := cfg.Category
	if category == "" {
		category = "default"
	}
	apiKeyHeader := cfg.APIKeyHeader
	if apiKeyHeader == "" {
		apiKeyHeader = "X-API-Key"
	}
	apiKeyCategory := cfg.APIKeyCategory
	if apiKeyCategory == "" {
		apiKeyCategory = "api_key"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Limiter == nil {
				next.ServeHTTP(w, r)
				return
			}
			principal, ok := PrincipalFromContext(r.Context())
			if !ok {
				writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
				return
			}
			apiKey := strings.TrimSpace(r.Header.Get(apiKeyHeader))
			if apiKey != "" {
				apiKeyLimit := cfg.APIKeyDefaultLimit
				if cfg.APIKeyLimitFor != nil {
					value, err := cfg.APIKeyLimitFor(r.Context(), principal.OrgID)
					if err != nil {
						writeError(w, r, http.StatusInternalServerError, "INTERNAL", "rate limit unavailable")
						return
					}
					apiKeyLimit = value
				}
				if apiKeyLimit > 0 {
					apiKeyHash := sha256.Sum256([]byte(apiKey))
					key := "rl:" + principal.OrgID.String() + ":" + apiKeyCategory + ":" + hex.EncodeToString(apiKeyHash[:])
					allowed, remaining, resetAt, err := cfg.Limiter.Allow(r.Context(), key, apiKeyLimit, cfg.Window)
					if err != nil {
						writeError(w, r, http.StatusInternalServerError, "INTERNAL", "rate limit error")
						return
					}
					w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
					w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(resetAt.Unix())))
					if !allowed {
						retryAfter := int(time.Until(resetAt).Seconds())
						if retryAfter < 0 {
							retryAfter = 0
						}
						w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
						observability.IncRateLimitRejection("api_key")
						writeError(w, r, http.StatusTooManyRequests, "RATE_LIMITED", "rate limit exceeded")
						return
					}
				}
			}
			limit := cfg.DefaultLimit
			if cfg.LimitFor != nil {
				value, err := cfg.LimitFor(r.Context(), principal.OrgID)
				if err != nil {
					writeError(w, r, http.StatusInternalServerError, "INTERNAL", "rate limit unavailable")
					return
				}
				limit = value
			}
			if limit <= 0 {
				next.ServeHTTP(w, r)
				return
			}
			key := "rl:" + principal.OrgID.String() + ":" + category
			allowed, remaining, resetAt, err := cfg.Limiter.Allow(r.Context(), key, limit, cfg.Window)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL", "rate limit error")
				return
			}
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(resetAt.Unix())))
			if !allowed {
				retryAfter := int(time.Until(resetAt).Seconds())
				if retryAfter < 0 {
					retryAfter = 0
				}
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				observability.IncRateLimitRejection("org")
				writeError(w, r, http.StatusTooManyRequests, "RATE_LIMITED", "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
