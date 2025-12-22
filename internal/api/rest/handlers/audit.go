package handlers

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type auditLogResponse struct {
	ID            uuid.UUID      `json:"id"`
	OrgID         uuid.UUID      `json:"org_id"`
	EntityType    string         `json:"entity_type"`
	EntityID      uuid.UUID      `json:"entity_id"`
	Action        string         `json:"action"`
	UserID        uuid.UUID      `json:"user_id"`
	RequestID     uuid.UUID      `json:"request_id"`
	IPAddress     string         `json:"ip_address"`
	UserAgent     string         `json:"user_agent"`
	EntityVersion int            `json:"entity_version"`
	Timestamp     time.Time      `json:"timestamp"`
	Details       map[string]any `json:"details"`
}

func ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	services, ok := servicesFromRequest(r)
	if !ok || services.AuditQuery == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}

	filter, err := parseAuditFilter(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}

	entries, err := services.AuditQuery.List(r.Context(), actor, filter)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	resp := make([]auditLogResponse, 0, len(entries))
	for _, entry := range entries {
		resp = append(resp, auditLogResponse{
			ID:            entry.ID,
			OrgID:         entry.OrgID,
			EntityType:    entry.EntityType,
			EntityID:      entry.EntityID,
			Action:        string(entry.Action),
			UserID:        entry.UserID,
			RequestID:     entry.RequestID,
			IPAddress:     entry.IPAddress,
			UserAgent:     entry.UserAgent,
			EntityVersion: entry.EntityVersion,
			Timestamp:     entry.Timestamp,
			Details:       entry.Details,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func ExportAuditLogs(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	if actor.Role != domain.RoleAuditor && actor.Role != domain.RoleAdmin {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "forbidden")
		return
	}
	services, ok := servicesFromRequest(r)
	if !ok || services.AuditQuery == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	filter, err := parseAuditFilter(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	entries, err := services.AuditQuery.List(r.Context(), actor, filter)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"audit-logs.csv\"")
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"id", "org_id", "entity_type", "entity_id", "action", "user_id", "request_id", "ip_address", "user_agent", "entity_version", "timestamp", "details"})
	for _, entry := range entries {
		details := ""
		if entry.Details != nil {
			if data, err := json.Marshal(entry.Details); err == nil {
				details = string(data)
			}
		}
		_ = writer.Write([]string{
			entry.ID.String(),
			entry.OrgID.String(),
			entry.EntityType,
			entry.EntityID.String(),
			string(entry.Action),
			entry.UserID.String(),
			entry.RequestID.String(),
			entry.IPAddress,
			entry.UserAgent,
			strconv.Itoa(entry.EntityVersion),
			entry.Timestamp.Format(time.RFC3339),
			details,
		})
	}
	writer.Flush()
}

func parseAuditFilter(r *http.Request) (ports.AuditLogFilter, error) {
	query := r.URL.Query()
	var filter ports.AuditLogFilter
	if org := query.Get("org_id"); org != "" {
		orgID, err := uuid.Parse(org)
		if err != nil {
			return filter, errInvalid("invalid org_id")
		}
		filter.OrgID = &orgID
	}
	filter.EntityType = query.Get("entity_type")
	if entityID := query.Get("entity_id"); entityID != "" {
		id, err := uuid.Parse(entityID)
		if err != nil {
			return filter, errInvalid("invalid entity_id")
		}
		filter.EntityID = &id
	}
	if userID := query.Get("user_id"); userID != "" {
		id, err := uuid.Parse(userID)
		if err != nil {
			return filter, errInvalid("invalid user_id")
		}
		filter.UserID = &id
	}
	if from := query.Get("from"); from != "" {
		value, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return filter, errInvalid("invalid from")
		}
		filter.From = &value
	}
	if to := query.Get("to"); to != "" {
		value, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return filter, errInvalid("invalid to")
		}
		filter.To = &value
	}
	if limit := query.Get("limit"); limit != "" {
		value, err := strconv.Atoi(limit)
		if err != nil {
			return filter, errInvalid("invalid limit")
		}
		filter.Limit = value
	}
	if offset := query.Get("offset"); offset != "" {
		value, err := strconv.Atoi(offset)
		if err != nil {
			return filter, errInvalid("invalid offset")
		}
		filter.Offset = value
	}
	return filter, nil
}

type invalidParam struct {
	message string
}

func (e invalidParam) Error() string {
	return e.message
}

func errInvalid(message string) error {
	return invalidParam{message: message}
}

func AuditLogsRoute(r chi.Router) {
	r.Get("/audit-logs", ListAuditLogs)
}
