package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportRepository struct {
	DB *pgxpool.Pool
}

func (r *ReportRepository) Summary(ctx context.Context, orgID uuid.UUID) (ports.ReportSummary, error) {
	if r == nil || r.DB == nil {
		return ports.ReportSummary{}, fmt.Errorf("report repository unavailable")
	}
	var summary ports.ReportSummary
	if err := r.DB.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE state='scheduled' AND deleted_at IS NULL) AS scheduled,
			COUNT(*) FILTER (WHERE state='in_progress' AND deleted_at IS NULL) AS in_progress,
			COUNT(*) FILTER (WHERE state='completed' AND deleted_at IS NULL) AS completed,
			COUNT(*) FILTER (WHERE state='cancelled' AND deleted_at IS NULL) AS cancelled
		FROM maintenance_tasks
		WHERE org_id=$1
	`, orgID).Scan(&summary.TasksScheduled, &summary.TasksInProgress, &summary.TasksCompleted, &summary.TasksCancelled); err != nil {
		return ports.ReportSummary{}, err
	}
	if err := r.DB.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM aircraft
		WHERE org_id=$1 AND deleted_at IS NULL
	`, orgID).Scan(&summary.AircraftTotal); err != nil {
		return ports.ReportSummary{}, err
	}
	if err := r.DB.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status='in_stock' AND deleted_at IS NULL) AS in_stock,
			COUNT(*) FILTER (WHERE status='used' AND deleted_at IS NULL) AS used,
			COUNT(*) FILTER (WHERE status='disposed' AND deleted_at IS NULL) AS disposed
		FROM part_items
		WHERE org_id=$1
	`, orgID).Scan(&summary.PartsInStock, &summary.PartsUsed, &summary.PartsDisposed); err != nil {
		return ports.ReportSummary{}, err
	}
	if err := r.DB.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE result='pending' AND deleted_at IS NULL) AS pending,
			COUNT(*) FILTER (WHERE sign_off_time IS NOT NULL AND deleted_at IS NULL) AS signed
		FROM compliance_items
		WHERE org_id=$1
	`, orgID).Scan(&summary.CompliancePending, &summary.ComplianceSigned); err != nil {
		return ports.ReportSummary{}, err
	}
	return summary, nil
}

func (r *ReportRepository) Compliance(ctx context.Context, filter ports.ComplianceReportFilter) (ports.ComplianceReport, error) {
	if r == nil || r.DB == nil {
		return ports.ComplianceReport{}, fmt.Errorf("report repository unavailable")
	}
	clauses := []string{"org_id=$1", "deleted_at IS NULL"}
	args := []any{filter.OrgID}
	add := func(condition string, value any) {
		args = append(args, value)
		clauses = append(clauses, condition+fmt.Sprintf("$%d", len(args)))
	}
	if filter.TaskID != nil {
		add("task_id=", *filter.TaskID)
	}
	if filter.Result != nil {
		add("result=", *filter.Result)
	}
	if filter.From != nil {
		add("created_at >=", *filter.From)
	}
	if filter.To != nil {
		add("created_at <=", *filter.To)
	}
	if filter.Signed != nil {
		if *filter.Signed {
			clauses = append(clauses, "sign_off_time IS NOT NULL")
		} else {
			clauses = append(clauses, "sign_off_time IS NULL")
		}
	}

	query := `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE result='pass') AS pass,
			COUNT(*) FILTER (WHERE result='fail') AS fail,
			COUNT(*) FILTER (WHERE result='pending') AS pending,
			COUNT(*) FILTER (WHERE sign_off_time IS NOT NULL) AS signed,
			COUNT(*) FILTER (WHERE sign_off_time IS NULL) AS unsigned
		FROM compliance_items
		WHERE ` + strings.Join(clauses, " AND ")

	var report ports.ComplianceReport
	if err := r.DB.QueryRow(ctx, query, args...).Scan(&report.Total, &report.Pass, &report.Fail, &report.Pending, &report.Signed, &report.Unsigned); err != nil {
		return ports.ComplianceReport{}, err
	}
	return report, nil
}
