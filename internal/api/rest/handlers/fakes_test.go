package handlers

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

func applyOffsetLimit[T any](items []T, offset, limit int) []T {
	if offset > 0 {
		if offset >= len(items) {
			return []T{}
		}
		items = items[offset:]
	}
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}
	return items
}

type fakeTaskRepo struct {
	mu    sync.Mutex
	tasks map[uuid.UUID]domain.MaintenanceTask
}

func newFakeTaskRepo() *fakeTaskRepo {
	return &fakeTaskRepo{tasks: make(map[uuid.UUID]domain.MaintenanceTask)}
}

func (f *fakeTaskRepo) GetByID(_ context.Context, orgID, id uuid.UUID) (domain.MaintenanceTask, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	task, ok := f.tasks[id]
	if !ok || task.OrgID != orgID || task.DeletedAt != nil {
		return domain.MaintenanceTask{}, domain.ErrNotFound
	}
	return task, nil
}

func (f *fakeTaskRepo) Create(_ context.Context, task domain.MaintenanceTask) (domain.MaintenanceTask, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tasks[task.ID] = task
	return task, nil
}

func (f *fakeTaskRepo) Update(_ context.Context, task domain.MaintenanceTask) (domain.MaintenanceTask, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.tasks[task.ID]; !ok {
		return domain.MaintenanceTask{}, domain.ErrNotFound
	}
	f.tasks[task.ID] = task
	return task, nil
}

func (f *fakeTaskRepo) SoftDelete(_ context.Context, orgID, id uuid.UUID, at time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	task, ok := f.tasks[id]
	if !ok || task.OrgID != orgID || task.DeletedAt != nil {
		return domain.ErrNotFound
	}
	task.DeletedAt = &at
	f.tasks[id] = task
	return nil
}

func (f *fakeTaskRepo) List(_ context.Context, filter ports.TaskFilter) ([]domain.MaintenanceTask, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.MaintenanceTask
	for _, task := range f.tasks {
		if task.DeletedAt != nil {
			continue
		}
		if filter.OrgID != nil && task.OrgID != *filter.OrgID {
			continue
		}
		if filter.AircraftID != nil && task.AircraftID != *filter.AircraftID {
			continue
		}
		if filter.State != nil && task.State != *filter.State {
			continue
		}
		if filter.Type != nil && task.Type != *filter.Type {
			continue
		}
		if filter.StartFrom != nil && task.StartTime.Before(*filter.StartFrom) {
			continue
		}
		if filter.StartTo != nil && task.StartTime.After(*filter.StartTo) {
			continue
		}
		out = append(out, task)
	}
	return applyOffsetLimit(out, filter.Offset, filter.Limit), nil
}

func (f *fakeTaskRepo) UpdateState(_ context.Context, orgID, id uuid.UUID, newState domain.TaskState, notes string, now time.Time) (domain.MaintenanceTask, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	task, ok := f.tasks[id]
	if !ok || task.OrgID != orgID || task.DeletedAt != nil {
		return domain.MaintenanceTask{}, domain.ErrNotFound
	}
	task.State = newState
	task.Notes = notes
	task.UpdatedAt = now
	f.tasks[id] = task
	return task, nil
}

func (f *fakeTaskRepo) HasActiveForProgram(_ context.Context, orgID, programID uuid.UUID) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, task := range f.tasks {
		if task.DeletedAt != nil || task.OrgID != orgID || task.ProgramID == nil || *task.ProgramID != programID {
			continue
		}
		if task.State == domain.TaskStateScheduled || task.State == domain.TaskStateInProgress {
			return true, nil
		}
	}
	return false, nil
}

type fakePartDefinitionRepo struct {
	mu   sync.Mutex
	defs map[uuid.UUID]domain.PartDefinition
}

func newFakePartDefinitionRepo() *fakePartDefinitionRepo {
	return &fakePartDefinitionRepo{defs: make(map[uuid.UUID]domain.PartDefinition)}
}

