package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestOutboxPublisherHandleFailureSchedulesRetry(t *testing.T) {
	ctx := context.Background()
	outboxRepo := newFakeOutboxRepo()
	publisher := &OutboxPublisher{
		Outbox:      outboxRepo,
		MaxAttempts: 3,
	}

	event := ports.OutboxEvent{
		ID:           uuid.New(),
		AttemptCount: 0,
	}
	before := time.Now().UTC()

	publisher.handleFailure(ctx, event, errors.New("publish failed"))

	if len(outboxRepo.scheduleCalls) != 1 {
		t.Fatalf("expected 1 schedule call, got %d", len(outboxRepo.scheduleCalls))
	}
	call := outboxRepo.scheduleCalls[0]
	if call.id != event.ID {
		t.Fatalf("expected event id %s, got %s", event.ID, call.id)
	}
	if call.attempt != 1 {
		t.Fatalf("expected attempt 1, got %d", call.attempt)
	}
	if call.lastError != "publish failed" {
		t.Fatalf("expected last error publish failed, got %s", call.lastError)
	}
	if !call.nextAttempt.After(before) {
		t.Fatalf("expected next attempt after now, got %s", call.nextAttempt)
	}
}

func TestOutboxPublisherProcessOnceCreatesDeliveriesAndMarksProcessed(t *testing.T) {
	ctx := context.Background()
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer redisServer.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	orgID := uuid.New()
	event := ports.OutboxEvent{
		ID:        uuid.New(),
		OrgID:     orgID,
		EventType: "task.created",
		Payload: map[string]any{
			"value": "ok",
		},
	}

	outboxRepo := newFakeOutboxRepo()
	outboxRepo.lockEvents = []ports.OutboxEvent{event}

	webhookRepo := newFakeWebhookRepo()
	deliveryRepo := &fakeWebhookDeliveryRepo{}
	_, _ = webhookRepo.Create(ctx, domain.Webhook{
		ID:        uuid.New(),
		OrgID:     orgID,
		URL:       "https://example.com/hook",
		Events:    []string{event.EventType},
		Secret:    "secret",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	publisher := &OutboxPublisher{
		Outbox:     outboxRepo,
		Webhooks:   webhookRepo,
		Deliveries: deliveryRepo,
		Redis:      redisClient,
		WorkerID:   "test-worker",
	}
	publisher.processOnce(ctx)

	outboxRepo.mu.Lock()
	markCount := len(outboxRepo.markCalls)
	outboxRepo.mu.Unlock()
	if markCount != 1 {
		t.Fatalf("expected 1 mark processed, got %d", markCount)
	}
	if len(deliveryRepo.created) != 1 {
		t.Fatalf("expected 1 webhook delivery, got %d", len(deliveryRepo.created))
	}
	streamLen, err := redisClient.XLen(ctx, outboxStream).Result()
	if err != nil {
		t.Fatalf("read stream len: %v", err)
	}
	if streamLen != 1 {
		t.Fatalf("expected stream length 1, got %d", streamLen)
	}
}
