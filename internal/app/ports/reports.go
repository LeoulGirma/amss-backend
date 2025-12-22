package ports

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type ReportRepository interface {
	Summary(ctx context.Context, orgID uuid.UUID) (ReportSummary, error)
	Compliance(ctx context.Context, filter ComplianceReportFilter) (ComplianceReport, error)
}

type ReportSummary struct {
	TasksScheduled    int
	TasksInProgress   int
	TasksCompleted    int
	TasksCancelled    int
	AircraftTotal     int
	PartsInStock      int
	PartsUsed         int
	PartsDisposed     int
	CompliancePending int
	ComplianceSigned  int
}

type ComplianceReportFilter struct {
	OrgID  uuid.UUID
	TaskID *uuid.UUID
	Result *domain.ComplianceResult
	Signed *bool
	From   *time.Time
	To     *time.Time
}

type ComplianceReport struct {
	Total    int
	Pass     int
	Fail     int
	Pending  int
	Signed   int
	Unsigned int
}