func (f *fakePartDefinitionRepo) GetByID(_ context.Context, orgID, id uuid.UUID) (domain.PartDefinition, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	def, ok := f.defs[id]
	if !ok || def.OrgID != orgID || def.DeletedAt != nil {
		return domain.PartDefinition{}, domain.ErrNotFound
	}
	return def, nil
}

func (f *fakePartDefinitionRepo) Create(_ context.Context, def domain.PartDefinition) (domain.PartDefinition, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.defs[def.ID] = def
	return def, nil
}

func (f *fakePartDefinitionRepo) Update(_ context.Context, def domain.PartDefinition) (domain.PartDefinition, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.defs[def.ID]; !ok {
		return domain.PartDefinition{}, domain.ErrNotFound
	}
	f.defs[def.ID] = def
	return def, nil
}

func (f *fakePartDefinitionRepo) List(_ context.Context, filter ports.PartDefinitionFilter) ([]domain.PartDefinition, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.PartDefinition
	for _, def := range f.defs {
		if def.DeletedAt != nil {
			continue
		}
		if filter.OrgID != nil && def.OrgID != *filter.OrgID {
			continue
		}
		if filter.Name != "" && def.Name != filter.Name {
			continue
		}
		out = append(out, def)
	}
	return applyOffsetLimit(out, filter.Offset, filter.Limit), nil
}

func (f *fakePartDefinitionRepo) SoftDelete(_ context.Context, orgID, id uuid.UUID, at time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	def, ok := f.defs[id]
	if !ok || def.OrgID != orgID || def.DeletedAt != nil {
		return domain.ErrNotFound
	}
	def.DeletedAt = &at
	f.defs[id] = def
	return nil
}

type fakePartItemRepo struct {
	mu    sync.Mutex
	items map[uuid.UUID]domain.PartItem
}

func newFakePartItemRepo() *fakePartItemRepo {
	return &fakePartItemRepo{items: make(map[uuid.UUID]domain.PartItem)}
}

func (f *fakePartItemRepo) GetByID(_ context.Context, orgID, id uuid.UUID) (domain.PartItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.items[id]
	if !ok || item.OrgID != orgID || item.DeletedAt != nil {
		return domain.PartItem{}, domain.ErrNotFound
	}
	return item, nil
}

func (f *fakePartItemRepo) GetBySerialNumber(_ context.Context, orgID uuid.UUID, serial string) (domain.PartItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, item := range f.items {
		if item.DeletedAt != nil || item.OrgID != orgID || item.SerialNumber != serial {
			continue
		}
		return item, nil
	}
	return domain.PartItem{}, domain.ErrNotFound
}

func (f *fakePartItemRepo) Create(_ context.Context, item domain.PartItem) (domain.PartItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.items[item.ID] = item
	return item, nil
}

func (f *fakePartItemRepo) Update(_ context.Context, item domain.PartItem) (domain.PartItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.items[item.ID]; !ok {
		return domain.PartItem{}, domain.ErrNotFound
	}
	f.items[item.ID] = item
	return item, nil
}

func (f *fakePartItemRepo) List(_ context.Context, filter ports.PartItemFilter) ([]domain.PartItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.PartItem
	for _, item := range f.items {
		if item.DeletedAt != nil {
			continue
		}
		if filter.OrgID != nil && item.OrgID != *filter.OrgID {
			continue
		}
		if filter.DefinitionID != nil && item.DefinitionID != *filter.DefinitionID {
			continue
		}
		if filter.Status != nil && item.Status != *filter.Status {
			continue
		}
		if filter.ExpiryBefore != nil && item.ExpiryDate != nil && item.ExpiryDate.After(*filter.ExpiryBefore) {
			continue
		}
		out = append(out, item)
	}
	return applyOffsetLimit(out, filter.Offset, filter.Limit), nil
}

func (f *fakePartItemRepo) SoftDelete(_ context.Context, orgID, id uuid.UUID, at time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.items[id]
	if !ok || item.OrgID != orgID || item.DeletedAt != nil {
		return domain.ErrNotFound
	}
	item.DeletedAt = &at
	f.items[id] = item
	return nil
}

