package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuditAction string

const (
	AuditActionCreate      AuditAction = "create"
	AuditActionUpdate      AuditAction = "update"
	AuditActionDelete      AuditAction = "delete"
	AuditActionStateChange AuditAction = "state_change"
)

type AuditLog struct {
	ID            uuid.UUID
	OrgID         uuid.UUID
	EntityType    string
	EntityID      uuid.UUID
	Action        AuditAction
	UserID        uuid.UUID
	RequestID     uuid.UUID
	IPAddress     string
	UserAgent     string
	EntityVersion int
	Timestamp     time.Time
	Details       map[string]any
}
