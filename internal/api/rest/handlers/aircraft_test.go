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

func TestCreateAircraft(t *testing.T) {
	orgID := uuid.New()
	aircraftRepo := newFakeAircraftRepo()
	aircraftService := &services.AircraftService{Aircraft: aircraftRepo}
	registry := middleware.ServiceRegistry{Aircraft: aircraftService}

	lastMaintenance := time.Now().Add(-24 * time.Hour).UTC()
	nextDue := time.Now().Add(24 * time.Hour).UTC()
	req := newJSONRequest(t, http.MethodPost, "/api/v1/aircraft", map[string]any{
		"tail_number":        "N123AM",
		"model":              "A320",
		"last_maintenance":   lastMaintenance.Format(time.RFC3339),
		"next_due":           nextDue.Format(time.RFC3339),
		"status":             string(domain.AircraftOperational),
		"capacity_slots":     3,
		"flight_hours_total": 100,
		"cycles_total":       10,
	})
	req = withPrincipal(req, orgID, domain.RoleScheduler)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreateAircraft))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	var resp aircraftResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.TailNumber != "N123AM" {
		t.Fatalf("expected tail number N123AM, got %s", resp.TailNumber)
	}
}

func TestCreateAircraftInvalidTime(t *testing.T) {
	orgID := uuid.New()
	aircraftRepo := newFakeAircraftRepo()
	aircraftService := &services.AircraftService{Aircraft: aircraftRepo}
	registry := middleware.ServiceRegistry{Aircraft: aircraftService}

	req := newJSONRequest(t, http.MethodPost, "/api/v1/aircraft", map[string]any{
		"tail_number":        "N123AM",
		"model":              "A320",
		"last_maintenance":   "not-a-date",
		"capacity_slots":     3,
		"flight_hours_total": 100,
		"cycles_total":       10,
	})
	req = withPrincipal(req, orgID, domain.RoleScheduler)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreateAircraft))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestListAircraft(t *testing.T) {
	orgID := uuid.New()
	aircraftRepo := newFakeAircraftRepo()
	aircraftService := &services.AircraftService{Aircraft: aircraftRepo}
	registry := middleware.ServiceRegistry{Aircraft: aircraftService}

	aircraft := domain.Aircraft{
		ID:            uuid.New(),
		OrgID:         orgID,
		TailNumber:    "N123AM",
		Model:         "A320",
		Status:        domain.AircraftOperational,
		CapacitySlots: 3,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	_, _ = aircraftRepo.Create(context.Background(), aircraft)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/aircraft", nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(ListAircraft))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp []aircraftResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 aircraft, got %d", len(resp))
	}
	if resp[0].ID != aircraft.ID {
		t.Fatalf("expected aircraft id %s, got %s", aircraft.ID, resp[0].ID)
	}
}

func TestGetAircraft(t *testing.T) {
	orgID := uuid.New()
	aircraftRepo := newFakeAircraftRepo()
	aircraftService := &services.AircraftService{Aircraft: aircraftRepo}
	registry := middleware.ServiceRegistry{Aircraft: aircraftService}

	aircraft := domain.Aircraft{
		ID:            uuid.New(),
		OrgID:         orgID,
		TailNumber:    "N123AM",
		Model:         "A320",
		Status:        domain.AircraftOperational,
		CapacitySlots: 3,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	_, _ = aircraftRepo.Create(context.Background(), aircraft)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/aircraft/"+aircraft.ID.String(), nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", aircraft.ID.String())

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(GetAircraft))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp aircraftResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != aircraft.ID {
		t.Fatalf("expected aircraft id %s, got %s", aircraft.ID, resp.ID)
	}
}

func TestListAircraftUnauthorized(t *testing.T) {
	req := newJSONRequest(t, http.MethodGet, "/api/v1/aircraft", nil)
	rr := httptest.NewRecorder()

	ListAircraft(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestUpdateAircraft(t *testing.T) {
	orgID := uuid.New()
	aircraftRepo := newFakeAircraftRepo()
	aircraftService := &services.AircraftService{Aircraft: aircraftRepo}
	registry := middleware.ServiceRegistry{Aircraft: aircraftService}

	aircraft := domain.Aircraft{
		ID:            uuid.New(),
		OrgID:         orgID,
		TailNumber:    "N123AM",
		Model:         "A320",
		Status:        domain.AircraftOperational,
		CapacitySlots: 3,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	_, _ = aircraftRepo.Create(context.Background(), aircraft)

	req := newJSONRequest(t, http.MethodPatch, "/api/v1/aircraft/"+aircraft.ID.String(), map[string]any{
		"status": string(domain.AircraftMaintenance),
	})
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", aircraft.ID.String())

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(UpdateAircraft))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp aircraftResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != domain.AircraftMaintenance {
		t.Fatalf("expected status maintenance, got %s", resp.Status)
	}
}

func TestDeleteAircraft(t *testing.T) {
	orgID := uuid.New()
	aircraftRepo := newFakeAircraftRepo()
	aircraftService := &services.AircraftService{Aircraft: aircraftRepo}
	registry := middleware.ServiceRegistry{Aircraft: aircraftService}

	aircraft := domain.Aircraft{
		ID:            uuid.New(),
		OrgID:         orgID,
		TailNumber:    "N123AM",
		Model:         "A320",
		Status:        domain.AircraftOperational,
		CapacitySlots: 3,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	_, _ = aircraftRepo.Create(context.Background(), aircraft)

	req := newJSONRequest(t, http.MethodDelete, "/api/v1/aircraft/"+aircraft.ID.String(), nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", aircraft.ID.String())

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(DeleteAircraft))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	aircraftRepo.mu.Lock()
	deleted := aircraftRepo.aircraft[aircraft.ID]
	aircraftRepo.mu.Unlock()
	if deleted.DeletedAt == nil {
		t.Fatalf("expected deleted_at to be set")
	}
}
