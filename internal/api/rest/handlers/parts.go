package handlers

import (
	"net/http"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type reserveRequest struct {
	TaskID     string `json:"task_id" validate:"required,uuid"`
	PartItemID string `json:"part_item_id" validate:"required,uuid"`
}

type reservationStateRequest struct {
	NewState string `json:"new_state" validate:"required,oneof=released used"`
}

type reservationResponse struct {
	ID         uuid.UUID                   `json:"id"`
	TaskID     uuid.UUID                   `json:"task_id"`
	PartItemID uuid.UUID                   `json:"part_item_id"`
	State      domain.PartReservationState `json:"state"`
}

func ReservePart(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	services, ok := servicesFromRequest(r)
	if !ok || services.Parts == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	var req reserveRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid task_id")
		return
	}
	partItemID, err := uuid.Parse(req.PartItemID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid part_item_id")
		return
	}

	reservation, err := services.Parts.Reserve(r.Context(), actor, taskID, partItemID)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, reservationResponse{
		ID:         reservation.ID,
		TaskID:     reservation.TaskID,
		PartItemID: reservation.PartItemID,
		State:      reservation.State,
	})
}

func UpdateReservationState(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	services, ok := servicesFromRequest(r)
	if !ok || services.Parts == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid reservation id")
		return
	}
	var req reservationStateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	state := domain.PartReservationState(req.NewState)
	if state != domain.ReservationUsed && state != domain.ReservationReleased {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid state")
		return
	}

	reservation, err := services.Parts.UpdateState(r.Context(), actor, id, state)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, reservationResponse{
		ID:         reservation.ID,
		TaskID:     reservation.TaskID,
		PartItemID: reservation.PartItemID,
		State:      reservation.State,
	})
}
