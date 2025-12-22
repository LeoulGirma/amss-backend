package grpcapi

import (
	"context"
	"time"

	amssv1 "github.com/aeromaintain/amss/api/proto/amssv1"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type TaskServiceServer struct {
	amssv1.UnimplementedTaskServiceServer
	Tasks *services.TaskService
}

func (s *TaskServiceServer) CreateTask(ctx context.Context, req *amssv1.CreateTaskRequest) (*amssv1.TaskResponse, error) {
	if req == nil {
		return nil, invalidArgument("missing request")
	}
	if s.Tasks == nil {
		return nil, mapError(domain.NewValidationError("task service unavailable"))
	}
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := resolveOrgID(actor, req.OrgId)
	if err != nil {
		return nil, invalidArgument("invalid org_id")
	}
	aircraftID, err := uuid.Parse(req.AircraftId)
	if err != nil {
		return nil, invalidArgument("invalid aircraft_id")
	}
	taskType := domain.TaskType(req.Type)
	if !validTaskType(taskType) {
		return nil, invalidArgument("invalid type")
	}
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, invalidArgument("invalid start_time")
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, invalidArgument("invalid end_time")
	}

	created, err := s.Tasks.Create(ctx, actor, services.TaskCreateInput{
		OrgID:      &orgID,
		AircraftID: aircraftID,
		Type:       taskType,
		StartTime:  startTime,
		EndTime:    endTime,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &amssv1.TaskResponse{
		TaskId: created.ID.String(),
		State:  string(created.State),
	}, nil
}

func (s *TaskServiceServer) TransitionState(ctx context.Context, req *amssv1.TransitionStateRequest) (*amssv1.TaskResponse, error) {
	if req == nil {
		return nil, invalidArgument("missing request")
	}
	if s.Tasks == nil {
		return nil, mapError(domain.NewValidationError("task service unavailable"))
	}
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	taskID, err := uuid.Parse(req.TaskId)
	if err != nil {
		return nil, invalidArgument("invalid task_id")
	}
	state := domain.TaskState(req.NewState)
	if !validTaskState(state) {
		return nil, invalidArgument("invalid new_state")
	}
	updated, err := s.Tasks.TransitionState(ctx, actor, taskID, state, services.TaskTransitionOptions{})
	if err != nil {
		return nil, mapError(err)
	}
	return &amssv1.TaskResponse{
		TaskId: updated.ID.String(),
		State:  string(updated.State),
	}, nil
}

func validTaskType(value domain.TaskType) bool {
	switch value {
	case domain.TaskTypeInspection, domain.TaskTypeRepair, domain.TaskTypeOverhaul:
		return true
	default:
		return false
	}
}

func validTaskState(state domain.TaskState) bool {
	switch state {
	case domain.TaskStateScheduled, domain.TaskStateInProgress, domain.TaskStateCompleted, domain.TaskStateCancelled:
		return true
	default:
		return false
	}
}
