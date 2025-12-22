package services

import (
	"context"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type AuditService struct {
	Repo  ports.AuditRepository
	Clock app.Clock
}

type AuditLogInput struct {
	OrgID      *uuid.UUID
	EntityType string
	EntityID   uuid.UUID
	Action     domain.AuditAction
	RequestID  *uuid.UUID
}

func (s *AuditService) Log(ctx context.Context, actor app.Actor, input AuditLogInput) (uuid.UUID, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if s.Repo == nil {
		return uuid.Nil, domain.NewValidationError("audit repository unavailable")
	}
	orgID := actor.OrgID
	if actor.IsAdmin() && input.OrgID != nil && *input.OrgID != uuid.Nil {
		orgID = *input.OrgID
	}
	requestID := uuid.Nil
	if input.RequestID != nil && *input.RequestID != uuid.Nil {
		requestID = *input.RequestID
	}
	entry := domain.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		Action:     input.Action,
		UserID:     actor.UserID,
		RequestID:  requestID,
		Timestamp:  s.Clock.Now(),
	}
	if err := s.Repo.Insert(ctx, entry); err != nil {
		return uuid.Nil, err
	}
	return entry.ID, nil
}
