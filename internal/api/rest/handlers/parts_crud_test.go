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

func TestCreatePartDefinition(t *testing.T) {
	orgID := uuid.New()
	defRepo := newFakePartDefinitionRepo()
	itemRepo := newFakePartItemRepo()
	catalogService := &services.PartCatalogService{Definitions: defRepo, Items: itemRepo}
	registry := middleware.ServiceRegistry{Catalog: catalogService}

	req := newJSONRequest(t, http.MethodPost, "/api/v1/part-definitions", map[string]any{
		"name":     "Filter",
		"category": "consumable",
	})
	req = withPrincipal(req, orgID, domain.RoleScheduler)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreatePartDefinition))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	var resp partDefinitionResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Name != "Filter" {
		t.Fatalf("expected name Filter, got %s", resp.Name)
	}
}

func TestCreatePartDefinitionUnauthorized(t *testing.T) {
	req := newJSONRequest(t, http.MethodPost, "/api/v1/part-definitions", map[string]any{
		"name":     "Filter",
		"category": "consumable",
	})

	rr := httptest.NewRecorder()
	CreatePartDefinition(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestCreatePartDefinitionForbidden(t *testing.T) {
	orgID := uuid.New()
	defRepo := newFakePartDefinitionRepo()
	itemRepo := newFakePartItemRepo()
	catalogService := &services.PartCatalogService{Definitions: defRepo, Items: itemRepo}
	registry := middleware.ServiceRegistry{Catalog: catalogService}

	req := newJSONRequest(t, http.MethodPost, "/api/v1/part-definitions", map[string]any{
		"name":     "Filter",
		"category": "consumable",
	})
	req = withPrincipal(req, orgID, domain.RoleAuditor)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreatePartDefinition))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rr.Code)
	}
}

func TestListPartDefinitions(t *testing.T) {
	orgID := uuid.New()
	defRepo := newFakePartDefinitionRepo()
	itemRepo := newFakePartItemRepo()
	catalogService := &services.PartCatalogService{Definitions: defRepo, Items: itemRepo}
	registry := middleware.ServiceRegistry{Catalog: catalogService}

	def := domain.PartDefinition{
		ID:        uuid.New(),
		OrgID:     orgID,
		Name:      "Filter",
		Category:  "consumable",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_, _ = defRepo.Create(context.Background(), def)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/part-definitions", nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(ListPartDefinitions))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp []partDefinitionResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(resp))
	}
	if resp[0].ID != def.ID {
		t.Fatalf("expected definition id %s, got %s", def.ID, resp[0].ID)
	}
}

func TestUpdatePartDefinition(t *testing.T) {
	orgID := uuid.New()
	defRepo := newFakePartDefinitionRepo()
	itemRepo := newFakePartItemRepo()
	catalogService := &services.PartCatalogService{Definitions: defRepo, Items: itemRepo}
	registry := middleware.ServiceRegistry{Catalog: catalogService}

	def := domain.PartDefinition{
		ID:        uuid.New(),
		OrgID:     orgID,
		Name:      "Filter",
		Category:  "consumable",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_, _ = defRepo.Create(context.Background(), def)

	req := newJSONRequest(t, http.MethodPatch, "/api/v1/part-definitions/"+def.ID.String(), map[string]any{
		"name":     "Filter-2",
		"category": "consumable",
	})
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", def.ID.String())
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(UpdatePartDefinition))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp partDefinitionResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Name != "Filter-2" {
		t.Fatalf("expected updated name, got %s", resp.Name)
	}
}

