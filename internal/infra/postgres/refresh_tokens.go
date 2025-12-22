package postgres

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepository struct {
	DB *pgxpool.Pool
}

func (r *RefreshTokenRepository) Insert(ctx context.Context, token ports.RefreshToken) error {
	if r == nil || r.DB == nil {
		return nil
	}
	_, err := r.DB.Exec(ctx, `
		INSERT INTO refresh_tokens (id, org_id, user_id, token_hash, token_id, expires_at, revoked_at, rotated_from, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, token.ID, token.OrgID, token.UserID, token.TokenHash, token.TokenID, token.ExpiresAt, token.RevokedAt, token.RotatedFrom, token.CreatedAt)
	return err
}

func (r *RefreshTokenRepository) GetByTokenID(ctx context.Context, orgID uuid.UUID, tokenID string) (ports.RefreshToken, error) {
	if r == nil || r.DB == nil {
		return ports.RefreshToken{}, domain.ErrUnauthorized
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, user_id, token_hash, token_id, expires_at, revoked_at, rotated_from, created_at
		FROM refresh_tokens
		WHERE org_id=$1 AND token_id=$2
	`, orgID, tokenID)
	var token ports.RefreshToken
	if err := row.Scan(&token.ID, &token.OrgID, &token.UserID, &token.TokenHash, &token.TokenID, &token.ExpiresAt, &token.RevokedAt, &token.RotatedFrom, &token.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return ports.RefreshToken{}, domain.ErrUnauthorized
		}
		return ports.RefreshToken{}, err
	}
	return token, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, orgID uuid.UUID, tokenID string, revokedAt time.Time) error {
	if r == nil || r.DB == nil {
		return nil
	}
	cmd, err := r.DB.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at=$1
		WHERE org_id=$2 AND token_id=$3 AND revoked_at IS NULL
	`, revokedAt, orgID, tokenID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrUnauthorized
	}
	return nil
}

func (r *RefreshTokenRepository) Rotate(ctx context.Context, oldToken ports.RefreshToken, newToken ports.RefreshToken, revokedAt time.Time) error {
	if r == nil || r.DB == nil {
		return nil
	}
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	cmd, err := tx.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at=$1
		WHERE id=$2 AND org_id=$3 AND revoked_at IS NULL
	`, revokedAt, oldToken.ID, oldToken.OrgID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrUnauthorized
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO refresh_tokens (id, org_id, user_id, token_hash, token_id, expires_at, revoked_at, rotated_from, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, newToken.ID, newToken.OrgID, newToken.UserID, newToken.TokenHash, newToken.TokenID, newToken.ExpiresAt, newToken.RevokedAt, newToken.RotatedFrom, newToken.CreatedAt)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
