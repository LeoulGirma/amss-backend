package handlers

import (
	"net/http"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type partDefinitionRequest struct {
	OrgID    string `json:"org_id" validate:"omitempty,uuid"`
	Name     string `json:"name" validate:"required"`
	Category string `json:"category" validate:"required"`
}

type partDefinitionResponse struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type partItemCreateRequest struct {
	OrgID            string `json:"org_id" validate:"omitempty,uuid"`
	PartDefinitionID string `json:"part_definition_id" validate:"required,uuid"`
	SerialNumber     string `json:"serial_number" validate:"required"`
	Status           string `json:"status" validate:"omitempty,oneof=in_stock used disposed"`
	ExpiryDate       string `json:"expiry_date" validate:"omitempty,rfc3339"`
}

type partItemUpdateRequest struct {
	Status     string `json:"status" validate:"omitempty,oneof=in_stock used disposed"`
	ExpiryDate string `json:"expiry_date" validate:"omitempty,rfc3339"`
	OrgID      string `json:"org_id" validate:"omitempty,uuid"`
}

type partItemResponse struct {
	ID               uuid.UUID             `json:"id"`
	OrgID            uuid.UUID             `json:"org_id"`
	PartDefinitionID uuid.UUID             `json:"part_definition_id"`
	SerialNumber     string                `json:"serial_number"`
	Status           domain.PartItemStatus `json:"status"`
	ExpiryDate       *time.Time            `json:"expiry_date,omitempty"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
}

func CreatePartDefinition(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Catalog == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	var req partDefinitionRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	orgID, err := resolveOrgID(actor, req.OrgID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
		return
	}

	created, err := servicesReg.Catalog.CreateDefinition(r.Context(), actor, orgID, req.Name, req.Category)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, mapPartDefinition(created))
}

func ListPartDefinitions(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Catalog == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	query := r.URL.Query()
	filter := ports.PartDefinitionFilter{
		Name: query.Get("name"),
	}
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

	defs, err := servicesReg.Catalog.ListDefinitions(r.Context(), actor, filter)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]partDefinitionResponse, 0, len(defs))
	for _, def := range defs {
		resp = append(resp, mapPartDefinition(def))
	}
	writeJSON(w, http.StatusOK, resp)
}

func UpdatePartDefinition(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Catalog == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid id")
		return
	}
	var req partDefinitionRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
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

	updated, err := servicesReg.Catalog.UpdateDefinition(r.Context(), actor, orgID, id, req.Name, req.Category)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapPartDefinition(updated))
}

func DeletePartDefinition(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Catalog == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid id")
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

	if err := servicesReg.Catalog.DeleteDefinition(r.Context(), actor, orgID, id); err != nil {
		writeDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func CreatePartItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Catalog == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	var req partItemCreateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	orgID, err := resolveOrgID(actor, req.OrgID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
		return
	}
	defID, err := uuid.Parse(req.PartDefinitionID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid part_definition_id")
		return
	}
	status := domain.PartItemStatus(req.Status)
	if req.Status != "" && status != domain.PartItemInStock && status != domain.PartItemUsed && status != domain.PartItemDisposed {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid status")
		return
	}
	var expiry *time.Time
	if req.ExpiryDate != "" {
		value, err := time.Parse(time.RFC3339, req.ExpiryDate)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid expiry_date")
			return
		}
		expiry = &value
	}

	created, err := servicesReg.Catalog.CreateItem(r.Context(), actor, orgID, defID, req.SerialNumber, status, expiry)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapPartItem(created))
}

func ListPartItems(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Catalog == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	query := r.URL.Query()
	filter := ports.PartItemFilter{}
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
	if defID := query.Get("definition_id"); defID != "" {
		parsed, err := uuid.Parse(defID)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid definition_id")
			return
		}
		filter.DefinitionID = &parsed
	}
	if status := query.Get("status"); status != "" {
		value := domain.PartItemStatus(status)
		if value != domain.PartItemInStock && value != domain.PartItemUsed && value != domain.PartItemDisposed {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid status")
			return
		}
		filter.Status = &value
	}
	if expiry := query.Get("expiry_before"); expiry != "" {
		value, err := time.Parse(time.RFC3339, expiry)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid expiry_before")
			return
		}
		filter.ExpiryBefore = &value
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

	items, err := servicesReg.Catalog.ListItems(r.Context(), actor, filter)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]partItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, mapPartItem(item))
	}
	writeJSON(w, http.StatusOK, resp)
}

func UpdatePartItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Catalog == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid id")
		return
	}
	var req partItemUpdateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	if req.Status == "" && req.ExpiryDate == "" {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "no changes provided")
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

	var status *domain.PartItemStatus
	if req.Status != "" {
		value := domain.PartItemStatus(req.Status)
		if value != domain.PartItemInStock && value != domain.PartItemUsed && value != domain.PartItemDisposed {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid status")
			return
		}
		status = &value
	}
	var expiry *time.Time
	if req.ExpiryDate != "" {
		value, err := time.Parse(time.RFC3339, req.ExpiryDate)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid expiry_date")
			return
		}
		expiry = &value
	}

	updated, err := servicesReg.Catalog.UpdateItem(r.Context(), actor, orgID, id, status, expiry)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapPartItem(updated))
}

func DeletePartItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Catalog == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid id")
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
	if err := servicesReg.Catalog.DeleteItem(r.Context(), actor, orgID, id); err != nil {
		writeDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func mapPartDefinition(def domain.PartDefinition) partDefinitionResponse {
	return partDefinitionResponse{
		ID:        def.ID,
		OrgID:     def.OrgID,
		Name:      def.Name,
		Category:  def.Category,
		CreatedAt: def.CreatedAt,
		UpdatedAt: def.UpdatedAt,
	}
}

func mapPartItem(item domain.PartItem) partItemResponse {
	return partItemResponse{
		ID:               item.ID,
		OrgID:            item.OrgID,
		PartDefinitionID: item.DefinitionID,
		SerialNumber:     item.SerialNumber,
		Status:           item.Status,
		ExpiryDate:       item.ExpiryDate,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}
