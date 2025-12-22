package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxRepository struct {
	DB *pgxpool.Pool
}

func (r *OutboxRepository) Enqueue(ctx context.Context, orgID uuid.UUID, eventType string, aggregateType string, aggregateID uuid.UUID, payload map[string]any, dedupeKey string) error {
	if r == nil || r.DB == nil {
		return nil
	}
	var jsonValue []byte
	if payload != nil {
		if data, err := json.Marshal(payload); err == nil {
			jsonValue = data
		}
	}
	_, err := r.DB.Exec(ctx, `
		INSERT INTO outbox_events (id, org_id, aggregate_type, aggregate_id, event_type, payload, dedupe_key, attempt_count, next_attempt_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,0,$8,$9)
	`, uuid.New(), orgID, aggregateType, aggregateID, eventType, jsonValue, dedupeKey, time.Now().UTC(), time.Now().UTC())
	return TranslateError(err)
}

func (r *OutboxRepository) LockPending(ctx context.Context, workerID string, limit int) ([]ports.OutboxEvent, error) {
	if r == nil || r.DB == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 100
	}
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id, org_id, aggregate_type, aggregate_id, event_type, payload, dedupe_key, attempt_count, last_error, next_attempt_at, locked_at, locked_by, created_at, processed_at
		FROM outbox_events
		WHERE processed_at IS NULL AND next_attempt_at <= now()
		ORDER BY created_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []ports.OutboxEvent
	for rows.Next() {
		event, err := scanOutbox(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if len(events) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return events, nil
	}

	ids := make([]uuid.UUID, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
	}
	_, err = tx.Exec(ctx, `
		UPDATE outbox_events
		SET locked_at=$1, locked_by=$2
		WHERE id = ANY($3)
	`, time.Now().UTC(), workerID, ids)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *OutboxRepository) MarkProcessed(ctx context.Context, id uuid.UUID, processedAt time.Time) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	_, err := r.DB.Exec(ctx, `
		UPDATE outbox_events
		SET processed_at=$1, locked_at=NULL, locked_by=NULL
		WHERE id=$2
	`, processedAt, id)
	return TranslateError(err)
}

func (r *OutboxRepository) ScheduleRetry(ctx context.Context, id uuid.UUID, attempt int, nextAttemptAt time.Time, lastError string) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	_, err := r.DB.Exec(ctx, `
		UPDATE outbox_events
		SET attempt_count=$1, last_error=$2, next_attempt_at=$3, locked_at=NULL, locked_by=NULL
		WHERE id=$4
	`, attempt, lastError, nextAttemptAt, id)
	return TranslateError(err)
}

func (r *OutboxRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (ports.OutboxEvent, error) {
	if r == nil || r.DB == nil {
		return ports.OutboxEvent{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, aggregate_type, aggregate_id, event_type, payload, dedupe_key, attempt_count, last_error, next_attempt_at, locked_at, locked_by, created_at, processed_at
		FROM outbox_events
		WHERE org_id=$1 AND id=$2
	`, orgID, id)
	return scanOutbox(row)
}

func scanOutbox(row pgx.Row) (ports.OutboxEvent, error) {
	var event ports.OutboxEvent
	var payload []byte
	if err := row.Scan(&event.ID, &event.OrgID, &event.AggregateType, &event.AggregateID, &event.EventType, &payload, &event.DedupeKey, &event.AttemptCount, &event.LastError, &event.NextAttemptAt, &event.LockedAt, &event.LockedBy, &event.CreatedAt, &event.ProcessedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.OutboxEvent{}, domain.ErrNotFound
		}
		return ports.OutboxEvent{}, err
	}
	if len(payload) > 0 {
		var data map[string]any
		if err := json.Unmarshal(payload, &data); err == nil {
			event.Payload = data
		}
	}
	return event, nil
}
