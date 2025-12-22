package services

import (
	"context"
	"crypto/rsa"
	"time"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/aeromaintain/amss/pkg/auth"
	"github.com/google/uuid"
)

type AuthService struct {
	Users         ports.AuthRepository
	RefreshTokens ports.RefreshTokenStore
	PrivateKey    *rsa.PrivateKey
	PublicKey     *rsa.PublicKey
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	Clock         app.Clock
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Duration
}

func (s *AuthService) Login(ctx context.Context, orgID uuid.UUID, email, password string) (TokenPair, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	user, err := s.Users.GetUserByEmail(ctx, orgID, email)
	if err != nil {
		return TokenPair{}, err
	}
	if user.DeletedAt != nil {
		return TokenPair{}, domain.ErrUnauthorized
	}
	if err := auth.ComparePassword(user.PasswordHash, password); err != nil {
		return TokenPair{}, domain.ErrUnauthorized
	}

	now := s.Clock.Now()
	pair, err := auth.GenerateTokenPair(user.ID.String(), user.OrgID.String(), string(user.Role), nil, now, s.AccessTTL, s.RefreshTTL, s.PrivateKey)
	if err != nil {
		return TokenPair{}, err
	}

	refreshHash := auth.HashToken(pair.RefreshToken)
	refreshToken := ports.RefreshToken{
		ID:        uuid.New(),
		OrgID:     user.OrgID,
		UserID:    user.ID,
		TokenHash: refreshHash,
		TokenID:   pair.RefreshTokenID,
		ExpiresAt: pair.RefreshExpiresAt,
		CreatedAt: now,
	}
	if err := s.RefreshTokens.Insert(ctx, refreshToken); err != nil {
		return TokenPair{}, err
	}
	_ = s.Users.UpdateLastLogin(ctx, user.OrgID, user.ID, now)

	return TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.AccessExpiresAt.Sub(now),
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	claims, err := auth.ParseToken(refreshToken, s.PublicKey)
	if err != nil {
		return TokenPair{}, domain.ErrUnauthorized
	}
	if claims.TokenType != "refresh" {
		return TokenPair{}, domain.ErrUnauthorized
	}
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		return TokenPair{}, domain.ErrUnauthorized
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return TokenPair{}, domain.ErrUnauthorized
	}

	existing, err := s.RefreshTokens.GetByTokenID(ctx, orgID, claims.ID)
	if err != nil {
		return TokenPair{}, err
	}
	if existing.RevokedAt != nil || existing.ExpiresAt.Before(s.Clock.Now()) {
		return TokenPair{}, domain.ErrUnauthorized
	}
	if !auth.SecureCompare(existing.TokenHash, auth.HashToken(refreshToken)) {
		return TokenPair{}, domain.ErrUnauthorized
	}

	now := s.Clock.Now()
	pair, err := auth.GenerateTokenPair(userID.String(), orgID.String(), claims.Role, claims.Scopes, now, s.AccessTTL, s.RefreshTTL, s.PrivateKey)
	if err != nil {
		return TokenPair{}, err
	}

	newToken := ports.RefreshToken{
		ID:          uuid.New(),
		OrgID:       orgID,
		UserID:      userID,
		TokenHash:   auth.HashToken(pair.RefreshToken),
		TokenID:     pair.RefreshTokenID,
		ExpiresAt:   pair.RefreshExpiresAt,
		RotatedFrom: &existing.ID,
		CreatedAt:   now,
	}
	if err := s.RefreshTokens.Rotate(ctx, existing, newToken, now); err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.AccessExpiresAt.Sub(now),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	claims, err := auth.ParseToken(refreshToken, s.PublicKey)
	if err != nil {
		return domain.ErrUnauthorized
	}
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		return domain.ErrUnauthorized
	}
	return s.RefreshTokens.Revoke(ctx, orgID, claims.ID, s.Clock.Now())
}
