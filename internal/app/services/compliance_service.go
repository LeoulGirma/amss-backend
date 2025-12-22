package services

import (
	"context"
	"fmt"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type ComplianceService struct {
	Compliance ports.ComplianceRepository
	Audit      ports.AuditRepository
	Outbox     ports.OutboxRepository
	Clock      app.Clock
}

func (s *ComplianceService) List(ctx context.Context, actor app.Actor, filter ports.ComplianceFilter) ([]domain.ComplianceItem, error) {
	if !actor.IsAdmin() {
		filter.OrgID = &actor.OrgID
	}
	return s.Compliance.List(ctx, filter)
}

func (s *ComplianceService) ListByTask(ctx context.Context, orgID, taskID uuid.UUID) ([]domain.ComplianceItem, error) {
	return s.Compliance.ListByTask(ctx, orgID, taskID)
}

func (s *ComplianceService) Create(ctx context.Context, actor app.Actor, item domain.ComplianceItem) (domain.ComplianceItem, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleMechanic && actor.Role != domain.RoleAdmin {
		return domain.ComplianceItem{}, domain.ErrForbidden
	}
	item.ID = uuid.New()
	item.OrgID = actor.OrgID
	item.CreatedAt = s.Clock.Now()
	item.UpdatedAt = s.Clock.Now()

	if err := s.Compliance.Create(ctx, item); err != nil {
		return domain.ComplianceItem{}, err
	}

	s.emitComplianceAudit(ctx, actor, item, domain.AuditActionCreate)
	s.emitComplianceOutbox(ctx, item, "compliance_created")

	return item, nil
}

func (s *ComplianceService) Update(ctx context.Context, actor app.Actor, item domain.ComplianceItem) (domain.ComplianceItem, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleMechanic && actor.Role != domain.RoleAdmin {
		return domain.ComplianceItem{}, domain.ErrForbidden
	}

	existing, err := s.Compliance.GetByID(ctx, actor.OrgID, item.ID)
	if err != nil {
		return domain.ComplianceItem{}, err
	}
	if existing.SignOffTime != nil {
		return domain.ComplianceItem{}, domain.NewConflictError("compliance item already signed off")
	}

	item.OrgID = actor.OrgID
	item.UpdatedAt = s.Clock.Now()

	if err := s.Compliance.Update(ctx, item); err != nil {
		return domain.ComplianceItem{}, err
	}

	s.emitComplianceAudit(ctx, actor, item, domain.AuditActionUpdate)
	s.emitComplianceOutbox(ctx, item, "compliance_updated")

	return item, nil
}

func (s *ComplianceService) SignOff(ctx context.Context, actor app.Actor, complianceID uuid.UUID) (domain.ComplianceItem, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	item, err := s.Compliance.GetByID(ctx, actor.OrgID, complianceID)
	if err != nil {
		return domain.ComplianceItem{}, err
	}
	if err := item.CanSignOff(actor.Role); err != nil {
		return domain.ComplianceItem{}, err
	}

	if err := s.Compliance.SignOff(ctx, actor.OrgID, item.ID, actor.UserID, s.Clock.Now()); err != nil {
		return domain.ComplianceItem{}, err
	}

	timestamp := s.Clock.Now()
	item.SignOffTime = &timestamp
	item.SignOffUserID = &actor.UserID

	s.emitComplianceAudit(ctx, actor, item, domain.AuditActionUpdate)
	s.emitComplianceOutbox(ctx, item, "compliance_signed")

	return item, nil
}

func (s *ComplianceService) emitComplianceAudit(ctx context.Context, actor app.Actor, item domain.ComplianceItem, action domain.AuditAction) {
	if s.Audit == nil {
		return
	}
	entry := domain.AuditLog{
		ID:            uuid.New(),
		OrgID:         item.OrgID,
		EntityType:    "compliance_item",
		EntityID:      item.ID,
		Action:        action,
		UserID:        actor.UserID,
		RequestID:     uuid.Nil,
		EntityVersion: 0,
		Timestamp:     s.Clock.Now(),
		Details: map[string]any{
			"result": item.Result,
		},
	}
	_ = s.Audit.Insert(ctx, entry)
}

func (s *ComplianceService) emitComplianceOutbox(ctx context.Context, item domain.ComplianceItem, eventType string) {
	if s.Outbox == nil {
		return
	}
	dedupeKey := fmt.Sprintf("%s:%s:%s", eventType, item.OrgID, item.ID)
	payload := map[string]any{
		"version":       1,
		"org_id":        item.OrgID,
		"compliance_id": item.ID,
		"task_id":       item.TaskID,
		"result":        item.Result,
		"timestamp":     s.Clock.Now(),
	}
	_ = s.Outbox.Enqueue(ctx, item.OrgID, eventType, "compliance_item", item.ID, payload, dedupeKey)
}
