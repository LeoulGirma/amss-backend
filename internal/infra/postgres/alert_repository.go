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

type AlertRepository struct {
	DB *pgxpool.Pool
}

func (r *AlertRepository) Create(ctx context.Context, alert domain.Alert) (domain.Alert, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO alerts
			(id, org_id, level, category, title, description, entity_type, entity_id,
			 threshold_value, current_value, acknowledged, resolved, escalation_level,
			 auto_escalate_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id, org_id, level, category, title, description, entity_type, entity_id,
		          threshold_value, current_value, acknowledged, acknowledged_by, acknowledged_at,
		          resolved, resolved_at, escalation_level, auto_escalate_at, created_at
	`, alert.ID, alert.OrgID, alert.Level, alert.Category, alert.Title, alert.Description,
		alert.EntityType, alert.EntityID, alert.ThresholdValue, alert.CurrentValue,
		alert.Acknowledged, alert.Resolved, alert.EscalationLevel, alert.AutoEscalateAt, alert.CreatedAt)
	return scanAlert(row)
}

func (r *AlertRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.Alert, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, level, category, title, description, entity_type, entity_id,
		       threshold_value, current_value, acknowledged, acknowledged_by, acknowledged_at,
		       resolved, resolved_at, escalation_level, auto_escalate_at, created_at
		FROM alerts
		WHERE org_id=$1 AND id=$2
	`, orgID, id)
	return scanAlert(row)
}

func (r *AlertRepository) List(ctx context.Context, filter ports.AlertFilter) ([]domain.Alert, error) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 6)
	add := func(condition string, value any) {
		args = append(args, value)
		clauses = append(clauses, condition+"$"+itoa(len(args)))
	}
	if filter.OrgID != nil {
		add("org_id=", *filter.OrgID)
	}
	if filter.Level != nil {
		add("level=", *filter.Level)
	}
	if filter.Category != "" {
		add("category=", filter.Category)
	}
	if filter.Resolved != nil {
		add("resolved=", *filter.Resolved)
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
		SELECT id, org_id, level, category, title, description, entity_type, entity_id,
		       threshold_value, current_value, acknowledged, acknowledged_by, acknowledged_at,
		       resolved, resolved_at, escalation_level, auto_escalate_at, created_at
		FROM alerts
		WHERE 1=1`
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

	var items []domain.Alert
	for rows.Next() {
		a, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

func (r *AlertRepository) Acknowledge(ctx context.Context, orgID, id, userID uuid.UUID, at time.Time) error {
	cmd, err := r.DB.Exec(ctx, `
		UPDATE alerts
		SET acknowledged=true, acknowledged_by=$1, acknowledged_at=$2
		WHERE org_id=$3 AND id=$4 AND acknowledged=false
	`, userID, at, orgID, id)
	if err != nil {
		return TranslateError(err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AlertRepository) Resolve(ctx context.Context, orgID, id uuid.UUID, at time.Time) error {
	cmd, err := r.DB.Exec(ctx, `
		UPDATE alerts
		SET resolved=true, resolved_at=$1
		WHERE org_id=$2 AND id=$3 AND resolved=false
	`, at, orgID, id)
	if err != nil {
		return TranslateError(err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AlertRepository) CountUnresolved(ctx context.Context, orgID uuid.UUID) (int, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM alerts WHERE org_id=$1 AND resolved=false
	`, orgID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func scanAlert(row pgx.Row) (domain.Alert, error) {
	var a domain.Alert
	if err := row.Scan(&a.ID, &a.OrgID, &a.Level, &a.Category, &a.Title, &a.Description,
		&a.EntityType, &a.EntityID, &a.ThresholdValue, &a.CurrentValue,
		&a.Acknowledged, &a.AcknowledgedBy, &a.AcknowledgedAt,
		&a.Resolved, &a.ResolvedAt, &a.EscalationLevel, &a.AutoEscalateAt, &a.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.Alert{}, domain.ErrNotFound
		}
		return domain.Alert{}, err
	}
	return a, nil
}
