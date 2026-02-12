package handlers

import (
	"net/http"

	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type taskStateRequest struct {
	NewState             string `json:"new_state" validate:"required,oneof=scheduled in_progress completed cancelled"`
	Notes                string `json:"notes"`
	AllowEarlyCompletion bool   `json:"allow_early_completion"`
	AllowLateCancel      bool   `json:"allow_late_cancel"`
	RequireAllPartsUsed  bool   `json:"require_all_parts_used"`
}

type taskStateResponse struct {
	ID    uuid.UUID        `json:"id"`
	State domain.TaskState `json:"state"`
}

func TransitionTaskState(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleMechanic && actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "forbidden")
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
	var req taskStateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	state := domain.TaskState(req.NewState)
	if !validTaskState(state) {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid state")
		return
	}

	opts := services.TaskTransitionOptions{
		AllowEarlyCompletion: req.AllowEarlyCompletion,
		AllowLateCancel:      req.AllowLateCancel,
		RequireAllPartsUsed:  req.RequireAllPartsUsed,
		Notes:                req.Notes,
	}

	updated, err := servicesReg.Tasks.TransitionState(r.Context(), actor, id, state, opts)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, taskStateResponse{ID: updated.ID, State: updated.State})
}

func validTaskState(state domain.TaskState) bool {
	switch state {
	case domain.TaskStateScheduled, domain.TaskStateInProgress, domain.TaskStateCompleted, domain.TaskStateCancelled:
		return true
	default:
		return false
	}
}
