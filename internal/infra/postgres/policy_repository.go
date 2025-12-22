package postgres

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrgPolicyRepository struct {
	DB *pgxpool.Pool
}

func (r *OrgPolicyRepository) GetByOrgID(ctx context.Context, orgID uuid.UUID) (domain.OrgPolicy, error) {
	if r == nil || r.DB == nil {
		return domain.OrgPolicy{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT org_id, retention_interval, max_webhook_attempts, webhook_replay_window_seconds, api_rate_limit_per_min, api_key_rate_limit_per_min, created_at, updated_at
		FROM org_policies
		WHERE org_id=$1
	`, orgID)
	return scanPolicy(row)
}

func (r *OrgPolicyRepository) Upsert(ctx context.Context, policy domain.OrgPolicy) (domain.OrgPolicy, error) {
	if r == nil || r.DB == nil {
		return domain.OrgPolicy{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		INSERT INTO org_policies
			(org_id, retention_interval, max_webhook_attempts, webhook_replay_window_seconds, api_rate_limit_per_min, api_key_rate_limit_per_min, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (org_id) DO UPDATE
		SET retention_interval=$2,
			max_webhook_attempts=$3,
			webhook_replay_window_seconds=$4,
			api_rate_limit_per_min=$5,
			api_key_rate_limit_per_min=$6,
			updated_at=$8
		RETURNING org_id, retention_interval, max_webhook_attempts, webhook_replay_window_seconds, api_rate_limit_per_min, api_key_rate_limit_per_min, created_at, updated_at
	`, policy.OrgID, durationToInterval(policy.RetentionInterval), policy.MaxWebhookAttempts, policy.WebhookReplayWindowSeconds, policy.APIRateLimitPerMin, policy.APIKeyRateLimitPerMin, policy.CreatedAt, policy.UpdatedAt)
	updated, err := scanPolicy(row)
	if err != nil {
		return domain.OrgPolicy{}, TranslateError(err)
	}
	return updated, nil
}

func scanPolicy(row pgx.Row) (domain.OrgPolicy, error) {
	var policy domain.OrgPolicy
	var retention pgtype.Interval
	if err := row.Scan(&policy.OrgID, &retention, &policy.MaxWebhookAttempts, &policy.WebhookReplayWindowSeconds, &policy.APIRateLimitPerMin, &policy.APIKeyRateLimitPerMin, &policy.CreatedAt, &policy.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.OrgPolicy{}, domain.ErrNotFound
		}
		return domain.OrgPolicy{}, err
	}
	policy.RetentionInterval = intervalToDuration(retention)
	return policy, nil
}

func intervalToDuration(value pgtype.Interval) time.Duration {
	if !value.Valid {
		return 0
	}
	duration := time.Duration(value.Microseconds) * time.Microsecond
	duration += time.Duration(value.Days) * 24 * time.Hour
	duration += time.Duration(value.Months) * 30 * 24 * time.Hour
	return duration
}

func durationToInterval(duration time.Duration) pgtype.Interval {
	if duration <= 0 {
		return pgtype.Interval{Valid: false}
	}
	return pgtype.Interval{Microseconds: int64(duration / time.Microsecond), Valid: true}
}
