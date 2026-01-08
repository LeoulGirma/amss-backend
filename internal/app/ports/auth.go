package ports

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type AuthUser struct {
	ID           uuid.UUID
	OrgID        uuid.UUID
	Email        string
	Role         domain.Role
	PasswordHash string
	LastLogin    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// OrgInfo contains organization details for email lookup
type OrgInfo struct {
	OrgID   uuid.UUID
	OrgName string
}

type AuthRepository interface {
	GetUserByEmail(ctx context.Context, orgID uuid.UUID, email string) (AuthUser, error)
	GetUserByID(ctx context.Context, orgID uuid.UUID, userID uuid.UUID) (AuthUser, error)
	UpdateLastLogin(ctx context.Context, orgID, userID uuid.UUID, at time.Time) error
	LookupOrgsByEmail(ctx context.Context, email string) ([]OrgInfo, error)
}

type RefreshToken struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	UserID      uuid.UUID
	TokenHash   string
	TokenID     string
	ExpiresAt   time.Time
	RevokedAt   *time.Time
	RotatedFrom *uuid.UUID
	CreatedAt   time.Time
}

type RefreshTokenStore interface {
	Insert(ctx context.Context, token RefreshToken) error
	GetByTokenID(ctx context.Context, orgID uuid.UUID, tokenID string) (RefreshToken, error)
	Revoke(ctx context.Context, orgID uuid.UUID, tokenID string, revokedAt time.Time) error
	Rotate(ctx context.Context, oldToken RefreshToken, newToken RefreshToken, revokedAt time.Time) error
}
