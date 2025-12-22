package jobs

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestWebhookDispatcherEventNotFoundMarksFailed(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	hookID := uuid.New()
	eventID := uuid.New()
	deliveryRepo := &fakeWebhookDeliveryRepo{}
	webhookRepo := newFakeWebhookRepo()
	outboxRepo := newFakeOutboxRepo()

	_, _ = webhookRepo.Create(ctx, domain.Webhook{
		ID:        hookID,
		OrgID:     orgID,
		URL:       "https://example.com/hook",
		Events:    []string{"task.created"},
		Secret:    "secret",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	dispatcher := &WebhookDispatcher{
		Deliveries: deliveryRepo,
		Webhooks:   webhookRepo,
		Outbox:     outboxRepo,
		Logger:     zerolog.Nop(),
	}

	delivery := domain.WebhookDelivery{
		ID:            uuid.New(),
		OrgID:         orgID,
		WebhookID:     hookID,
		EventID:       eventID,
		AttemptCount:  0,
		Status:        domain.WebhookDeliveryPending,
		NextAttemptAt: time.Now().UTC(),
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	dispatcher.handleDelivery(ctx, delivery)

	if len(deliveryRepo.updated) != 1 {
		t.Fatalf("expected 1 update, got %d", len(deliveryRepo.updated))
	}
	updated := deliveryRepo.updated[0]
	if updated.Status != domain.WebhookDeliveryFailed {
		t.Fatalf("expected status failed, got %s", updated.Status)
	}
	if updated.AttemptCount != 1 {
		t.Fatalf("expected attempt count 1, got %d", updated.AttemptCount)
	}
	if updated.LastError == nil || *updated.LastError != "event not found" {
		t.Fatalf("expected last error event not found, got %v", updated.LastError)
	}
}

func TestWebhookDispatcherHttpErrorSchedulesRetry(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	hookID := uuid.New()
	eventID := uuid.New()
	deliveryRepo := &fakeWebhookDeliveryRepo{}
	webhookRepo := newFakeWebhookRepo()
	outboxRepo := newFakeOutboxRepo()

	_, _ = webhookRepo.Create(ctx, domain.Webhook{
		ID:        hookID,
		OrgID:     orgID,
		URL:       "https://example.com/hook",
		Events:    []string{"task.created"},
		Secret:    "secret",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	outboxRepo.events[eventID] = ports.OutboxEvent{
		ID:        eventID,
		OrgID:     orgID,
		EventType: "task.created",
		Payload: map[string]any{
			"value": "test",
		},
	}

	dispatcher := &WebhookDispatcher{
		Deliveries: deliveryRepo,
		Webhooks:   webhookRepo,
		Outbox:     outboxRepo,
		Logger:     zerolog.Nop(),
		HTTPClient: &http.Client{
			Transport: roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
				return nil, errors.New("network down")
			}),
		},
	}

	delivery := domain.WebhookDelivery{
		ID:            uuid.New(),
		OrgID:         orgID,
		WebhookID:     hookID,
		EventID:       eventID,
		AttemptCount:  0,
		Status:        domain.WebhookDeliveryPending,
		NextAttemptAt: time.Now().UTC(),
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	before := time.Now().UTC()

	dispatcher.handleDelivery(ctx, delivery)

	if len(deliveryRepo.updated) != 1 {
		t.Fatalf("expected 1 update, got %d", len(deliveryRepo.updated))
	}
	updated := deliveryRepo.updated[0]
	if updated.Status != domain.WebhookDeliveryPending {
		t.Fatalf("expected status pending, got %s", updated.Status)
	}
	if updated.AttemptCount != 1 {
		t.Fatalf("expected attempt count 1, got %d", updated.AttemptCount)
	}
	if updated.LastError == nil || !strings.Contains(*updated.LastError, "network down") {
		t.Fatalf("expected last error to contain network down, got %v", updated.LastError)
	}
	if !updated.NextAttemptAt.After(before) {
		t.Fatalf("expected next attempt after now, got %s", updated.NextAttemptAt)
	}
}

func TestWebhookDispatcherSuccessMarksDelivered(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	hookID := uuid.New()
	eventID := uuid.New()
	deliveryRepo := &fakeWebhookDeliveryRepo{}
	webhookRepo := newFakeWebhookRepo()
	outboxRepo := newFakeOutboxRepo()

	_, _ = webhookRepo.Create(ctx, domain.Webhook{
		ID:        hookID,
		OrgID:     orgID,
		URL:       "https://example.com/hook",
		Events:    []string{"task.created"},
		Secret:    "secret",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	outboxRepo.events[eventID] = ports.OutboxEvent{
		ID:        eventID,
		OrgID:     orgID,
		EventType: "task.created",
		Payload: map[string]any{
			"value": "ok",
		},
	}

	dispatcher := &WebhookDispatcher{
		Deliveries: deliveryRepo,
		Webhooks:   webhookRepo,
		Outbox:     outboxRepo,
		Logger:     zerolog.Nop(),
		HTTPClient: &http.Client{
			Transport: roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("ok")),
					Header:     http.Header{},
				}, nil
			}),
		},
	}

	delivery := domain.WebhookDelivery{
		ID:            uuid.New(),
		OrgID:         orgID,
		WebhookID:     hookID,
		EventID:       eventID,
		AttemptCount:  0,
		Status:        domain.WebhookDeliveryPending,
		NextAttemptAt: time.Now().UTC(),
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	dispatcher.handleDelivery(ctx, delivery)

	if len(deliveryRepo.updated) != 1 {
		t.Fatalf("expected 1 update, got %d", len(deliveryRepo.updated))
	}
	updated := deliveryRepo.updated[0]
	if updated.Status != domain.WebhookDeliveryDelivered {
		t.Fatalf("expected status delivered, got %s", updated.Status)
	}
	if updated.AttemptCount != 1 {
		t.Fatalf("expected attempt count 1, got %d", updated.AttemptCount)
	}
	if updated.LastResponseCode == nil || *updated.LastResponseCode != http.StatusOK {
		t.Fatalf("expected response code 200, got %v", updated.LastResponseCode)
	}
	if updated.LastResponseBody == nil || *updated.LastResponseBody != "ok" {
		t.Fatalf("expected response body ok, got %v", updated.LastResponseBody)
	}
}
