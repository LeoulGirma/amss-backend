package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aeromaintain/amss/internal/api/rest/middleware"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

func TestGetReportSummary(t *testing.T) {
	orgID := uuid.New()
	repo := &fakeReportRepo{
		summary: ports.ReportSummary{
			TasksScheduled:    2,
			TasksInProgress:   1,
			TasksCompleted:    3,
			TasksCancelled:    1,
			AircraftTotal:     5,
			PartsInStock:      7,
			PartsUsed:         4,
			PartsDisposed:     2,
			CompliancePending: 6,
			ComplianceSigned:  9,
		},
	}
	service := &services.ReportService{Reports: repo}
	registry := middleware.ServiceRegistry{Reports: service}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/summary?org_id="+orgID.String(), nil)
	req = withPrincipal(req, uuid.New(), domain.RoleAdmin)
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(GetReportSummary))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp reportSummaryResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Tasks.Scheduled != 2 || resp.Tasks.InProgress != 1 || resp.Tasks.Completed != 3 || resp.Tasks.Cancelled != 1 {
		t.Fatalf("unexpected task summary: %+v", resp.Tasks)
	}
	if resp.Aircraft.Total != 5 {
		t.Fatalf("unexpected aircraft total %d", resp.Aircraft.Total)
	}
	if resp.Parts.InStock != 7 || resp.Parts.Used != 4 || resp.Parts.Disposed != 2 {
		t.Fatalf("unexpected parts summary: %+v", resp.Parts)
	}
	if resp.Compliance.Pending != 6 || resp.Compliance.Signed != 9 {
		t.Fatalf("unexpected compliance summary: %+v", resp.Compliance)
	}
	if repo.lastSummaryOrg != orgID {
		t.Fatalf("expected summary org_id %s, got %s", orgID, repo.lastSummaryOrg)
	}
}

func TestGetComplianceReport(t *testing.T) {
	orgID := uuid.New()
	repo := &fakeReportRepo{
		compliance: ports.ComplianceReport{
			Total:    10,
			Pass:     6,
			Fail:     2,
			Pending:  2,
			Signed:   8,
			Unsigned: 2,
		},
	}
	service := &services.ReportService{Reports: repo}
	registry := middleware.ServiceRegistry{Reports: service}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/compliance?result=pass&signed=true", nil)
	req = withPrincipal(req, orgID, domain.RoleAuditor)
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(GetComplianceReport))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp reportComplianceResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Total != 10 || resp.Pass != 6 || resp.Fail != 2 || resp.Pending != 2 || resp.Signed != 8 || resp.Unsigned != 2 {
		t.Fatalf("unexpected compliance report: %+v", resp)
	}
	if repo.lastComplianceFilter.OrgID != orgID {
		t.Fatalf("expected org_id %s, got %s", orgID, repo.lastComplianceFilter.OrgID)
	}
	if repo.lastComplianceFilter.Result == nil || *repo.lastComplianceFilter.Result != domain.CompliancePass {
		t.Fatalf("expected result filter pass")
	}
	if repo.lastComplianceFilter.Signed == nil || *repo.lastComplianceFilter.Signed != true {
		t.Fatalf("expected signed filter true")
	}
}

func TestGetComplianceReportForbidden(t *testing.T) {
	orgID := uuid.New()
	repo := &fakeReportRepo{}
	service := &services.ReportService{Reports: repo}
	registry := middleware.ServiceRegistry{Reports: service}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/compliance", nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(GetComplianceReport))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rr.Code)
	}
}

func TestGetReportSummaryInvalidOrgID(t *testing.T) {
	repo := &fakeReportRepo{}
	service := &services.ReportService{Reports: repo}
	registry := middleware.ServiceRegistry{Reports: service}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/summary?org_id=invalid", nil)
	req = withPrincipal(req, uuid.New(), domain.RoleAdmin)
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(GetReportSummary))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestGetComplianceReportInvalidResult(t *testing.T) {
	orgID := uuid.New()
	repo := &fakeReportRepo{}
	service := &services.ReportService{Reports: repo}
	registry := middleware.ServiceRegistry{Reports: service}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/compliance?result=bad", nil)
	req = withPrincipal(req, orgID, domain.RoleAuditor)
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(GetComplianceReport))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestGetComplianceReportInvalidSigned(t *testing.T) {
	orgID := uuid.New()
	repo := &fakeReportRepo{}
	service := &services.ReportService{Reports: repo}
	registry := middleware.ServiceRegistry{Reports: service}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/compliance?signed=maybe", nil)
	req = withPrincipal(req, orgID, domain.RoleAuditor)
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(GetComplianceReport))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestGetComplianceReportAdminOrgSwitching(t *testing.T) {
	adminOrg := uuid.New()
	targetOrg := uuid.New()
	repo := &fakeReportRepo{
		compliance: ports.ComplianceReport{Total: 1},
	}
	service := &services.ReportService{Reports: repo}
	registry := middleware.ServiceRegistry{Reports: service}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/compliance?org_id="+targetOrg.String(), nil)
	req = withPrincipal(req, adminOrg, domain.RoleAdmin)
	rr := httptest.NewRecorder()

	handler := middleware.InjectServices(registry)(http.HandlerFunc(GetComplianceReport))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if repo.lastComplianceFilter.OrgID != targetOrg {
		t.Fatalf("expected org_id %s, got %s", targetOrg, repo.lastComplianceFilter.OrgID)
	}
}
