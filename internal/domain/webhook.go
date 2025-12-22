package domain

import (
	"time"

	"github.com/google/uuid"
)

type WebhookDeliveryStatus string

const (
	WebhookDeliveryPending   WebhookDeliveryStatus = "pending"
	WebhookDeliveryDelivered WebhookDeliveryStatus = "delivered"
	WebhookDeliveryFailed    WebhookDeliveryStatus = "failed"
)

type Webhook struct {
	ID        uuid.UUID
	OrgID     uuid.UUID
	URL       string
	Events    []string
	Secret    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WebhookDelivery struct {
	ID               uuid.UUID
	OrgID            uuid.UUID
	WebhookID        uuid.UUID
	EventID          uuid.UUID
	AttemptCount     int
	LastError        *string
	NextAttemptAt    time.Time
	Status           WebhookDeliveryStatus
	LastResponseCode *int
	LastResponseBody *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
