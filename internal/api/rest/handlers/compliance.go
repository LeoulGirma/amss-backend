package handlers

import (
	"net/http"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type complianceCreateRequest struct {
	TaskID      string `json:"task_id" validate:"required,uuid"`
	Description string `json:"description" validate:"required"`
	Result      string `json:"result" validate:"required,oneof=pass fail pending"`
}

type complianceUpdateRequest struct {
	Description string `json:"description" validate:"required"`
	Result      string `json:"result" validate:"required,oneof=pass fail pending"`
}

type complianceResponse struct {
	ID          uuid.UUID               `json:"id"`
	TaskID      uuid.UUID               `json:"task_id"`
	Description string                  `json:"description"`
	Result      domain.ComplianceResult `json:"result"`
	SignedOff   bool                    `json:"signed_off"`
}

func CreateComplianceItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	services, ok := servicesFromRequest(r)
	if !ok || services.Compliance == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	var req complianceCreateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid task_id")
		return
	}
	result := domain.ComplianceResult(req.Result)
	if !validComplianceResult(result) {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid result")
		return
	}

	item := domain.ComplianceItem{
		TaskID:      taskID,
		Description: req.Description,
		Result:      result,
	}

	created, err := services.Compliance.Create(r.Context(), actor, item)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, complianceResponse{
		ID:          created.ID,
		TaskID:      created.TaskID,
		Description: created.Description,
		Result:      created.Result,
		SignedOff:   created.SignOffTime != nil,
	})
}

func UpdateComplianceItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	services, ok := servicesFromRequest(r)
	if !ok || services.Compliance == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid compliance id")
		return
	}
	var req complianceUpdateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	result := domain.ComplianceResult(req.Result)
	if !validComplianceResult(result) {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid result")
		return
	}

	item := domain.ComplianceItem{
		ID:          id,
		Description: req.Description,
		Result:      result,
	}
	updated, err := services.Compliance.Update(r.Context(), actor, item)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, complianceResponse{
		ID:          updated.ID,
		TaskID:      updated.TaskID,
		Description: updated.Description,
		Result:      updated.Result,
		SignedOff:   updated.SignOffTime != nil,
	})
}

func SignOffComplianceItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	services, ok := servicesFromRequest(r)
	if !ok || services.Compliance == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid compliance id")
		return
	}

	updated, err := services.Compliance.SignOff(r.Context(), actor, id)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, complianceResponse{
		ID:          updated.ID,
		TaskID:      updated.TaskID,
		Description: updated.Description,
		Result:      updated.Result,
		SignedOff:   updated.SignOffTime != nil,
	})
}

func ListComplianceItems(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	services, ok := servicesFromRequest(r)
	if !ok || services.Compliance == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	filter, err := parseComplianceFilter(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	items, err := services.Compliance.List(r.Context(), actor, filter)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]complianceResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, complianceResponse{
			ID:          item.ID,
			TaskID:      item.TaskID,
			Description: item.Description,
			Result:      item.Result,
			SignedOff:   item.SignOffTime != nil,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func parseComplianceFilter(r *http.Request) (ports.ComplianceFilter, error) {
	query := r.URL.Query()
	var filter ports.ComplianceFilter
	if org := query.Get("org_id"); org != "" {
		orgID, err := uuid.Parse(org)
		if err != nil {
			return filter, errInvalid("invalid org_id")
		}
		filter.OrgID = &orgID
	}
	if taskID := query.Get("task_id"); taskID != "" {
		id, err := uuid.Parse(taskID)
		if err != nil {
			return filter, errInvalid("invalid task_id")
		}
		filter.TaskID = &id
	}
	if result := query.Get("result"); result != "" {
		value := domain.ComplianceResult(result)
		if !validComplianceResult(value) {
			return filter, errInvalid("invalid result")
		}
		filter.Result = &value
	}
	if signed := query.Get("signed"); signed != "" {
		value, err := parseBool(signed)
		if err != nil {
			return filter, errInvalid("invalid signed")
		}
		filter.Signed = &value
	}
	if limit := query.Get("limit"); limit != "" {
		value, err := parseInt(limit)
		if err != nil {
			return filter, errInvalid("invalid limit")
		}
		filter.Limit = value
	}
	if offset := query.Get("offset"); offset != "" {
		value, err := parseInt(offset)
		if err != nil {
			return filter, errInvalid("invalid offset")
		}
		filter.Offset = value
	}
	return filter, nil
}

func validComplianceResult(result domain.ComplianceResult) bool {
	switch result {
	case domain.CompliancePass, domain.ComplianceFail, domain.CompliancePending:
		return true
	default:
		return false
	}
}
