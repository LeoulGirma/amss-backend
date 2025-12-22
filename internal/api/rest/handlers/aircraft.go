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

type aircraftCreateRequest struct {
	OrgID            string `json:"org_id" validate:"omitempty,uuid"`
	TailNumber       string `json:"tail_number" validate:"required"`
	Model            string `json:"model" validate:"required"`
	LastMaintenance  string `json:"last_maintenance" validate:"omitempty,rfc3339"`
	NextDue          string `json:"next_due" validate:"omitempty,rfc3339"`
	Status           string `json:"status" validate:"omitempty,oneof=operational maintenance grounded"`
	CapacitySlots    int    `json:"capacity_slots" validate:"min=1"`
	FlightHoursTotal int    `json:"flight_hours_total" validate:"min=0"`
	CyclesTotal      int    `json:"cycles_total" validate:"min=0"`
}

type aircraftUpdateRequest struct {
	OrgID            string  `json:"org_id" validate:"omitempty,uuid"`
	TailNumber       *string `json:"tail_number" validate:"omitempty,min=1"`
	Model            *string `json:"model" validate:"omitempty,min=1"`
	LastMaintenance  *string `json:"last_maintenance" validate:"omitempty,rfc3339"`
	NextDue          *string `json:"next_due" validate:"omitempty,rfc3339"`
	Status           *string `json:"status" validate:"omitempty,oneof=operational maintenance grounded"`
	CapacitySlots    *int    `json:"capacity_slots" validate:"omitempty,min=1"`
	FlightHoursTotal *int    `json:"flight_hours_total" validate:"omitempty,min=0"`
	CyclesTotal      *int    `json:"cycles_total" validate:"omitempty,min=0"`
}

type aircraftResponse struct {
	ID               uuid.UUID             `json:"id"`
	OrgID            uuid.UUID             `json:"org_id"`
	TailNumber       string                `json:"tail_number"`
	Model            string                `json:"model"`
	LastMaintenance  *time.Time            `json:"last_maintenance,omitempty"`
	NextDue          *time.Time            `json:"next_due,omitempty"`
	Status           domain.AircraftStatus `json:"status"`
	CapacitySlots    int                   `json:"capacity_slots"`
	FlightHoursTotal int                   `json:"flight_hours_total"`
	CyclesTotal      int                   `json:"cycles_total"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
}

func CreateAircraft(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Aircraft == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	var req aircraftCreateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	orgID, err := resolveOrgID(actor, req.OrgID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
		return
	}
	var lastMaintenance *time.Time
	if req.LastMaintenance != "" {
		value, err := time.Parse(time.RFC3339, req.LastMaintenance)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid last_maintenance")
			return
		}
		lastMaintenance = &value
	}
	var nextDue *time.Time
	if req.NextDue != "" {
		value, err := time.Parse(time.RFC3339, req.NextDue)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid next_due")
			return
		}
		nextDue = &value
	}
	status := domain.AircraftStatus(req.Status)
	input := services.AircraftCreateInput{
		OrgID:            &orgID,
		TailNumber:       req.TailNumber,
		Model:            req.Model,
		LastMaintenance:  lastMaintenance,
		NextDue:          nextDue,
		Status:           status,
		CapacitySlots:    req.CapacitySlots,
		FlightHoursTotal: req.FlightHoursTotal,
		CyclesTotal:      req.CyclesTotal,
	}
	created, err := servicesReg.Aircraft.Create(r.Context(), actor, input)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapAircraft(created))
}

func ListAircraft(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Aircraft == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	query := r.URL.Query()
	filter := ports.AircraftFilter{}
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
	if status := query.Get("status"); status != "" {
		value := domain.AircraftStatus(status)
		filter.Status = &value
	}
	filter.Model = query.Get("model")
	filter.TailNumber = query.Get("tail_number")
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

	items, err := servicesReg.Aircraft.List(r.Context(), actor, filter)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]aircraftResponse, 0, len(items))
	for _, aircraft := range items {
		resp = append(resp, mapAircraft(aircraft))
	}
	writeJSON(w, http.StatusOK, resp)
}

func GetAircraft(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Aircraft == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid aircraft id")
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
	aircraft, err := servicesReg.Aircraft.Get(r.Context(), actor, orgID, id)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapAircraft(aircraft))
}

func UpdateAircraft(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Aircraft == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid aircraft id")
		return
	}
	var req aircraftUpdateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	orgID, err := resolveOrgID(actor, req.OrgID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
		return
	}
	var lastMaintenance *time.Time
	if req.LastMaintenance != nil {
		if *req.LastMaintenance != "" {
			value, err := time.Parse(time.RFC3339, *req.LastMaintenance)
			if err != nil {
				writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid last_maintenance")
				return
			}
			lastMaintenance = &value
		}
	}
	var nextDue *time.Time
	if req.NextDue != nil {
		if *req.NextDue != "" {
			value, err := time.Parse(time.RFC3339, *req.NextDue)
			if err != nil {
				writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid next_due")
				return
			}
			nextDue = &value
		}
	}
	var status *domain.AircraftStatus
	if req.Status != nil {
		value := domain.AircraftStatus(*req.Status)
		status = &value
	}
	input := services.AircraftUpdateInput{
		TailNumber:       req.TailNumber,
		Model:            req.Model,
		LastMaintenance:  lastMaintenance,
		NextDue:          nextDue,
		Status:           status,
		CapacitySlots:    req.CapacitySlots,
		FlightHoursTotal: req.FlightHoursTotal,
		CyclesTotal:      req.CyclesTotal,
	}
	updated, err := servicesReg.Aircraft.Update(r.Context(), actor, orgID, id, input)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapAircraft(updated))
}

func DeleteAircraft(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Aircraft == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid aircraft id")
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
	if err := servicesReg.Aircraft.Delete(r.Context(), actor, orgID, id); err != nil {
		writeDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func mapAircraft(aircraft domain.Aircraft) aircraftResponse {
	return aircraftResponse{
		ID:               aircraft.ID,
		OrgID:            aircraft.OrgID,
		TailNumber:       aircraft.TailNumber,
		Model:            aircraft.Model,
		LastMaintenance:  aircraft.LastMaintenance,
		NextDue:          aircraft.NextDue,
		Status:           aircraft.Status,
		CapacitySlots:    aircraft.CapacitySlots,
		FlightHoursTotal: aircraft.FlightHoursTotal,
		CyclesTotal:      aircraft.CyclesTotal,
		CreatedAt:        aircraft.CreatedAt,
		UpdatedAt:        aircraft.UpdatedAt,
	}
}
