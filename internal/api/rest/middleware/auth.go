package middleware

import (
	"context"
	"crypto/rsa"
	"net/http"
	"strings"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/aeromaintain/amss/pkg/auth"
	"github.com/google/uuid"
)

type principalKey struct{}

type Principal struct {
	UserID  uuid.UUID
	OrgID   uuid.UUID
	Role    domain.Role
	Scopes  []string
	TokenID string
}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	value := ctx.Value(principalKey{})
	if value == nil {
		return Principal{}, false
	}
	principal, ok := value.(Principal)
	return principal, ok
}

type Authenticator struct {
	PublicKey *rsa.PublicKey
}

func (a Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			writeError(w, r, http.StatusUnauthorized, "AUTH", "missing authorization")
			return
		}
		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeError(w, r, http.StatusUnauthorized, "AUTH", "invalid authorization header")
			return
		}
		claims, err := auth.ParseToken(parts[1], a.PublicKey)
		if err != nil {
			writeError(w, r, http.StatusUnauthorized, "AUTH", "invalid token")
			return
		}
		if claims.TokenType != "access" {
			writeError(w, r, http.StatusUnauthorized, "AUTH", "invalid token type")
			return
		}
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			writeError(w, r, http.StatusUnauthorized, "AUTH", "invalid subject")
			return
		}
		orgID, err := uuid.Parse(claims.OrgID)
		if err != nil {
			writeError(w, r, http.StatusUnauthorized, "AUTH", "invalid org")
			return
		}
		principal := Principal{
			UserID:  userID,
			OrgID:   orgID,
			Role:    domain.Role(claims.Role),
			Scopes:  claims.Scopes,
			TokenID: claims.ID,
		}
		ctx := WithPrincipal(r.Context(), principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
