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

type OrganizationRepository struct {
	DB *pgxpool.Pool
}

func (r *OrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Organization, error) {
	if r == nil || r.DB == nil {
		return domain.Organization{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, name, created_at, updated_at, deleted_at
		FROM organizations
		WHERE id=$1 AND deleted_at IS NULL
	`, id)
	return scanOrganization(row)
}

func (r *OrganizationRepository) Create(ctx context.Context, org domain.Organization) (domain.Organization, error) {
	if r == nil || r.DB == nil {
		return domain.Organization{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		INSERT INTO organizations (id, name, created_at, updated_at, deleted_at)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, name, created_at, updated_at, deleted_at
	`, org.ID, org.Name, org.CreatedAt, org.UpdatedAt, org.DeletedAt)
	created, err := scanOrganization(row)
	if err != nil {
		return domain.Organization{}, TranslateError(err)
	}
	return created, nil
}

func (r *OrganizationRepository) Update(ctx context.Context, org domain.Organization) (domain.Organization, error) {
	if r == nil || r.DB == nil {
		return domain.Organization{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		UPDATE organizations
		SET name=$1, updated_at=$2
		WHERE id=$3 AND deleted_at IS NULL
		RETURNING id, name, created_at, updated_at, deleted_at
	`, org.Name, org.UpdatedAt, org.ID)
	updated, err := scanOrganization(row)
	if err != nil {
		return domain.Organization{}, TranslateError(err)
	}
	return updated, nil
}

func (r *OrganizationRepository) SoftDelete(ctx context.Context, id uuid.UUID, at time.Time) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	cmd, err := r.DB.Exec(ctx, `
		UPDATE organizations
		SET deleted_at=$1, updated_at=$1
		WHERE id=$2 AND deleted_at IS NULL
	`, at, id)
	if err != nil {
		return TranslateError(err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *OrganizationRepository) List(ctx context.Context, filter ports.OrganizationFilter) ([]domain.Organization, error) {
	if r == nil || r.DB == nil {
		return nil, nil
	}
	clauses := make([]string, 0, 2)
	args := make([]any, 0, 4)
	add := func(condition string, value any) {
		args = append(args, value)
		clauses = append(clauses, condition+"$"+itoa(len(args)))
	}
	if filter.Name != "" {
		add("name ILIKE ", "%"+filter.Name+"%")
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
		SELECT id, name, created_at, updated_at, deleted_at
		FROM organizations
		WHERE deleted_at IS NULL`
	if len(clauses) > 0 {
		query += " AND " + strings.Join(clauses, " AND ")
	}
	args = append(args, limit, offset)
	query += " ORDER BY name ASC LIMIT $" + itoa(len(args)-1) + " OFFSET $" + itoa(len(args))

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []domain.Organization
	for rows.Next() {
		org, err := scanOrganization(rows)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, rows.Err()
}

func scanOrganization(row pgx.Row) (domain.Organization, error) {
	var org domain.Organization
	if err := row.Scan(&org.ID, &org.Name, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.Organization{}, domain.ErrNotFound
		}
		return domain.Organization{}, err
	}
	return org, nil
}
