package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aeromaintain/amss/internal/api/rest/middleware"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

func TestCreateUser(t *testing.T) {
	orgID := uuid.New()
	userRepo := newFakeUserRepo()
	userService := &services.UserService{Users: userRepo}
	registry := middleware.ServiceRegistry{Users: userService}

	req := newJSONRequest(t, http.MethodPost, "/api/v1/users", map[string]any{
		"email":    "tech@example.com",
		"role":     string(domain.RoleScheduler),
		"password": "Secret123!",
	})
	req = withPrincipal(req, orgID, domain.RoleTenantAdmin)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreateUser))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	var resp userResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Email != "tech@example.com" {
		t.Fatalf("expected email tech@example.com, got %s", resp.Email)
	}
	if resp.Role != domain.RoleScheduler {
		t.Fatalf("expected role scheduler, got %s", resp.Role)
	}
}

func TestCreateUserInvalidJSON(t *testing.T) {
	orgID := uuid.New()
	userRepo := newFakeUserRepo()
	userService := &services.UserService{Users: userRepo}
	registry := middleware.ServiceRegistry{Users: userService}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	req = withPrincipal(req, orgID, domain.RoleTenantAdmin)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreateUser))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestListUsers(t *testing.T) {
	orgID := uuid.New()
	userRepo := newFakeUserRepo()
	userService := &services.UserService{Users: userRepo}
	registry := middleware.ServiceRegistry{Users: userService}

	user := domain.User{
		ID:           uuid.New(),
		OrgID:        orgID,
		Email:        "tech@example.com",
		Role:         domain.RoleScheduler,
		PasswordHash: "hash",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	_, _ = userRepo.Create(context.Background(), user)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/users", nil)
	req = withPrincipal(req, orgID, domain.RoleTenantAdmin)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(ListUsers))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp []userResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 user, got %d", len(resp))
	}
	if resp[0].ID != user.ID {
		t.Fatalf("expected user id %s, got %s", user.ID, resp[0].ID)
	}
}

func TestListUsersUnauthorized(t *testing.T) {
	req := newJSONRequest(t, http.MethodGet, "/api/v1/users", nil)
	rr := httptest.NewRecorder()
	ListUsers(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestGetUser(t *testing.T) {
	orgID := uuid.New()
	userRepo := newFakeUserRepo()
	userService := &services.UserService{Users: userRepo}
	registry := middleware.ServiceRegistry{Users: userService}

	user := domain.User{
		ID:           uuid.New(),
		OrgID:        orgID,
		Email:        "tech@example.com",
		Role:         domain.RoleScheduler,
		PasswordHash: "hash",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	_, _ = userRepo.Create(context.Background(), user)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/users/"+user.ID.String(), nil)
	req = withPrincipal(req, orgID, domain.RoleTenantAdmin)
	req = withRouteParam(req, "id", user.ID.String())

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(GetUser))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp userResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != user.ID {
		t.Fatalf("expected user id %s, got %s", user.ID, resp.ID)
	}
}

func TestGetUserUnauthorized(t *testing.T) {
	id := uuid.New()
	req := newJSONRequest(t, http.MethodGet, "/api/v1/users/"+id.String(), nil)
	req = withRouteParam(req, "id", id.String())

	rr := httptest.NewRecorder()
	GetUser(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestUpdateUser(t *testing.T) {
	orgID := uuid.New()
	userRepo := newFakeUserRepo()
	userService := &services.UserService{Users: userRepo}
	registry := middleware.ServiceRegistry{Users: userService}

	user := domain.User{
		ID:           uuid.New(),
		OrgID:        orgID,
		Email:        "tech@example.com",
		Role:         domain.RoleScheduler,
		PasswordHash: "hash",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	_, _ = userRepo.Create(context.Background(), user)

	req := newJSONRequest(t, http.MethodPatch, "/api/v1/users/"+user.ID.String(), map[string]any{
		"email": "planner@example.com",
	})
	req = withPrincipal(req, orgID, domain.RoleTenantAdmin)
	req = withRouteParam(req, "id", user.ID.String())

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(UpdateUser))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp userResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Email != "planner@example.com" {
		t.Fatalf("expected updated email, got %s", resp.Email)
	}
}

func TestDeleteUser(t *testing.T) {
	orgID := uuid.New()
	userRepo := newFakeUserRepo()
	userService := &services.UserService{Users: userRepo}
	registry := middleware.ServiceRegistry{Users: userService}

	user := domain.User{
		ID:           uuid.New(),
		OrgID:        orgID,
		Email:        "tech@example.com",
		Role:         domain.RoleScheduler,
		PasswordHash: "hash",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	_, _ = userRepo.Create(context.Background(), user)

	req := newJSONRequest(t, http.MethodDelete, "/api/v1/users/"+user.ID.String(), nil)
	req = withPrincipal(req, orgID, domain.RoleTenantAdmin)
	req = withRouteParam(req, "id", user.ID.String())

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(DeleteUser))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	userRepo.mu.Lock()
	deleted := userRepo.users[user.ID]
	userRepo.mu.Unlock()
	if deleted.DeletedAt == nil {
		t.Fatalf("expected deleted_at to be set")
	}
}
