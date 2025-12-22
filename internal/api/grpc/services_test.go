package grpcapi

import (
	"context"
	"testing"
	"time"

	amssv1 "github.com/aeromaintain/amss/api/proto/amssv1"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTaskServiceCreateTask(t *testing.T) {
	orgID := uuid.New()
	taskRepo := newFakeTaskRepo()
	taskSvc := &services.TaskService{Tasks: taskRepo}
	server := &TaskServiceServer{Tasks: taskSvc}

	start := time.Now().UTC().Add(1 * time.Hour)
	end := start.Add(2 * time.Hour)
	ctx := contextWithPrincipal(orgID, domain.RoleScheduler)
	resp, err := server.CreateTask(ctx, &amssv1.CreateTaskRequest{
		OrgId:      orgID.String(),
		AircraftId: uuid.New().String(),
		Type:       string(domain.TaskTypeInspection),
		StartTime:  start.Format(time.RFC3339),
		EndTime:    end.Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.State != string(domain.TaskStateScheduled) {
		t.Fatalf("expected state scheduled, got %s", resp.State)
	}
}

func TestTaskServiceCreateTaskInvalidArgs(t *testing.T) {
	orgID := uuid.New()
	taskRepo := newFakeTaskRepo()
	taskSvc := &services.TaskService{Tasks: taskRepo}
	server := &TaskServiceServer{Tasks: taskSvc}

	start := time.Now().UTC().Add(1 * time.Hour)
	end := start.Add(2 * time.Hour)
	ctx := contextWithPrincipal(orgID, domain.RoleScheduler)
	_, err := server.CreateTask(ctx, &amssv1.CreateTaskRequest{
		OrgId:      orgID.String(),
		AircraftId: uuid.New().String(),
		Type:       "invalid",
		StartTime:  start.Format(time.RFC3339),
		EndTime:    end.Format(time.RFC3339),
	})
	requireStatusCode(t, err, codes.InvalidArgument)
}

func TestTaskServiceCreateTaskForbidden(t *testing.T) {
	orgID := uuid.New()
	taskRepo := newFakeTaskRepo()
	taskSvc := &services.TaskService{Tasks: taskRepo}
	server := &TaskServiceServer{Tasks: taskSvc}

	start := time.Now().UTC().Add(1 * time.Hour)
	end := start.Add(2 * time.Hour)
	ctx := contextWithPrincipal(orgID, domain.RoleMechanic)
	_, err := server.CreateTask(ctx, &amssv1.CreateTaskRequest{
		OrgId:      orgID.String(),
		AircraftId: uuid.New().String(),
		Type:       string(domain.TaskTypeInspection),
		StartTime:  start.Format(time.RFC3339),
		EndTime:    end.Format(time.RFC3339),
	})
	requireStatusCode(t, err, codes.PermissionDenied)
}

func TestTaskServiceTransitionState(t *testing.T) {
	orgID := uuid.New()
	aircraftID := uuid.New()
	mechanicID := uuid.New()
	taskRepo := newFakeTaskRepo()
	aircraftRepo := &fakeAircraftRepo{statuses: map[uuid.UUID]domain.AircraftStatus{
		aircraftID: domain.AircraftGrounded,
	}}
	taskSvc := &services.TaskService{
		Tasks:    taskRepo,
		Aircraft: aircraftRepo,
	}
	server := &TaskServiceServer{Tasks: taskSvc}

	now := time.Now().UTC()
	task := domain.MaintenanceTask{
		ID:                 uuid.New(),
		OrgID:              orgID,
		AircraftID:         aircraftID,
		Type:               domain.TaskTypeInspection,
		State:              domain.TaskStateScheduled,
		StartTime:          now.Add(-1 * time.Minute),
		EndTime:            now.Add(1 * time.Hour),
		AssignedMechanicID: &mechanicID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	taskRepo.tasks[task.ID] = task

	ctx := contextWithPrincipalForUser(orgID, mechanicID, domain.RoleMechanic)
	resp, err := server.TransitionState(ctx, &amssv1.TransitionStateRequest{
		TaskId:   task.ID.String(),
		NewState: string(domain.TaskStateInProgress),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.State != string(domain.TaskStateInProgress) {
		t.Fatalf("expected state in_progress, got %s", resp.State)
	}
}

func TestTaskServiceTransitionStateNotFound(t *testing.T) {
	orgID := uuid.New()
	taskRepo := newFakeTaskRepo()
	taskSvc := &services.TaskService{Tasks: taskRepo}
	server := &TaskServiceServer{Tasks: taskSvc}

	ctx := contextWithPrincipal(orgID, domain.RoleScheduler)
	_, err := server.TransitionState(ctx, &amssv1.TransitionStateRequest{
		TaskId:   uuid.New().String(),
		NewState: string(domain.TaskStateInProgress),
	})
	requireStatusCode(t, err, codes.NotFound)
}

func TestInventoryServiceReserveParts(t *testing.T) {
	orgID := uuid.New()
	taskID := uuid.New()
	partID := uuid.New()
	resRepo := newFakeReservationRepo()
	itemRepo := newFakePartItemRepo()
	itemRepo.items[partID] = domain.PartItem{
		ID:     partID,
		OrgID:  orgID,
		Status: domain.PartItemInStock,
	}
	partSvc := &services.PartReservationService{
		Reservations: resRepo,
		PartItems:    itemRepo,
		Locker:       fakeLocker{},
	}
	server := &InventoryServiceServer{Parts: partSvc}

	ctx := contextWithPrincipal(orgID, domain.RoleScheduler)
	resp, err := server.ReserveParts(ctx, &amssv1.ReservePartsRequest{
		OrgId:      orgID.String(),
		TaskId:     taskID.String(),
		PartItemId: partID.String(),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.State != string(domain.ReservationReserved) {
		t.Fatalf("expected state reserved, got %s", resp.State)
	}
}

func TestInventoryServiceReservePartsInvalidArgs(t *testing.T) {
	orgID := uuid.New()
	partSvc := &services.PartReservationService{}
	server := &InventoryServiceServer{Parts: partSvc}

	ctx := contextWithPrincipal(orgID, domain.RoleScheduler)
	_, err := server.ReserveParts(ctx, &amssv1.ReservePartsRequest{
		OrgId:      orgID.String(),
		TaskId:     "invalid",
		PartItemId: uuid.New().String(),
	})
	requireStatusCode(t, err, codes.InvalidArgument)
}

func TestInventoryServiceReservePartsForbidden(t *testing.T) {
	orgID := uuid.New()
	otherOrg := uuid.New()
	partSvc := &services.PartReservationService{}
	server := &InventoryServiceServer{Parts: partSvc}

	ctx := contextWithPrincipal(orgID, domain.RoleScheduler)
	_, err := server.ReserveParts(ctx, &amssv1.ReservePartsRequest{
		OrgId:      otherOrg.String(),
		TaskId:     uuid.New().String(),
		PartItemId: uuid.New().String(),
	})
	requireStatusCode(t, err, codes.PermissionDenied)
}

func TestInventoryServiceReleaseParts(t *testing.T) {
	orgID := uuid.New()
	taskID := uuid.New()
	resID := uuid.New()
	partID := uuid.New()
	resRepo := newFakeReservationRepo()
	resRepo.items[resID] = domain.PartReservation{
		ID:         resID,
		OrgID:      orgID,
		TaskID:     taskID,
		PartItemID: partID,
		State:      domain.ReservationReserved,
	}
	taskRepo := newFakeTaskRepo()
	taskRepo.tasks[taskID] = domain.MaintenanceTask{
		ID:    taskID,
		OrgID: orgID,
		State: domain.TaskStateScheduled,
	}
	partSvc := &services.PartReservationService{
		Reservations: resRepo,
		Tasks:        taskRepo,
	}
	server := &InventoryServiceServer{Parts: partSvc}

	ctx := contextWithPrincipal(orgID, domain.RoleScheduler)
	resp, err := server.ReleaseParts(ctx, &amssv1.ReleasePartsRequest{
		OrgId:         orgID.String(),
		ReservationId: resID.String(),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.State != string(domain.ReservationReleased) {
		t.Fatalf("expected state released, got %s", resp.State)
	}
}

func TestInventoryServiceReleasePartsNotFound(t *testing.T) {
	orgID := uuid.New()
	resRepo := newFakeReservationRepo()
	partSvc := &services.PartReservationService{
		Reservations: resRepo,
	}
	server := &InventoryServiceServer{Parts: partSvc}

	ctx := contextWithPrincipal(orgID, domain.RoleScheduler)
	_, err := server.ReleaseParts(ctx, &amssv1.ReleasePartsRequest{
		OrgId:         orgID.String(),
		ReservationId: uuid.New().String(),
	})
	requireStatusCode(t, err, codes.NotFound)
}

func TestAuditServiceLogAction(t *testing.T) {
	orgID := uuid.New()
	entityID := uuid.New()
	repo := &fakeAuditRepo{}
	auditSvc := &services.AuditService{Repo: repo}
	server := &AuditServiceServer{Audit: auditSvc}

	ctx := contextWithPrincipal(orgID, domain.RoleScheduler)
	resp, err := server.LogAction(ctx, &amssv1.AuditLogRequest{
		OrgId:      orgID.String(),
		EntityType: "maintenance_task",
		EntityId:   entityID.String(),
		Action:     string(domain.AuditActionCreate),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AuditLogId == "" {
		t.Fatalf("expected audit_log_id")
	}
	if repo.last.OrgID != orgID {
		t.Fatalf("expected org_id %s, got %s", orgID, repo.last.OrgID)
	}
}

func TestAuditServiceLogActionInvalidArgs(t *testing.T) {
	orgID := uuid.New()
	repo := &fakeAuditRepo{}
	auditSvc := &services.AuditService{Repo: repo}
	server := &AuditServiceServer{Audit: auditSvc}

	ctx := contextWithPrincipal(orgID, domain.RoleScheduler)
	_, err := server.LogAction(ctx, &amssv1.AuditLogRequest{
		OrgId:    orgID.String(),
		EntityId: uuid.New().String(),
		Action:   string(domain.AuditActionCreate),
	})
	requireStatusCode(t, err, codes.InvalidArgument)
}

func TestAuditServiceLogActionForbidden(t *testing.T) {
	orgID := uuid.New()
	otherOrg := uuid.New()
	repo := &fakeAuditRepo{}
	auditSvc := &services.AuditService{Repo: repo}
	server := &AuditServiceServer{Audit: auditSvc}

	ctx := contextWithPrincipal(orgID, domain.RoleScheduler)
	_, err := server.LogAction(ctx, &amssv1.AuditLogRequest{
		OrgId:      otherOrg.String(),
		EntityType: "maintenance_task",
		EntityId:   uuid.New().String(),
		Action:     string(domain.AuditActionCreate),
	})
	requireStatusCode(t, err, codes.PermissionDenied)
}

func TestProgramServiceGenerateTasks(t *testing.T) {
	orgID := uuid.New()
	aircraftID := uuid.New()
	programRepo := &fakeProgramRepo{
		due: []domain.MaintenanceProgram{
			{
				ID:            uuid.New(),
				OrgID:         orgID,
				AircraftID:    &aircraftID,
				Name:          "A-Check",
				IntervalType:  domain.ProgramIntervalCalendar,
				IntervalValue: 30,
				CreatedAt:     time.Now().UTC(),
				UpdatedAt:     time.Now().UTC(),
			},
		},
	}
	taskRepo := newFakeTaskRepo()
	taskSvc := &services.TaskService{Tasks: taskRepo}
	programSvc := &services.MaintenanceProgramService{
		Programs: programRepo,
		Tasks:    taskRepo,
		TaskSvc:  taskSvc,
	}
	server := &ProgramServiceServer{Programs: programSvc}

	ctx := contextWithPrincipal(orgID, domain.RoleAdmin)
	resp, err := server.GenerateTasks(ctx, &amssv1.GenerateTasksRequest{OrgId: orgID.String()})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Created != 1 {
		t.Fatalf("expected created 1, got %d", resp.Created)
	}
}

func TestProgramServiceGenerateTasksForbidden(t *testing.T) {
	orgID := uuid.New()
	programSvc := &services.MaintenanceProgramService{Programs: &fakeProgramRepo{}}
	server := &ProgramServiceServer{Programs: programSvc}

	ctx := contextWithPrincipal(orgID, domain.RoleScheduler)
	_, err := server.GenerateTasks(ctx, &amssv1.GenerateTasksRequest{OrgId: orgID.String()})
	requireStatusCode(t, err, codes.PermissionDenied)
}

func contextWithPrincipal(orgID uuid.UUID, role domain.Role) context.Context {
	return contextWithPrincipalForUser(orgID, uuid.New(), role)
}

func contextWithPrincipalForUser(orgID, userID uuid.UUID, role domain.Role) context.Context {
	return WithPrincipal(context.Background(), Principal{
		UserID: userID,
		OrgID:  orgID,
		Role:   role,
	})
}

func requireStatusCode(t *testing.T, err error, code codes.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error")
	}
	statusErr, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error, got %v", err)
	}
	if statusErr.Code() != code {
		t.Fatalf("expected code %v, got %v", code, statusErr.Code())
	}
}

type fakeTaskRepo struct {
	tasks map[uuid.UUID]domain.MaintenanceTask
}

func newFakeTaskRepo() *fakeTaskRepo {
	return &fakeTaskRepo{tasks: make(map[uuid.UUID]domain.MaintenanceTask)}
}

func (f *fakeTaskRepo) GetByID(_ context.Context, orgID, id uuid.UUID) (domain.MaintenanceTask, error) {
	task, ok := f.tasks[id]
	if !ok || task.OrgID != orgID || task.DeletedAt != nil {
		return domain.MaintenanceTask{}, domain.ErrNotFound
	}
	return task, nil
}

func (f *fakeTaskRepo) Create(_ context.Context, task domain.MaintenanceTask) (domain.MaintenanceTask, error) {
	f.tasks[task.ID] = task
	return task, nil
}

func (f *fakeTaskRepo) Update(_ context.Context, task domain.MaintenanceTask) (domain.MaintenanceTask, error) {
	f.tasks[task.ID] = task
	return task, nil
}

func (f *fakeTaskRepo) SoftDelete(_ context.Context, orgID, id uuid.UUID, at time.Time) error {
	task, ok := f.tasks[id]
	if !ok || task.OrgID != orgID {
		return domain.ErrNotFound
	}
	task.DeletedAt = &at
	f.tasks[id] = task
	return nil
}

func (f *fakeTaskRepo) List(_ context.Context, _ ports.TaskFilter) ([]domain.MaintenanceTask, error) {
	return nil, nil
}

func (f *fakeTaskRepo) UpdateState(_ context.Context, orgID, id uuid.UUID, newState domain.TaskState, notes string, now time.Time) (domain.MaintenanceTask, error) {
	task, ok := f.tasks[id]
	if !ok || task.OrgID != orgID {
		return domain.MaintenanceTask{}, domain.ErrNotFound
	}
	task.State = newState
	task.Notes = notes
	task.UpdatedAt = now
	f.tasks[id] = task
	return task, nil
}

func (f *fakeTaskRepo) HasActiveForProgram(_ context.Context, _ uuid.UUID, _ uuid.UUID) (bool, error) {
	return false, nil
}

type fakeAircraftRepo struct {
	statuses map[uuid.UUID]domain.AircraftStatus
}

func (f *fakeAircraftRepo) GetStatus(_ context.Context, _ uuid.UUID, id uuid.UUID) (domain.AircraftStatus, error) {
	status, ok := f.statuses[id]
	if !ok {
		return "", domain.ErrNotFound
	}
	return status, nil
}

func (f *fakeAircraftRepo) GetByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (domain.Aircraft, error) {
	return domain.Aircraft{}, domain.ErrNotFound
}

func (f *fakeAircraftRepo) GetByTailNumber(_ context.Context, _ uuid.UUID, _ string) (domain.Aircraft, error) {
	return domain.Aircraft{}, domain.ErrNotFound
}

func (f *fakeAircraftRepo) Create(_ context.Context, aircraft domain.Aircraft) (domain.Aircraft, error) {
	return aircraft, nil
}

func (f *fakeAircraftRepo) Update(_ context.Context, aircraft domain.Aircraft) (domain.Aircraft, error) {
	return aircraft, nil
}

func (f *fakeAircraftRepo) SoftDelete(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time) error {
	return nil
}

func (f *fakeAircraftRepo) List(_ context.Context, _ ports.AircraftFilter) ([]domain.Aircraft, error) {
	return nil, nil
}

type fakeReservationRepo struct {
	items map[uuid.UUID]domain.PartReservation
}

func newFakeReservationRepo() *fakeReservationRepo {
	return &fakeReservationRepo{items: make(map[uuid.UUID]domain.PartReservation)}
}

func (f *fakeReservationRepo) Create(_ context.Context, reservation domain.PartReservation) error {
	f.items[reservation.ID] = reservation
	return nil
}

func (f *fakeReservationRepo) GetByID(_ context.Context, orgID, id uuid.UUID) (domain.PartReservation, error) {
	item, ok := f.items[id]
	if !ok || item.OrgID != orgID {
		return domain.PartReservation{}, domain.ErrNotFound
	}
	return item, nil
}

func (f *fakeReservationRepo) ListByTask(_ context.Context, _ uuid.UUID, _ uuid.UUID) ([]domain.PartReservation, error) {
	return nil, nil
}

func (f *fakeReservationRepo) UpdateState(_ context.Context, orgID, id uuid.UUID, state domain.PartReservationState, now time.Time) error {
	item, ok := f.items[id]
	if !ok || item.OrgID != orgID {
		return domain.ErrNotFound
	}
	item.State = state
	item.UpdatedAt = now
	f.items[id] = item
	return nil
}

func (f *fakeReservationRepo) ReleaseByTask(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time) error {
	return nil
}

type fakePartItemRepo struct {
	items map[uuid.UUID]domain.PartItem
}

func newFakePartItemRepo() *fakePartItemRepo {
	return &fakePartItemRepo{items: make(map[uuid.UUID]domain.PartItem)}
}

func (f *fakePartItemRepo) GetByID(_ context.Context, orgID, id uuid.UUID) (domain.PartItem, error) {
	item, ok := f.items[id]
	if !ok || item.OrgID != orgID {
		return domain.PartItem{}, domain.ErrNotFound
	}
	return item, nil
}

func (f *fakePartItemRepo) GetBySerialNumber(_ context.Context, _ uuid.UUID, _ string) (domain.PartItem, error) {
	return domain.PartItem{}, domain.ErrNotFound
}

func (f *fakePartItemRepo) Create(_ context.Context, item domain.PartItem) (domain.PartItem, error) {
	f.items[item.ID] = item
	return item, nil
}

func (f *fakePartItemRepo) Update(_ context.Context, item domain.PartItem) (domain.PartItem, error) {
	f.items[item.ID] = item
	return item, nil
}

func (f *fakePartItemRepo) List(_ context.Context, _ ports.PartItemFilter) ([]domain.PartItem, error) {
	return nil, nil
}

func (f *fakePartItemRepo) SoftDelete(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time) error {
	return nil
}

func (f *fakePartItemRepo) UpdateStatus(_ context.Context, orgID, id uuid.UUID, status domain.PartItemStatus, now time.Time) error {
	item, ok := f.items[id]
	if !ok || item.OrgID != orgID {
		return domain.ErrNotFound
	}
	item.Status = status
	item.UpdatedAt = now
	f.items[id] = item
	return nil
}

type fakeLocker struct{}

type fakeLock struct{}

func (fakeLocker) Acquire(_ context.Context, _ string, _ time.Duration) (ports.Lock, error) {
	return fakeLock{}, nil
}

func (fakeLock) Release(_ context.Context) error {
	return nil
}

type fakeAuditRepo struct {
	last domain.AuditLog
}

func (f *fakeAuditRepo) Insert(_ context.Context, entry domain.AuditLog) error {
	f.last = entry
	return nil
}

type fakeProgramRepo struct {
	due []domain.MaintenanceProgram
}

func (f *fakeProgramRepo) GetByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (domain.MaintenanceProgram, error) {
	return domain.MaintenanceProgram{}, domain.ErrNotFound
}

func (f *fakeProgramRepo) Create(_ context.Context, program domain.MaintenanceProgram) (domain.MaintenanceProgram, error) {
	return program, nil
}

func (f *fakeProgramRepo) Update(_ context.Context, program domain.MaintenanceProgram) (domain.MaintenanceProgram, error) {
	return program, nil
}

func (f *fakeProgramRepo) SoftDelete(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time) error {
	return nil
}

func (f *fakeProgramRepo) List(_ context.Context, _ ports.MaintenanceProgramFilter) ([]domain.MaintenanceProgram, error) {
	return nil, nil
}

func (f *fakeProgramRepo) ListDueCalendar(_ context.Context, _ time.Time, limit int) ([]domain.MaintenanceProgram, error) {
	if limit > 0 && len(f.due) > limit {
		return f.due[:limit], nil
	}
	return f.due, nil
}

func (f *fakeProgramRepo) GetByName(_ context.Context, _ uuid.UUID, _ string, _ *uuid.UUID) (domain.MaintenanceProgram, error) {
	return domain.MaintenanceProgram{}, domain.ErrNotFound
}
