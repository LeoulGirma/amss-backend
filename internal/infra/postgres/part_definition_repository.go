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

type PartDefinitionRepository struct {
	DB *pgxpool.Pool
}

func (r *PartDefinitionRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.PartDefinition, error) {
	if r == nil || r.DB == nil {
		return domain.PartDefinition{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, name, category, deleted_at, created_at, updated_at
		FROM part_definitions
		WHERE org_id=$1 AND id=$2 AND deleted_at IS NULL
	`, orgID, id)
	var def domain.PartDefinition
	if err := row.Scan(&def.ID, &def.OrgID, &def.Name, &def.Category, &def.DeletedAt, &def.CreatedAt, &def.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.PartDefinition{}, domain.ErrNotFound
		}
		return domain.PartDefinition{}, err
	}
	return def, nil
}

func (r *PartDefinitionRepository) Create(ctx context.Context, def domain.PartDefinition) (domain.PartDefinition, error) {
	if r == nil || r.DB == nil {
		return domain.PartDefinition{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		INSERT INTO part_definitions (id, org_id, name, category, created_at, updated_at, deleted_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, org_id, name, category, deleted_at, created_at, updated_at
	`, def.ID, def.OrgID, def.Name, def.Category, def.CreatedAt, def.UpdatedAt, def.DeletedAt)
	var created domain.PartDefinition
	if err := row.Scan(&created.ID, &created.OrgID, &created.Name, &created.Category, &created.DeletedAt, &created.CreatedAt, &created.UpdatedAt); err != nil {
		return domain.PartDefinition{}, TranslateError(err)
	}
	return created, nil
}

func (r *PartDefinitionRepository) Update(ctx context.Context, def domain.PartDefinition) (domain.PartDefinition, error) {
	if r == nil || r.DB == nil {
		return domain.PartDefinition{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		UPDATE part_definitions
		SET name=$1, category=$2, updated_at=$3
		WHERE org_id=$4 AND id=$5 AND deleted_at IS NULL
		RETURNING id, org_id, name, category, deleted_at, created_at, updated_at
	`, def.Name, def.Category, def.UpdatedAt, def.OrgID, def.ID)
	var updated domain.PartDefinition
	if err := row.Scan(&updated.ID, &updated.OrgID, &updated.Name, &updated.Category, &updated.DeletedAt, &updated.CreatedAt, &updated.UpdatedAt); err != nil {
		return domain.PartDefinition{}, TranslateError(err)
	}
	return updated, nil
}

func (r *PartDefinitionRepository) List(ctx context.Context, filter ports.PartDefinitionFilter) ([]domain.PartDefinition, error) {
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
	if strings.TrimSpace(filter.Name) != "" {
		args = append(args, "%"+strings.TrimSpace(filter.Name)+"%")
		clauses = append(clauses, "name ILIKE $"+itoa(len(args)))
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
		SELECT id, org_id, name, category, deleted_at, created_at, updated_at
		FROM part_definitions
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

	var defs []domain.PartDefinition
	for rows.Next() {
		var def domain.PartDefinition
		if err := rows.Scan(&def.ID, &def.OrgID, &def.Name, &def.Category, &def.DeletedAt, &def.CreatedAt, &def.UpdatedAt); err != nil {
			return nil, err
		}
		defs = append(defs, def)
	}
	return defs, rows.Err()
}

func (r *PartDefinitionRepository) SoftDelete(ctx context.Context, orgID, id uuid.UUID, at time.Time) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	cmd, err := r.DB.Exec(ctx, `
		UPDATE part_definitions
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
