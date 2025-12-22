package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type IdempotencyRecord struct {
	OrgID       uuid.UUID
	Key         string
	Endpoint    string
	RequestHash string
	Response    []byte
	StatusCode  int
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

type IdempotencyStore interface {
	Get(ctx context.Context, orgID uuid.UUID, key, endpoint string) (IdempotencyRecord, bool, error)
	CreatePlaceholder(ctx context.Context, record IdempotencyRecord) error
	UpdateResponse(ctx context.Context, orgID uuid.UUID, key, endpoint string, statusCode int, response []byte) error
}
