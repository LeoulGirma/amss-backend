package handlers

import (
	"net/http"
	"time"

	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// --- Request/Response Types ---

type dependencyCreateRequest struct {
	DependsOnTaskID string `json:"depends_on_task_id" validate:"required,uuid"`
	DependencyType  string `json:"dependency_type" validate:"omitempty,oneof=finish_to_start start_to_start finish_to_finish"`
}

type rescheduleRequest struct {
	NewStartTime string `json:"new_start_time" validate:"required,rfc3339"`
	NewEndTime   string `json:"new_end_time" validate:"required,rfc3339"`
	Reason       string `json:"reason" validate:"required"`
	Cascade      bool   `json:"cascade"`
}

type dependencyResponse struct {
	ID              uuid.UUID            `json:"id"`
	OrgID           uuid.UUID            `json:"org_id"`
	TaskID          uuid.UUID            `json:"task_id"`
	DependsOnTaskID uuid.UUID            `json:"depends_on_task_id"`
	DependencyType  domain.DependencyType `json:"dependency_type"`
	CreatedAt       time.Time            `json:"created_at"`
}

type scheduleChangeResponse struct {
	ID              uuid.UUID                 `json:"id"`
	OrgID           uuid.UUID                 `json:"org_id"`
	TaskID          uuid.UUID                 `json:"task_id"`
	ChangeType      domain.ScheduleChangeType `json:"change_type"`
	Reason          string                    `json:"reason"`
	OldStartTime    *time.Time                `json:"old_start_time,omitempty"`
	NewStartTime    *time.Time                `json:"new_start_time,omitempty"`
	OldEndTime      *time.Time                `json:"old_end_time,omitempty"`
	NewEndTime      *time.Time                `json:"new_end_time,omitempty"`
	TriggeredBy     *uuid.UUID                `json:"triggered_by,omitempty"`
	AffectedTaskIDs []uuid.UUID               `json:"affected_task_ids,omitempty"`
	CreatedAt       time.Time                 `json:"created_at"`
}

// --- Handlers ---

func CreateTaskDependency(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Scheduling == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid task id")
		return
	}
	var req dependencyCreateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	dependsOnID, _ := uuid.Parse(req.DependsOnTaskID)

	created, err := servicesReg.Scheduling.CreateDependency(r.Context(), actor, services.DependencyCreateInput{
		TaskID:          taskID,
		DependsOnTaskID: dependsOnID,
		DependencyType:  domain.DependencyType(req.DependencyType),
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, dependencyResponse{
		ID:              created.ID,
		OrgID:           created.OrgID,
		TaskID:          created.TaskID,
		DependsOnTaskID: created.DependsOnTaskID,
		DependencyType:  created.DependencyType,
		CreatedAt:       created.CreatedAt,
	})
}

func ListTaskDependencies(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Scheduling == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid task id")
		return
	}
	deps, err := servicesReg.Scheduling.ListDependencies(r.Context(), actor, taskID)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]dependencyResponse, 0, len(deps))
	for _, d := range deps {
		resp = append(resp, dependencyResponse{
			ID:              d.ID,
			OrgID:           d.OrgID,
			TaskID:          d.TaskID,
			DependsOnTaskID: d.DependsOnTaskID,
			DependencyType:  d.DependencyType,
			CreatedAt:       d.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func DeleteTaskDependency(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Scheduling == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	depID, err := uuid.Parse(chi.URLParam(r, "depId"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid dependency id")
		return
	}
	if err := servicesReg.Scheduling.DeleteDependency(r.Context(), actor, depID); err != nil {
		writeDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func RescheduleTask(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Scheduling == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid task id")
		return
	}
	var req rescheduleRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	newStart, _ := time.Parse(time.RFC3339, req.NewStartTime)
	newEnd, _ := time.Parse(time.RFC3339, req.NewEndTime)

	event, err := servicesReg.Scheduling.RescheduleTask(r.Context(), actor, services.RescheduleInput{
		TaskID:       taskID,
		NewStartTime: newStart,
		NewEndTime:   newEnd,
		Reason:       req.Reason,
		Cascade:      req.Cascade,
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, scheduleChangeResponse{
		ID:              event.ID,
		OrgID:           event.OrgID,
		TaskID:          event.TaskID,
		ChangeType:      event.ChangeType,
		Reason:          event.Reason,
		OldStartTime:    event.OldStartTime,
		NewStartTime:    event.NewStartTime,
		OldEndTime:      event.OldEndTime,
		NewEndTime:      event.NewEndTime,
		TriggeredBy:     event.TriggeredBy,
		AffectedTaskIDs: event.AffectedTaskIDs,
		CreatedAt:       event.CreatedAt,
	})
}

func ListScheduleChanges(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Scheduling == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid task id")
		return
	}
	events, err := servicesReg.Scheduling.ListScheduleChanges(r.Context(), actor, taskID)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]scheduleChangeResponse, 0, len(events))
	for _, e := range events {
		resp = append(resp, scheduleChangeResponse{
			ID:              e.ID,
			OrgID:           e.OrgID,
			TaskID:          e.TaskID,
			ChangeType:      e.ChangeType,
			Reason:          e.Reason,
			OldStartTime:    e.OldStartTime,
			NewStartTime:    e.NewStartTime,
			OldEndTime:      e.OldEndTime,
			NewEndTime:      e.NewEndTime,
			TriggeredBy:     e.TriggeredBy,
			AffectedTaskIDs: e.AffectedTaskIDs,
			CreatedAt:       e.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func DetectScheduleConflicts(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Scheduling == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	conflicts, err := servicesReg.Scheduling.DetectConflicts(r.Context(), actor)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, conflicts)
}
