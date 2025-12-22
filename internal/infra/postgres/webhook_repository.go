package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookRepository struct {
	DB *pgxpool.Pool
}

func (r *WebhookRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.Webhook, error) {
	if r == nil || r.DB == nil {
		return domain.Webhook{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, url, events, secret, created_at, updated_at
		FROM webhooks
		WHERE org_id=$1 AND id=$2
	`, orgID, id)
	return scanWebhook(row)
}

func (r *WebhookRepository) Create(ctx context.Context, webhook domain.Webhook) (domain.Webhook, error) {
	if r == nil || r.DB == nil {
		return domain.Webhook{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		INSERT INTO webhooks
			(id, org_id, url, events, secret, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, org_id, url, events, secret, created_at, updated_at
	`, webhook.ID, webhook.OrgID, webhook.URL, webhook.Events, webhook.Secret, webhook.CreatedAt, webhook.UpdatedAt)
	created, err := scanWebhook(row)
	if err != nil {
		return domain.Webhook{}, TranslateError(err)
	}
	return created, nil
}

func (r *WebhookRepository) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	cmd, err := r.DB.Exec(ctx, `
		DELETE FROM webhooks
		WHERE org_id=$1 AND id=$2
	`, orgID, id)
	if err != nil {
		return TranslateError(err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *WebhookRepository) List(ctx context.Context, orgID uuid.UUID) ([]domain.Webhook, error) {
	if r == nil || r.DB == nil {
		return nil, nil
	}
	rows, err := r.DB.Query(ctx, `
		SELECT id, org_id, url, events, secret, created_at, updated_at
		FROM webhooks
		WHERE org_id=$1
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hooks []domain.Webhook
	for rows.Next() {
		hook, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, hook)
	}
	return hooks, rows.Err()
}

func (r *WebhookRepository) ListByEvent(ctx context.Context, orgID uuid.UUID, eventType string) ([]domain.Webhook, error) {
	if r == nil || r.DB == nil {
		return nil, nil
	}
	rows, err := r.DB.Query(ctx, `
		SELECT id, org_id, url, events, secret, created_at, updated_at
		FROM webhooks
		WHERE org_id=$1 AND $2 = ANY(events)
		ORDER BY created_at DESC
	`, orgID, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hooks []domain.Webhook
	for rows.Next() {
		hook, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, hook)
	}
	return hooks, rows.Err()
}

type WebhookDeliveryRepository struct {
	DB *pgxpool.Pool
}

func (r *WebhookDeliveryRepository) Create(ctx context.Context, delivery domain.WebhookDelivery) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	_, err := r.DB.Exec(ctx, `
		INSERT INTO webhook_deliveries
			(id, org_id, webhook_id, event_id, attempt_count, last_error, next_attempt_at, status, last_response_code, last_response_body, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`, delivery.ID, delivery.OrgID, delivery.WebhookID, delivery.EventID, delivery.AttemptCount, delivery.LastError, delivery.NextAttemptAt, delivery.Status, delivery.LastResponseCode, delivery.LastResponseBody, delivery.CreatedAt, delivery.UpdatedAt)
	return TranslateError(err)
}

func (r *WebhookDeliveryRepository) ClaimPending(ctx context.Context, limit int, lockUntil time.Time) ([]domain.WebhookDelivery, error) {
	if r == nil || r.DB == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 50
	}
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id, org_id, webhook_id, event_id, attempt_count, last_error, next_attempt_at, status, last_response_code, last_response_body, created_at, updated_at
		FROM webhook_deliveries
		WHERE status='pending' AND next_attempt_at <= now()
		ORDER BY next_attempt_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []domain.WebhookDelivery
	for rows.Next() {
		delivery, err := scanWebhookDelivery(rows)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, delivery)
	}
	if len(deliveries) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return deliveries, nil
	}

	ids := make([]uuid.UUID, 0, len(deliveries))
	for _, delivery := range deliveries {
		ids = append(ids, delivery.ID)
	}
	_, err = tx.Exec(ctx, `
		UPDATE webhook_deliveries
		SET next_attempt_at=$1, updated_at=$1
		WHERE id = ANY($2)
	`, lockUntil, ids)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return deliveries, nil
}

func (r *WebhookDeliveryRepository) Update(ctx context.Context, delivery domain.WebhookDelivery) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	_, err := r.DB.Exec(ctx, `
		UPDATE webhook_deliveries
		SET attempt_count=$1, last_error=$2, next_attempt_at=$3, status=$4, last_response_code=$5, last_response_body=$6, updated_at=$7
		WHERE org_id=$8 AND id=$9
	`, delivery.AttemptCount, delivery.LastError, delivery.NextAttemptAt, delivery.Status, delivery.LastResponseCode, delivery.LastResponseBody, delivery.UpdatedAt, delivery.OrgID, delivery.ID)
	return TranslateError(err)
}

func scanWebhook(row pgx.Row) (domain.Webhook, error) {
	var webhook domain.Webhook
	var events []string
	if err := row.Scan(&webhook.ID, &webhook.OrgID, &webhook.URL, &events, &webhook.Secret, &webhook.CreatedAt, &webhook.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.Webhook{}, domain.ErrNotFound
		}
		return domain.Webhook{}, err
	}
	webhook.Events = normalizeEvents(events)
	return webhook, nil
}

func scanWebhookDelivery(row pgx.Row) (domain.WebhookDelivery, error) {
	var delivery domain.WebhookDelivery
	if err := row.Scan(&delivery.ID, &delivery.OrgID, &delivery.WebhookID, &delivery.EventID, &delivery.AttemptCount, &delivery.LastError, &delivery.NextAttemptAt, &delivery.Status, &delivery.LastResponseCode, &delivery.LastResponseBody, &delivery.CreatedAt, &delivery.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.WebhookDelivery{}, domain.ErrNotFound
		}
		return domain.WebhookDelivery{}, err
	}
	return delivery, nil
}

func normalizeEvents(events []string) []string {
	out := make([]string, 0, len(events))
	for _, event := range events {
		event = strings.TrimSpace(event)
		if event == "" {
			continue
		}
		out = append(out, event)
	}
	return out
}
