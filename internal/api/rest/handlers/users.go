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

type userCreateRequest struct {
	OrgID    string `json:"org_id" validate:"omitempty,uuid"`
	Email    string `json:"email" validate:"required,email"`
	Role     string `json:"role" validate:"required,oneof=admin tenant_admin scheduler mechanic auditor"`
	Password string `json:"password" validate:"required"`
}

type userUpdateRequest struct {
	OrgID    string  `json:"org_id" validate:"omitempty,uuid"`
	Email    *string `json:"email" validate:"omitempty,email"`
	Role     *string `json:"role" validate:"omitempty,oneof=admin tenant_admin scheduler mechanic auditor"`
	Password *string `json:"password" validate:"omitempty,min=1"`
}

type userResponse struct {
	ID        uuid.UUID   `json:"id"`
	OrgID     uuid.UUID   `json:"org_id"`
	Email     string      `json:"email"`
	Role      domain.Role `json:"role"`
	LastLogin *time.Time  `json:"last_login,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Users == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	var req userCreateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	orgID, err := resolveOrgID(actor, req.OrgID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
		return
	}
	role := domain.Role(req.Role)
	input := services.UserCreateInput{
		OrgID:    &orgID,
		Email:    req.Email,
		Role:     role,
		Password: req.Password,
	}
	created, err := servicesReg.Users.Create(r.Context(), actor, input)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapUser(created))
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Users == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	filter := ports.UserFilter{}
	query := r.URL.Query()
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
	if role := query.Get("role"); role != "" {
		value := domain.Role(role)
		filter.Role = &value
	}
	filter.Email = query.Get("email")
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

	users, err := servicesReg.Users.List(r.Context(), actor, filter)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]userResponse, 0, len(users))
	for _, user := range users {
		resp = append(resp, mapUser(user))
	}
	writeJSON(w, http.StatusOK, resp)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Users == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid user id")
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

	user, err := servicesReg.Users.Get(r.Context(), actor, orgID, id)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapUser(user))
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Users == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid user id")
		return
	}
	var req userUpdateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	orgID, err := resolveOrgID(actor, req.OrgID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
		return
	}

	var role *domain.Role
	if req.Role != nil {
		value := domain.Role(*req.Role)
		role = &value
	}
	input := services.UserUpdateInput{
		Email:    req.Email,
		Role:     role,
		Password: req.Password,
	}
	updated, err := servicesReg.Users.Update(r.Context(), actor, orgID, id, input)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapUser(updated))
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Users == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid user id")
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
	if err := servicesReg.Users.Delete(r.Context(), actor, orgID, id); err != nil {
		writeDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func mapUser(user domain.User) userResponse {
	return userResponse{
		ID:        user.ID,
		OrgID:     user.OrgID,
		Email:     user.Email,
		Role:      user.Role,
		LastLogin: user.LastLogin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