func TestDeletePartDefinition(t *testing.T) {
	orgID := uuid.New()
	defRepo := newFakePartDefinitionRepo()
	itemRepo := newFakePartItemRepo()
	catalogService := &services.PartCatalogService{Definitions: defRepo, Items: itemRepo}
	registry := middleware.ServiceRegistry{Catalog: catalogService}

	def := domain.PartDefinition{
		ID:        uuid.New(),
		OrgID:     orgID,
		Name:      "Filter",
		Category:  "consumable",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_, _ = defRepo.Create(context.Background(), def)

	req := newJSONRequest(t, http.MethodDelete, "/api/v1/part-definitions/"+def.ID.String(), nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", def.ID.String())
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(DeletePartDefinition))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	defRepo.mu.Lock()
	updated := defRepo.defs[def.ID]
	defRepo.mu.Unlock()
	if updated.DeletedAt == nil {
		t.Fatalf("expected deleted_at to be set")
	}
}

func TestCreatePartItem(t *testing.T) {
	orgID := uuid.New()
	defRepo := newFakePartDefinitionRepo()
	itemRepo := newFakePartItemRepo()
	catalogService := &services.PartCatalogService{Definitions: defRepo, Items: itemRepo}
	registry := middleware.ServiceRegistry{Catalog: catalogService}

	defID := uuid.New()
	req := newJSONRequest(t, http.MethodPost, "/api/v1/part-items", map[string]any{
		"part_definition_id": defID.String(),
		"serial_number":      "SN-123",
		"status":             string(domain.PartItemInStock),
	})
	req = withPrincipal(req, orgID, domain.RoleScheduler)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreatePartItem))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	var resp partItemResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.PartDefinitionID != defID {
		t.Fatalf("expected definition id %s, got %s", defID, resp.PartDefinitionID)
	}
}

func TestListPartItems(t *testing.T) {
	orgID := uuid.New()
	defRepo := newFakePartDefinitionRepo()
	itemRepo := newFakePartItemRepo()
	catalogService := &services.PartCatalogService{Definitions: defRepo, Items: itemRepo}
	registry := middleware.ServiceRegistry{Catalog: catalogService}

	item := domain.PartItem{
		ID:           uuid.New(),
		OrgID:        orgID,
		DefinitionID: uuid.New(),
		SerialNumber: "SN-123",
		Status:       domain.PartItemInStock,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	_, _ = itemRepo.Create(context.Background(), item)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/part-items", nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(ListPartItems))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp []partItemResponse
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

func TestUpdatePartItem(t *testing.T) {
	orgID := uuid.New()
	defRepo := newFakePartDefinitionRepo()
	itemRepo := newFakePartItemRepo()
	catalogService := &services.PartCatalogService{Definitions: defRepo, Items: itemRepo}
	registry := middleware.ServiceRegistry{Catalog: catalogService}

	item := domain.PartItem{
		ID:           uuid.New(),
		OrgID:        orgID,
		DefinitionID: uuid.New(),
		SerialNumber: "SN-123",
		Status:       domain.PartItemInStock,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	_, _ = itemRepo.Create(context.Background(), item)

	req := newJSONRequest(t, http.MethodPatch, "/api/v1/part-items/"+item.ID.String(), map[string]any{
		"status": string(domain.PartItemUsed),
	})
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", item.ID.String())
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(UpdatePartItem))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp partItemResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != domain.PartItemUsed {
		t.Fatalf("expected status used, got %s", resp.Status)
	}
}

func TestDeletePartItem(t *testing.T) {
	orgID := uuid.New()
	defRepo := newFakePartDefinitionRepo()
	itemRepo := newFakePartItemRepo()
	catalogService := &services.PartCatalogService{Definitions: defRepo, Items: itemRepo}
	registry := middleware.ServiceRegistry{Catalog: catalogService}

	item := domain.PartItem{
		ID:           uuid.New(),
		OrgID:        orgID,
		DefinitionID: uuid.New(),
		SerialNumber: "SN-123",
		Status:       domain.PartItemInStock,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	_, _ = itemRepo.Create(context.Background(), item)

	req := newJSONRequest(t, http.MethodDelete, "/api/v1/part-items/"+item.ID.String(), nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", item.ID.String())
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(DeletePartItem))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	itemRepo.mu.Lock()
	updated := itemRepo.items[item.ID]
	itemRepo.mu.Unlock()
	if updated.DeletedAt == nil {
		t.Fatalf("expected deleted_at to be set")
	}
}
