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

func TestCreateProgram(t *testing.T) {
	orgID := uuid.New()
	programRepo := newFakeProgramRepo()
	programService := &services.MaintenanceProgramService{Programs: programRepo}
	registry := middleware.ServiceRegistry{Programs: programService}

	req := newJSONRequest(t, http.MethodPost, "/api/v1/maintenance-programs", map[string]any{
		"name":           "A-Check",
		"interval_type":  string(domain.ProgramIntervalCalendar),
		"interval_value": 30,
	})
	req = withPrincipal(req, orgID, domain.RoleScheduler)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreateProgram))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	var resp programResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Name != "A-Check" {
		t.Fatalf("expected name A-Check, got %s", resp.Name)
	}
}

func TestCreateProgramInvalidInterval(t *testing.T) {
	orgID := uuid.New()
	programRepo := newFakeProgramRepo()
	programService := &services.MaintenanceProgramService{Programs: programRepo}
	registry := middleware.ServiceRegistry{Programs: programService}

	req := newJSONRequest(t, http.MethodPost, "/api/v1/maintenance-programs", map[string]any{
		"name":           "A-Check",
		"interval_type":  string(domain.ProgramIntervalCalendar),
		"interval_value": 0,
	})
	req = withPrincipal(req, orgID, domain.RoleScheduler)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreateProgram))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestListPrograms(t *testing.T) {
	orgID := uuid.New()
	programRepo := newFakeProgramRepo()
	programService := &services.MaintenanceProgramService{Programs: programRepo}
	registry := middleware.ServiceRegistry{Programs: programService}

	program := domain.MaintenanceProgram{
		ID:            uuid.New(),
		OrgID:         orgID,
		Name:          "A-Check",
		IntervalType:  domain.ProgramIntervalCalendar,
		IntervalValue: 30,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	_, _ = programRepo.Create(context.Background(), program)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/maintenance-programs", nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(ListPrograms))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp []programResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 program, got %d", len(resp))
	}
	if resp[0].ID != program.ID {
		t.Fatalf("expected program id %s, got %s", program.ID, resp[0].ID)
	}
}

func TestGetProgram(t *testing.T) {
	orgID := uuid.New()
	programRepo := newFakeProgramRepo()
	programService := &services.MaintenanceProgramService{Programs: programRepo}
	registry := middleware.ServiceRegistry{Programs: programService}

	program := domain.MaintenanceProgram{
		ID:            uuid.New(),
		OrgID:         orgID,
		Name:          "A-Check",
		IntervalType:  domain.ProgramIntervalCalendar,
		IntervalValue: 30,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	_, _ = programRepo.Create(context.Background(), program)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/maintenance-programs/"+program.ID.String(), nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", program.ID.String())

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(GetProgram))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp programResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != program.ID {
		t.Fatalf("expected program id %s, got %s", program.ID, resp.ID)
	}
}

func TestListProgramsUnauthorized(t *testing.T) {
	req := newJSONRequest(t, http.MethodGet, "/api/v1/maintenance-programs", nil)
	rr := httptest.NewRecorder()

	ListPrograms(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestUpdateProgram(t *testing.T) {
	orgID := uuid.New()
	programRepo := newFakeProgramRepo()
	programService := &services.MaintenanceProgramService{Programs: programRepo}
	registry := middleware.ServiceRegistry{Programs: programService}

	program := domain.MaintenanceProgram{
		ID:            uuid.New(),
		OrgID:         orgID,
		Name:          "A-Check",
		IntervalType:  domain.ProgramIntervalCalendar,
		IntervalValue: 30,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	_, _ = programRepo.Create(context.Background(), program)

	req := newJSONRequest(t, http.MethodPatch, "/api/v1/maintenance-programs/"+program.ID.String(), map[string]any{
		"name": "B-Check",
	})
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", program.ID.String())

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(UpdateProgram))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp programResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Name != "B-Check" {
		t.Fatalf("expected name B-Check, got %s", resp.Name)
	}
}

func TestDeleteProgram(t *testing.T) {
	orgID := uuid.New()
	programRepo := newFakeProgramRepo()
	programService := &services.MaintenanceProgramService{Programs: programRepo}
	registry := middleware.ServiceRegistry{Programs: programService}

	program := domain.MaintenanceProgram{
		ID:            uuid.New(),
		OrgID:         orgID,
		Name:          "A-Check",
		IntervalType:  domain.ProgramIntervalCalendar,
		IntervalValue: 30,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	_, _ = programRepo.Create(context.Background(), program)

	req := newJSONRequest(t, http.MethodDelete, "/api/v1/maintenance-programs/"+program.ID.String(), nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", program.ID.String())

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(DeleteProgram))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	programRepo.mu.Lock()
	deleted := programRepo.programs[program.ID]
	programRepo.mu.Unlock()
	if deleted.DeletedAt == nil {
		t.Fatalf("expected deleted_at to be set")
	}
}