func (f *fakePartItemRepo) UpdateStatus(_ context.Context, orgID, id uuid.UUID, status domain.PartItemStatus, now time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.items[id]
	if !ok || item.OrgID != orgID || item.DeletedAt != nil {
		return domain.ErrNotFound
	}
	item.Status = status
	item.UpdatedAt = now
	f.items[id] = item
	return nil
}

type fakeComplianceRepo struct {
	mu    sync.Mutex
	items map[uuid.UUID]domain.ComplianceItem
}

func newFakeComplianceRepo() *fakeComplianceRepo {
	return &fakeComplianceRepo{items: make(map[uuid.UUID]domain.ComplianceItem)}
}

func (f *fakeComplianceRepo) GetByID(_ context.Context, orgID, id uuid.UUID) (domain.ComplianceItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.items[id]
	if !ok || item.OrgID != orgID || item.DeletedAt != nil {
		return domain.ComplianceItem{}, domain.ErrNotFound
	}
	return item, nil
}

func (f *fakeComplianceRepo) ListByTask(_ context.Context, orgID, taskID uuid.UUID) ([]domain.ComplianceItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.ComplianceItem
	for _, item := range f.items {
		if item.DeletedAt != nil {
			continue
		}
		if item.OrgID != orgID || item.TaskID != taskID {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

func (f *fakeComplianceRepo) List(_ context.Context, filter ports.ComplianceFilter) ([]domain.ComplianceItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.ComplianceItem
	for _, item := range f.items {
		if item.DeletedAt != nil {
			continue
		}
		if filter.OrgID != nil && item.OrgID != *filter.OrgID {
			continue
		}
		if filter.TaskID != nil && item.TaskID != *filter.TaskID {
			continue
		}
		if filter.Result != nil && item.Result != *filter.Result {
			continue
		}
		if filter.Signed != nil {
			isSigned := item.SignOffTime != nil
			if isSigned != *filter.Signed {
				continue
			}
		}
		out = append(out, item)
	}
	return applyOffsetLimit(out, filter.Offset, filter.Limit), nil
}

func (f *fakeComplianceRepo) Create(_ context.Context, item domain.ComplianceItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.items[item.ID] = item
	return nil
}

func (f *fakeComplianceRepo) Update(_ context.Context, item domain.ComplianceItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.items[item.ID]; !ok {
		return domain.ErrNotFound
	}
	f.items[item.ID] = item
	return nil
}

func (f *fakeComplianceRepo) SignOff(_ context.Context, orgID, id, userID uuid.UUID, at time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.items[id]
	if !ok || item.OrgID != orgID || item.DeletedAt != nil {
		return domain.ErrNotFound
	}
	item.SignOffUserID = &userID
	item.SignOffTime = &at
	item.UpdatedAt = at
	f.items[id] = item
	return nil
}

type fakeAuditQueryRepo struct {
	mu      sync.Mutex
	entries []domain.AuditLog
}

func (f *fakeAuditQueryRepo) List(_ context.Context, filter ports.AuditLogFilter) ([]domain.AuditLog, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.AuditLog
	for _, entry := range f.entries {
		if filter.OrgID != nil && entry.OrgID != *filter.OrgID {
			continue
		}
		if filter.EntityType != "" && entry.EntityType != filter.EntityType {
			continue
		}
		if filter.EntityID != nil && entry.EntityID != *filter.EntityID {
			continue
		}
		if filter.UserID != nil && entry.UserID != *filter.UserID {
			continue
		}
		if filter.From != nil && entry.Timestamp.Before(*filter.From) {
			continue
		}
		if filter.To != nil && entry.Timestamp.After(*filter.To) {
			continue
		}
		out = append(out, entry)
	}
	return applyOffsetLimit(out, filter.Offset, filter.Limit), nil
}

type fakeOrganizationRepo struct {
	mu   sync.Mutex
	orgs map[uuid.UUID]domain.Organization
}

func newFakeOrganizationRepo() *fakeOrganizationRepo {
	return &fakeOrganizationRepo{orgs: make(map[uuid.UUID]domain.Organization)}
}

func (f *fakeOrganizationRepo) GetByID(_ context.Context, id uuid.UUID) (domain.Organization, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	org, ok := f.orgs[id]
	if !ok || org.DeletedAt != nil {
		return domain.Organization{}, domain.ErrNotFound
	}
	return org, nil
}

func (f *fakeOrganizationRepo) Create(_ context.Context, org domain.Organization) (domain.Organization, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.orgs[org.ID] = org
	return org, nil
}

func (f *fakeOrganizationRepo) Update(_ context.Context, org domain.Organization) (domain.Organization, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.orgs[org.ID]; !ok {
		return domain.Organization{}, domain.ErrNotFound
	}
	f.orgs[org.ID] = org
	return org, nil
}

func (f *fakeOrganizationRepo) SoftDelete(_ context.Context, id uuid.UUID, at time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	org, ok := f.orgs[id]
	if !ok || org.DeletedAt != nil {
		return domain.ErrNotFound
	}
	org.DeletedAt = &at
	f.orgs[id] = org
	return nil
}

func (f *fakeOrganizationRepo) List(_ context.Context, filter ports.OrganizationFilter) ([]domain.Organization, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.Organization
	for _, org := range f.orgs {
		if org.DeletedAt != nil {
			continue
		}
		if filter.Name != "" && !strings.EqualFold(org.Name, filter.Name) {
			continue
		}
		out = append(out, org)
	}
	return applyOffsetLimit(out, filter.Offset, filter.Limit), nil
}

type fakeUserRepo struct {
	mu    sync.Mutex
	users map[uuid.UUID]domain.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[uuid.UUID]domain.User)}
}

func (f *fakeUserRepo) GetByID(_ context.Context, orgID, id uuid.UUID) (domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	user, ok := f.users[id]
	if !ok || user.OrgID != orgID || user.DeletedAt != nil {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (f *fakeUserRepo) Create(_ context.Context, user domain.User) (domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.users[user.ID] = user
	return user, nil
}

func (f *fakeUserRepo) Update(_ context.Context, user domain.User) (domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.users[user.ID]; !ok {
		return domain.User{}, domain.ErrNotFound
	}
	f.users[user.ID] = user
	return user, nil
}

func (f *fakeUserRepo) SoftDelete(_ context.Context, orgID, id uuid.UUID, at time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	user, ok := f.users[id]
	if !ok || user.OrgID != orgID || user.DeletedAt != nil {
		return domain.ErrNotFound
	}
	user.DeletedAt = &at
	f.users[id] = user
	return nil
}

func (f *fakeUserRepo) List(_ context.Context, filter ports.UserFilter) ([]domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.User
	for _, user := range f.users {
		if user.DeletedAt != nil {
			continue
		}
		if filter.OrgID != nil && user.OrgID != *filter.OrgID {
			continue
		}
		if filter.Role != nil && user.Role != *filter.Role {
			continue
		}
		if filter.Email != "" && !strings.EqualFold(user.Email, filter.Email) {
			continue
		}
		out = append(out, user)
	}
	return applyOffsetLimit(out, filter.Offset, filter.Limit), nil
}

type fakeAircraftRepo struct {
	mu       sync.Mutex
	aircraft map[uuid.UUID]domain.Aircraft
}

func newFakeAircraftRepo() *fakeAircraftRepo {
	return &fakeAircraftRepo{aircraft: make(map[uuid.UUID]domain.Aircraft)}
}

func (f *fakeAircraftRepo) GetStatus(_ context.Context, orgID, id uuid.UUID) (domain.AircraftStatus, error) {
	item, err := f.GetByID(context.Background(), orgID, id)
	if err != nil {
		return "", err
	}
	return item.Status, nil
}

func (f *fakeAircraftRepo) GetByID(_ context.Context, orgID, id uuid.UUID) (domain.Aircraft, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.aircraft[id]
	if !ok || item.OrgID != orgID || item.DeletedAt != nil {
		return domain.Aircraft{}, domain.ErrNotFound
	}
	return item, nil
}

func (f *fakeAircraftRepo) GetByTailNumber(_ context.Context, orgID uuid.UUID, tailNumber string) (domain.Aircraft, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, item := range f.aircraft {
		if item.DeletedAt != nil || item.OrgID != orgID {
			continue
		}
		if strings.EqualFold(item.TailNumber, tailNumber) {
			return item, nil
		}
	}
	return domain.Aircraft{}, domain.ErrNotFound
}

func (f *fakeAircraftRepo) Create(_ context.Context, aircraft domain.Aircraft) (domain.Aircraft, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.aircraft[aircraft.ID] = aircraft
	return aircraft, nil
}

func (f *fakeAircraftRepo) Update(_ context.Context, aircraft domain.Aircraft) (domain.Aircraft, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.aircraft[aircraft.ID]; !ok {
		return domain.Aircraft{}, domain.ErrNotFound
	}
	f.aircraft[aircraft.ID] = aircraft
	return aircraft, nil
}

func (f *fakeAircraftRepo) SoftDelete(_ context.Context, orgID, id uuid.UUID, at time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.aircraft[id]
	if !ok || item.OrgID != orgID || item.DeletedAt != nil {
		return domain.ErrNotFound
	}
	item.DeletedAt = &at
	f.aircraft[id] = item
	return nil
}

func (f *fakeAircraftRepo) List(_ context.Context, filter ports.AircraftFilter) ([]domain.Aircraft, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.Aircraft
	for _, item := range f.aircraft {
		if item.DeletedAt != nil {
			continue
		}
		if filter.OrgID != nil && item.OrgID != *filter.OrgID {
			continue
		}
		if filter.Status != nil && item.Status != *filter.Status {
			continue
		}
		if filter.Model != "" && !strings.EqualFold(item.Model, filter.Model) {
			continue
		}
		if filter.TailNumber != "" && !strings.EqualFold(item.TailNumber, filter.TailNumber) {
			continue
		}
		out = append(out, item)
	}
	return applyOffsetLimit(out, filter.Offset, filter.Limit), nil
}

type fakeProgramRepo struct {
	mu       sync.Mutex
	programs map[uuid.UUID]domain.MaintenanceProgram
	due      []domain.MaintenanceProgram
}

func newFakeProgramRepo() *fakeProgramRepo {
	return &fakeProgramRepo{programs: make(map[uuid.UUID]domain.MaintenanceProgram)}
}

func (f *fakeProgramRepo) GetByID(_ context.Context, orgID, id uuid.UUID) (domain.MaintenanceProgram, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	program, ok := f.programs[id]
	if !ok || program.OrgID != orgID || program.DeletedAt != nil {
		return domain.MaintenanceProgram{}, domain.ErrNotFound
	}
	return program, nil
}

func (f *fakeProgramRepo) Create(_ context.Context, program domain.MaintenanceProgram) (domain.MaintenanceProgram, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.programs[program.ID] = program
	return program, nil
}

func (f *fakeProgramRepo) Update(_ context.Context, program domain.MaintenanceProgram) (domain.MaintenanceProgram, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.programs[program.ID]; !ok {
		return domain.MaintenanceProgram{}, domain.ErrNotFound
	}
	f.programs[program.ID] = program
	return program, nil
}

func (f *fakeProgramRepo) SoftDelete(_ context.Context, orgID, id uuid.UUID, at time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	program, ok := f.programs[id]
	if !ok || program.OrgID != orgID || program.DeletedAt != nil {
		return domain.ErrNotFound
	}
	program.DeletedAt = &at
	f.programs[id] = program
	return nil
}

func (f *fakeProgramRepo) List(_ context.Context, filter ports.MaintenanceProgramFilter) ([]domain.MaintenanceProgram, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.MaintenanceProgram
	for _, program := range f.programs {
		if program.DeletedAt != nil {
			continue
		}
		if filter.OrgID != nil && program.OrgID != *filter.OrgID {
			continue
		}
		if filter.AircraftID != nil {
			if program.AircraftID == nil || *program.AircraftID != *filter.AircraftID {
				continue
			}
		}
		out = append(out, program)
	}
	return applyOffsetLimit(out, filter.Offset, filter.Limit), nil
}

func (f *fakeProgramRepo) ListDueCalendar(_ context.Context, _ time.Time, limit int) ([]domain.MaintenanceProgram, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.due) == 0 {
		var out []domain.MaintenanceProgram
		for _, program := range f.programs {
			if program.DeletedAt != nil {
				continue
			}
			if program.IntervalType != domain.ProgramIntervalCalendar {
				continue
			}
			out = append(out, program)
		}
		return applyOffsetLimit(out, 0, limit), nil
	}
	return applyOffsetLimit(f.due, 0, limit), nil
}

func (f *fakeProgramRepo) GetByName(_ context.Context, orgID uuid.UUID, name string, aircraftID *uuid.UUID) (domain.MaintenanceProgram, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, program := range f.programs {
		if program.DeletedAt != nil || program.OrgID != orgID {
			continue
		}
		if !strings.EqualFold(program.Name, name) {
			continue
		}
		if aircraftID == nil && program.AircraftID != nil {
			continue
		}
		if aircraftID != nil {
			if program.AircraftID == nil || *program.AircraftID != *aircraftID {
				continue
			}
		}
		return program, nil
	}
	return domain.MaintenanceProgram{}, domain.ErrNotFound
}

type fakeImportRepo struct {
	mu      sync.Mutex
	imports map[uuid.UUID]domain.Import
}

func newFakeImportRepo() *fakeImportRepo {
	return &fakeImportRepo{imports: make(map[uuid.UUID]domain.Import)}
}

func (f *fakeImportRepo) GetByID(_ context.Context, orgID, id uuid.UUID) (domain.Import, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	imp, ok := f.imports[id]
	if !ok {
		return domain.Import{}, domain.ErrNotFound
	}
	if orgID != uuid.Nil && imp.OrgID != orgID {
		return domain.Import{}, domain.ErrNotFound
	}
	return imp, nil
}

func (f *fakeImportRepo) Create(_ context.Context, imp domain.Import) (domain.Import, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.imports[imp.ID] = imp
	return imp, nil
}

func (f *fakeImportRepo) Update(_ context.Context, imp domain.Import) (domain.Import, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.imports[imp.ID]; !ok {
		return domain.Import{}, domain.ErrNotFound
	}
	f.imports[imp.ID] = imp
	return imp, nil
}

func (f *fakeImportRepo) UpdateStatus(_ context.Context, orgID, id uuid.UUID, status domain.ImportStatus, summary map[string]any, updatedAt time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	imp, ok := f.imports[id]
	if !ok || (orgID != uuid.Nil && imp.OrgID != orgID) {
		return domain.ErrNotFound
	}
	imp.Status = status
	imp.Summary = summary
	imp.UpdatedAt = updatedAt
	f.imports[id] = imp
	return nil
}

type fakeImportRowRepo struct {
	mu   sync.Mutex
	rows map[uuid.UUID]domain.ImportRow
}

func newFakeImportRowRepo() *fakeImportRowRepo {
	return &fakeImportRowRepo{rows: make(map[uuid.UUID]domain.ImportRow)}
}

func (f *fakeImportRowRepo) Create(_ context.Context, row domain.ImportRow) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rows[row.ID] = row
	return nil
}

func (f *fakeImportRowRepo) Update(_ context.Context, row domain.ImportRow) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rows[row.ID] = row
	return nil
}

