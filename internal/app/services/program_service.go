package services

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type MaintenanceProgramService struct {
	Programs ports.MaintenanceProgramRepository
	Tasks    ports.TaskRepository
	TaskSvc  *TaskService
	Clock    app.Clock
}

type ProgramCreateInput struct {
	OrgID         *uuid.UUID
	AircraftID    *uuid.UUID
	Name          string
	IntervalType  domain.MaintenanceProgramIntervalType
	IntervalValue int
	LastPerformed *time.Time
}

type ProgramUpdateInput struct {
	AircraftID    *uuid.UUID
	Name          *string
	IntervalType  *domain.MaintenanceProgramIntervalType
	IntervalValue *int
	LastPerformed *time.Time
}

func (s *MaintenanceProgramService) Create(ctx context.Context, actor app.Actor, input ProgramCreateInput) (domain.MaintenanceProgram, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleAdmin {
		return domain.MaintenanceProgram{}, domain.ErrForbidden
	}
	if input.Name == "" || input.IntervalType == "" || input.IntervalValue <= 0 {
		return domain.MaintenanceProgram{}, domain.NewValidationError("name, interval_type, and interval_value are required")
	}
	orgID := actor.OrgID
	if actor.IsAdmin() && input.OrgID != nil && *input.OrgID != uuid.Nil {
		orgID = *input.OrgID
	}
	program := domain.MaintenanceProgram{
		ID:            uuid.New(),
		OrgID:         orgID,
		AircraftID:    input.AircraftID,
		Name:          input.Name,
		IntervalType:  input.IntervalType,
		IntervalValue: input.IntervalValue,
		LastPerformed: input.LastPerformed,
		CreatedAt:     s.Clock.Now(),
		UpdatedAt:     s.Clock.Now(),
	}
	return s.Programs.Create(ctx, program)
}

func (s *MaintenanceProgramService) List(ctx context.Context, actor app.Actor, filter ports.MaintenanceProgramFilter) ([]domain.MaintenanceProgram, error) {
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleAdmin && actor.Role != domain.RoleAuditor && actor.Role != domain.RoleTenantAdmin {
		return nil, domain.ErrForbidden
	}
	if !actor.IsAdmin() {
		filter.OrgID = &actor.OrgID
	}
	return s.Programs.List(ctx, filter)
}

func (s *MaintenanceProgramService) Get(ctx context.Context, actor app.Actor, orgID, id uuid.UUID) (domain.MaintenanceProgram, error) {
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleAdmin && actor.Role != domain.RoleAuditor && actor.Role != domain.RoleTenantAdmin {
		return domain.MaintenanceProgram{}, domain.ErrForbidden
	}
	if !actor.IsAdmin() && orgID != actor.OrgID {
		return domain.MaintenanceProgram{}, domain.ErrForbidden
	}
	return s.Programs.GetByID(ctx, orgID, id)
}

func (s *MaintenanceProgramService) Update(ctx context.Context, actor app.Actor, orgID, id uuid.UUID, input ProgramUpdateInput) (domain.MaintenanceProgram, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleAdmin {
		return domain.MaintenanceProgram{}, domain.ErrForbidden
	}
	if !actor.IsAdmin() {
		orgID = actor.OrgID
	}
	program, err := s.Programs.GetByID(ctx, orgID, id)
	if err != nil {
		return domain.MaintenanceProgram{}, err
	}
	if input.AircraftID != nil {
		program.AircraftID = input.AircraftID
	}
	if input.Name != nil {
		if *input.Name == "" {
			return domain.MaintenanceProgram{}, domain.NewValidationError("name is required")
		}
		program.Name = *input.Name
	}
	if input.IntervalType != nil {
		program.IntervalType = *input.IntervalType
	}
	if input.IntervalValue != nil {
		if *input.IntervalValue <= 0 {
			return domain.MaintenanceProgram{}, domain.NewValidationError("interval_value must be greater than 0")
		}
		program.IntervalValue = *input.IntervalValue
	}
	if input.LastPerformed != nil {
		program.LastPerformed = input.LastPerformed
	}
	program.UpdatedAt = s.Clock.Now()
	return s.Programs.Update(ctx, program)
}

func (s *MaintenanceProgramService) Delete(ctx context.Context, actor app.Actor, orgID, id uuid.UUID) error {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	if !actor.IsAdmin() {
		orgID = actor.OrgID
	}
	return s.Programs.SoftDelete(ctx, orgID, id, s.Clock.Now())
}

func (s *MaintenanceProgramService) GenerateDueTasks(ctx context.Context, actor app.Actor, limit int) (int, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleScheduler {
		return 0, domain.ErrForbidden
	}
	if s.TaskSvc == nil {
		return 0, domain.NewValidationError("task service unavailable")
	}
	programs, err := s.Programs.ListDueCalendar(ctx, s.Clock.Now(), limit)
	if err != nil {
		return 0, err
	}
	created := 0
	for _, program := range programs {
		if program.AircraftID == nil {
			continue
		}
		hasActive := false
		if s.Tasks != nil {
			active, err := s.Tasks.HasActiveForProgram(ctx, program.OrgID, program.ID)
			if err != nil {
				return created, err
			}
			hasActive = active
		}
		if hasActive {
			continue
		}
		due := s.Clock.Now()
		if program.LastPerformed != nil {
			due = program.LastPerformed.Add(time.Duration(program.IntervalValue) * 24 * time.Hour)
		} else {
			due = s.Clock.Now().Add(time.Duration(program.IntervalValue) * 24 * time.Hour)
		}
		if due.Before(s.Clock.Now()) {
			due = s.Clock.Now().Add(1 * time.Hour)
		}
		end := due.Add(2 * time.Hour)
		_, err := s.TaskSvc.Create(ctx, actor, TaskCreateInput{
			OrgID:              &program.OrgID,
			AircraftID:         *program.AircraftID,
			ProgramID:          &program.ID,
			Type:               domain.TaskTypeInspection,
			StartTime:          due,
			EndTime:            end,
			AssignedMechanicID: nil,
			Notes:              "Generated from maintenance program",
		})
		if err != nil {
			continue
		}
		created++
	}
	return created, nil
}
