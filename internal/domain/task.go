package domain

import (
	"time"

	"github.com/google/uuid"
)

type TaskType string

type TaskState string

const (
	TaskTypeInspection TaskType = "inspection"
	TaskTypeRepair     TaskType = "repair"
	TaskTypeOverhaul   TaskType = "overhaul"
)

const (
	TaskStateScheduled  TaskState = "scheduled"
	TaskStateInProgress TaskState = "in_progress"
	TaskStateCompleted  TaskState = "completed"
	TaskStateCancelled  TaskState = "cancelled"
)

type MaintenanceTask struct {
	ID                 uuid.UUID
	OrgID              uuid.UUID
	AircraftID         uuid.UUID
	ProgramID          *uuid.UUID
	Type               TaskType
	State              TaskState
	StartTime          time.Time
	EndTime            time.Time
	AssignedMechanicID *uuid.UUID
	Notes              string
	DeletedAt          *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type TaskTransitionContext struct {
	Now                   time.Time
	AircraftStatus        AircraftStatus
	ActorRole             Role
	ActorID               uuid.UUID
	AllowEarlyCompletion  bool
	AllowLateCancel       bool
	AllReservationsClosed bool
	RequiredPartsUsed     bool
	ComplianceSignedOff   bool
	Notes                 string
}

func (t MaintenanceTask) ValidateCreate() error {
	if t.EndTime.Before(t.StartTime) || t.EndTime.Equal(t.StartTime) {
		return NewValidationError("end_time must be after start_time")
	}
	return nil
}

func (t MaintenanceTask) CanTransition(newState TaskState, ctx TaskTransitionContext) error {
	if t.State == newState {
		return nil
	}

	notes := ctx.Notes
	if notes == "" {
		notes = t.Notes
	}

	switch newState {
	case TaskStateInProgress:
		if t.State != TaskStateScheduled {
			return NewConflictError("task must be scheduled")
		}
		if t.AssignedMechanicID == nil {
			return NewValidationError("assigned_mechanic_id is required")
		}
		if ctx.ActorRole == RoleMechanic && ctx.ActorID != *t.AssignedMechanicID {
			return ErrForbidden
		}
		if ctx.AircraftStatus != AircraftGrounded {
			return NewConflictError("aircraft must be grounded")
		}
		if ctx.Now.Before(t.StartTime.Add(-5 * time.Minute)) {
			return NewConflictError("too early to start task")
		}
		return nil
	case TaskStateCompleted:
		if t.State != TaskStateInProgress {
			return NewConflictError("task must be in progress")
		}
		if t.AssignedMechanicID == nil {
			return NewValidationError("assigned_mechanic_id is required")
		}
		if ctx.ActorRole == RoleMechanic && ctx.ActorID != *t.AssignedMechanicID {
			return ErrForbidden
		}
		if ctx.Now.Before(t.EndTime) && !ctx.AllowEarlyCompletion {
			return NewConflictError("task cannot be completed early")
		}
		if !ctx.AllReservationsClosed {
			return NewConflictError("part reservations must be used or released")
		}
		if !ctx.RequiredPartsUsed {
			return NewConflictError("required parts must be used")
		}
		if !ctx.ComplianceSignedOff {
			return NewConflictError("compliance items must be signed off")
		}
		if notes == "" {
			return NewValidationError("notes are required")
		}
		return nil
	case TaskStateCancelled:
		if t.State == TaskStateCompleted {
			return NewConflictError("completed tasks cannot be cancelled")
		}
		if ctx.ActorRole != RoleScheduler && ctx.ActorRole != RoleAdmin {
			return ErrForbidden
		}
		if ctx.Now.After(t.StartTime.Add(24*time.Hour)) && !ctx.AllowLateCancel {
			return NewConflictError("task can no longer be cancelled without override")
		}
		return nil
	default:
		return NewValidationError("invalid task state transition")
	}
}
