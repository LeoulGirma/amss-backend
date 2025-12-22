package postgres

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditQueryRepository struct {
	DB *pgxpool.Pool
}

func (r *AuditQueryRepository) List(ctx context.Context, filter ports.AuditLogFilter) ([]domain.AuditLog, error) {
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
	if filter.EntityType != "" {
		add("entity_type=", filter.EntityType)
	}
	if filter.EntityID != nil {
		add("entity_id=", *filter.EntityID)
	}
	if filter.UserID != nil {
		add("user_id=", *filter.UserID)
	}
	if filter.From != nil {
		add("timestamp >= ", *filter.From)
	}
	if filter.To != nil {
		add("timestamp <= ", *filter.To)
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
		SELECT id, org_id, entity_type, entity_id, action, user_id, request_id, ip_address, user_agent, entity_version, timestamp, details
		FROM audit_logs`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	args = append(args, limit, offset)
	query += " ORDER BY timestamp DESC LIMIT $" + itoa(len(args)-1) + " OFFSET $" + itoa(len(args))

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.AuditLog
	for rows.Next() {
		var entry domain.AuditLog
		var details []byte
		if err := rows.Scan(&entry.ID, &entry.OrgID, &entry.EntityType, &entry.EntityID, &entry.Action, &entry.UserID, &entry.RequestID, &entry.IPAddress, &entry.UserAgent, &entry.EntityVersion, &entry.Timestamp, &details); err != nil {
			return nil, err
		}
		if len(details) > 0 {
			var payload map[string]any
			if err := json.Unmarshal(details, &payload); err == nil {
				entry.Details = payload
			}
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
