package middleware

import (
	"net/http"

	"github.com/aeromaintain/amss/internal/domain"
)

func RequireRoles(roles ...domain.Role) func(http.Handler) http.Handler {
	allowed := map[domain.Role]struct{}{}
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := PrincipalFromContext(r.Context())
			if !ok {
				writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
				return
			}
			if principal.Role == domain.RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}
			if _, exists := allowed[principal.Role]; !exists {
				writeError(w, r, http.StatusForbidden, "FORBIDDEN", "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
