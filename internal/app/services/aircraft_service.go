package services

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type AircraftService struct {
	Aircraft ports.AircraftRepository
	Clock    app.Clock
}

type AircraftCreateInput struct {
	OrgID            *uuid.UUID
	TailNumber       string
	Model            string
	LastMaintenance  *time.Time
	NextDue          *time.Time
	Status           domain.AircraftStatus
	CapacitySlots    int
	FlightHoursTotal int
	CyclesTotal      int
}

type AircraftUpdateInput struct {
	TailNumber       *string
	Model            *string
	LastMaintenance  *time.Time
	NextDue          *time.Time
	Status           *domain.AircraftStatus
	CapacitySlots    *int
	FlightHoursTotal *int
	CyclesTotal      *int
}

func (s *AircraftService) Create(ctx context.Context, actor app.Actor, input AircraftCreateInput) (domain.Aircraft, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleMechanic && actor.Role != domain.RoleAdmin {
		return domain.Aircraft{}, domain.ErrForbidden
	}
	if input.TailNumber == "" || input.Model == "" {
		return domain.Aircraft{}, domain.NewValidationError("tail_number and model are required")
	}
	if input.CapacitySlots <= 0 {
		return domain.Aircraft{}, domain.NewValidationError("capacity_slots must be greater than 0")
	}
	if input.NextDue != nil && input.LastMaintenance != nil && !input.NextDue.After(*input.LastMaintenance) {
		return domain.Aircraft{}, domain.NewValidationError("next_due must be after last_maintenance")
	}
	status := input.Status
	if status == "" {
		status = domain.AircraftOperational
	}
	orgID := actor.OrgID
	if actor.IsAdmin() && input.OrgID != nil && *input.OrgID != uuid.Nil {
		orgID = *input.OrgID
	}

	aircraft := domain.Aircraft{
		ID:               uuid.New(),
		OrgID:            orgID,
		TailNumber:       input.TailNumber,
		Model:            input.Model,
		LastMaintenance:  input.LastMaintenance,
		NextDue:          input.NextDue,
		Status:           status,
		CapacitySlots:    input.CapacitySlots,
		FlightHoursTotal: input.FlightHoursTotal,
		CyclesTotal:      input.CyclesTotal,
		CreatedAt:        s.Clock.Now(),
		UpdatedAt:        s.Clock.Now(),
	}
	return s.Aircraft.Create(ctx, aircraft)
}

func (s *AircraftService) List(ctx context.Context, actor app.Actor, filter ports.AircraftFilter) ([]domain.Aircraft, error) {
	if !actor.IsAdmin() {
		filter.OrgID = &actor.OrgID
	}
	return s.Aircraft.List(ctx, filter)
}

func (s *AircraftService) Get(ctx context.Context, actor app.Actor, orgID, id uuid.UUID) (domain.Aircraft, error) {
	if !actor.IsAdmin() && actor.OrgID != orgID {
		return domain.Aircraft{}, domain.ErrForbidden
	}
	return s.Aircraft.GetByID(ctx, orgID, id)
}

func (s *AircraftService) Update(ctx context.Context, actor app.Actor, orgID, id uuid.UUID, input AircraftUpdateInput) (domain.Aircraft, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleMechanic && actor.Role != domain.RoleAdmin {
		return domain.Aircraft{}, domain.ErrForbidden
	}
	if !actor.IsAdmin() {
		orgID = actor.OrgID
	}
	aircraft, err := s.Aircraft.GetByID(ctx, orgID, id)
	if err != nil {
		return domain.Aircraft{}, err
	}
	if input.TailNumber != nil {
		aircraft.TailNumber = *input.TailNumber
	}
	if input.Model != nil {
		aircraft.Model = *input.Model
	}
	if input.LastMaintenance != nil {
		aircraft.LastMaintenance = input.LastMaintenance
	}
	if input.NextDue != nil {
		aircraft.NextDue = input.NextDue
	}
	if input.Status != nil {
		aircraft.Status = *input.Status
	}
	if input.CapacitySlots != nil {
		if *input.CapacitySlots <= 0 {
			return domain.Aircraft{}, domain.NewValidationError("capacity_slots must be greater than 0")
		}
		aircraft.CapacitySlots = *input.CapacitySlots
	}
	if input.FlightHoursTotal != nil {
		aircraft.FlightHoursTotal = *input.FlightHoursTotal
	}
	if input.CyclesTotal != nil {
		aircraft.CyclesTotal = *input.CyclesTotal
	}
	if aircraft.NextDue != nil && aircraft.LastMaintenance != nil && !aircraft.NextDue.After(*aircraft.LastMaintenance) {
		return domain.Aircraft{}, domain.NewValidationError("next_due must be after last_maintenance")
	}
	aircraft.UpdatedAt = s.Clock.Now()
	return s.Aircraft.Update(ctx, aircraft)
}

func (s *AircraftService) Delete(ctx context.Context, actor app.Actor, orgID, id uuid.UUID) error {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleMechanic && actor.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	if !actor.IsAdmin() {
		orgID = actor.OrgID
	}
	return s.Aircraft.SoftDelete(ctx, orgID, id, s.Clock.Now())
}
