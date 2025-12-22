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

func TestCreateComplianceItem(t *testing.T) {
	orgID := uuid.New()
	repo := newFakeComplianceRepo()
	service := &services.ComplianceService{Compliance: repo}
	registry := middleware.ServiceRegistry{Compliance: service}

	req := newJSONRequest(t, http.MethodPost, "/api/v1/compliance-items", map[string]any{
		"task_id":     uuid.New().String(),
		"description": "Check torque",
		"result":      string(domain.CompliancePass),
	})
	req = withPrincipal(req, orgID, domain.RoleMechanic)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreateComplianceItem))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	var resp complianceResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Result != domain.CompliancePass {
		t.Fatalf("expected result pass, got %s", resp.Result)
	}
}

func TestListComplianceItemsUnauthorized(t *testing.T) {
	req := newJSONRequest(t, http.MethodGet, "/api/v1/compliance-items", nil)
	rr := httptest.NewRecorder()

	ListComplianceItems(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestCreateComplianceItemForbidden(t *testing.T) {
	orgID := uuid.New()
	repo := newFakeComplianceRepo()
	service := &services.ComplianceService{Compliance: repo}
	registry := middleware.ServiceRegistry{Compliance: service}

	req := newJSONRequest(t, http.MethodPost, "/api/v1/compliance-items", map[string]any{
		"task_id":     uuid.New().String(),
		"description": "Check torque",
		"result":      string(domain.CompliancePass),
	})
	req = withPrincipal(req, orgID, domain.RoleScheduler)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreateComplianceItem))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rr.Code)
	}
}

func TestListComplianceItems(t *testing.T) {
	orgID := uuid.New()
	repo := newFakeComplianceRepo()
	service := &services.ComplianceService{Compliance: repo}
	registry := middleware.ServiceRegistry{Compliance: service}

	item := domain.ComplianceItem{
		ID:          uuid.New(),
		OrgID:       orgID,
		TaskID:      uuid.New(),
		Description: "Check torque",
		Result:      domain.CompliancePass,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	_ = repo.Create(context.Background(), item)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/compliance-items", nil)
	req = withPrincipal(req, orgID, domain.RoleMechanic)
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(ListComplianceItems))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp []complianceResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp))
	}
	if resp[0].ID != item.ID {
		t.Fatalf("expected item id %s, got %s", item.ID, resp[0].ID)
	}
}

func TestUpdateComplianceItem(t *testing.T) {
	orgID := uuid.New()
	repo := newFakeComplianceRepo()
	service := &services.ComplianceService{Compliance: repo}
	registry := middleware.ServiceRegistry{Compliance: service}

	item := domain.ComplianceItem{
		ID:          uuid.New(),
		OrgID:       orgID,
		TaskID:      uuid.New(),
		Description: "Check torque",
		Result:      domain.CompliancePending,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	_ = repo.Create(context.Background(), item)

	req := newJSONRequest(t, http.MethodPatch, "/api/v1/compliance-items/"+item.ID.String(), map[string]any{
		"description": "Check torque updated",
		"result":      string(domain.CompliancePass),
	})
	req = withPrincipal(req, orgID, domain.RoleMechanic)
	req = withRouteParam(req, "id", item.ID.String())
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(UpdateComplianceItem))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp complianceResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Description != "Check torque updated" {
		t.Fatalf("expected updated description, got %s", resp.Description)
	}
}

func TestSignOffComplianceItem(t *testing.T) {
	orgID := uuid.New()
	repo := newFakeComplianceRepo()
	service := &services.ComplianceService{Compliance: repo}
	registry := middleware.ServiceRegistry{Compliance: service}

	item := domain.ComplianceItem{
		ID:          uuid.New(),
		OrgID:       orgID,
		TaskID:      uuid.New(),
		Description: "Check torque",
		Result:      domain.CompliancePass,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	_ = repo.Create(context.Background(), item)

	req := newJSONRequest(t, http.MethodPatch, "/api/v1/compliance-items/"+item.ID.String()+"/sign-off", nil)
	req = withPrincipal(req, orgID, domain.RoleMechanic)
	req = withRouteParam(req, "id", item.ID.String())
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(SignOffComplianceItem))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp complianceResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.SignedOff {
		t.Fatalf("expected signed_off true")
	}
}
