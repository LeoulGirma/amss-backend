package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type taskCreateRequest struct {
	OrgID              string `json:"org_id" validate:"omitempty,uuid"`
	AircraftID         string `json:"aircraft_id" validate:"required,uuid"`
	ProgramID          string `json:"program_id" validate:"omitempty,uuid"`
	Type               string `json:"type" validate:"required,oneof=inspection repair overhaul"`
	StartTime          string `json:"start_time" validate:"required,rfc3339"`
	EndTime            string `json:"end_time" validate:"required,rfc3339"`
	AssignedMechanicID string `json:"assigned_mechanic_id" validate:"omitempty,uuid"`
	Notes              string `json:"notes"`
}

type taskUpdateRequest struct {
	ProgramID          string  `json:"program_id" validate:"omitempty,uuid"`
	Type               string  `json:"type" validate:"omitempty,oneof=inspection repair overhaul"`
	StartTime          string  `json:"start_time" validate:"omitempty,rfc3339"`
	EndTime            string  `json:"end_time" validate:"omitempty,rfc3339"`
	AssignedMechanicID string  `json:"assigned_mechanic_id" validate:"omitempty,uuid"`
	Notes              *string `json:"notes"`
	OrgID              string  `json:"org_id" validate:"omitempty,uuid"`
}

