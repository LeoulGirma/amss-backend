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

type SchedulingService struct {
	Tasks          ports.TaskRepository
	Dependencies   ports.TaskDependencyRepository
	ScheduleEvents ports.ScheduleChangeRepository
	Outbox         ports.OutboxRepository
	Clock          app.Clock
}

// --- Task Dependencies ---

type DependencyCreateInput struct {
	TaskID          uuid.UUID
	DependsOnTaskID uuid.UUID
	DependencyType  domain.DependencyType
}

func (s *SchedulingService) CreateDependency(ctx context.Context, actor app.Actor, input DependencyCreateInput) (domain.TaskDependency, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return domain.TaskDependency{}, domain.ErrForbidden
	}

	if input.TaskID == input.DependsOnTaskID {
		return domain.TaskDependency{}, domain.NewValidationError("task cannot depend on itself")
	}

	// Verify both tasks exist
	if _, err := s.Tasks.GetByID(ctx, actor.OrgID, input.TaskID); err != nil {
		return domain.TaskDependency{}, fmt.Errorf("task: %w", err)
	}
	if _, err := s.Tasks.GetByID(ctx, actor.OrgID, input.DependsOnTaskID); err != nil {
		return domain.TaskDependency{}, fmt.Errorf("depends_on_task: %w", err)
	}

	// Check for circular dependency
	if err := s.detectCycle(ctx, actor.OrgID, input.DependsOnTaskID, input.TaskID); err != nil {
		return domain.TaskDependency{}, err
	}

	depType := input.DependencyType
	if depType == "" {
		depType = domain.DependencyFinishToStart
	}

	dep := domain.TaskDependency{
		ID:              uuid.New(),
		OrgID:           actor.OrgID,
		TaskID:          input.TaskID,
		DependsOnTaskID: input.DependsOnTaskID,
		DependencyType:  depType,
		CreatedAt:       s.Clock.Now(),
	}

	return s.Dependencies.Create(ctx, dep)
}

func (s *SchedulingService) DeleteDependency(ctx context.Context, actor app.Actor, depID uuid.UUID) error {
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return domain.ErrForbidden
	}
	return s.Dependencies.Delete(ctx, actor.OrgID, depID)
}

func (s *SchedulingService) ListDependencies(ctx context.Context, actor app.Actor, taskID uuid.UUID) ([]domain.TaskDependency, error) {
	return s.Dependencies.ListByTask(ctx, actor.OrgID, taskID)
}

func (s *SchedulingService) ListDependents(ctx context.Context, actor app.Actor, taskID uuid.UUID) ([]domain.TaskDependency, error) {
	return s.Dependencies.ListDependents(ctx, actor.OrgID, taskID)
}

// detectCycle checks if adding an edge from sourceID -> targetID would create a cycle
func (s *SchedulingService) detectCycle(ctx context.Context, orgID, sourceID, targetID uuid.UUID) error {
	visited := make(map[uuid.UUID]bool)
	return s.dfs(ctx, orgID, sourceID, targetID, visited)
}

func (s *SchedulingService) dfs(ctx context.Context, orgID, current, target uuid.UUID, visited map[uuid.UUID]bool) error {
	if current == target {
		return domain.NewConflictError("circular dependency detected")
	}
	if visited[current] {
		return nil
	}
	visited[current] = true

	deps, err := s.Dependencies.ListByTask(ctx, orgID, current)
	if err != nil {
		return err
	}
	for _, dep := range deps {
		if err := s.dfs(ctx, orgID, dep.DependsOnTaskID, target, visited); err != nil {
			return err
		}
	}
	return nil
}

// --- Rescheduling ---

type RescheduleInput struct {
	TaskID       uuid.UUID
	NewStartTime time.Time
	NewEndTime   time.Time
	Reason       string
	Cascade      bool
}