func (f *fakeImportRowRepo) ListByImport(_ context.Context, filter ports.ImportRowFilter) ([]domain.ImportRow, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.ImportRow
	for _, row := range f.rows {
		if row.OrgID != filter.OrgID || row.ImportID != filter.ImportID {
			continue
		}
		if filter.Status != nil && row.Status != *filter.Status {
			continue
		}
		out = append(out, row)
	}
	return applyOffsetLimit(out, filter.Offset, filter.Limit), nil
}

type fakeImportJobQueue struct {
	mu     sync.Mutex
	queued []uuid.UUID
}

func (f *fakeImportJobQueue) Enqueue(_ context.Context, importID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.queued = append(f.queued, importID)
	return nil
}

type fakeWebhookRepo struct {
	mu    sync.Mutex
	hooks map[uuid.UUID]domain.Webhook
}

func newFakeWebhookRepo() *fakeWebhookRepo {
	return &fakeWebhookRepo{hooks: make(map[uuid.UUID]domain.Webhook)}
}

func (f *fakeWebhookRepo) GetByID(_ context.Context, orgID, id uuid.UUID) (domain.Webhook, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	hook, ok := f.hooks[id]
	if !ok || hook.OrgID != orgID {
		return domain.Webhook{}, domain.ErrNotFound
	}
	return hook, nil
}

