package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPostgresOrganizationAndUserRepositories(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	userRepo := &UserRepository{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Sky Ops",
		CreatedAt: now,
		UpdatedAt: now,
	}
	createdOrg, err := orgRepo.Create(ctx, org)
	if err != nil {
		t.Fatalf("create organization: %v", err)
	}
	if createdOrg.ID != org.ID {
		t.Fatalf("expected org id %s, got %s", org.ID, createdOrg.ID)
	}
	listedOrgs, err := orgRepo.List(ctx, ports.OrganizationFilter{Name: "Sky"})
	if err != nil {
		t.Fatalf("list organizations: %v", err)
	}
	if len(listedOrgs) == 0 {
		t.Fatalf("expected org list to include org")
	}

	user := domain.User{
		ID:           uuid.New(),
		OrgID:        org.ID,
		Email:        "tech@example.com",
		Role:         domain.RoleScheduler,
		PasswordHash: "hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	createdUser, err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if createdUser.ID != user.ID {
		t.Fatalf("expected user id %s, got %s", user.ID, createdUser.ID)
	}
	gotUser, err := userRepo.GetByID(ctx, org.ID, user.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if gotUser.Email != user.Email {
		t.Fatalf("expected email %s, got %s", user.Email, gotUser.Email)
	}
	users, err := userRepo.List(ctx, ports.UserFilter{OrgID: &org.ID})
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
}

func TestPostgresAircraftAndProgramRepositories(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	aircraftRepo := &AircraftRepository{DB: pool}
	programRepo := &MaintenanceProgramRepository{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Ops",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	aircraft := domain.Aircraft{
		ID:            uuid.New(),
		OrgID:         org.ID,
		TailNumber:    "N123AM",
		Model:         "A320",
		Status:        domain.AircraftOperational,
		CapacitySlots: 3,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	createdAircraft, err := aircraftRepo.Create(ctx, aircraft)
	if err != nil {
		t.Fatalf("create aircraft: %v", err)
	}
	gotByTail, err := aircraftRepo.GetByTailNumber(ctx, org.ID, "N123AM")
	if err != nil {
		t.Fatalf("get by tail number: %v", err)
	}
	if gotByTail.ID != createdAircraft.ID {
		t.Fatalf("expected aircraft id %s, got %s", createdAircraft.ID, gotByTail.ID)
	}

	lastPerformed := now.Add(-10 * 24 * time.Hour)
	program := domain.MaintenanceProgram{
		ID:            uuid.New(),
		OrgID:         org.ID,
		AircraftID:    &createdAircraft.ID,
		Name:          "A-Check",
		IntervalType:  domain.ProgramIntervalCalendar,
		IntervalValue: 7,
		LastPerformed: &lastPerformed,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	_, err = programRepo.Create(ctx, program)
	if err != nil {
		t.Fatalf("create program: %v", err)
	}
	gotProgram, err := programRepo.GetByName(ctx, org.ID, "A-Check", &createdAircraft.ID)
	if err != nil {
		t.Fatalf("get program by name: %v", err)
	}
	if gotProgram.ID != program.ID {
		t.Fatalf("expected program id %s, got %s", program.ID, gotProgram.ID)
	}
	due, err := programRepo.ListDueCalendar(ctx, now, 10)
	if err != nil {
		t.Fatalf("list due programs: %v", err)
	}
	if len(due) == 0 {
		t.Fatalf("expected due program list to include program")
	}
}

func TestPostgresWebhookAndOutboxRepositories(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	webhookRepo := &WebhookRepository{DB: pool}
	outboxRepo := &OutboxRepository{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Ops",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	hook := domain.Webhook{
		ID:        uuid.New(),
		OrgID:     org.ID,
		URL:       "https://example.com/hook",
		Events:    []string{"task.created"},
		Secret:    "secret",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := webhookRepo.Create(ctx, hook); err != nil {
		t.Fatalf("create webhook: %v", err)
	}
	listed, err := webhookRepo.ListByEvent(ctx, org.ID, "task.created")
	if err != nil {
		t.Fatalf("list by event: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(listed))
	}

	aggregateID := uuid.New()
	payload := map[string]any{"value": "ok"}
	if err := outboxRepo.Enqueue(ctx, org.ID, "task_created", "maintenance_task", aggregateID, payload, "dedupe-1"); err != nil {
		t.Fatalf("enqueue outbox: %v", err)
	}
	events, err := outboxRepo.LockPending(ctx, "worker-1", 10)
	if err != nil {
		t.Fatalf("lock pending: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 outbox event, got %d", len(events))
	}
	if err := outboxRepo.MarkProcessed(ctx, events[0].ID, time.Now().UTC()); err != nil {
		t.Fatalf("mark processed: %v", err)
	}
}

func TestPostgresTaskRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	userRepo := &UserRepository{DB: pool}
	aircraftRepo := &AircraftRepository{DB: pool}
	programRepo := &MaintenanceProgramRepository{DB: pool}
	taskRepo := &TaskRepository{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Task Ops",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	user := domain.User{
		ID:           uuid.New(),
		OrgID:        org.ID,
		Email:        "mechanic@ops.local",
		Role:         domain.RoleMechanic,
		PasswordHash: "hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	aircraft := domain.Aircraft{
		ID:            uuid.New(),
		OrgID:         org.ID,
		TailNumber:    "N400TS",
		Model:         "A321",
		Status:        domain.AircraftGrounded,
		CapacitySlots: 2,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if _, err := aircraftRepo.Create(ctx, aircraft); err != nil {
		t.Fatalf("create aircraft: %v", err)
	}

	program := domain.MaintenanceProgram{
		ID:            uuid.New(),
		OrgID:         org.ID,
		AircraftID:    &aircraft.ID,
		Name:          "B-Check",
		IntervalType:  domain.ProgramIntervalCalendar,
		IntervalValue: 60,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if _, err := programRepo.Create(ctx, program); err != nil {
		t.Fatalf("create program: %v", err)
	}

	task := domain.MaintenanceTask{
		ID:                 uuid.New(),
		OrgID:              org.ID,
		AircraftID:         aircraft.ID,
		ProgramID:          &program.ID,
		Type:               domain.TaskTypeInspection,
		State:              domain.TaskStateScheduled,
		StartTime:          now.Add(2 * time.Hour),
		EndTime:            now.Add(5 * time.Hour),
		AssignedMechanicID: &user.ID,
		Notes:              "initial task",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	created, err := taskRepo.Create(ctx, task)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if created.ID != task.ID {
		t.Fatalf("expected task id %s, got %s", task.ID, created.ID)
	}

	got, err := taskRepo.GetByID(ctx, org.ID, task.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if got.Notes != task.Notes {
		t.Fatalf("expected notes %q, got %q", task.Notes, got.Notes)
	}

	listed, err := taskRepo.List(ctx, ports.TaskFilter{OrgID: &org.ID, AircraftID: &aircraft.ID})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 task, got %d", len(listed))
	}

	task.Notes = "updated task"
	task.UpdatedAt = now.Add(2 * time.Minute)
	updated, err := taskRepo.Update(ctx, task)
	if err != nil {
		t.Fatalf("update task: %v", err)
	}
	if updated.Notes != "updated task" {
		t.Fatalf("expected updated notes, got %q", updated.Notes)
	}

	stateUpdated, err := taskRepo.UpdateState(ctx, org.ID, task.ID, domain.TaskStateInProgress, "started", now.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("update state: %v", err)
	}
	if stateUpdated.State != domain.TaskStateInProgress {
		t.Fatalf("expected state %s, got %s", domain.TaskStateInProgress, stateUpdated.State)
	}

	hasActive, err := taskRepo.HasActiveForProgram(ctx, org.ID, program.ID)
	if err != nil {
		t.Fatalf("has active: %v", err)
	}
	if !hasActive {
		t.Fatalf("expected active task for program")
	}

	if err := taskRepo.SoftDelete(ctx, org.ID, task.ID, now.Add(4*time.Minute)); err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	_, err = taskRepo.GetByID(ctx, org.ID, task.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
	hasActive, err = taskRepo.HasActiveForProgram(ctx, org.ID, program.ID)
	if err != nil {
		t.Fatalf("has active after delete: %v", err)
	}
	if hasActive {
		t.Fatalf("expected no active tasks after delete")
	}
}

func TestPostgresPartRepositories(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	aircraftRepo := &AircraftRepository{DB: pool}
	taskRepo := &TaskRepository{DB: pool}
	defRepo := &PartDefinitionRepository{DB: pool}
	itemRepo := &PartItemRepository{DB: pool}
	reservationRepo := &PartReservationRepository{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Parts Ops",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	aircraft := domain.Aircraft{
		ID:            uuid.New(),
		OrgID:         org.ID,
		TailNumber:    "N100PT",
		Model:         "737",
		Status:        domain.AircraftGrounded,
		CapacitySlots: 1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if _, err := aircraftRepo.Create(ctx, aircraft); err != nil {
		t.Fatalf("create aircraft: %v", err)
	}

	task := domain.MaintenanceTask{
		ID:         uuid.New(),
		OrgID:      org.ID,
		AircraftID: aircraft.ID,
		Type:       domain.TaskTypeRepair,
		State:      domain.TaskStateScheduled,
		StartTime:  now.Add(1 * time.Hour),
		EndTime:    now.Add(3 * time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if _, err := taskRepo.Create(ctx, task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	def := domain.PartDefinition{
		ID:        uuid.New(),
		OrgID:     org.ID,
		Name:      "Hydraulic Pump",
		Category:  "Hydraulics",
		CreatedAt: now,
		UpdatedAt: now,
	}
	createdDef, err := defRepo.Create(ctx, def)
	if err != nil {
		t.Fatalf("create part definition: %v", err)
	}
	if createdDef.ID != def.ID {
		t.Fatalf("expected definition id %s, got %s", def.ID, createdDef.ID)
	}

	expiry := now.Add(30 * 24 * time.Hour)
	item := domain.PartItem{
		ID:           uuid.New(),
		OrgID:        org.ID,
		DefinitionID: def.ID,
		SerialNumber: "SN-100",
		Status:       domain.PartItemInStock,
		ExpiryDate:   &expiry,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	createdItem, err := itemRepo.Create(ctx, item)
	if err != nil {
		t.Fatalf("create part item: %v", err)
	}
	gotItem, err := itemRepo.GetByID(ctx, org.ID, item.ID)
	if err != nil {
		t.Fatalf("get part item: %v", err)
	}
	if gotItem.SerialNumber != item.SerialNumber {
		t.Fatalf("expected serial %s, got %s", item.SerialNumber, gotItem.SerialNumber)
	}
	bySerial, err := itemRepo.GetBySerialNumber(ctx, org.ID, item.SerialNumber)
	if err != nil {
		t.Fatalf("get by serial: %v", err)
	}
	if bySerial.ID != createdItem.ID {
		t.Fatalf("expected item id %s, got %s", createdItem.ID, bySerial.ID)
	}

	items, err := itemRepo.List(ctx, ports.PartItemFilter{OrgID: &org.ID})
	if err != nil {
		t.Fatalf("list part items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item.Status = domain.PartItemUsed
	item.UpdatedAt = now.Add(2 * time.Minute)
	updatedItem, err := itemRepo.Update(ctx, item)
	if err != nil {
		t.Fatalf("update part item: %v", err)
	}
	if updatedItem.Status != domain.PartItemUsed {
		t.Fatalf("expected status %s, got %s", domain.PartItemUsed, updatedItem.Status)
	}

	item2 := domain.PartItem{
		ID:           uuid.New(),
		OrgID:        org.ID,
		DefinitionID: def.ID,
		SerialNumber: "SN-200",
		Status:       domain.PartItemInStock,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := itemRepo.Create(ctx, item2); err != nil {
		t.Fatalf("create part item 2: %v", err)
	}

	reservation := domain.PartReservation{
		ID:         uuid.New(),
		OrgID:      org.ID,
		TaskID:     task.ID,
		PartItemID: item.ID,
		State:      domain.ReservationReserved,
		Quantity:   1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := reservationRepo.Create(ctx, reservation); err != nil {
		t.Fatalf("create reservation: %v", err)
	}
	gotReservation, err := reservationRepo.GetByID(ctx, org.ID, reservation.ID)
	if err != nil {
		t.Fatalf("get reservation: %v", err)
	}
	if gotReservation.State != domain.ReservationReserved {
		t.Fatalf("expected reservation state %s, got %s", domain.ReservationReserved, gotReservation.State)
	}
	listedReservations, err := reservationRepo.ListByTask(ctx, org.ID, task.ID)
	if err != nil {
		t.Fatalf("list reservations: %v", err)
	}
	if len(listedReservations) != 1 {
		t.Fatalf("expected 1 reservation, got %d", len(listedReservations))
	}

	if err := reservationRepo.UpdateState(ctx, org.ID, reservation.ID, domain.ReservationUsed, now.Add(3*time.Minute)); err != nil {
		t.Fatalf("update reservation state: %v", err)
	}
	gotReservation, err = reservationRepo.GetByID(ctx, org.ID, reservation.ID)
	if err != nil {
		t.Fatalf("get reservation after update: %v", err)
	}
	if gotReservation.State != domain.ReservationUsed {
		t.Fatalf("expected reservation state %s, got %s", domain.ReservationUsed, gotReservation.State)
	}

	reservation2 := domain.PartReservation{
		ID:         uuid.New(),
		OrgID:      org.ID,
		TaskID:     task.ID,
		PartItemID: item2.ID,
		State:      domain.ReservationReserved,
		Quantity:   1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := reservationRepo.Create(ctx, reservation2); err != nil {
		t.Fatalf("create reservation 2: %v", err)
	}
	if err := reservationRepo.ReleaseByTask(ctx, org.ID, task.ID, now.Add(4*time.Minute)); err != nil {
		t.Fatalf("release by task: %v", err)
	}
	released, err := reservationRepo.GetByID(ctx, org.ID, reservation2.ID)
	if err != nil {
		t.Fatalf("get reservation after release: %v", err)
	}
	if released.State != domain.ReservationReleased {
		t.Fatalf("expected reservation state %s, got %s", domain.ReservationReleased, released.State)
	}

	if err := defRepo.SoftDelete(ctx, org.ID, def.ID, now.Add(5*time.Minute)); err != nil {
		t.Fatalf("soft delete definition: %v", err)
	}
	_, err = defRepo.GetByID(ctx, org.ID, def.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected definition not found after delete, got %v", err)
	}
	if err := itemRepo.SoftDelete(ctx, org.ID, item2.ID, now.Add(6*time.Minute)); err != nil {
		t.Fatalf("soft delete item: %v", err)
	}
	_, err = itemRepo.GetByID(ctx, org.ID, item2.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected item not found after delete, got %v", err)
	}
}

func TestPostgresComplianceRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	userRepo := &UserRepository{DB: pool}
	aircraftRepo := &AircraftRepository{DB: pool}
	taskRepo := &TaskRepository{DB: pool}
	complianceRepo := &ComplianceRepository{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Compliance Ops",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}
	user := domain.User{
		ID:           uuid.New(),
		OrgID:        org.ID,
		Email:        "mech@ops.local",
		Role:         domain.RoleMechanic,
		PasswordHash: "hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	aircraft := domain.Aircraft{
		ID:            uuid.New(),
		OrgID:         org.ID,
		TailNumber:    "N900CM",
		Model:         "A220",
		Status:        domain.AircraftGrounded,
		CapacitySlots: 1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if _, err := aircraftRepo.Create(ctx, aircraft); err != nil {
		t.Fatalf("create aircraft: %v", err)
	}
	task := domain.MaintenanceTask{
		ID:         uuid.New(),
		OrgID:      org.ID,
		AircraftID: aircraft.ID,
		Type:       domain.TaskTypeInspection,
		State:      domain.TaskStateScheduled,
		StartTime:  now.Add(1 * time.Hour),
		EndTime:    now.Add(2 * time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if _, err := taskRepo.Create(ctx, task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	item := domain.ComplianceItem{
		ID:          uuid.New(),
		OrgID:       org.ID,
		TaskID:      task.ID,
		Description: "Torque check",
		Result:      domain.CompliancePending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := complianceRepo.Create(ctx, item); err != nil {
		t.Fatalf("create compliance item: %v", err)
	}
	got, err := complianceRepo.GetByID(ctx, org.ID, item.ID)
	if err != nil {
		t.Fatalf("get compliance item: %v", err)
	}
	if got.Description != item.Description {
		t.Fatalf("expected description %q, got %q", item.Description, got.Description)
	}
	byTask, err := complianceRepo.ListByTask(ctx, org.ID, task.ID)
	if err != nil {
		t.Fatalf("list compliance by task: %v", err)
	}
	if len(byTask) != 1 {
		t.Fatalf("expected 1 compliance item, got %d", len(byTask))
	}
	result := domain.CompliancePending
	listed, err := complianceRepo.List(ctx, ports.ComplianceFilter{OrgID: &org.ID, Result: &result})
	if err != nil {
		t.Fatalf("list compliance: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 compliance item, got %d", len(listed))
	}

	item.Description = "Torque check updated"
	item.Result = domain.CompliancePass
	item.UpdatedAt = now.Add(2 * time.Minute)
	if err := complianceRepo.Update(ctx, item); err != nil {
		t.Fatalf("update compliance item: %v", err)
	}

	signAt := now.Add(3 * time.Minute)
	if err := complianceRepo.SignOff(ctx, org.ID, item.ID, user.ID, signAt); err != nil {
		t.Fatalf("sign off: %v", err)
	}
	signed, err := complianceRepo.GetByID(ctx, org.ID, item.ID)
	if err != nil {
		t.Fatalf("get compliance after sign off: %v", err)
	}
	if signed.SignOffUserID == nil || *signed.SignOffUserID != user.ID {
		t.Fatalf("expected sign off user %s", user.ID)
	}
	if signed.SignOffTime == nil || signed.SignOffTime.UTC().Truncate(time.Microsecond) != signAt.UTC().Truncate(time.Microsecond) {
		t.Fatalf("expected sign off time %v", signAt)
	}
}

func TestPostgresImportRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	userRepo := &UserRepository{DB: pool}
	importRepo := &ImportRepository{DB: pool}
	rowRepo := &ImportRowRepository{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Import Ops",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}
	user := domain.User{
		ID:           uuid.New(),
		OrgID:        org.ID,
		Email:        "importer@ops.local",
		Role:         domain.RoleScheduler,
		PasswordHash: "hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	imp := domain.Import{
		ID:        uuid.New(),
		OrgID:     org.ID,
		Type:      domain.ImportTypeAircraft,
		Status:    domain.ImportStatusPending,
		FileName:  "aircraft.csv",
		FilePath:  "/tmp/aircraft.csv",
		CreatedBy: user.ID,
		Summary:   map[string]any{"rows": 1},
		CreatedAt: now,
		UpdatedAt: now,
	}
	created, err := importRepo.Create(ctx, imp)
	if err != nil {
		t.Fatalf("create import: %v", err)
	}
	if created.ID != imp.ID {
		t.Fatalf("expected import id %s, got %s", imp.ID, created.ID)
	}
	got, err := importRepo.GetByID(ctx, org.ID, imp.ID)
	if err != nil {
		t.Fatalf("get import: %v", err)
	}
	if got.FilePath != imp.FilePath {
		t.Fatalf("expected file path %s, got %s", imp.FilePath, got.FilePath)
	}

	imp.Status = domain.ImportStatusApplying
	imp.FileName = "aircraft_v2.csv"
	imp.FilePath = "/tmp/aircraft_v2.csv"
	imp.Summary = map[string]any{"rows": 2}
	imp.UpdatedAt = now.Add(2 * time.Minute)
	updated, err := importRepo.Update(ctx, imp)
	if err != nil {
		t.Fatalf("update import: %v", err)
	}
	if updated.Status != domain.ImportStatusApplying {
		t.Fatalf("expected status %s, got %s", domain.ImportStatusApplying, updated.Status)
	}

	if err := importRepo.UpdateStatus(ctx, org.ID, imp.ID, domain.ImportStatusCompleted, map[string]any{"rows": 2}, now.Add(3*time.Minute)); err != nil {
		t.Fatalf("update status: %v", err)
	}
	got, err = importRepo.GetByID(ctx, org.ID, imp.ID)
	if err != nil {
		t.Fatalf("get import after update status: %v", err)
	}
	if got.Status != domain.ImportStatusCompleted {
		t.Fatalf("expected status %s, got %s", domain.ImportStatusCompleted, got.Status)
	}

	row1 := domain.ImportRow{
		ID:        uuid.New(),
		OrgID:     org.ID,
		ImportID:  imp.ID,
		RowNumber: 1,
		Raw:       map[string]any{"tail_number": "N1"},
		Status:    domain.ImportRowValid,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := rowRepo.Create(ctx, row1); err != nil {
		t.Fatalf("create import row: %v", err)
	}
	row2 := domain.ImportRow{
		ID:        uuid.New(),
		OrgID:     org.ID,
		ImportID:  imp.ID,
		RowNumber: 2,
		Raw:       map[string]any{"tail_number": ""},
		Status:    domain.ImportRowInvalid,
		Errors:    []string{"missing tail number"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := rowRepo.Create(ctx, row2); err != nil {
		t.Fatalf("create import row 2: %v", err)
	}

	rows, err := rowRepo.ListByImport(ctx, ports.ImportRowFilter{OrgID: org.ID, ImportID: imp.ID})
	if err != nil {
		t.Fatalf("list import rows: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	status := domain.ImportRowInvalid
	filtered, err := rowRepo.ListByImport(ctx, ports.ImportRowFilter{OrgID: org.ID, ImportID: imp.ID, Status: &status})
	if err != nil {
		t.Fatalf("list import rows by status: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("expected 1 invalid row, got %d", len(filtered))
	}
}

func TestPostgresOutboxRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	outboxRepo := &OutboxRepository{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Outbox Ops",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	aggregateID := uuid.New()
	if err := outboxRepo.Enqueue(ctx, org.ID, "task_created", "maintenance_task", aggregateID, map[string]any{"value": "ok"}, "dedupe-2"); err != nil {
		t.Fatalf("enqueue outbox: %v", err)
	}
	events, err := outboxRepo.LockPending(ctx, "worker-2", 10)
	if err != nil {
		t.Fatalf("lock pending: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 outbox event, got %d", len(events))
	}
	event := events[0]
	got, err := outboxRepo.GetByID(ctx, org.ID, event.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.DedupeKey != "dedupe-2" {
		t.Fatalf("expected dedupe key dedupe-2, got %s", got.DedupeKey)
	}

	nextAttempt := now.Add(10 * time.Minute)
	if err := outboxRepo.ScheduleRetry(ctx, event.ID, 2, nextAttempt, "boom"); err != nil {
		t.Fatalf("schedule retry: %v", err)
	}
	got, err = outboxRepo.GetByID(ctx, org.ID, event.ID)
	if err != nil {
		t.Fatalf("get after retry: %v", err)
	}
	if got.AttemptCount != 2 {
		t.Fatalf("expected attempt count 2, got %d", got.AttemptCount)
	}
	if got.LastError == nil || *got.LastError != "boom" {
		t.Fatalf("expected last error boom")
	}

	if err := outboxRepo.MarkProcessed(ctx, event.ID, now.Add(20*time.Minute)); err != nil {
		t.Fatalf("mark processed: %v", err)
	}
	got, err = outboxRepo.GetByID(ctx, org.ID, event.ID)
	if err != nil {
		t.Fatalf("get after processed: %v", err)
	}
	if got.ProcessedAt == nil {
		t.Fatalf("expected processed_at to be set")
	}
}

func TestPostgresWebhookDeliveryRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	webhookRepo := &WebhookRepository{DB: pool}
	deliveryRepo := &WebhookDeliveryRepository{DB: pool}
	outboxRepo := &OutboxRepository{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Webhook Ops",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	hook := domain.Webhook{
		ID:        uuid.New(),
		OrgID:     org.ID,
		URL:       "https://example.com/hook",
		Events:    []string{"task.created"},
		Secret:    "secret",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := webhookRepo.Create(ctx, hook); err != nil {
		t.Fatalf("create webhook: %v", err)
	}

	aggregateID := uuid.New()
	if err := outboxRepo.Enqueue(ctx, org.ID, "task_created", "maintenance_task", aggregateID, map[string]any{"value": "ok"}, "dedupe-3"); err != nil {
		t.Fatalf("enqueue outbox: %v", err)
	}
	events, err := outboxRepo.LockPending(ctx, "worker-3", 10)
	if err != nil {
		t.Fatalf("lock pending: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 outbox event, got %d", len(events))
	}

	delivery := domain.WebhookDelivery{
		ID:            uuid.New(),
		OrgID:         org.ID,
		WebhookID:     hook.ID,
		EventID:       events[0].ID,
		AttemptCount:  0,
		NextAttemptAt: now.Add(-1 * time.Minute),
		Status:        domain.WebhookDeliveryPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := deliveryRepo.Create(ctx, delivery); err != nil {
		t.Fatalf("create delivery: %v", err)
	}

	lockUntil := now.Add(5 * time.Minute)
	pending, err := deliveryRepo.ClaimPending(ctx, 5, lockUntil)
	if err != nil {
		t.Fatalf("claim pending: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending delivery, got %d", len(pending))
	}
	if pending[0].ID != delivery.ID {
		t.Fatalf("expected delivery id %s, got %s", delivery.ID, pending[0].ID)
	}
	var nextAttempt time.Time
	if err := pool.QueryRow(ctx, "SELECT next_attempt_at FROM webhook_deliveries WHERE id=$1", delivery.ID).Scan(&nextAttempt); err != nil {
		t.Fatalf("select next_attempt_at: %v", err)
	}
	if nextAttempt.UTC().Truncate(time.Microsecond) != lockUntil.UTC().Truncate(time.Microsecond) {
		t.Fatalf("expected next attempt %v, got %v", lockUntil, nextAttempt)
	}

	statusCode := 200
	body := "ok"
	delivery.AttemptCount = 1
	delivery.Status = domain.WebhookDeliveryDelivered
	delivery.LastResponseCode = &statusCode
	delivery.LastResponseBody = &body
	delivery.UpdatedAt = now.Add(6 * time.Minute)
	if err := deliveryRepo.Update(ctx, delivery); err != nil {
		t.Fatalf("update delivery: %v", err)
	}

	var status string
	var attempts int
	var responseCode *int
	if err := pool.QueryRow(ctx, "SELECT status, attempt_count, last_response_code FROM webhook_deliveries WHERE id=$1", delivery.ID).Scan(&status, &attempts, &responseCode); err != nil {
		t.Fatalf("select delivery: %v", err)
	}
	if status != string(domain.WebhookDeliveryDelivered) {
		t.Fatalf("expected status %s, got %s", domain.WebhookDeliveryDelivered, status)
	}
	if attempts != 1 {
		t.Fatalf("expected attempt count 1, got %d", attempts)
	}
	if responseCode == nil || *responseCode != 200 {
		t.Fatalf("expected response code 200")
	}
}

func TestPostgresTaskOverlapConflict(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	aircraftRepo := &AircraftRepository{DB: pool}
	taskRepo := &TaskRepository{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Overlap Ops",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	aircraft := domain.Aircraft{
		ID:            uuid.New(),
		OrgID:         org.ID,
		TailNumber:    "N777OV",
		Model:         "A320",
		Status:        domain.AircraftOperational,
		CapacitySlots: 2,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if _, err := aircraftRepo.Create(ctx, aircraft); err != nil {
		t.Fatalf("create aircraft: %v", err)
	}

	task1 := domain.MaintenanceTask{
		ID:         uuid.New(),
		OrgID:      org.ID,
		AircraftID: aircraft.ID,
		Type:       domain.TaskTypeInspection,
		State:      domain.TaskStateScheduled,
		StartTime:  now.Add(1 * time.Hour),
		EndTime:    now.Add(3 * time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if _, err := taskRepo.Create(ctx, task1); err != nil {
		t.Fatalf("create task1: %v", err)
	}

	task2 := task1
	task2.ID = uuid.New()
	task2.StartTime = now.Add(2 * time.Hour)
	task2.EndTime = now.Add(4 * time.Hour)
	if _, err := taskRepo.Create(ctx, task2); err == nil || !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected overlap conflict, got %v", err)
	}

	task3 := task1
	task3.ID = uuid.New()
	task3.StartTime = now.Add(4 * time.Hour)
	task3.EndTime = now.Add(5 * time.Hour)
	if _, err := taskRepo.Create(ctx, task3); err != nil {
		t.Fatalf("create non-overlapping task: %v", err)
	}
}

func TestPostgresAuditLogsImmutable(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	userRepo := &UserRepository{DB: pool}
	auditRepo := &AuditRepository{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Audit Ops",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	user := domain.User{
		ID:           uuid.New(),
		OrgID:        org.ID,
		Email:        "audit@ops.local",
		Role:         domain.RoleAuditor,
		PasswordHash: "hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	entry := domain.AuditLog{
		ID:            uuid.New(),
		OrgID:         org.ID,
		EntityType:    "maintenance_task",
		EntityID:      uuid.New(),
		Action:        domain.AuditActionCreate,
		UserID:        user.ID,
		RequestID:     uuid.New(),
		IPAddress:     "127.0.0.1",
		UserAgent:     "integration-test",
		EntityVersion: 1,
		Timestamp:     now,
		Details:       map[string]any{"note": "created"},
	}
	if err := auditRepo.Insert(ctx, entry); err != nil {
		t.Fatalf("insert audit log: %v", err)
	}

	if _, err := pool.Exec(ctx, "UPDATE audit_logs SET action='update' WHERE id=$1", entry.ID); err == nil {
		t.Fatalf("expected update to be rejected for immutable audit log")
	}
	if _, err := pool.Exec(ctx, "DELETE FROM audit_logs WHERE id=$1", entry.ID); err == nil {
		t.Fatalf("expected delete to be rejected for immutable audit log")
	}

	var count int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs WHERE id=$1", entry.ID).Scan(&count); err != nil {
		t.Fatalf("count audit logs: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected audit log to remain, got count %d", count)
	}
}

func TestPostgresComplianceImmutableAfterSignOff(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	userRepo := &UserRepository{DB: pool}
	aircraftRepo := &AircraftRepository{DB: pool}
	taskRepo := &TaskRepository{DB: pool}
	complianceRepo := &ComplianceRepository{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Compliance Lock",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	user := domain.User{
		ID:           uuid.New(),
		OrgID:        org.ID,
		Email:        "mechanic@ops.local",
		Role:         domain.RoleMechanic,
		PasswordHash: "hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	aircraft := domain.Aircraft{
		ID:            uuid.New(),
		OrgID:         org.ID,
		TailNumber:    "N998CM",
		Model:         "A321",
		Status:        domain.AircraftMaintenance,
		CapacitySlots: 1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if _, err := aircraftRepo.Create(ctx, aircraft); err != nil {
		t.Fatalf("create aircraft: %v", err)
	}

	task := domain.MaintenanceTask{
		ID:         uuid.New(),
		OrgID:      org.ID,
		AircraftID: aircraft.ID,
		Type:       domain.TaskTypeInspection,
		State:      domain.TaskStateScheduled,
		StartTime:  now.Add(1 * time.Hour),
		EndTime:    now.Add(2 * time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if _, err := taskRepo.Create(ctx, task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	item := domain.ComplianceItem{
		ID:          uuid.New(),
		OrgID:       org.ID,
		TaskID:      task.ID,
		Description: "Torque check",
		Result:      domain.CompliancePending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := complianceRepo.Create(ctx, item); err != nil {
		t.Fatalf("create compliance item: %v", err)
	}
	signAt := now.Add(2 * time.Minute)
	if err := complianceRepo.SignOff(ctx, org.ID, item.ID, user.ID, signAt); err != nil {
		t.Fatalf("sign off: %v", err)
	}

	item.Description = "Updated description"
	item.Result = domain.ComplianceFail
	item.UpdatedAt = now.Add(3 * time.Minute)
	if err := complianceRepo.Update(ctx, item); err == nil {
		t.Fatalf("expected update to be rejected after sign off")
	}
}

func TestPostgresIdempotencyStoreConflict(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	store := &IdempotencyStore{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Idempotency Ops",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	record := ports.IdempotencyRecord{
		OrgID:       org.ID,
		Key:         "idem-key",
		Endpoint:    "POST /maintenance-tasks",
		RequestHash: "hash-a",
		CreatedAt:   now,
		ExpiresAt:   now.Add(24 * time.Hour),
	}
	if err := store.CreatePlaceholder(ctx, record); err != nil {
		t.Fatalf("create placeholder: %v", err)
	}

	record.RequestHash = "hash-b"
	if err := store.CreatePlaceholder(ctx, record); err == nil {
		t.Fatalf("expected unique conflict on duplicate idempotency key")
	}
}

func TestPostgresWebhookDeliveryClaimPendingSkipsFuture(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	webhookRepo := &WebhookRepository{DB: pool}
	deliveryRepo := &WebhookDeliveryRepository{DB: pool}
	outboxRepo := &OutboxRepository{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Webhook Pending",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	hook := domain.Webhook{
		ID:        uuid.New(),
		OrgID:     org.ID,
		URL:       "https://example.com/hook",
		Events:    []string{"task.created"},
		Secret:    "secret",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := webhookRepo.Create(ctx, hook); err != nil {
		t.Fatalf("create webhook: %v", err)
	}

	if err := outboxRepo.Enqueue(ctx, org.ID, "task_created", "maintenance_task", uuid.New(), map[string]any{"value": "ok"}, "dedupe-pending"); err != nil {
		t.Fatalf("enqueue outbox: %v", err)
	}
	var eventID uuid.UUID
	if err := pool.QueryRow(ctx, "SELECT id FROM outbox_events WHERE org_id=$1 AND dedupe_key=$2", org.ID, "dedupe-pending").Scan(&eventID); err != nil {
		t.Fatalf("fetch event id: %v", err)
	}

	due := domain.WebhookDelivery{
		ID:            uuid.New(),
		OrgID:         org.ID,
		WebhookID:     hook.ID,
		EventID:       eventID,
		AttemptCount:  0,
		NextAttemptAt: now.Add(-1 * time.Minute),
		Status:        domain.WebhookDeliveryPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := deliveryRepo.Create(ctx, due); err != nil {
		t.Fatalf("create due delivery: %v", err)
	}
	future := domain.WebhookDelivery{
		ID:            uuid.New(),
		OrgID:         org.ID,
		WebhookID:     hook.ID,
		EventID:       eventID,
		AttemptCount:  0,
		NextAttemptAt: now.Add(2 * time.Hour),
		Status:        domain.WebhookDeliveryPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := deliveryRepo.Create(ctx, future); err != nil {
		t.Fatalf("create future delivery: %v", err)
	}

	lockUntil := now.Add(10 * time.Minute)
	pending, err := deliveryRepo.ClaimPending(ctx, 10, lockUntil)
	if err != nil {
		t.Fatalf("claim pending: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending delivery, got %d", len(pending))
	}
	if pending[0].ID != due.ID {
		t.Fatalf("expected due delivery, got %s", pending[0].ID)
	}

	var dueNext time.Time
	if err := pool.QueryRow(ctx, "SELECT next_attempt_at FROM webhook_deliveries WHERE id=$1", due.ID).Scan(&dueNext); err != nil {
		t.Fatalf("read due next_attempt_at: %v", err)
	}
	if dueNext.UTC().Truncate(time.Microsecond) != lockUntil.UTC().Truncate(time.Microsecond) {
		t.Fatalf("expected due next_attempt_at %v, got %v", lockUntil, dueNext)
	}

	var futureNext time.Time
	if err := pool.QueryRow(ctx, "SELECT next_attempt_at FROM webhook_deliveries WHERE id=$1", future.ID).Scan(&futureNext); err != nil {
		t.Fatalf("read future next_attempt_at: %v", err)
	}
	if futureNext.UTC().Truncate(time.Microsecond) != future.NextAttemptAt.UTC().Truncate(time.Microsecond) {
		t.Fatalf("expected future next_attempt_at %v, got %v", future.NextAttemptAt, futureNext)
	}
}

func TestPostgresOutboxLockPendingSkipsFuture(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	outboxRepo := &OutboxRepository{DB: pool}
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Outbox Pending",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	if err := outboxRepo.Enqueue(ctx, org.ID, "task_created", "maintenance_task", uuid.New(), map[string]any{"value": "ok"}, "dedupe-due"); err != nil {
		t.Fatalf("enqueue due: %v", err)
	}
	if err := outboxRepo.Enqueue(ctx, org.ID, "task_created", "maintenance_task", uuid.New(), map[string]any{"value": "ok"}, "dedupe-future"); err != nil {
		t.Fatalf("enqueue future: %v", err)
	}
	future := now.Add(2 * time.Hour)
	if _, err := pool.Exec(ctx, "UPDATE outbox_events SET next_attempt_at=$1 WHERE org_id=$2 AND dedupe_key=$3", future, org.ID, "dedupe-future"); err != nil {
		t.Fatalf("set future next_attempt_at: %v", err)
	}

	events, err := outboxRepo.LockPending(ctx, "worker-skip", 10)
	if err != nil {
		t.Fatalf("lock pending: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 due event, got %d", len(events))
	}
	if events[0].DedupeKey != "dedupe-due" {
		t.Fatalf("expected dedupe-due event, got %s", events[0].DedupeKey)
	}
}

func TestPostgresRetentionCleanup(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orgRepo := &OrganizationRepository{DB: pool}
	userRepo := &UserRepository{DB: pool}
	aircraftRepo := &AircraftRepository{DB: pool}
	programRepo := &MaintenanceProgramRepository{DB: pool}
	taskRepo := &TaskRepository{DB: pool}
	defRepo := &PartDefinitionRepository{DB: pool}
	itemRepo := &PartItemRepository{DB: pool}
	reservationRepo := &PartReservationRepository{DB: pool}
	complianceRepo := &ComplianceRepository{DB: pool}
	retentionRepo := &RetentionRepository{DB: pool}

	now := time.Now().UTC()
	deletedAt := now.Add(-48 * time.Hour)
	cutoff := now.Add(-24 * time.Hour)

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Retention Ops",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	user := domain.User{
		ID:           uuid.New(),
		OrgID:        org.ID,
		Email:        "cleanup@ops.local",
		Role:         domain.RoleMechanic,
		PasswordHash: "hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	aircraft := domain.Aircraft{
		ID:            uuid.New(),
		OrgID:         org.ID,
		TailNumber:    "N100RC",
		Model:         "A320",
		Status:        domain.AircraftMaintenance,
		CapacitySlots: 2,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if _, err := aircraftRepo.Create(ctx, aircraft); err != nil {
		t.Fatalf("create aircraft: %v", err)
	}

	program := domain.MaintenanceProgram{
		ID:            uuid.New(),
		OrgID:         org.ID,
		AircraftID:    &aircraft.ID,
		Name:          "C-Check",
		IntervalType:  domain.ProgramIntervalCalendar,
		IntervalValue: 90,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if _, err := programRepo.Create(ctx, program); err != nil {
		t.Fatalf("create program: %v", err)
	}

	task := domain.MaintenanceTask{
		ID:                 uuid.New(),
		OrgID:              org.ID,
		AircraftID:         aircraft.ID,
		ProgramID:          &program.ID,
		Type:               domain.TaskTypeInspection,
		State:              domain.TaskStateScheduled,
		StartTime:          now.Add(2 * time.Hour),
		EndTime:            now.Add(5 * time.Hour),
		AssignedMechanicID: &user.ID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if _, err := taskRepo.Create(ctx, task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	definition := domain.PartDefinition{
		ID:        uuid.New(),
		OrgID:     org.ID,
		Name:      "Fuel Pump",
		Category:  "Fuel",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := defRepo.Create(ctx, definition); err != nil {
		t.Fatalf("create part definition: %v", err)
	}

	item := domain.PartItem{
		ID:           uuid.New(),
		OrgID:        org.ID,
		DefinitionID: definition.ID,
		SerialNumber: "RC-100",
		Status:       domain.PartItemInStock,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := itemRepo.Create(ctx, item); err != nil {
		t.Fatalf("create part item: %v", err)
	}

	reservation := domain.PartReservation{
		ID:         uuid.New(),
		OrgID:      org.ID,
		TaskID:     task.ID,
		PartItemID: item.ID,
		State:      domain.ReservationReserved,
		Quantity:   1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := reservationRepo.Create(ctx, reservation); err != nil {
		t.Fatalf("create reservation: %v", err)
	}

	compliance := domain.ComplianceItem{
		ID:          uuid.New(),
		OrgID:       org.ID,
		TaskID:      task.ID,
		Description: "Leak check",
		Result:      domain.CompliancePending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := complianceRepo.Create(ctx, compliance); err != nil {
		t.Fatalf("create compliance item: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE compliance_items
		SET deleted_at=$1, updated_at=$1
		WHERE id=$2
	`, deletedAt, compliance.ID); err != nil {
		t.Fatalf("soft delete compliance item: %v", err)
	}

	if err := taskRepo.SoftDelete(ctx, org.ID, task.ID, deletedAt); err != nil {
		t.Fatalf("soft delete task: %v", err)
	}
	if err := programRepo.SoftDelete(ctx, org.ID, program.ID, deletedAt); err != nil {
		t.Fatalf("soft delete program: %v", err)
	}
	if err := itemRepo.SoftDelete(ctx, org.ID, item.ID, deletedAt); err != nil {
		t.Fatalf("soft delete part item: %v", err)
	}
	if err := defRepo.SoftDelete(ctx, org.ID, definition.ID, deletedAt); err != nil {
		t.Fatalf("soft delete part definition: %v", err)
	}
	if err := aircraftRepo.SoftDelete(ctx, org.ID, aircraft.ID, deletedAt); err != nil {
		t.Fatalf("soft delete aircraft: %v", err)
	}
	if err := userRepo.SoftDelete(ctx, org.ID, user.ID, deletedAt); err != nil {
		t.Fatalf("soft delete user: %v", err)
	}

	stats, err := retentionRepo.Cleanup(ctx, org.ID, cutoff)
	if err != nil {
		t.Fatalf("retention cleanup: %v", err)
	}
	if stats.PartReservations != 1 ||
		stats.ComplianceItems != 1 ||
		stats.MaintenanceTasks != 1 ||
		stats.Programs != 1 ||
		stats.PartItems != 1 ||
		stats.PartDefinitions != 1 ||
		stats.Aircraft != 1 ||
		stats.Users != 1 {
		t.Fatalf("unexpected retention stats: %+v", stats)
	}

	assertDeleted := func(query string, args ...any) {
		t.Helper()
		var count int
		if err := pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
			t.Fatalf("count rows: %v", err)
		}
		if count != 0 {
			t.Fatalf("expected 0 rows for %s, got %d", query, count)
		}
	}

	assertDeleted("SELECT COUNT(*) FROM part_reservations WHERE id=$1", reservation.ID)
	assertDeleted("SELECT COUNT(*) FROM compliance_items WHERE id=$1", compliance.ID)
	assertDeleted("SELECT COUNT(*) FROM maintenance_tasks WHERE id=$1", task.ID)
	assertDeleted("SELECT COUNT(*) FROM maintenance_programs WHERE id=$1", program.ID)
	assertDeleted("SELECT COUNT(*) FROM part_items WHERE id=$1", item.ID)
	assertDeleted("SELECT COUNT(*) FROM part_definitions WHERE id=$1", definition.ID)
	assertDeleted("SELECT COUNT(*) FROM aircraft WHERE id=$1", aircraft.ID)
	assertDeleted("SELECT COUNT(*) FROM users WHERE id=$1", user.ID)
}

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	if os.Getenv("AMSS_INTEGRATION") != "1" {
		t.Skip("set AMSS_INTEGRATION=1 to run integration tests")
	}
	baseURL := os.Getenv("AMSS_TEST_DB_URL")
	if baseURL == "" {
		var err error
		var cleanupContainer func()
		baseURL, cleanupContainer, err = startPostgresContainer(context.Background())
		if err != nil {
			t.Fatalf("start postgres container: %v", err)
		}
		t.Cleanup(cleanupContainer)
	}

	baseCfg, err := pgxpool.ParseConfig(baseURL)
	if err != nil {
		t.Fatalf("parse db url: %v", err)
	}
	adminCfg := baseCfg.ConnConfig.Copy()
	adminCfg.Database = "postgres"

	ctx := context.Background()
	adminConn, err := pgx.ConnectConfig(ctx, adminCfg)
	if err != nil {
		t.Fatalf("connect admin: %v", err)
	}

	dbName := "amss_test_" + strings.ReplaceAll(uuid.New().String(), "-", "")
	if _, err := adminConn.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{dbName}.Sanitize()); err != nil {
		_ = adminConn.Close(ctx)
		t.Fatalf("create test db: %v", err)
	}
	_ = adminConn.Close(ctx)

	testCfg := *baseCfg
	testCfg.ConnConfig = baseCfg.ConnConfig.Copy()
	testCfg.ConnConfig.Database = dbName
	if testCfg.ConnConfig.RuntimeParams == nil {
		testCfg.ConnConfig.RuntimeParams = map[string]string{}
	}
	testCfg.ConnConfig.RuntimeParams["search_path"] = "public"
	pool, err := pgxpool.NewWithConfig(ctx, &testCfg)
	if err != nil {
		dropTestDB(t, adminCfg, dbName)
		t.Fatalf("connect test db: %v", err)
	}
	if err := applyMigrations(ctx, pool, findMigrationsDir(t)); err != nil {
		pool.Close()
		dropTestDB(t, adminCfg, dbName)
		t.Fatalf("apply migrations: %v", err)
	}

	cleanup := func() {
		pool.Close()
		dropTestDB(t, adminCfg, dbName)
	}
	return pool, cleanup
}

var (
	postgresContainer     testcontainers.Container
	postgresContainerURL  string
	postgresContainerUses int
	postgresMu            sync.Mutex
)

func startPostgresContainer(ctx context.Context) (string, func(), error) {
	postgresMu.Lock()
	defer postgresMu.Unlock()
	if postgresContainer != nil {
		postgresContainerUses++
		return postgresContainerURL, func() { releasePostgresContainer() }, nil
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "amss",
			"POSTGRES_PASSWORD": "amss",
			"POSTGRES_DB":       "postgres",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return "", nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return "", nil, err
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		return "", nil, err
	}
	postgresContainer = container
	postgresContainerURL = fmt.Sprintf("postgres://amss:amss@%s:%s/postgres?sslmode=disable", host, port.Port())
	postgresContainerUses = 1
	return postgresContainerURL, func() { releasePostgresContainer() }, nil
}

func releasePostgresContainer() {
	postgresMu.Lock()
	defer postgresMu.Unlock()
	if postgresContainer == nil {
		return
	}
	postgresContainerUses--
	if postgresContainerUses > 0 {
		return
	}
	_ = postgresContainer.Terminate(context.Background())
	postgresContainer = nil
	postgresContainerURL = ""
	postgresContainerUses = 0
}

func dropTestDB(t *testing.T, adminCfg *pgx.ConnConfig, dbName string) {
	t.Helper()
	ctx := context.Background()
	adminConn, err := pgx.ConnectConfig(ctx, adminCfg)
	if err != nil {
		t.Fatalf("connect admin for cleanup: %v", err)
	}
	_, _ = adminConn.Exec(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{dbName}.Sanitize()+" WITH (FORCE)")
	_ = adminConn.Close(ctx)
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SET search_path TO public"); err != nil {
		return fmt.Errorf("set search_path: %w", err)
	}
	if _, err := conn.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public"); err != nil {
		return fmt.Errorf("create citext extension: %w", err)
	}
	var hasCitext bool
	if err := conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'citext')").Scan(&hasCitext); err != nil {
		return fmt.Errorf("verify citext type: %w", err)
	}
	if !hasCitext {
		return fmt.Errorf("citext type not available after extension creation")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(dir, name)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sqlText := string(sqlBytes)
		if idx := strings.Index(sqlText, "-- +goose Down"); idx != -1 {
			sqlText = sqlText[:idx]
		}
		if _, err := conn.Exec(ctx, sqlText); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
	}
	return nil
}

func findMigrationsDir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	for i := 0; i < 6; i++ {
		path := filepath.Join(dir, "migrations")
		if stat, err := os.Stat(path); err == nil && stat.IsDir() {
			return path
		}
		dir = filepath.Dir(dir)
	}
	t.Fatalf("migrations directory not found")
	return ""
}
