package ports

import (
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID            uuid.UUID
	OrgID         uuid.UUID
	AggregateType string
	AggregateID   uuid.UUID
	EventType     string
	Payload       map[string]any
	DedupeKey     string
	AttemptCount  int
	LastError     *string
	NextAttemptAt time.Time
	LockedAt      *time.Time
	LockedBy      *string
	CreatedAt     time.Time
	ProcessedAt   *time.Time
}
