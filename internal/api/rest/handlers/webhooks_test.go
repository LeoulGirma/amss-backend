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
	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

func TestCreateWebhook(t *testing.T) {
	orgID := uuid.New()
	webhookRepo := newFakeWebhookRepo()
	webhookService := &services.WebhookService{Webhooks: webhookRepo}
	registry := middleware.ServiceRegistry{Webhooks: webhookService}

	req := newJSONRequest(t, http.MethodPost, "/api/v1/webhooks", map[string]any{
		"url":    "https://example.com/hook",
		"events": []string{"task.created"},
	})
	req = withPrincipal(req, orgID, domain.RoleTenantAdmin)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreateWebhook))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	var resp webhookResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Secret == "" {
		t.Fatalf("expected secret to be set")
	}
	if resp.URL != "https://example.com/hook" {
		t.Fatalf("expected url https://example.com/hook, got %s", resp.URL)
	}
}

func TestCreateWebhookUnauthorized(t *testing.T) {
	req := newJSONRequest(t, http.MethodPost, "/api/v1/webhooks", map[string]any{
		"url":    "https://example.com/hook",
		"events": []string{"task.created"},
	})

	rr := httptest.NewRecorder()
	CreateWebhook(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestCreateWebhookInvalidJSON(t *testing.T) {
	orgID := uuid.New()
	webhookRepo := newFakeWebhookRepo()
	webhookService := &services.WebhookService{Webhooks: webhookRepo}
	registry := middleware.ServiceRegistry{Webhooks: webhookService}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	req = withPrincipal(req, orgID, domain.RoleTenantAdmin)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(CreateWebhook))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestListWebhooks(t *testing.T) {
	orgID := uuid.New()
	webhookRepo := newFakeWebhookRepo()
	webhookService := &services.WebhookService{Webhooks: webhookRepo}
	registry := middleware.ServiceRegistry{Webhooks: webhookService}

	hook := domain.Webhook{
		ID:        uuid.New(),
		OrgID:     orgID,
		URL:       "https://example.com/hook",
		Events:    []string{"task.created"},
		Secret:    "secret",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_, _ = webhookRepo.Create(context.Background(), hook)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/webhooks", nil)
	req = withPrincipal(req, orgID, domain.RoleTenantAdmin)

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(ListWebhooks))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp []webhookResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(resp))
	}
	if resp[0].ID != hook.ID {
		t.Fatalf("expected webhook id %s, got %s", hook.ID, resp[0].ID)
	}
}

func TestDeleteWebhook(t *testing.T) {
	orgID := uuid.New()
	webhookRepo := newFakeWebhookRepo()
	webhookService := &services.WebhookService{Webhooks: webhookRepo}
	registry := middleware.ServiceRegistry{Webhooks: webhookService}

	hook := domain.Webhook{
		ID:        uuid.New(),
		OrgID:     orgID,
		URL:       "https://example.com/hook",
		Events:    []string{"task.created"},
		Secret:    "secret",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_, _ = webhookRepo.Create(context.Background(), hook)

	req := newJSONRequest(t, http.MethodDelete, "/api/v1/webhooks/"+hook.ID.String(), nil)
	req = withPrincipal(req, orgID, domain.RoleTenantAdmin)
	req = withRouteParam(req, "id", hook.ID.String())

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(DeleteWebhook))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
}

func TestWebhookTestEndpoint(t *testing.T) {
	orgID := uuid.New()
	outbox := &fakeOutboxRepo{}
	webhookService := &services.WebhookService{Outbox: outbox, Clock: app.RealClock{}}
	registry := middleware.ServiceRegistry{Webhooks: webhookService}

	hookID := uuid.New()
	req := newJSONRequest(t, http.MethodPost, "/api/v1/webhooks/"+hookID.String()+"/test", nil)
	req = withPrincipal(req, orgID, domain.RoleTenantAdmin)
	req = withRouteParam(req, "id", hookID.String())

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(TestWebhook))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", rr.Code)
	}
	outbox.mu.Lock()
	count := len(outbox.events)
	outbox.mu.Unlock()
	if count != 1 {
		t.Fatalf("expected 1 outbox event, got %d", count)
	}
}

func TestWebhookTestEndpointOutboxMissing(t *testing.T) {
	orgID := uuid.New()
	webhookService := &services.WebhookService{Clock: app.RealClock{}}
	registry := middleware.ServiceRegistry{Webhooks: webhookService}

	hookID := uuid.New()
	req := newJSONRequest(t, http.MethodPost, "/api/v1/webhooks/"+hookID.String()+"/test", nil)
	req = withPrincipal(req, orgID, domain.RoleTenantAdmin)
	req = withRouteParam(req, "id", hookID.String())

	rr := httptest.NewRecorder()
	handler := middleware.InjectServices(registry)(http.HandlerFunc(TestWebhook))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}
