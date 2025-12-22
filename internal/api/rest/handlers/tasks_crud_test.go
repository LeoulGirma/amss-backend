package handlers

import (
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

func TestCreateTask(t *testing.T) {
	orgID := uuid.New()
	taskRepo := newFakeTaskRepo()
	taskService := &services.TaskService{Tasks: taskRepo}
	registry := middleware.ServiceRegistry{Tasks: taskService}

	start := time.Now().Add(1 * time.Hour).UTC()
	end := start.Add(2 * time.Hour)
	req := newJSONRequest(t, http.MethodPost, "/api/v1/maintenance-tasks", map[string]any{
		"aircraft_id": uuid.New().String(),
		"type":        string(domain.TaskTypeInspection),
		"start_time":  start.Format(time.RFC3339),
		"end_time":    end.Format(time.RFC3339),
	})
	req = withPrincipal(req, orgID, domain.RoleScheduler)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreateTask))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	var resp taskResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.OrgID != orgID {
		t.Fatalf("expected org_id %s, got %s", orgID, resp.OrgID)
	}
	if resp.Type != domain.TaskTypeInspection {
		t.Fatalf("expected type inspection, got %s", resp.Type)
	}
}

func TestListTasksUnauthorized(t *testing.T) {
	req := newJSONRequest(t, http.MethodGet, "/api/v1/maintenance-tasks", nil)
	rr := httptest.NewRecorder()

	ListTasks(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestCreateTaskForbidden(t *testing.T) {
	orgID := uuid.New()
	taskRepo := newFakeTaskRepo()
	taskService := &services.TaskService{Tasks: taskRepo}
	registry := middleware.ServiceRegistry{Tasks: taskService}

	start := time.Now().Add(1 * time.Hour).UTC()
	end := start.Add(2 * time.Hour)
	req := newJSONRequest(t, http.MethodPost, "/api/v1/maintenance-tasks", map[string]any{
		"aircraft_id": uuid.New().String(),
		"type":        string(domain.TaskTypeInspection),
		"start_time":  start.Format(time.RFC3339),
		"end_time":    end.Format(time.RFC3339),
	})
	req = withPrincipal(req, orgID, domain.RoleMechanic)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreateTask))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rr.Code)
	}
}

func TestListTasks(t *testing.T) {
	orgID := uuid.New()
	taskRepo := newFakeTaskRepo()
	taskService := &services.TaskService{Tasks: taskRepo}
	registry := middleware.ServiceRegistry{Tasks: taskService}

	task := domain.MaintenanceTask{
		ID:         uuid.New(),
		OrgID:      orgID,
		AircraftID: uuid.New(),
		Type:       domain.TaskTypeRepair,
		State:      domain.TaskStateScheduled,
		StartTime:  time.Now().UTC(),
		EndTime:    time.Now().Add(1 * time.Hour).UTC(),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	_, _ = taskRepo.Create(context.Background(), task)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/maintenance-tasks", nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(ListTasks))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp []taskResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 task, got %d", len(resp))
	}
	if resp[0].ID != task.ID {
		t.Fatalf("expected task id %s, got %s", task.ID, resp[0].ID)
	}
}

func TestGetTask(t *testing.T) {
	orgID := uuid.New()
	taskRepo := newFakeTaskRepo()
	taskService := &services.TaskService{Tasks: taskRepo}
	registry := middleware.ServiceRegistry{Tasks: taskService}

	task := domain.MaintenanceTask{
		ID:         uuid.New(),
		OrgID:      orgID,
		AircraftID: uuid.New(),
		Type:       domain.TaskTypeRepair,
		State:      domain.TaskStateScheduled,
		StartTime:  time.Now().UTC(),
		EndTime:    time.Now().Add(1 * time.Hour).UTC(),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	_, _ = taskRepo.Create(context.Background(), task)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/maintenance-tasks/"+task.ID.String(), nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", task.ID.String())
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(GetTask))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp taskDetailResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Task.ID != task.ID {
		t.Fatalf("expected task id %s, got %s", task.ID, resp.Task.ID)
	}
}

func TestUpdateTask(t *testing.T) {
	orgID := uuid.New()
	taskRepo := newFakeTaskRepo()
	taskService := &services.TaskService{Tasks: taskRepo}
	registry := middleware.ServiceRegistry{Tasks: taskService}

	task := domain.MaintenanceTask{
		ID:         uuid.New(),
		OrgID:      orgID,
		AircraftID: uuid.New(),
		Type:       domain.TaskTypeInspection,
		State:      domain.TaskStateScheduled,
		StartTime:  time.Now().UTC(),
		EndTime:    time.Now().Add(1 * time.Hour).UTC(),
		Notes:      "before",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	_, _ = taskRepo.Create(context.Background(), task)

	req := newJSONRequest(t, http.MethodPatch, "/api/v1/maintenance-tasks/"+task.ID.String(), map[string]any{
		"notes": "after",
	})
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", task.ID.String())
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(UpdateTask))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp taskResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Notes != "after" {
		t.Fatalf("expected updated notes, got %s", resp.Notes)
	}
}

func TestDeleteTask(t *testing.T) {
	orgID := uuid.New()
	taskRepo := newFakeTaskRepo()
	taskService := &services.TaskService{Tasks: taskRepo}
	registry := middleware.ServiceRegistry{Tasks: taskService}

	task := domain.MaintenanceTask{
		ID:         uuid.New(),
		OrgID:      orgID,
		AircraftID: uuid.New(),
		Type:       domain.TaskTypeInspection,
		State:      domain.TaskStateScheduled,
		StartTime:  time.Now().UTC(),
		EndTime:    time.Now().Add(1 * time.Hour).UTC(),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	_, _ = taskRepo.Create(context.Background(), task)

	req := newJSONRequest(t, http.MethodDelete, "/api/v1/maintenance-tasks/"+task.ID.String(), nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", task.ID.String())
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(DeleteTask))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	taskRepo.mu.Lock()
	updated := taskRepo.tasks[task.ID]
	taskRepo.mu.Unlock()
	if updated.DeletedAt == nil {
		t.Fatalf("expected deleted_at to be set")
	}
}
