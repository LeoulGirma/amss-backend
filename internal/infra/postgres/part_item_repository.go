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

type PartItemRepository struct {
	DB *pgxpool.Pool
}

func (r *PartItemRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.PartItem, error) {
	if r == nil || r.DB == nil {
		return domain.PartItem{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, part_definition_id, serial_number, status, expiry_date, deleted_at, created_at, updated_at
		FROM part_items
		WHERE org_id=$1 AND id=$2 AND deleted_at IS NULL
	`, orgID, id)
	return scanPartItem(row)
}

func (r *PartItemRepository) Create(ctx context.Context, item domain.PartItem) (domain.PartItem, error) {
	if r == nil || r.DB == nil {
		return domain.PartItem{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		INSERT INTO part_items (id, org_id, part_definition_id, serial_number, status, expiry_date, created_at, updated_at, deleted_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, org_id, part_definition_id, serial_number, status, expiry_date, deleted_at, created_at, updated_at
	`, item.ID, item.OrgID, item.DefinitionID, item.SerialNumber, item.Status, item.ExpiryDate, item.CreatedAt, item.UpdatedAt, item.DeletedAt)
	created, err := scanPartItem(row)
	if err != nil {
		return domain.PartItem{}, TranslateError(err)
	}
	return created, nil
}

func (r *PartItemRepository) Update(ctx context.Context, item domain.PartItem) (domain.PartItem, error) {
	if r == nil || r.DB == nil {
		return domain.PartItem{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		UPDATE part_items
		SET status=$1, expiry_date=$2, updated_at=$3
		WHERE org_id=$4 AND id=$5 AND deleted_at IS NULL
		RETURNING id, org_id, part_definition_id, serial_number, status, expiry_date, deleted_at, created_at, updated_at
	`, item.Status, item.ExpiryDate, item.UpdatedAt, item.OrgID, item.ID)
	updated, err := scanPartItem(row)
	if err != nil {
		return domain.PartItem{}, TranslateError(err)
	}
	return updated, nil
}

func (r *PartItemRepository) List(ctx context.Context, filter ports.PartItemFilter) ([]domain.PartItem, error) {
	if r == nil || r.DB == nil {
		return nil, nil
	}
	clauses := make([]string, 0, 6)
	args := make([]any, 0, 8)
	add := func(condition string, value any) {
		args = append(args, value)
		clauses = append(clauses, condition+"$"+itoa(len(args)))
	}
	if filter.OrgID != nil {
		add("org_id=", *filter.OrgID)
	}
	if filter.DefinitionID != nil {
		add("part_definition_id=", *filter.DefinitionID)
	}
	if filter.Status != nil {
		add("status=", *filter.Status)
	}
	if filter.ExpiryBefore != nil {
		add("expiry_date <= ", *filter.ExpiryBefore)
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
		SELECT id, org_id, part_definition_id, serial_number, status, expiry_date, deleted_at, created_at, updated_at
		FROM part_items
		WHERE deleted_at IS NULL`
	if len(clauses) > 0 {
		query += " AND " + strings.Join(clauses, " AND ")
	}
	args = append(args, limit, offset)
	query += " ORDER BY created_at DESC LIMIT $" + itoa(len(args)-1) + " OFFSET $" + itoa(len(args))

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.PartItem
	for rows.Next() {
		item, err := scanPartItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PartItemRepository) SoftDelete(ctx context.Context, orgID, id uuid.UUID, at time.Time) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	cmd, err := r.DB.Exec(ctx, `
		UPDATE part_items
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

func (r *PartItemRepository) UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status domain.PartItemStatus, now time.Time) error {
	if r == nil || r.DB == nil {
		return nil
	}
	_, err := r.DB.Exec(ctx, `
		UPDATE part_items
		SET status=$1, updated_at=$2
		WHERE org_id=$3 AND id=$4 AND deleted_at IS NULL
	`, status, now, orgID, id)
	return TranslateError(err)
}

func scanPartItem(row pgx.Row) (domain.PartItem, error) {
	var item domain.PartItem
	if err := row.Scan(&item.ID, &item.OrgID, &item.DefinitionID, &item.SerialNumber, &item.Status, &item.ExpiryDate, &item.DeletedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.PartItem{}, domain.ErrNotFound
		}
		return domain.PartItem{}, err
	}
	return item, nil
}

func (r *PartItemRepository) GetBySerialNumber(ctx context.Context, orgID uuid.UUID, serial string) (domain.PartItem, error) {
	if r == nil || r.DB == nil {
		return domain.PartItem{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, part_definition_id, serial_number, status, expiry_date, deleted_at, created_at, updated_at
		FROM part_items
		WHERE org_id=$1 AND serial_number=$2 AND deleted_at IS NULL
	`, orgID, serial)
	var item domain.PartItem
	if err := row.Scan(&item.ID, &item.OrgID, &item.DefinitionID, &item.SerialNumber, &item.Status, &item.ExpiryDate, &item.DeletedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.PartItem{}, domain.ErrNotFound
		}
		return domain.PartItem{}, err
	}
	return item, nil
}
