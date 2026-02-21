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

// --- Request/Response Types ---

type directiveCreateRequest struct {
	AuthorityID           string   `json:"authority_id" validate:"required,uuid"`
	DirectiveType         string   `json:"directive_type" validate:"required,oneof=ad sb eo tcds stc"`
	ReferenceNumber       string   `json:"reference_number" validate:"required"`
	Title                 string   `json:"title" validate:"required"`
	Description           string   `json:"description"`
	Applicability         string   `json:"applicability" validate:"required,oneof=mandatory recommended optional"`
	AffectedAircraftTypes []string `json:"affected_aircraft_types"`
	EffectiveDate         string   `json:"effective_date" validate:"required,rfc3339"`
	ComplianceDeadline    string   `json:"compliance_deadline" validate:"omitempty,rfc3339"`
	RecurrenceInterval    string   `json:"recurrence_interval"`
	SourceURL             string   `json:"source_url"`
}

type complianceUpdateRequest struct {
	AircraftID  string  `json:"aircraft_id" validate:"required,uuid"`
	DirectiveID string  `json:"directive_id" validate:"required,uuid"`
	Status      string  `json:"status" validate:"required,oneof=pending in_progress compliant not_applicable overdue"`
	TaskID      *string `json:"task_id" validate:"omitempty,uuid"`
	Notes       string  `json:"notes"`
}

type directiveResponse struct {
	ID                    uuid.UUID                     `json:"id"`
	OrgID                 uuid.UUID                     `json:"org_id"`
	AuthorityID           uuid.UUID                     `json:"authority_id"`
	DirectiveType         domain.DirectiveType          `json:"directive_type"`
	ReferenceNumber       string                        `json:"reference_number"`
	Title                 string                        `json:"title"`
	Description           string                        `json:"description,omitempty"`
	Applicability         domain.DirectiveApplicability `json:"applicability"`
	AffectedAircraftTypes []uuid.UUID                   `json:"affected_aircraft_types,omitempty"`
	EffectiveDate         time.Time                     `json:"effective_date"`
	ComplianceDeadline    *time.Time                    `json:"compliance_deadline,omitempty"`
	RecurrenceInterval    string                        `json:"recurrence_interval,omitempty"`
	SourceURL             string                        `json:"source_url,omitempty"`
	CreatedAt             time.Time                     `json:"created_at"`
	UpdatedAt             time.Time                     `json:"updated_at"`
}

type authorityResponse struct {
	ID                   uuid.UUID `json:"id"`
	Code                 string    `json:"code"`
	Name                 string    `json:"name"`
	Country              string    `json:"country"`
	RecordRetentionYears int       `json:"record_retention_years"`
}

type aircraftComplianceResponse struct {
	ID             uuid.UUID                        `json:"id"`
	OrgID          uuid.UUID                        `json:"org_id"`
	AircraftID     uuid.UUID                        `json:"aircraft_id"`
	DirectiveID    uuid.UUID                        `json:"directive_id"`
	Status         domain.DirectiveComplianceStatus `json:"status"`
	ComplianceDate *time.Time                       `json:"compliance_date,omitempty"`
	NextDueDate    *time.Time                       `json:"next_due_date,omitempty"`
	TaskID         *uuid.UUID                       `json:"task_id,omitempty"`
	SignedOffBy    *uuid.UUID                       `json:"signed_off_by,omitempty"`
	SignedOffAt    *time.Time                       `json:"signed_off_at,omitempty"`
	Notes          string                           `json:"notes,omitempty"`
	CreatedAt      time.Time                        `json:"created_at"`
	UpdatedAt      time.Time                        `json:"updated_at"`
}

// --- Handlers ---

