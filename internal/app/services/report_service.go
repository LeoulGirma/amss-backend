package services

import (
	"context"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type ReportService struct {
	Reports ports.ReportRepository
}

func (s *ReportService) Summary(ctx context.Context, actor app.Actor, orgID uuid.UUID) (ports.ReportSummary, error) {
	if s.Reports == nil {
		return ports.ReportSummary{}, domain.NewValidationError("report repository unavailable")
	}
	if !actor.IsAdmin() && actor.OrgID != orgID {
		return ports.ReportSummary{}, domain.ErrForbidden
	}
	if orgID == uuid.Nil {
		return ports.ReportSummary{}, domain.NewValidationError("org_id is required")
	}
	return s.Reports.Summary(ctx, orgID)
}

func (s *ReportService) Compliance(ctx context.Context, actor app.Actor, filter ports.ComplianceReportFilter) (ports.ComplianceReport, error) {
	if s.Reports == nil {
		return ports.ComplianceReport{}, domain.NewValidationError("report repository unavailable")
	}
	if actor.Role != domain.RoleAuditor && actor.Role != domain.RoleAdmin {
		return ports.ComplianceReport{}, domain.ErrForbidden
	}
	if !actor.IsAdmin() && actor.OrgID != filter.OrgID {
		return ports.ComplianceReport{}, domain.ErrForbidden
	}
	if filter.OrgID == uuid.Nil {
		return ports.ComplianceReport{}, domain.NewValidationError("org_id is required")
	}
	return s.Reports.Compliance(ctx, filter)
}
