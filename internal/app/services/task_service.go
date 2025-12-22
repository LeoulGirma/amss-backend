package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type TaskService struct {
	Tasks        ports.TaskRepository
	Aircraft     ports.AircraftRepository
	Reservations ports.PartReservationRepository
	Compliance   ports.ComplianceRepository
	Audit        ports.AuditRepository
	Outbox       ports.OutboxRepository
	Clock        app.Clock
}

type TaskTransitionOptions struct {
	AllowEarlyCompletion bool
	AllowLateCancel      bool
	RequireAllPartsUsed  bool
	Notes                string
}

type TaskCreateInput struct {
	OrgID              *uuid.UUID
	AircraftID         uuid.UUID
	ProgramID          *uuid.UUID
	Type               domain.TaskType
	StartTime          time.Time
	EndTime            time.Time
	AssignedMechanicID *uuid.UUID
	Notes              string
}

type TaskUpdateInput struct {
	OrgID              *uuid.UUID
	ProgramID          *uuid.UUID
	Type               *domain.TaskType
	StartTime          *time.Time
	EndTime            *time.Time
	AssignedMechanicID *uuid.UUID
	Notes              *string
}

func (s *TaskService) Create(ctx context.Context, actor app.Actor, input TaskCreateInput) (domain.MaintenanceTask, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleAdmin {
		return domain.MaintenanceTask{}, domain.ErrForbidden
	}
	orgID := actor.OrgID
	if actor.IsAdmin() && input.OrgID != nil && *input.OrgID != uuid.Nil {
		orgID = *input.OrgID
	}

	task := domain.MaintenanceTask{
		ID:                 uuid.New(),
		OrgID:              orgID,
		AircraftID:         input.AircraftID,
		ProgramID:          input.ProgramID,
		Type:               input.Type,
		State:              domain.TaskStateScheduled,
		StartTime:          input.StartTime.UTC(),
		EndTime:            input.EndTime.UTC(),
		AssignedMechanicID: input.AssignedMechanicID,
		Notes:              input.Notes,
		CreatedAt:          s.Clock.Now(),
		UpdatedAt:          s.Clock.Now(),
	}

	if err := task.ValidateCreate(); err != nil {
		return domain.MaintenanceTask{}, err
	}

	created, err := s.Tasks.Create(ctx, task)
	if err != nil {
		return domain.MaintenanceTask{}, err
	}

	s.emitTaskCreateAudit(ctx, actor, created)
	s.emitTaskCreated(ctx, created)
	return created, nil
}

func (s *TaskService) Get(ctx context.Context, actor app.Actor, orgID uuid.UUID, id uuid.UUID) (domain.MaintenanceTask, error) {
	if !actor.IsAdmin() && actor.OrgID != orgID {
		return domain.MaintenanceTask{}, domain.ErrForbidden
	}
	return s.Tasks.GetByID(ctx, orgID, id)
}

func (s *TaskService) List(ctx context.Context, actor app.Actor, filter ports.TaskFilter) ([]domain.MaintenanceTask, error) {
	if !actor.IsAdmin() {
		filter.OrgID = &actor.OrgID
	}
	return s.Tasks.List(ctx, filter)
}

func (s *TaskService) Update(ctx context.Context, actor app.Actor, orgID uuid.UUID, id uuid.UUID, input TaskUpdateInput) (domain.MaintenanceTask, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleAdmin {
		return domain.MaintenanceTask{}, domain.ErrForbidden
	}
	if !actor.IsAdmin() && actor.OrgID != orgID {
		return domain.MaintenanceTask{}, domain.ErrForbidden
	}

	task, err := s.Tasks.GetByID(ctx, orgID, id)
	if err != nil {
		return domain.MaintenanceTask{}, err
	}
	if input.ProgramID != nil {
		task.ProgramID = input.ProgramID
	}
	if input.Type != nil {
		task.Type = *input.Type
	}
	if input.StartTime != nil {
		task.StartTime = input.StartTime.UTC()
	}
	if input.EndTime != nil {
		task.EndTime = input.EndTime.UTC()
	}
	if input.AssignedMechanicID != nil {
		task.AssignedMechanicID = input.AssignedMechanicID
	}
	if input.Notes != nil {
		task.Notes = *input.Notes
	}
	task.UpdatedAt = s.Clock.Now()

	if err := task.ValidateCreate(); err != nil {
		return domain.MaintenanceTask{}, err
	}

	updated, err := s.Tasks.Update(ctx, task)
	if err != nil {
		return domain.MaintenanceTask{}, err
	}

	s.emitTaskAudit(ctx, actor, updated, updated.State)
	return updated, nil
}

