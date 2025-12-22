package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/aeromaintain/amss/pkg/observability"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

const (
	outboxStream = "amss.events"
	outboxDLQ    = "amss.dlq"
)

type OutboxPublisher struct {
	Outbox      ports.OutboxRepository
	Webhooks    ports.WebhookRepository
	Deliveries  ports.WebhookDeliveryRepository
	Redis       *redis.Client
	Logger      zerolog.Logger
	WorkerID    string
	MaxAttempts int
}

func (p *OutboxPublisher) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.processOnce(ctx)
		}
	}
}

func (p *OutboxPublisher) processOnce(ctx context.Context) {
	if p.Outbox == nil || p.Redis == nil {
		return
	}
	observability.IncJobRun("outbox_publisher")
	events, err := p.Outbox.LockPending(ctx, p.WorkerID, 50)
	if err != nil {
		observability.IncJobFailure("outbox_publisher")
		p.Logger.Error().Err(err).Msg("outbox lock failed")
		return
	}
	for _, event := range events {
		if err := p.publish(ctx, event); err != nil {
			observability.IncOutboxFailed(event.EventType)
			p.handleFailure(ctx, event, err)
			continue
		}
		observability.IncOutboxProcessed(event.EventType)
		if p.Webhooks != nil && p.Deliveries != nil {
			if err := p.enqueueWebhooks(ctx, event); err != nil {
				p.Logger.Error().Err(err).Str("event_id", event.ID.String()).Msg("webhook delivery enqueue failed")
			}
		}
		_ = p.Outbox.MarkProcessed(ctx, event.ID, time.Now().UTC())
	}
}

func (p *OutboxPublisher) publish(ctx context.Context, event ports.OutboxEvent) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}
	_, err = p.Redis.XAdd(ctx, &redis.XAddArgs{
		Stream: outboxStream,
		Values: map[string]any{
			"event_type": event.EventType,
			"org_id":     event.OrgID.String(),
			"payload":    string(payload),
		},
	}).Result()
	return err
}

func (p *OutboxPublisher) handleFailure(ctx context.Context, event ports.OutboxEvent, err error) {
	maxAttempts := p.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 10
	}
	attempt := event.AttemptCount + 1
	if attempt >= maxAttempts {
		_ = p.sendToDLQ(ctx, event, err)
		_ = p.Outbox.MarkProcessed(ctx, event.ID, time.Now().UTC())
		return
	}
	nextAttempt := time.Now().UTC().Add(backoffDuration(attempt))
	_ = p.Outbox.ScheduleRetry(ctx, event.ID, attempt, nextAttempt, err.Error())
}

func (p *OutboxPublisher) sendToDLQ(ctx context.Context, event ports.OutboxEvent, err error) error {
	payload, marshalErr := json.Marshal(event.Payload)
	if marshalErr != nil {
		payload = []byte("{}")
	}
	_, dlqErr := p.Redis.XAdd(ctx, &redis.XAddArgs{
		Stream: outboxDLQ,
		Values: map[string]any{
			"event_type": event.EventType,
			"org_id":     event.OrgID.String(),
			"payload":    string(payload),
			"error":      err.Error(),
		},
	}).Result()
	return dlqErr
}

func (p *OutboxPublisher) enqueueWebhooks(ctx context.Context, event ports.OutboxEvent) error {
	webhooks, err := p.Webhooks.ListByEvent(ctx, event.OrgID, event.EventType)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, hook := range webhooks {
		delivery := domain.WebhookDelivery{
			ID:            uuidNew(),
			OrgID:         hook.OrgID,
			WebhookID:     hook.ID,
			EventID:       event.ID,
			AttemptCount:  0,
			NextAttemptAt: now,
			Status:        domain.WebhookDeliveryPending,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := p.Deliveries.Create(ctx, delivery); err != nil {
			p.Logger.Error().Err(err).Str("webhook_id", hook.ID.String()).Msg("webhook delivery insert failed")
		}
	}
	return nil
}

func backoffDuration(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	delay := time.Duration(1<<attempt) * time.Second
	if delay > time.Hour {
		return time.Hour
	}
	return delay
}
