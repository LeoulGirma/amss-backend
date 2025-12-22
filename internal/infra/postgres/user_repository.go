package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	DB *pgxpool.Pool
}

func (r *UserRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.User, error) {
	if r == nil || r.DB == nil {
		return domain.User{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, email, role, password_hash, last_login, created_at, updated_at, deleted_at
		FROM users
		WHERE org_id=$1 AND id=$2 AND deleted_at IS NULL
	`, orgID, id)
	return scanUser(row)
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	if r == nil || r.DB == nil {
		return domain.User{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		INSERT INTO users
			(id, org_id, email, role, password_hash, last_login, created_at, updated_at, deleted_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, org_id, email, role, password_hash, last_login, created_at, updated_at, deleted_at
	`, user.ID, user.OrgID, user.Email, user.Role, user.PasswordHash, user.LastLogin, user.CreatedAt, user.UpdatedAt, user.DeletedAt)
	created, err := scanUser(row)
	if err != nil {
		return domain.User{}, TranslateError(err)
	}
	return created, nil
}

func (r *UserRepository) Update(ctx context.Context, user domain.User) (domain.User, error) {
	if r == nil || r.DB == nil {
		return domain.User{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		UPDATE users
		SET email=$1, role=$2, password_hash=$3, last_login=$4, updated_at=$5
		WHERE org_id=$6 AND id=$7 AND deleted_at IS NULL
		RETURNING id, org_id, email, role, password_hash, last_login, created_at, updated_at, deleted_at
	`, user.Email, user.Role, user.PasswordHash, user.LastLogin, user.UpdatedAt, user.OrgID, user.ID)
	updated, err := scanUser(row)
	if err != nil {
		return domain.User{}, TranslateError(err)
	}
	return updated, nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, orgID, id uuid.UUID, at time.Time) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	cmd, err := r.DB.Exec(ctx, `
		UPDATE users
		SET deleted_at=$1, updated_at=$1
		WHERE org_id=$2 AND id=$3 AND deleted_at IS NULL
	`, at, orgID, id)
	if err != nil {
		return TranslateError(err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepository) List(ctx context.Context, filter ports.UserFilter) ([]domain.User, error) {
	if r == nil || r.DB == nil {
		return nil, nil
	}
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 6)
	add := func(condition string, value any) {
		args = append(args, value)
		clauses = append(clauses, condition+"$"+itoa(len(args)))
	}
	if filter.OrgID != nil {
		add("org_id=", *filter.OrgID)
	}
	if filter.Role != nil {
		add("role=", *filter.Role)
	}
	if filter.Email != "" {
		add("email ILIKE ", "%"+filter.Email+"%")
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, org_id, email, role, password_hash, last_login, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL`
	if len(clauses) > 0 {
		query += " AND " + strings.Join(clauses, " AND ")
	}
	args = append(args, limit, offset)
	query += " ORDER BY email ASC LIMIT $" + itoa(len(args)-1) + " OFFSET $" + itoa(len(args))

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func scanUser(row pgx.Row) (domain.User, error) {
	var user domain.User
	if err := row.Scan(&user.ID, &user.OrgID, &user.Email, &user.Role, &user.PasswordHash, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return user, nil
}
