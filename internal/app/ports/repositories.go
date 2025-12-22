package ports

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type TaskRepository interface {
	GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.MaintenanceTask, error)
	Create(ctx context.Context, task domain.MaintenanceTask) (domain.MaintenanceTask, error)
	Update(ctx context.Context, task domain.MaintenanceTask) (domain.MaintenanceTask, error)
	SoftDelete(ctx context.Context, orgID, id uuid.UUID, at time.Time) error
	List(ctx context.Context, filter TaskFilter) ([]domain.MaintenanceTask, error)
	UpdateState(ctx context.Context, orgID, id uuid.UUID, newState domain.TaskState, notes string, now time.Time) (domain.MaintenanceTask, error)
	HasActiveForProgram(ctx context.Context, orgID, programID uuid.UUID) (bool, error)
}

type AircraftRepository interface {
	GetStatus(ctx context.Context, orgID, id uuid.UUID) (domain.AircraftStatus, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.Aircraft, error)
	GetByTailNumber(ctx context.Context, orgID uuid.UUID, tailNumber string) (domain.Aircraft, error)
	Create(ctx context.Context, aircraft domain.Aircraft) (domain.Aircraft, error)
	Update(ctx context.Context, aircraft domain.Aircraft) (domain.Aircraft, error)
	SoftDelete(ctx context.Context, orgID, id uuid.UUID, at time.Time) error
	List(ctx context.Context, filter AircraftFilter) ([]domain.Aircraft, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.User, error)
	Create(ctx context.Context, user domain.User) (domain.User, error)
	Update(ctx context.Context, user domain.User) (domain.User, error)
	SoftDelete(ctx context.Context, orgID, id uuid.UUID, at time.Time) error
	List(ctx context.Context, filter UserFilter) ([]domain.User, error)
}

type PartItemRepository interface {
	GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.PartItem, error)
	GetBySerialNumber(ctx context.Context, orgID uuid.UUID, serial string) (domain.PartItem, error)
	Create(ctx context.Context, item domain.PartItem) (domain.PartItem, error)
	Update(ctx context.Context, item domain.PartItem) (domain.PartItem, error)
	List(ctx context.Context, filter PartItemFilter) ([]domain.PartItem, error)
	SoftDelete(ctx context.Context, orgID, id uuid.UUID, at time.Time) error
	UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status domain.PartItemStatus, now time.Time) error
}

type PartDefinitionRepository interface {
	GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.PartDefinition, error)
	Create(ctx context.Context, def domain.PartDefinition) (domain.PartDefinition, error)
	Update(ctx context.Context, def domain.PartDefinition) (domain.PartDefinition, error)
	List(ctx context.Context, filter PartDefinitionFilter) ([]domain.PartDefinition, error)
	SoftDelete(ctx context.Context, orgID, id uuid.UUID, at time.Time) error
}

type PartReservationRepository interface {
	Create(ctx context.Context, reservation domain.PartReservation) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.PartReservation, error)
	ListByTask(ctx context.Context, orgID, taskID uuid.UUID) ([]domain.PartReservation, error)
	UpdateState(ctx context.Context, orgID, id uuid.UUID, state domain.PartReservationState, now time.Time) error
	ReleaseByTask(ctx context.Context, orgID, taskID uuid.UUID, now time.Time) error
}

type ComplianceRepository interface {
	GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.ComplianceItem, error)
	ListByTask(ctx context.Context, orgID, taskID uuid.UUID) ([]domain.ComplianceItem, error)
	List(ctx context.Context, filter ComplianceFilter) ([]domain.ComplianceItem, error)
	Create(ctx context.Context, item domain.ComplianceItem) error
	Update(ctx context.Context, item domain.ComplianceItem) error
	SignOff(ctx context.Context, orgID, id, userID uuid.UUID, at time.Time) error
}

type AuditRepository interface {
	Insert(ctx context.Context, entry domain.AuditLog) error
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, orgID uuid.UUID, eventType string, aggregateType string, aggregateID uuid.UUID, payload map[string]any, dedupeKey string) error
	LockPending(ctx context.Context, workerID string, limit int) ([]OutboxEvent, error)
	MarkProcessed(ctx context.Context, id uuid.UUID, processedAt time.Time) error
	ScheduleRetry(ctx context.Context, id uuid.UUID, attempt int, nextAttemptAt time.Time, lastError string) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (OutboxEvent, error)
}

type TaskFilter struct {
	OrgID      *uuid.UUID
	AircraftID *uuid.UUID
	State      *domain.TaskState
	Type       *domain.TaskType
	StartFrom  *time.Time
	StartTo    *time.Time
	Limit      int
	Offset     int
}

type AircraftFilter struct {
	OrgID      *uuid.UUID
	Status     *domain.AircraftStatus
	Model      string
	TailNumber string
	Limit      int
	Offset     int
}

type PartDefinitionFilter struct {
	OrgID  *uuid.UUID
	Name   string
	Limit  int
	Offset int
}

type PartItemFilter struct {
	OrgID        *uuid.UUID
	DefinitionID *uuid.UUID
	Status       *domain.PartItemStatus
	ExpiryBefore *time.Time
	Limit        int
	Offset       int
}

type ComplianceFilter struct {
	OrgID  *uuid.UUID
	TaskID *uuid.UUID
	Result *domain.ComplianceResult
	Signed *bool
	Limit  int
	Offset int
}

type UserFilter struct {
	OrgID  *uuid.UUID
	Role   *domain.Role
	Email  string
	Limit  int
	Offset int
}
