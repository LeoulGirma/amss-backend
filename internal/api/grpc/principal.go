package grpcapi

import (
	"context"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func actorFromContext(ctx context.Context) (app.Actor, error) {
	principal, ok := PrincipalFromContext(ctx)
	if !ok {
		return app.Actor{}, status.Error(codes.Unauthenticated, "missing principal")
	}
	return app.Actor{
		UserID: principal.UserID,
		OrgID:  principal.OrgID,
		Role:   principal.Role,
	}, nil
}
