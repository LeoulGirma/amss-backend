package handlers

import (
	"net/http"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type alertResponse struct {
	ID              uuid.UUID        `json:"id"`
	OrgID           uuid.UUID        `json:"org_id"`
	Level           domain.AlertLevel `json:"level"`
	Category        string           `json:"category"`
	Title           string           `json:"title"`
	Description     string           `json:"description,omitempty"`
	EntityType      string           `json:"entity_type"`
	EntityID        uuid.UUID        `json:"entity_id"`
	ThresholdValue  *float64         `json:"threshold_value,omitempty"`
	CurrentValue    *float64         `json:"current_value,omitempty"`
	Acknowledged    bool             `json:"acknowledged"`
	AcknowledgedBy  *uuid.UUID       `json:"acknowledged_by,omitempty"`
	AcknowledgedAt  *time.Time       `json:"acknowledged_at,omitempty"`
	Resolved        bool             `json:"resolved"`
	ResolvedAt      *time.Time       `json:"resolved_at,omitempty"`
	EscalationLevel int              `json:"escalation_level"`
	CreatedAt       time.Time        `json:"created_at"`
}

func ListAlerts(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Alerts == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	query := r.URL.Query()
	filter := ports.AlertFilter{}
	if level := query.Get("level"); level != "" {
		v := domain.AlertLevel(level)
		filter.Level = &v
	}
	filter.Category = query.Get("category")
	if resolved := query.Get("resolved"); resolved != "" {
		v, err := parseBool(resolved)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid resolved")
			return
		}
		filter.Resolved = &v
	}
	if limit := query.Get("limit"); limit != "" {
		v, err := parseInt(limit)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid limit")
			return
		}
		filter.Limit = v
	}
	if offset := query.Get("offset"); offset != "" {
		v, err := parseInt(offset)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid offset")
			return
		}
		filter.Offset = v
	}

	items, err := servicesReg.Alerts.List(r.Context(), actor, filter)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]alertResponse, 0, len(items))
	for _, a := range items {
		resp = append(resp, mapAlert(a))
	}
	writeJSON(w, http.StatusOK, resp)
}

func AcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Alerts == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid alert id")
		return
	}
	if err := servicesReg.Alerts.Acknowledge(r.Context(), actor, id); err != nil {
		writeDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func ResolveAlert(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Alerts == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid alert id")
		return
	}
	if err := servicesReg.Alerts.Resolve(r.Context(), actor, id); err != nil {
		writeDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func mapAlert(a domain.Alert) alertResponse {
	return alertResponse{
		ID:              a.ID,
		OrgID:           a.OrgID,
		Level:           a.Level,
		Category:        a.Category,
		Title:           a.Title,
		Description:     a.Description,
		EntityType:      a.EntityType,
		EntityID:        a.EntityID,
		ThresholdValue:  a.ThresholdValue,
		CurrentValue:    a.CurrentValue,
		Acknowledged:    a.Acknowledged,
		AcknowledgedBy:  a.AcknowledgedBy,
		AcknowledgedAt:  a.AcknowledgedAt,
		Resolved:        a.Resolved,
		ResolvedAt:      a.ResolvedAt,
		EscalationLevel: a.EscalationLevel,
		CreatedAt:       a.CreatedAt,
	}
}
