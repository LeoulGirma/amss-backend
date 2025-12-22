package handlers

import (
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aeromaintain/amss/internal/api/rest/middleware"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

func TestExportAuditLogs(t *testing.T) {
	orgID := uuid.New()
	entry := domain.AuditLog{
		ID:            uuid.New(),
		OrgID:         orgID,
		EntityType:    "maintenance_task",
		EntityID:      uuid.New(),
		Action:        domain.AuditActionCreate,
		UserID:        uuid.New(),
		RequestID:     uuid.New(),
		EntityVersion: 1,
		Timestamp:     time.Now().UTC(),
		Details: map[string]any{
			"state": "scheduled",
		},
	}
	repo := &fakeAuditQueryRepo{entries: []domain.AuditLog{entry}}
	service := &services.AuditQueryService{Repo: repo}
	registry := middleware.ServiceRegistry{AuditQuery: service}

	req := newJSONRequest(t, http.MethodGet, "/api/v1/audit-logs/export", nil)
	req = withPrincipal(req, orgID, domain.RoleAdmin)
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(ExportAuditLogs))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	reader := csv.NewReader(strings.NewReader(rr.Body.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if len(records) < 2 {
		t.Fatalf("expected csv header and one row, got %d rows", len(records))
	}
	if records[0][0] != "id" {
		t.Fatalf("expected header id, got %s", records[0][0])
	}
	if records[1][0] != entry.ID.String() {
		t.Fatalf("expected entry id %s, got %s", entry.ID, records[1][0])
	}
}

func TestListAuditLogsUnauthorized(t *testing.T) {
	req := newJSONRequest(t, http.MethodGet, "/api/v1/audit-logs", nil)
	rr := httptest.NewRecorder()

	ListAuditLogs(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestExportAuditLogsForbidden(t *testing.T) {
	req := newJSONRequest(t, http.MethodGet, "/api/v1/audit-logs/export", nil)
	req = withPrincipal(req, uuid.New(), domain.RoleScheduler)
	rr := httptest.NewRecorder()

	ExportAuditLogs(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rr.Code)
	}
}
