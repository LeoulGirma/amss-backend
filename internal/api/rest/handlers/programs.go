package handlers

import (
	"net/http"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type programCreateRequest struct {
	OrgID         string `json:"org_id" validate:"omitempty,uuid"`
	AircraftID    string `json:"aircraft_id" validate:"omitempty,uuid"`
	Name          string `json:"name" validate:"required"`
	IntervalType  string `json:"interval_type" validate:"required,oneof=flight_hours cycles calendar"`
	IntervalValue int    `json:"interval_value" validate:"min=1"`
	LastPerformed string `json:"last_performed" validate:"omitempty,rfc3339"`
}

type programUpdateRequest struct {
	OrgID         string  `json:"org_id" validate:"omitempty,uuid"`
	AircraftID    *string `json:"aircraft_id" validate:"omitempty,uuid"`
	Name          *string `json:"name" validate:"omitempty,min=1"`
	IntervalType  *string `json:"interval_type" validate:"omitempty,oneof=flight_hours cycles calendar"`
	IntervalValue *int    `json:"interval_value" validate:"omitempty,min=1"`
	LastPerformed *string `json:"last_performed" validate:"omitempty,rfc3339"`
}

type programResponse struct {
	ID            uuid.UUID                             `json:"id"`
	OrgID         uuid.UUID                             `json:"org_id"`
	AircraftID    *uuid.UUID                            `json:"aircraft_id,omitempty"`
	Name          string                                `json:"name"`
	IntervalType  domain.MaintenanceProgramIntervalType `json:"interval_type"`
	IntervalValue int                                   `json:"interval_value"`
	LastPerformed *time.Time                            `json:"last_performed,omitempty"`
	CreatedAt     time.Time                             `json:"created_at"`
	UpdatedAt     time.Time                             `json:"updated_at"`
}

func CreateProgram(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Programs == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	var req programCreateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	orgID, err := resolveOrgID(actor, req.OrgID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
		return
	}
	var aircraftID *uuid.UUID
	if req.AircraftID != "" {
		parsed, err := uuid.Parse(req.AircraftID)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid aircraft_id")
			return
		}
		aircraftID = &parsed
	}
	var lastPerformed *time.Time
	if req.LastPerformed != "" {
		value, err := time.Parse(time.RFC3339, req.LastPerformed)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid last_performed")
			return
		}
		lastPerformed = &value
	}
	input := services.ProgramCreateInput{
		OrgID:         &orgID,
		AircraftID:    aircraftID,
		Name:          req.Name,
		IntervalType:  domain.MaintenanceProgramIntervalType(req.IntervalType),
		IntervalValue: req.IntervalValue,
		LastPerformed: lastPerformed,
	}
	created, err := servicesReg.Programs.Create(r.Context(), actor, input)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapProgram(created))
}

func ListPrograms(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Programs == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	query := r.URL.Query()
	filter := ports.MaintenanceProgramFilter{}
	if actor.IsAdmin() {
		if org := query.Get("org_id"); org != "" {
			orgID, err := uuid.Parse(org)
			if err != nil {
				writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
				return
			}
			filter.OrgID = &orgID
		}
	}
	if aircraft := query.Get("aircraft_id"); aircraft != "" {
		id, err := uuid.Parse(aircraft)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid aircraft_id")
			return
		}
		filter.AircraftID = &id
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

	items, err := servicesReg.Programs.List(r.Context(), actor, filter)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]programResponse, 0, len(items))
	for _, program := range items {
		resp = append(resp, mapProgram(program))
	}
	writeJSON(w, http.StatusOK, resp)
}

func GetProgram(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Programs == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid program id")
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
	program, err := servicesReg.Programs.Get(r.Context(), actor, orgID, id)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapProgram(program))
}

func UpdateProgram(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Programs == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid program id")
		return
	}
	var req programUpdateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	orgID, err := resolveOrgID(actor, req.OrgID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
		return
	}
	var aircraftID *uuid.UUID
	if req.AircraftID != nil {
		if *req.AircraftID != "" {
			parsed, err := uuid.Parse(*req.AircraftID)
			if err != nil {
				writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid aircraft_id")
				return
			}
			aircraftID = &parsed
		}
	}
	var intervalType *domain.MaintenanceProgramIntervalType
	if req.IntervalType != nil {
		value := domain.MaintenanceProgramIntervalType(*req.IntervalType)
		intervalType = &value
	}
	var lastPerformed *time.Time
	if req.LastPerformed != nil {
		if *req.LastPerformed != "" {
			value, err := time.Parse(time.RFC3339, *req.LastPerformed)
			if err != nil {
				writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid last_performed")
				return
			}
			lastPerformed = &value
		}
	}

	input := services.ProgramUpdateInput{
		AircraftID:    aircraftID,
		Name:          req.Name,
		IntervalType:  intervalType,
		IntervalValue: req.IntervalValue,
		LastPerformed: lastPerformed,
	}
	updated, err := servicesReg.Programs.Update(r.Context(), actor, orgID, id, input)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapProgram(updated))
}

func DeleteProgram(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Programs == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid program id")
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
	if err := servicesReg.Programs.Delete(r.Context(), actor, orgID, id); err != nil {
		writeDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func mapProgram(program domain.MaintenanceProgram) programResponse {
	return programResponse{
		ID:            program.ID,
		OrgID:         program.OrgID,
		AircraftID:    program.AircraftID,
		Name:          program.Name,
		IntervalType:  program.IntervalType,
		IntervalValue: program.IntervalValue,
		LastPerformed: program.LastPerformed,
		CreatedAt:     program.CreatedAt,
		UpdatedAt:     program.UpdatedAt,
	}
}
