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

type ComplianceRepository struct {
	DB *pgxpool.Pool
}

func (r *ComplianceRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.ComplianceItem, error) {
	if r == nil || r.DB == nil {
		return domain.ComplianceItem{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, task_id, description, result, sign_off_user_id, sign_off_time, deleted_at, created_at, updated_at
		FROM compliance_items
		WHERE org_id=$1 AND id=$2 AND deleted_at IS NULL
	`, orgID, id)
	var item domain.ComplianceItem
	if err := row.Scan(&item.ID, &item.OrgID, &item.TaskID, &item.Description, &item.Result, &item.SignOffUserID, &item.SignOffTime, &item.DeletedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.ComplianceItem{}, domain.ErrNotFound
		}
		return domain.ComplianceItem{}, err
	}
	return item, nil
}

func (r *ComplianceRepository) ListByTask(ctx context.Context, orgID, taskID uuid.UUID) ([]domain.ComplianceItem, error) {
	if r == nil || r.DB == nil {
		return nil, nil
	}
	rows, err := r.DB.Query(ctx, `
		SELECT id, org_id, task_id, description, result, sign_off_user_id, sign_off_time, deleted_at, created_at, updated_at
		FROM compliance_items
		WHERE org_id=$1 AND task_id=$2 AND deleted_at IS NULL
	`, orgID, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ComplianceItem
	for rows.Next() {
		var item domain.ComplianceItem
		if err := rows.Scan(&item.ID, &item.OrgID, &item.TaskID, &item.Description, &item.Result, &item.SignOffUserID, &item.SignOffTime, &item.DeletedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ComplianceRepository) List(ctx context.Context, filter ports.ComplianceFilter) ([]domain.ComplianceItem, error) {
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
	if filter.TaskID != nil {
		add("task_id=", *filter.TaskID)
	}
	if filter.Result != nil {
		add("result=", *filter.Result)
	}
	if filter.Signed != nil {
		if *filter.Signed {
			clauses = append(clauses, "sign_off_time IS NOT NULL")
		} else {
			clauses = append(clauses, "sign_off_time IS NULL")
		}
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
		SELECT id, org_id, task_id, description, result, sign_off_user_id, sign_off_time, deleted_at, created_at, updated_at
		FROM compliance_items
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

	var items []domain.ComplianceItem
	for rows.Next() {
		var item domain.ComplianceItem
		if err := rows.Scan(&item.ID, &item.OrgID, &item.TaskID, &item.Description, &item.Result, &item.SignOffUserID, &item.SignOffTime, &item.DeletedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ComplianceRepository) Create(ctx context.Context, item domain.ComplianceItem) error {
	if r == nil || r.DB == nil {
		return nil
	}
	_, err := r.DB.Exec(ctx, `
		INSERT INTO compliance_items (id, org_id, task_id, description, result, sign_off_user_id, sign_off_time, created_at, updated_at, deleted_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`, item.ID, item.OrgID, item.TaskID, item.Description, item.Result, item.SignOffUserID, item.SignOffTime, item.CreatedAt, item.UpdatedAt, item.DeletedAt)
	return TranslateError(err)
}

func (r *ComplianceRepository) Update(ctx context.Context, item domain.ComplianceItem) error {
	if r == nil || r.DB == nil {
		return nil
	}
	_, err := r.DB.Exec(ctx, `
		UPDATE compliance_items
		SET description=$1, result=$2, updated_at=$3
		WHERE org_id=$4 AND id=$5 AND deleted_at IS NULL
	`, item.Description, item.Result, item.UpdatedAt, item.OrgID, item.ID)
	return TranslateError(err)
}

func (r *ComplianceRepository) SignOff(ctx context.Context, orgID, id, userID uuid.UUID, at time.Time) error {
	if r == nil || r.DB == nil {
		return nil
	}
	_, err := r.DB.Exec(ctx, `
		UPDATE compliance_items
		SET sign_off_user_id=$1, sign_off_time=$2, updated_at=$2
		WHERE org_id=$3 AND id=$4 AND deleted_at IS NULL
	`, userID, at, orgID, id)
	return TranslateError(err)
}
