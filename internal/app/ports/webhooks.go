package ports

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type WebhookRepository interface {
	GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.Webhook, error)
	Create(ctx context.Context, webhook domain.Webhook) (domain.Webhook, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	List(ctx context.Context, orgID uuid.UUID) ([]domain.Webhook, error)
	ListByEvent(ctx context.Context, orgID uuid.UUID, eventType string) ([]domain.Webhook, error)
}

type WebhookDeliveryRepository interface {
	Create(ctx context.Context, delivery domain.WebhookDelivery) error
	ClaimPending(ctx context.Context, limit int, lockUntil time.Time) ([]domain.WebhookDelivery, error)
	Update(ctx context.Context, delivery domain.WebhookDelivery) error
}