func (s *TaskService) Delete(ctx context.Context, actor app.Actor, orgID uuid.UUID, id uuid.UUID) error {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	if !actor.IsAdmin() && actor.OrgID != orgID {
		return domain.ErrForbidden
	}
	if err := s.Tasks.SoftDelete(ctx, orgID, id, s.Clock.Now()); err != nil {
		return err
	}
	if s.Audit != nil {
		entry := domain.AuditLog{
			ID:            uuid.New(),
			OrgID:         orgID,
			EntityType:    "maintenance_task",
			EntityID:      id,
			Action:        domain.AuditActionDelete,
			UserID:        actor.UserID,
			RequestID:     uuid.Nil,
			EntityVersion: 0,
			Timestamp:     s.Clock.Now(),
		}
		_ = s.Audit.Insert(ctx, entry)
	}
	return nil
}

func (s *TaskService) TransitionState(ctx context.Context, actor app.Actor, taskID uuid.UUID, newState domain.TaskState, opts TaskTransitionOptions) (domain.MaintenanceTask, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}

	task, err := s.Tasks.GetByID(ctx, actor.OrgID, taskID)
	if err != nil {
		return domain.MaintenanceTask{}, err
	}
	if !actor.IsAdmin() && actor.OrgID != task.OrgID {
		return domain.MaintenanceTask{}, domain.ErrForbidden
	}

	aircraftStatus, err := s.Aircraft.GetStatus(ctx, task.OrgID, task.AircraftID)
	if err != nil {
		return domain.MaintenanceTask{}, err
	}

	reservations := []domain.PartReservation{}
	if s.Reservations != nil {
		list, err := s.Reservations.ListByTask(ctx, task.OrgID, task.ID)
		if err != nil {
			return domain.MaintenanceTask{}, err
		}
		reservations = list
	}
	allClosed, allUsed := summarizeReservations(reservations)

	complianceSignedOff := true
	if s.Compliance != nil {
		items, err := s.Compliance.ListByTask(ctx, task.OrgID, task.ID)
		if err != nil {
			return domain.MaintenanceTask{}, err
		}
		complianceSignedOff = allComplianceSignedOff(items)
	}

	notes := opts.Notes
	if notes == "" {
		notes = task.Notes
	}

	ctxTransition := domain.TaskTransitionContext{
		Now:                   s.Clock.Now(),
		AircraftStatus:        aircraftStatus,
		ActorRole:             actor.Role,
		ActorID:               actor.UserID,
		AllowEarlyCompletion:  opts.AllowEarlyCompletion && (actor.Role == domain.RoleScheduler || actor.Role == domain.RoleAdmin),
		AllowLateCancel:       opts.AllowLateCancel && (actor.Role == domain.RoleScheduler || actor.Role == domain.RoleAdmin),
		AllReservationsClosed: allClosed,
		RequiredPartsUsed:     !opts.RequireAllPartsUsed || allUsed,
		ComplianceSignedOff:   complianceSignedOff,
		Notes:                 notes,
	}

	if err := task.CanTransition(newState, ctxTransition); err != nil {
		return domain.MaintenanceTask{}, err
	}

	if newState == domain.TaskStateCancelled && s.Reservations != nil {
		if err := s.Reservations.ReleaseByTask(ctx, task.OrgID, task.ID, s.Clock.Now()); err != nil {
			return domain.MaintenanceTask{}, err
		}
	}

	updated, err := s.Tasks.UpdateState(ctx, task.OrgID, task.ID, newState, notes, s.Clock.Now())
	if err != nil {
		return domain.MaintenanceTask{}, err
	}

	s.emitTaskAudit(ctx, actor, updated, newState)
	s.emitTaskOutbox(ctx, updated, newState)

	return updated, nil
}

