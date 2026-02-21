package services

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type AlertService struct {
	Alerts ports.AlertRepository
	Clock  app.Clock
}

type AlertCreateInput struct {
	Level          domain.AlertLevel
	Category       string
	Title          string
	Description    string
	EntityType     string
	EntityID       uuid.UUID
	ThresholdValue *float64
	CurrentValue   *float64
}

func (s *AlertService) Create(ctx context.Context, orgID uuid.UUID, input AlertCreateInput) (domain.Alert, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}

	now := s.Clock.Now()
	var escalateAt *time.Time
	if input.Level == domain.AlertCritical {
		t := now.Add(1 * time.Hour)
		escalateAt = &t
	}

	alert := domain.Alert{
		ID:              uuid.New(),
		OrgID:           orgID,
		Level:           input.Level,
		Category:        input.Category,
		Title:           input.Title,
		Description:     input.Description,
		EntityType:      input.EntityType,
		EntityID:        input.EntityID,
		ThresholdValue:  input.ThresholdValue,
		CurrentValue:    input.CurrentValue,
		EscalationLevel: 0,
		AutoEscalateAt:  escalateAt,
		CreatedAt:       now,
	}

	return s.Alerts.Create(ctx, alert)
}

func (s *AlertService) List(ctx context.Context, actor app.Actor, filter ports.AlertFilter) ([]domain.Alert, error) {
	if !actor.IsAdmin() {
		filter.OrgID = &actor.OrgID
	}
	return s.Alerts.List(ctx, filter)
}

func (s *AlertService) Acknowledge(ctx context.Context, actor app.Actor, alertID uuid.UUID) error {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	return s.Alerts.Acknowledge(ctx, actor.OrgID, alertID, actor.UserID, s.Clock.Now())
}

func (s *AlertService) Resolve(ctx context.Context, actor app.Actor, alertID uuid.UUID) error {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	return s.Alerts.Resolve(ctx, actor.OrgID, alertID, s.Clock.Now())
}

func (s *AlertService) CountUnresolved(ctx context.Context, orgID uuid.UUID) (int, error) {
	return s.Alerts.CountUnresolved(ctx, orgID)
}
