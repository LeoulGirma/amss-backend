package grpcapi

import (
	"context"
	"crypto/rsa"
	"strings"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/aeromaintain/amss/pkg/auth"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func authUnaryInterceptor(publicKey *rsa.PublicKey) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if publicKey == nil {
			return handler(ctx, req)
		}
		md, _ := metadata.FromIncomingContext(ctx)
		authHeader := ""
		if md != nil {
			values := md.Get("authorization")
			if len(values) == 0 {
				values = md.Get("Authorization")
			}
			if len(values) > 0 {
				authHeader = values[0]
			}
		}
		if authHeader == "" {
			return nil, status.Error(codes.Unauthenticated, "missing authorization")
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
		}
		claims, err := auth.ParseToken(parts[1], publicKey)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		if claims.TokenType != "access" {
			return nil, status.Error(codes.Unauthenticated, "invalid token type")
		}
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid subject")
		}
		orgID, err := uuid.Parse(claims.OrgID)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid org")
		}
		principal := Principal{
			UserID:  userID,
			OrgID:   orgID,
			Role:    domain.Role(claims.Role),
			Scopes:  claims.Scopes,
			TokenID: claims.ID,
		}
		return handler(WithPrincipal(ctx, principal), req)
	}
}