func summarizeReservations(reservations []domain.PartReservation) (allClosed bool, allUsed bool) {
	if len(reservations) == 0 {
		return true, true
	}
	allClosed = true
	allUsed = true
	for _, res := range reservations {
		switch res.State {
		case domain.ReservationReserved:
			allClosed = false
			allUsed = false
		case domain.ReservationReleased:
			allUsed = false
		}
	}
	return allClosed, allUsed
}

func allComplianceSignedOff(items []domain.ComplianceItem) bool {
	if len(items) == 0 {
		return true
	}
	for _, item := range items {
		if item.SignOffTime == nil {
			return false
		}
		if item.Result == domain.CompliancePending {
			return false
		}
	}
	return true
}

func (s *TaskService) emitTaskAudit(ctx context.Context, actor app.Actor, task domain.MaintenanceTask, newState domain.TaskState) {
	if s.Audit == nil {
		return
	}
	entry := domain.AuditLog{
		ID:            uuid.New(),
		OrgID:         task.OrgID,
		EntityType:    "maintenance_task",
		EntityID:      task.ID,
		Action:        domain.AuditActionStateChange,
		UserID:        actor.UserID,
		RequestID:     uuid.Nil,
		EntityVersion: 0,
		Timestamp:     s.Clock.Now(),
		Details: map[string]any{
			"new_state": newState,
		},
	}
	_ = s.Audit.Insert(ctx, entry)
}

func (s *TaskService) emitTaskCreateAudit(ctx context.Context, actor app.Actor, task domain.MaintenanceTask) {
	if s.Audit == nil {
		return
	}
	entry := domain.AuditLog{
		ID:            uuid.New(),
		OrgID:         task.OrgID,
		EntityType:    "maintenance_task",
		EntityID:      task.ID,
		Action:        domain.AuditActionCreate,
		UserID:        actor.UserID,
		RequestID:     uuid.Nil,
		EntityVersion: 0,
		Timestamp:     s.Clock.Now(),
	}
	_ = s.Audit.Insert(ctx, entry)
}

func (s *TaskService) emitTaskOutbox(ctx context.Context, task domain.MaintenanceTask, newState domain.TaskState) {
	if s.Outbox == nil {
		return
	}
	eventType := "task_state_changed"
	dedupeKey := fmt.Sprintf("task_state_changed:%s:%s:%s", task.OrgID, task.ID, newState)
	payload := map[string]any{
		"version":   1,
		"org_id":    task.OrgID,
		"task_id":   task.ID,
		"new_state": newState,
		"timestamp": s.Clock.Now(),
	}
	_ = s.Outbox.Enqueue(ctx, task.OrgID, eventType, "maintenance_task", task.ID, payload, dedupeKey)
}

func (s *TaskService) emitTaskCreated(ctx context.Context, task domain.MaintenanceTask) {
	if s.Outbox == nil {
		return
	}
	eventType := "task_created"
	dedupeKey := fmt.Sprintf("task_created:%s:%s", task.OrgID, task.ID)
	payload := map[string]any{
		"version":   1,
		"org_id":    task.OrgID,
		"task_id":   task.ID,
		"state":     task.State,
		"timestamp": s.Clock.Now(),
	}
	_ = s.Outbox.Enqueue(ctx, task.OrgID, eventType, "maintenance_task", task.ID, payload, dedupeKey)
}
