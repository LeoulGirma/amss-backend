package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrgPolicy struct {
	OrgID                      uuid.UUID
	RetentionInterval          time.Duration
	MaxWebhookAttempts         int
	WebhookReplayWindowSeconds int
	APIRateLimitPerMin         int
	APIKeyRateLimitPerMin      int
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}
