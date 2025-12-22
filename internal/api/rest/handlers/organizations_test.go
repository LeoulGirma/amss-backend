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

func TestCreateOrganization(t *testing.T) {
	orgRepo := newFakeOrganizationRepo()
	orgService := &services.OrganizationService{Organizations: orgRepo}
	registry := middleware.ServiceRegistry{Organizations: orgService}

	req := newJSONRequest(t, http.MethodPost, "/api/v1/organizations", map[string]any{
		"name": "Sky Ops",
	})
	req = withPrincipal(req, uuid.New(), domain.RoleAdmin)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreateOrganization))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	var resp organizationResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Name != "Sky Ops" {
		t.Fatalf("expected name Sky Ops, got %s", resp.Name)
	}
}

func TestCreateOrganizationUnauthorized(t *testing.T) {
	req := newJSONRequest(t, http.MethodPost, "/api/v1/organizations", map[string]any{
		"name": "Sky Ops",
	})

	rr := httptest.NewRecorder()
	CreateOrganization(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestCreateOrganizationForbidden(t *testing.T) {
	orgRepo := newFakeOrganizationRepo()
	orgService := &services.OrganizationService{Organizations: orgRepo}
	registry := middleware.ServiceRegistry{Organizations: orgService}

	req := newJSONRequest(t, http.MethodPost, "/api/v1/organizations", map[string]any{
		"name": "Sky Ops",
	})
	req = withPrincipal(req, uuid.New(), domain.RoleScheduler)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreateOrganization))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rr.Code)
	}
}

func TestListOrganizations(t *testing.T) {
	orgRepo := newFakeOrganizationRepo()
	orgService := &services.OrganizationService{Organizations: orgRepo}
	registry := middleware.ServiceRegistry{Organizations: orgService}

	orgID := uuid.New()
	_, _ = orgRepo.Create(context.Background(), domain.Organization{
		ID:        orgID,
		Name:      "Sky Ops",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	req := newJSONRequest(t, http.MethodGet, "/api/v1/organizations", nil)
	req = withPrincipal(req, uuid.New(), domain.RoleAdmin)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(ListOrganizations))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp []organizationResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 organization, got %d", len(resp))
	}
	if resp[0].ID != orgID {
		t.Fatalf("expected org id %s, got %s", orgID, resp[0].ID)
	}
}

func TestListOrganizationsUnauthorized(t *testing.T) {
	req := newJSONRequest(t, http.MethodGet, "/api/v1/organizations", nil)
	rr := httptest.NewRecorder()

	ListOrganizations(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestGetOrganization(t *testing.T) {
	orgRepo := newFakeOrganizationRepo()
	orgService := &services.OrganizationService{Organizations: orgRepo}
	registry := middleware.ServiceRegistry{Organizations: orgService}

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Sky Ops",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_, _ = orgRepo.Create(context.Background(), org)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/organizations/"+org.ID.String(), nil)
	req = withPrincipal(req, org.ID, domain.RoleTenantAdmin)
	req = withRouteParam(req, "id", org.ID.String())

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(GetOrganization))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp organizationResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != org.ID {
		t.Fatalf("expected org id %s, got %s", org.ID, resp.ID)
	}
}

func TestUpdateOrganization(t *testing.T) {
	orgRepo := newFakeOrganizationRepo()
	orgService := &services.OrganizationService{Organizations: orgRepo}
	registry := middleware.ServiceRegistry{Organizations: orgService}

	org := domain.Organization{
		ID:        uuid.New(),
		Name:      "Sky Ops",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_, _ = orgRepo.Create(context.Background(), org)

	req := newJSONRequest(t, http.MethodPatch, "/api/v1/organizations/"+org.ID.String(), map[string]any{
		"name": "Sky Ops West",
	})
	req = withPrincipal(req, org.ID, domain.RoleTenantAdmin)
	req = withRouteParam(req, "id", org.ID.String())

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(UpdateOrganization))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp organizationResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Name != "Sky Ops West" {
		t.Fatalf("expected name Sky Ops West, got %s", resp.Name)
	}
}
