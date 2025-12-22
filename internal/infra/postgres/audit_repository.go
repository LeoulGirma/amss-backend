package postgres

import (
	"context"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository struct {
	DB *pgxpool.Pool
}

func (r *AuditRepository) Insert(ctx context.Context, entry domain.AuditLog) error {
	if r == nil || r.DB == nil {
		return nil
	}
	_, err := r.DB.Exec(ctx, `
		INSERT INTO audit_logs (id, org_id, entity_type, entity_id, action, user_id, request_id, ip_address, user_agent, entity_version, timestamp, details)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`, entry.ID, entry.OrgID, entry.EntityType, entry.EntityID, entry.Action, entry.UserID, entry.RequestID, entry.IPAddress, entry.UserAgent, entry.EntityVersion, entry.Timestamp, entry.Details)
	return TranslateError(err)
}
