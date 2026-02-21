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

// --- Certification & Skill Repositories ---

type CertificationRepository interface {
	// Certification types (reference data)
	ListCertTypes(ctx context.Context) ([]domain.CertificationType, error)
	GetCertTypeByID(ctx context.Context, id uuid.UUID) (domain.CertificationType, error)

	// Employee certifications
	GetCertByID(ctx context.Context, orgID, id uuid.UUID) (domain.EmployeeCertification, error)
	ListCertsByUser(ctx context.Context, orgID, userID uuid.UUID) ([]domain.EmployeeCertification, error)
	CreateCert(ctx context.Context, cert domain.EmployeeCertification) (domain.EmployeeCertification, error)
	UpdateCert(ctx context.Context, cert domain.EmployeeCertification) (domain.EmployeeCertification, error)
	ListExpiringCerts(ctx context.Context, orgID uuid.UUID, before time.Time) ([]domain.EmployeeCertification, error)

	// Type ratings
	ListTypeRatingsByUser(ctx context.Context, orgID, userID uuid.UUID) ([]domain.EmployeeTypeRating, error)
	CreateTypeRating(ctx context.Context, rating domain.EmployeeTypeRating) (domain.EmployeeTypeRating, error)
	HasTypeRating(ctx context.Context, orgID, userID, aircraftTypeID uuid.UUID) (bool, error)

	// Skills
	ListSkillsByUser(ctx context.Context, orgID, userID uuid.UUID) ([]domain.EmployeeSkill, error)
	CreateSkill(ctx context.Context, skill domain.EmployeeSkill) (domain.EmployeeSkill, error)
	UpdateSkill(ctx context.Context, skill domain.EmployeeSkill) (domain.EmployeeSkill, error)

	// Recency
	LogRecency(ctx context.Context, entry domain.EmployeeRecencyLog) error
	GetRecencyHours(ctx context.Context, orgID, userID, aircraftTypeID uuid.UUID, since time.Time) (float64, error)

	// Task skill requirements
	ListRequirements(ctx context.Context, orgID uuid.UUID, taskType domain.TaskType, aircraftTypeID *uuid.UUID) ([]domain.TaskSkillRequirement, error)
	CreateRequirement(ctx context.Context, req domain.TaskSkillRequirement) (domain.TaskSkillRequirement, error)

	// Qualification check
	GetQualifiedMechanics(ctx context.Context, orgID uuid.UUID, taskType domain.TaskType, aircraftTypeID *uuid.UUID) ([]uuid.UUID, error)
}

type AircraftTypeRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.AircraftType, error)
	GetByICAOCode(ctx context.Context, code string) (domain.AircraftType, error)
	List(ctx context.Context) ([]domain.AircraftType, error)
	Create(ctx context.Context, at domain.AircraftType) (domain.AircraftType, error)
}

// --- Directive & Compliance Repositories ---

type DirectiveRepository interface {
	// Regulatory authorities
	ListAuthorities(ctx context.Context) ([]domain.RegulatoryAuthority, error)
	GetAuthorityByID(ctx context.Context, id uuid.UUID) (domain.RegulatoryAuthority, error)

	// Compliance directives
	GetDirectiveByID(ctx context.Context, id uuid.UUID) (domain.ComplianceDirective, error)
	ListDirectives(ctx context.Context, filter DirectiveFilter) ([]domain.ComplianceDirective, error)
	CreateDirective(ctx context.Context, d domain.ComplianceDirective) (domain.ComplianceDirective, error)
	UpdateDirective(ctx context.Context, d domain.ComplianceDirective) (domain.ComplianceDirective, error)

	// Aircraft directive compliance
	GetAircraftCompliance(ctx context.Context, orgID, aircraftID, directiveID uuid.UUID) (domain.AircraftDirectiveCompliance, error)
	ListAircraftCompliance(ctx context.Context, filter AircraftComplianceFilter) ([]domain.AircraftDirectiveCompliance, error)
	UpsertAircraftCompliance(ctx context.Context, c domain.AircraftDirectiveCompliance) (domain.AircraftDirectiveCompliance, error)

	// Templates
	ListTemplates(ctx context.Context, authorityID uuid.UUID) ([]domain.ComplianceTemplate, error)
	GetTemplateByCode(ctx context.Context, authorityID uuid.UUID, code string) (domain.ComplianceTemplate, error)
}

type DirectiveFilter struct {
	OrgID         *uuid.UUID
	AuthorityID   *uuid.UUID
	DirectiveType *domain.DirectiveType
	Limit         int
	Offset        int
}

type AircraftComplianceFilter struct {
	OrgID      *uuid.UUID
	AircraftID *uuid.UUID
	Status     *domain.DirectiveComplianceStatus
	Limit      int
	Offset     int
}

// --- Alert Repository ---

type AlertRepository interface {
	Create(ctx context.Context, alert domain.Alert) (domain.Alert, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.Alert, error)
	List(ctx context.Context, filter AlertFilter) ([]domain.Alert, error)
	Acknowledge(ctx context.Context, orgID, id, userID uuid.UUID, at time.Time) error
	Resolve(ctx context.Context, orgID, id uuid.UUID, at time.Time) error
	CountUnresolved(ctx context.Context, orgID uuid.UUID) (int, error)
}

type AlertFilter struct {
	OrgID    *uuid.UUID
	Level    *domain.AlertLevel
	Category string
	Resolved *bool
	Limit    int
	Offset   int
}

// --- Scheduling Repository ---

type TaskDependencyRepository interface {
	Create(ctx context.Context, dep domain.TaskDependency) (domain.TaskDependency, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	ListByTask(ctx context.Context, orgID, taskID uuid.UUID) ([]domain.TaskDependency, error)
	ListDependents(ctx context.Context, orgID, taskID uuid.UUID) ([]domain.TaskDependency, error)
}

type ScheduleChangeRepository interface {
	Create(ctx context.Context, event domain.ScheduleChangeEvent) (domain.ScheduleChangeEvent, error)
	ListByTask(ctx context.Context, orgID, taskID uuid.UUID) ([]domain.ScheduleChangeEvent, error)
}
