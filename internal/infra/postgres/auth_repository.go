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

type AuthRepository struct {
	DB *pgxpool.Pool
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, orgID uuid.UUID, email string) (ports.AuthUser, error) {
	if r == nil || r.DB == nil {
		return ports.AuthUser{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, email, role, password_hash, deleted_at
		FROM users
		WHERE org_id=$1 AND email=$2 AND deleted_at IS NULL
	`, orgID, email)

	var user ports.AuthUser
	var role string
	if err := row.Scan(&user.ID, &user.OrgID, &user.Email, &role, &user.PasswordHash, &user.DeletedAt); err != nil {
		if err == pgx.ErrNoRows {
			return ports.AuthUser{}, domain.ErrUnauthorized
		}
		return ports.AuthUser{}, err
	}
	user.Role = domain.Role(role)
	return user, nil
}

func (r *AuthRepository) UpdateLastLogin(ctx context.Context, orgID, userID uuid.UUID, at time.Time) error {
	if r == nil || r.DB == nil {
		return nil
	}
	_, err := r.DB.Exec(ctx, `
		UPDATE users
		SET last_login=$1, updated_at=$1
		WHERE org_id=$2 AND id=$3
	`, at, orgID, userID)
	return err
}