type taskResponse struct {
	ID                 uuid.UUID        `json:"id"`
	OrgID              uuid.UUID        `json:"org_id"`
	AircraftID         uuid.UUID        `json:"aircraft_id"`
	ProgramID          *uuid.UUID       `json:"program_id,omitempty"`
	Type               domain.TaskType  `json:"type"`
	State              domain.TaskState `json:"state"`
	StartTime          time.Time        `json:"start_time"`
	EndTime            time.Time        `json:"end_time"`
	AssignedMechanicID *uuid.UUID       `json:"assigned_mechanic_id,omitempty"`
	Notes              string           `json:"notes"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

type taskDetailResponse struct {
	Task         taskResponse          `json:"task"`
	Reservations []reservationResponse `json:"reservations"`
	Compliance   []complianceResponse  `json:"compliance"`
}

func CreateTask(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Tasks == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	var req taskCreateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	orgID, err := resolveOrgID(actor, req.OrgID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
		return
	}
	aircraftID, err := uuid.Parse(req.AircraftID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid aircraft_id")
		return
	}
	if req.Type == "" {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "type is required")
		return
	}
	taskType := domain.TaskType(req.Type)
	if !validTaskType(taskType) {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid type")
		return
	}
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid start_time")
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid end_time")
		return
	}
	var programID *uuid.UUID
	if req.ProgramID != "" {
		parsed, err := uuid.Parse(req.ProgramID)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid program_id")
			return
		}
		programID = &parsed
	}
	var mechanicID *uuid.UUID
	if req.AssignedMechanicID != "" {
		parsed, err := uuid.Parse(req.AssignedMechanicID)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid assigned_mechanic_id")
			return
		}
		mechanicID = &parsed
	}

	created, err := servicesReg.Tasks.Create(r.Context(), actor, services.TaskCreateInput{
		OrgID:              &orgID,
		AircraftID:         aircraftID,
		ProgramID:          programID,
		Type:               taskType,
		StartTime:          startTime,
		EndTime:            endTime,
		AssignedMechanicID: mechanicID,
		Notes:              req.Notes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			writeError(w, r, http.StatusConflict, "CONFLICT", "maintenance window overlaps existing task")
			return
		}
		writeDomainError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, mapTask(created))
}

func ListTasks(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Tasks == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	query := r.URL.Query()
	filter := ports.TaskFilter{}
	if actor.IsAdmin() {
		if org := query.Get("org_id"); org != "" {
			orgID, err := uuid.Parse(org)
			if err != nil {
				writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
				return
			}
			filter.OrgID = &orgID
		}
	} else {
		filter.OrgID = &actor.OrgID
	}
	if aircraft := query.Get("aircraft_id"); aircraft != "" {
		id, err := uuid.Parse(aircraft)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid aircraft_id")
			return
		}
		filter.AircraftID = &id
	}
	if state := query.Get("state"); state != "" {
		value := domain.TaskState(state)
		if !validTaskState(value) {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid state")
			return
		}
		filter.State = &value
	}
	if taskType := query.Get("type"); taskType != "" {
		value := domain.TaskType(taskType)
		if !validTaskType(value) {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid type")
			return
		}
		filter.Type = &value
	}
	if startFrom := query.Get("start_from"); startFrom != "" {
		value, err := time.Parse(time.RFC3339, startFrom)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid start_from")
			return
		}
		filter.StartFrom = &value
	}
	if startTo := query.Get("start_to"); startTo != "" {
		value, err := time.Parse(time.RFC3339, startTo)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid start_to")
			return
		}
		filter.StartTo = &value
	}
	if limit := query.Get("limit"); limit != "" {
		value, err := parseInt(limit)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid limit")
			return
		}
		filter.Limit = value
	}
	if offset := query.Get("offset"); offset != "" {
		value, err := parseInt(offset)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid offset")
			return
		}
		filter.Offset = value
	}

	tasks, err := servicesReg.Tasks.List(r.Context(), actor, filter)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]taskResponse, 0, len(tasks))
	for _, task := range tasks {
		resp = append(resp, mapTask(task))
	}
	writeJSON(w, http.StatusOK, resp)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Tasks == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid task id")
		return
	}
	orgID := actor.OrgID
	if actor.IsAdmin() {
		if org := r.URL.Query().Get("org_id"); org != "" {
			parsed, err := uuid.Parse(org)
			if err != nil {
				writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
				return
			}
			orgID = parsed
		}
	}

	task, err := servicesReg.Tasks.Get(r.Context(), actor, orgID, id)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	reservations := []reservationResponse{}
	if servicesReg.Parts != nil {
		list, err := servicesReg.Parts.ListByTask(r.Context(), orgID, task.ID)
		if err != nil {
			writeDomainError(w, r, err)
			return
		}
		for _, res := range list {
			reservations = append(reservations, reservationResponse{
				ID:         res.ID,
				TaskID:     res.TaskID,
				PartItemID: res.PartItemID,
				State:      res.State,
			})
		}
	}

	complianceItems := []complianceResponse{}
	if servicesReg.Compliance != nil {
		list, err := servicesReg.Compliance.ListByTask(r.Context(), orgID, task.ID)
		if err != nil {
			writeDomainError(w, r, err)
			return
		}
		for _, item := range list {
			complianceItems = append(complianceItems, complianceResponse{
				ID:          item.ID,
				TaskID:      item.TaskID,
				Description: item.Description,
				Result:      item.Result,
				SignedOff:   item.SignOffTime != nil,
			})
		}
	}

	writeJSON(w, http.StatusOK, taskDetailResponse{
		Task:         mapTask(task),
		Reservations: reservations,
		Compliance:   complianceItems,
	})
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Tasks == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid task id")
		return
	}
	var req taskUpdateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	if req.ProgramID == "" && req.Type == "" && req.StartTime == "" && req.EndTime == "" && req.AssignedMechanicID == "" && req.Notes == nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "no changes provided")
		return
	}
	orgID, err := resolveOrgID(actor, req.OrgID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
		return
	}

	input := services.TaskUpdateInput{}
	if req.ProgramID != "" {
		parsed, err := uuid.Parse(req.ProgramID)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid program_id")
			return
		}
		input.ProgramID = &parsed
	}
	if req.Type != "" {
		value := domain.TaskType(req.Type)
		if !validTaskType(value) {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid type")
			return
		}
		input.Type = &value
	}
	if req.StartTime != "" {
		value, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid start_time")
			return
		}
		input.StartTime = &value
	}
	if req.EndTime != "" {
		value, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid end_time")
			return
		}
		input.EndTime = &value
	}
	if req.AssignedMechanicID != "" {
		parsed, err := uuid.Parse(req.AssignedMechanicID)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid assigned_mechanic_id")
			return
		}
		input.AssignedMechanicID = &parsed
	}
	if req.Notes != nil {
		input.Notes = req.Notes
	}

	updated, err := servicesReg.Tasks.Update(r.Context(), actor, orgID, id, input)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			writeError(w, r, http.StatusConflict, "CONFLICT", "maintenance window overlaps existing task")
			return
		}
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapTask(updated))
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Tasks == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid task id")
		return
	}
	orgID := actor.OrgID
	if actor.IsAdmin() {
		if org := r.URL.Query().Get("org_id"); org != "" {
			parsed, err := uuid.Parse(org)
			if err != nil {
				writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
				return
			}
			orgID = parsed
		}
	}

	if err := servicesReg.Tasks.Delete(r.Context(), actor, orgID, id); err != nil {
		writeDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func mapTask(task domain.MaintenanceTask) taskResponse {
	return taskResponse{
		ID:                 task.ID,
		OrgID:              task.OrgID,
		AircraftID:         task.AircraftID,
		ProgramID:          task.ProgramID,
		Type:               task.Type,
		State:              task.State,
		StartTime:          task.StartTime.UTC(),
		EndTime:            task.EndTime.UTC(),
		AssignedMechanicID: task.AssignedMechanicID,
		Notes:              task.Notes,
		CreatedAt:          task.CreatedAt,
		UpdatedAt:          task.UpdatedAt,
	}
}

func validTaskType(value domain.TaskType) bool {
	switch value {
	case domain.TaskTypeInspection, domain.TaskTypeRepair, domain.TaskTypeOverhaul:
		return true
	default:
		return false
	}
}
