package ports

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type AuditLogFilter struct {
	OrgID      *uuid.UUID
	EntityType string
	EntityID   *uuid.UUID
	UserID     *uuid.UUID
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}

type AuditQueryRepository interface {
	List(ctx context.Context, filter AuditLogFilter) ([]domain.AuditLog, error)
}
