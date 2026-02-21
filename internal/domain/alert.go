package domain

import (
	"time"

	"github.com/google/uuid"
)

// AlertLevel represents the severity of an alert
type AlertLevel string

const (
	AlertInfo     AlertLevel = "info"
	AlertWarning  AlertLevel = "warning"
	AlertCritical AlertLevel = "critical"
)

// Alert represents a system alert for an organization
type Alert struct {
	ID              uuid.UUID
	OrgID           uuid.UUID
	Level           AlertLevel
	Category        string
	Title           string
	Description     string
	EntityType      string
	EntityID        uuid.UUID
	ThresholdValue  *float64
	CurrentValue    *float64
	Acknowledged    bool
	AcknowledgedBy  *uuid.UUID
	AcknowledgedAt  *time.Time
	Resolved        bool
	ResolvedAt      *time.Time
	EscalationLevel int
	AutoEscalateAt  *time.Time
	CreatedAt       time.Time
}

// CanAcknowledge checks if the alert can be acknowledged
func (a Alert) CanAcknowledge() error {
	if a.Acknowledged {
		return NewConflictError("alert already acknowledged")
	}
	if a.Resolved {
		return NewConflictError("alert already resolved")
	}
	return nil
}

// CanResolve checks if the alert can be resolved
func (a Alert) CanResolve() error {
	if a.Resolved {
		return NewConflictError("alert already resolved")
	}
	return nil
}