func ListAuthorities(w http.ResponseWriter, r *http.Request) {
	_, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Directives == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	authorities, err := servicesReg.Directives.ListAuthorities(r.Context())
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]authorityResponse, 0, len(authorities))
	for _, a := range authorities {
		resp = append(resp, authorityResponse{
			ID:                   a.ID,
			Code:                 a.Code,
			Name:                 a.Name,
			Country:              a.Country,
			RecordRetentionYears: a.RecordRetentionYears,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func CreateDirective(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Directives == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	var req directiveCreateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	authorityID, _ := uuid.Parse(req.AuthorityID)
	effectiveDate, _ := time.Parse(time.RFC3339, req.EffectiveDate)

	var complianceDeadline *time.Time
	if req.ComplianceDeadline != "" {
		v, _ := time.Parse(time.RFC3339, req.ComplianceDeadline)
		complianceDeadline = &v
	}

	var affectedTypes []uuid.UUID
	for _, id := range req.AffectedAircraftTypes {
		parsed, err := uuid.Parse(id)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid aircraft type id")
			return
		}
		affectedTypes = append(affectedTypes, parsed)
	}

	created, err := servicesReg.Directives.CreateDirective(r.Context(), actor, services.DirectiveCreateInput{
		AuthorityID:           authorityID,
		DirectiveType:         domain.DirectiveType(req.DirectiveType),
		ReferenceNumber:       req.ReferenceNumber,
		Title:                 req.Title,
		Description:           req.Description,
		Applicability:         domain.DirectiveApplicability(req.Applicability),
		AffectedAircraftTypes: affectedTypes,
		EffectiveDate:         effectiveDate,
		ComplianceDeadline:    complianceDeadline,
		RecurrenceInterval:    req.RecurrenceInterval,
		SourceURL:             req.SourceURL,
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapDirective(created))
}

func ListDirectives(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Directives == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	query := r.URL.Query()
	filter := ports.DirectiveFilter{}
	if authID := query.Get("authority_id"); authID != "" {
		parsed, err := uuid.Parse(authID)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid authority_id")
			return
		}
		filter.AuthorityID = &parsed
	}
	if dt := query.Get("directive_type"); dt != "" {
		v := domain.DirectiveType(dt)
		filter.DirectiveType = &v
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

	items, err := servicesReg.Directives.ListDirectives(r.Context(), actor, filter)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]directiveResponse, 0, len(items))
	for _, d := range items {
		resp = append(resp, mapDirective(d))
	}
	writeJSON(w, http.StatusOK, resp)
}

func GetDirective(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Directives == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid directive id")
		return
	}
	d, err := servicesReg.Directives.GetDirective(r.Context(), actor, id)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapDirective(d))
}

func ListAircraftDirectiveCompliance(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Directives == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	aircraftID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid aircraft id")
		return
	}
	filter := ports.AircraftComplianceFilter{
		AircraftID: &aircraftID,
	}
	if status := r.URL.Query().Get("status"); status != "" {
		v := domain.DirectiveComplianceStatus(status)
		filter.Status = &v
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		v, _ := parseInt(limit)
		filter.Limit = v
	}

	items, err := servicesReg.Directives.ListAircraftCompliance(r.Context(), actor, filter)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]aircraftComplianceResponse, 0, len(items))
	for _, c := range items {
		resp = append(resp, mapAircraftCompliance(c))
	}
	writeJSON(w, http.StatusOK, resp)
}

func UpdateAircraftDirectiveCompliance(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Directives == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	var req complianceUpdateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	aircraftID, _ := uuid.Parse(req.AircraftID)
	directiveID, _ := uuid.Parse(req.DirectiveID)
	var taskID *uuid.UUID
	if req.TaskID != nil {
		v, _ := uuid.Parse(*req.TaskID)
		taskID = &v
	}

	result, err := servicesReg.Directives.UpdateAircraftCompliance(r.Context(), actor, services.ComplianceUpdateInput{
		AircraftID:  aircraftID,
		DirectiveID: directiveID,
		Status:      domain.DirectiveComplianceStatus(req.Status),
		TaskID:      taskID,
		Notes:       req.Notes,
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapAircraftCompliance(result))
}

func ScanFleetForDirective(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Directives == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid directive id")
		return
	}
	count, err := servicesReg.Directives.ScanFleetForDirective(r.Context(), actor, id)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"affected_aircraft_count": count})
}

func ListComplianceTemplates(w http.ResponseWriter, r *http.Request) {
	_, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Directives == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	authorityID, err := uuid.Parse(chi.URLParam(r, "authorityId"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid authority id")
		return
	}
	templates, err := servicesReg.Directives.ListTemplates(r.Context(), authorityID)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, templates)
}

// --- Mappers ---

func mapDirective(d domain.ComplianceDirective) directiveResponse {
	return directiveResponse{
		ID:                    d.ID,
		OrgID:                 d.OrgID,
		AuthorityID:           d.AuthorityID,
		DirectiveType:         d.DirectiveType,
		ReferenceNumber:       d.ReferenceNumber,
		Title:                 d.Title,
		Description:           d.Description,
		Applicability:         d.Applicability,
		AffectedAircraftTypes: d.AffectedAircraftTypes,
		EffectiveDate:         d.EffectiveDate,
		ComplianceDeadline:    d.ComplianceDeadline,
		RecurrenceInterval:    d.RecurrenceInterval,
		SourceURL:             d.SourceURL,
		CreatedAt:             d.CreatedAt,
		UpdatedAt:             d.UpdatedAt,
	}
}

func mapAircraftCompliance(c domain.AircraftDirectiveCompliance) aircraftComplianceResponse {
	return aircraftComplianceResponse{
		ID:             c.ID,
		OrgID:          c.OrgID,
		AircraftID:     c.AircraftID,
		DirectiveID:    c.DirectiveID,
		Status:         c.Status,
		ComplianceDate: c.ComplianceDate,
		NextDueDate:    c.NextDueDate,
		TaskID:         c.TaskID,
		SignedOffBy:    c.SignedOffBy,
		SignedOffAt:    c.SignedOffAt,
		Notes:          c.Notes,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}
