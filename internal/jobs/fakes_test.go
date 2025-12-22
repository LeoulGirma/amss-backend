package jobs

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
		if filter.Name != "" && !strings.EqualFold(def.Name, filter.Name) {
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
		if item.DeletedAt != nil || item.OrgID != orgID {
			continue
		}
		if strings.EqualFold(item.SerialNumber, serial) {
			return item, nil
		}
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
			if program.DeletedAt != nil || program.IntervalType != domain.ProgramIntervalCalendar {
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

type fakeWebhookDeliveryRepo struct {
	mu      sync.Mutex
	updated []domain.WebhookDelivery
	created []domain.WebhookDelivery
}

func (f *fakeWebhookDeliveryRepo) Create(_ context.Context, delivery domain.WebhookDelivery) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.created = append(f.created, delivery)
	return nil
}

func (f *fakeWebhookDeliveryRepo) ClaimPending(_ context.Context, _ int, _ time.Time) ([]domain.WebhookDelivery, error) {
	return nil, nil
}

func (f *fakeWebhookDeliveryRepo) Update(_ context.Context, delivery domain.WebhookDelivery) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.updated = append(f.updated, delivery)
	return nil
}

type outboxScheduleCall struct {
	id          uuid.UUID
	attempt     int
	nextAttempt time.Time
	lastError   string
}

type fakeOutboxRepo struct {
	mu            sync.Mutex
	events        map[uuid.UUID]ports.OutboxEvent
	scheduleCalls []outboxScheduleCall
	markCalls     []uuid.UUID
	lockEvents    []ports.OutboxEvent
	lockErr       error
}

func newFakeOutboxRepo() *fakeOutboxRepo {
	return &fakeOutboxRepo{events: make(map[uuid.UUID]ports.OutboxEvent)}
}

func (f *fakeOutboxRepo) Enqueue(_ context.Context, _ uuid.UUID, _ string, _ string, _ uuid.UUID, _ map[string]any, _ string) error {
	return nil
}

func (f *fakeOutboxRepo) LockPending(_ context.Context, _ string, _ int) ([]ports.OutboxEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.lockErr != nil {
		return nil, f.lockErr
	}
	events := make([]ports.OutboxEvent, len(f.lockEvents))
	copy(events, f.lockEvents)
	f.lockEvents = nil
	return events, nil
}

func (f *fakeOutboxRepo) MarkProcessed(_ context.Context, id uuid.UUID, _ time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.markCalls = append(f.markCalls, id)
	return nil
}

func (f *fakeOutboxRepo) ScheduleRetry(_ context.Context, id uuid.UUID, attempt int, nextAttemptAt time.Time, lastError string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.scheduleCalls = append(f.scheduleCalls, outboxScheduleCall{
		id:          id,
		attempt:     attempt,
		nextAttempt: nextAttemptAt,
		lastError:   lastError,
	})
	return nil
}

func (f *fakeOutboxRepo) GetByID(_ context.Context, _ uuid.UUID, id uuid.UUID) (ports.OutboxEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	event, ok := f.events[id]
	if !ok {
		return ports.OutboxEvent{}, domain.ErrNotFound
	}
	return event, nil
}