func (s *SchedulingService) RescheduleTask(ctx context.Context, actor app.Actor, input RescheduleInput) (domain.ScheduleChangeEvent, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return domain.ScheduleChangeEvent{}, domain.ErrForbidden
	}

	task, err := s.Tasks.GetByID(ctx, actor.OrgID, input.TaskID)
	if err != nil {
		return domain.ScheduleChangeEvent{}, err
	}

	if task.State == domain.TaskStateCompleted || task.State == domain.TaskStateCancelled {
		return domain.ScheduleChangeEvent{}, domain.NewConflictError("cannot reschedule completed or cancelled task")
	}

	now := s.Clock.Now()

	// Record schedule change event
	event := domain.ScheduleChangeEvent{
		ID:           uuid.New(),
		OrgID:        actor.OrgID,
		TaskID:       input.TaskID,
		ChangeType:   domain.ScheduleChangeRescheduled,
		Reason:       input.Reason,
		OldStartTime: &task.StartTime,
		NewStartTime: &input.NewStartTime,
		OldEndTime:   &task.EndTime,
		NewEndTime:   &input.NewEndTime,
		TriggeredBy:  &actor.UserID,
		CreatedAt:    now,
	}

	// Update the task times
	task.StartTime = input.NewStartTime.UTC()
	task.EndTime = input.NewEndTime.UTC()
	task.UpdatedAt = now

	if _, err := s.Tasks.Update(ctx, task); err != nil {
		return domain.ScheduleChangeEvent{}, err
	}

	// Cascade to dependent tasks if requested
	if input.Cascade {
		affectedIDs, err := s.cascadeReschedule(ctx, actor.OrgID, input.TaskID, input.NewEndTime, *event.OldEndTime)
		if err != nil {
			return domain.ScheduleChangeEvent{}, err
		}
		event.AffectedTaskIDs = affectedIDs
	}

	created, err := s.ScheduleEvents.Create(ctx, event)
	if err != nil {
		return domain.ScheduleChangeEvent{}, err
	}

	// Emit outbox event for WebSocket notification
	if s.Outbox != nil {
		dedupeKey := fmt.Sprintf("task_rescheduled:%s:%s", actor.OrgID, input.TaskID)
		_ = s.Outbox.Enqueue(ctx, actor.OrgID, "task_rescheduled", "maintenance_task", input.TaskID, map[string]any{
			"version":        1,
			"org_id":         actor.OrgID,
			"task_id":        input.TaskID,
			"old_start_time": event.OldStartTime,
			"new_start_time": event.NewStartTime,
			"old_end_time":   event.OldEndTime,
			"new_end_time":   event.NewEndTime,
			"reason":         input.Reason,
			"affected_tasks": event.AffectedTaskIDs,
			"timestamp":      now,
		}, dedupeKey)
	}

	return created, nil
}

// cascadeReschedule shifts all downstream dependent tasks by the same time delta
func (s *SchedulingService) cascadeReschedule(ctx context.Context, orgID, taskID uuid.UUID, newEndTime, oldEndTime time.Time) ([]uuid.UUID, error) {
	delta := newEndTime.Sub(oldEndTime)
	if delta == 0 {
		return nil, nil
	}

	dependents, err := s.Dependencies.ListDependents(ctx, orgID, taskID)
	if err != nil {
		return nil, err
	}

	var affected []uuid.UUID
	for _, dep := range dependents {
		depTask, err := s.Tasks.GetByID(ctx, orgID, dep.TaskID)
		if err != nil {
			continue
		}
		if depTask.State == domain.TaskStateCompleted || depTask.State == domain.TaskStateCancelled {
			continue
		}

		depTask.StartTime = depTask.StartTime.Add(delta)
		depTask.EndTime = depTask.EndTime.Add(delta)
		if _, err := s.Tasks.Update(ctx, depTask); err != nil {
			continue
		}
		affected = append(affected, dep.TaskID)

		// Recursively cascade
		childAffected, err := s.cascadeReschedule(ctx, orgID, dep.TaskID, depTask.EndTime, depTask.EndTime.Add(-delta))
		if err == nil {
			affected = append(affected, childAffected...)
		}
	}

	return affected, nil
}

// --- Schedule Change History ---

func (s *SchedulingService) ListScheduleChanges(ctx context.Context, actor app.Actor, taskID uuid.UUID) ([]domain.ScheduleChangeEvent, error) {
	return s.ScheduleEvents.ListByTask(ctx, actor.OrgID, taskID)
}

// --- Conflict Detection ---

type ScheduleConflict struct {
	TaskID          uuid.UUID `json:"task_id"`
	ConflictType    string    `json:"conflict_type"`
	Description     string    `json:"description"`
	BlockingTaskIDs []uuid.UUID `json:"blocking_task_ids,omitempty"`
}

func (s *SchedulingService) DetectConflicts(ctx context.Context, actor app.Actor) ([]ScheduleConflict, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}

	// Get all active tasks
	scheduledState := domain.TaskStateScheduled
	inProgressState := domain.TaskStateInProgress
	scheduledTasks, err := s.Tasks.List(ctx, ports.TaskFilter{OrgID: &actor.OrgID, State: &scheduledState, Limit: 200})
	if err != nil {
		return nil, err
	}
	inProgressTasks, err := s.Tasks.List(ctx, ports.TaskFilter{OrgID: &actor.OrgID, State: &inProgressState, Limit: 200})
	if err != nil {
		return nil, err
	}

	allTasks := append(scheduledTasks, inProgressTasks...)
	var conflicts []ScheduleConflict

	for _, task := range allTasks {
		// Check for unmet dependencies
		deps, err := s.Dependencies.ListByTask(ctx, actor.OrgID, task.ID)
		if err != nil {
			continue
		}
		for _, dep := range deps {
			depTask, err := s.Tasks.GetByID(ctx, actor.OrgID, dep.DependsOnTaskID)
			if err != nil {
				continue
			}
			if dep.DependencyType == domain.DependencyFinishToStart && depTask.State != domain.TaskStateCompleted {
				if task.StartTime.Before(depTask.EndTime) {
					conflicts = append(conflicts, ScheduleConflict{
						TaskID:          task.ID,
						ConflictType:    "unmet_dependency",
						Description:     fmt.Sprintf("Task starts before dependency %s completes", dep.DependsOnTaskID),
						BlockingTaskIDs: []uuid.UUID{dep.DependsOnTaskID},
					})
				}
			}
		}
	}

	return conflicts, nil
}
