package jobs

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/aeromaintain/amss/pkg/observability"
	"github.com/rs/zerolog"
)

type WebhookDispatcher struct {
	Deliveries ports.WebhookDeliveryRepository
	Webhooks   ports.WebhookRepository
	Outbox     ports.OutboxRepository
	Policies   *services.OrgPolicyService
	HTTPClient *http.Client
	Logger     zerolog.Logger
}

func (d *WebhookDispatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.processOnce(ctx)
		}
	}
}

func (d *WebhookDispatcher) processOnce(ctx context.Context) {
	if d.Deliveries == nil || d.Outbox == nil || d.Webhooks == nil {
		return
	}
	observability.IncJobRun("webhook_dispatcher")
	lockUntil := time.Now().UTC().Add(2 * time.Minute)
	deliveries, err := d.Deliveries.ClaimPending(ctx, 25, lockUntil)
	if err != nil {
		observability.IncJobFailure("webhook_dispatcher")
		d.Logger.Error().Err(err).Msg("webhook claim failed")
		return
	}
	for _, delivery := range deliveries {
		d.handleDelivery(ctx, delivery)
	}
}

func (d *WebhookDispatcher) handleDelivery(ctx context.Context, delivery domain.WebhookDelivery) {
	policy := defaultPolicy()
	if d.Policies != nil {
		if p, err := d.Policies.Get(ctx, delivery.OrgID); err == nil {
			policy = p
		}
	}

	hook, err := d.Webhooks.GetByID(ctx, delivery.OrgID, delivery.WebhookID)
	if err != nil {
		d.markFailed(ctx, delivery, 0, "webhook not found", policy.MaxWebhookAttempts)
		return
	}
	event, err := d.Outbox.GetByID(ctx, delivery.OrgID, delivery.EventID)
	if err != nil {
		d.markFailed(ctx, delivery, 0, "event not found", policy.MaxWebhookAttempts)
		return
	}

	body := map[string]any{
		"event_type": event.EventType,
		"event_id":   event.ID.String(),
		"org_id":     event.OrgID.String(),
		"data":       event.Payload,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		d.markFailed(ctx, delivery, 0, "payload marshal failed", policy.MaxWebhookAttempts)
		return
	}
	timestamp := strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)
	signature := signWebhook(hook.Secret, timestamp, payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hook.URL, bytes.NewReader(payload))
	if err != nil {
		d.markFailed(ctx, delivery, 0, "request build failed", policy.MaxWebhookAttempts)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Timestamp", timestamp)
	req.Header.Set("X-Webhook-Signature", "v1="+signature)
	req.Header.Set("X-Webhook-Event", event.EventType)
	req.Header.Set("X-Webhook-Delivery-Id", delivery.ID.String())

	client := d.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		d.scheduleRetry(ctx, delivery, 0, err.Error(), policy.MaxWebhookAttempts, 0)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	statusCode := resp.StatusCode

	if statusCode >= 200 && statusCode < 300 {
		delivery.Status = domain.WebhookDeliveryDelivered
		delivery.AttemptCount++
		delivery.LastError = nil
		delivery.LastResponseCode = &statusCode
		if len(bodyBytes) > 0 {
			bodyStr := string(bodyBytes)
			delivery.LastResponseBody = &bodyStr
		}
		delivery.NextAttemptAt = time.Now().UTC()
		delivery.UpdatedAt = time.Now().UTC()
		_ = d.Deliveries.Update(ctx, delivery)
		observability.IncWebhookDelivery("delivered")
		return
	}

	retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
	d.scheduleRetry(ctx, delivery, statusCode, string(bodyBytes), policy.MaxWebhookAttempts, retryAfter)
}

func (d *WebhookDispatcher) scheduleRetry(ctx context.Context, delivery domain.WebhookDelivery, statusCode int, errBody string, maxAttempts int, retryAfter time.Duration) {
	attempt := delivery.AttemptCount + 1
	if attempt >= maxAttempts {
		d.markFailed(ctx, delivery, statusCode, errBody, maxAttempts)
		return
	}
	delay := retryAfter
	if delay == 0 {
		delay = backoffDuration(attempt)
	}
	delivery.AttemptCount = attempt
	if errBody != "" {
		delivery.LastError = &errBody
	}
	delivery.Status = domain.WebhookDeliveryPending
	if statusCode != 0 {
		delivery.LastResponseCode = &statusCode
		if errBody != "" {
			delivery.LastResponseBody = &errBody
		}
	}
	delivery.NextAttemptAt = time.Now().UTC().Add(delay)
	delivery.UpdatedAt = time.Now().UTC()
	_ = d.Deliveries.Update(ctx, delivery)
	observability.IncWebhookDelivery("retry")
}

func (d *WebhookDispatcher) markFailed(ctx context.Context, delivery domain.WebhookDelivery, statusCode int, errBody string, maxAttempts int) {
	attempt := delivery.AttemptCount + 1
	delivery.AttemptCount = attempt
	delivery.Status = domain.WebhookDeliveryFailed
	delivery.NextAttemptAt = time.Now().UTC()
	delivery.UpdatedAt = time.Now().UTC()
	if errBody != "" {
		delivery.LastError = &errBody
	}
	if statusCode != 0 {
		delivery.LastResponseCode = &statusCode
		if errBody != "" {
			delivery.LastResponseBody = &errBody
		}
	}
	_ = d.Deliveries.Update(ctx, delivery)
	observability.IncWebhookDelivery("failed")
}

func signWebhook(secret string, timestamp string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func defaultPolicy() domain.OrgPolicy {
	return domain.OrgPolicy{
		MaxWebhookAttempts:         10,
		WebhookReplayWindowSeconds: 300,
	}
}