func (f *fakeWebhookRepo) Create(_ context.Context, webhook domain.Webhook) (domain.Webhook, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.hooks[webhook.ID] = webhook
	return webhook, nil
}

func (f *fakeWebhookRepo) Delete(_ context.Context, orgID, id uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	hook, ok := f.hooks[id]
	if !ok || hook.OrgID != orgID {
		return domain.ErrNotFound
	}
	delete(f.hooks, id)
	return nil
}

func (f *fakeWebhookRepo) List(_ context.Context, orgID uuid.UUID) ([]domain.Webhook, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.Webhook
	for _, hook := range f.hooks {
		if hook.OrgID != orgID {
			continue
		}
		out = append(out, hook)
	}
	return out, nil
}

func (f *fakeWebhookRepo) ListByEvent(_ context.Context, orgID uuid.UUID, eventType string) ([]domain.Webhook, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.Webhook
	for _, hook := range f.hooks {
		if hook.OrgID != orgID {
			continue
		}
		for _, event := range hook.Events {
			if event == eventType {
				out = append(out, hook)
				break
			}
		}
	}
	return out, nil
}

type fakeOutboxRepo struct {
	mu     sync.Mutex
	events []ports.OutboxEvent
}

func (f *fakeOutboxRepo) Enqueue(_ context.Context, orgID uuid.UUID, eventType string, aggregateType string, aggregateID uuid.UUID, payload map[string]any, dedupeKey string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, ports.OutboxEvent{
		ID:            uuid.New(),
		OrgID:         orgID,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Payload:       payload,
		DedupeKey:     dedupeKey,
		CreatedAt:     time.Now().UTC(),
	})
	return nil
}

func (f *fakeOutboxRepo) LockPending(_ context.Context, _ string, _ int) ([]ports.OutboxEvent, error) {
	return nil, nil
}

func (f *fakeOutboxRepo) MarkProcessed(_ context.Context, _ uuid.UUID, _ time.Time) error {
	return nil
}

func (f *fakeOutboxRepo) ScheduleRetry(_ context.Context, _ uuid.UUID, _ int, _ time.Time, _ string) error {
	return nil
}

func (f *fakeOutboxRepo) GetByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (ports.OutboxEvent, error) {
	return ports.OutboxEvent{}, domain.ErrNotFound
}

type fakeReportRepo struct {
	summary              ports.ReportSummary
	compliance           ports.ComplianceReport
	lastSummaryOrg       uuid.UUID
	lastComplianceFilter ports.ComplianceReportFilter
}

func (f *fakeReportRepo) Summary(_ context.Context, orgID uuid.UUID) (ports.ReportSummary, error) {
	f.lastSummaryOrg = orgID
	return f.summary, nil
}

func (f *fakeReportRepo) Compliance(_ context.Context, filter ports.ComplianceReportFilter) (ports.ComplianceReport, error) {
	f.lastComplianceFilter = filter
	return f.compliance, nil
}
