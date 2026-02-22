package postgres

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MetricsRepository struct {
	DB *pgxpool.Pool
}

func (r *MetricsRepository) GetDashboardMetrics(ctx context.Context, orgID uuid.UUID) (domain.DashboardMetrics, error) {
	now := time.Now().UTC()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	var m domain.DashboardMetrics
	m.ComputedAt = now

	// Fleet metrics
	err := r.DB.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'operational'),
			COUNT(*) FILTER (WHERE status = 'maintenance'),
			COUNT(*) FILTER (WHERE status = 'grounded')
		FROM aircraft
		WHERE org_id = $1 AND deleted_at IS NULL
	`, orgID).Scan(&m.FleetTotal, &m.FleetOperational, &m.FleetMaintenance, &m.FleetGrounded)
	if err != nil {
		return m, err
	}
	if m.FleetTotal > 0 {
		m.FleetAvailRate = float64(m.FleetOperational) / float64(m.FleetTotal)
	}

	// Task metrics
	err = r.DB.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE state = 'scheduled' AND deleted_at IS NULL),
			COUNT(*) FILTER (WHERE state = 'in_progress' AND deleted_at IS NULL),
			COUNT(*) FILTER (WHERE state = 'completed' AND deleted_at IS NULL AND updated_at >= $2),
			COUNT(*) FILTER (WHERE state IN ('scheduled', 'in_progress') AND deleted_at IS NULL AND end_time < $3)
		FROM maintenance_tasks
		WHERE org_id = $1
	`, orgID, thirtyDaysAgo, now).Scan(
		&m.TasksScheduled, &m.TasksInProgress, &m.TasksCompleted, &m.TasksOverdue,
	)
	if err != nil {
		return m, err
	}

	// On-time rate and average TAT (last 30 days completed tasks)
	err = r.DB.QueryRow(ctx, `
		SELECT
			COALESCE(AVG(CASE WHEN updated_at <= end_time THEN 1.0 ELSE 0.0 END), 0),
			COALESCE(AVG(EXTRACT(EPOCH FROM (updated_at - start_time)) / 3600.0), 0)
		FROM maintenance_tasks
		WHERE org_id = $1 AND state = 'completed' AND deleted_at IS NULL AND updated_at >= $2
	`, orgID, thirtyDaysAgo).Scan(&m.OnTimeRate, &m.AvgTATHours)
	if err != nil {
		return m, err
	}

	// Parts metrics
	err = r.DB.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE pi.status = 'in_stock'),
			COUNT(*) FILTER (WHERE pi.status = 'in_stock' AND pi.expiry_date IS NOT NULL AND pi.expiry_date <= $2)
		FROM part_items pi
		WHERE pi.org_id = $1 AND pi.deleted_at IS NULL
	`, orgID, now.AddDate(0, 0, 30)).Scan(&m.PartsInStock, &m.PartsExpiringSoon)
	if err != nil {
		return m, err
	}

	// Low stock: definitions where in_stock count < min_stock_level
	err = r.DB.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM (
			SELECT pd.id
			FROM part_definitions pd
			LEFT JOIN part_items pi ON pi.definition_id = pd.id AND pi.org_id = pd.org_id AND pi.status = 'in_stock' AND pi.deleted_at IS NULL
			WHERE pd.org_id = $1 AND pd.deleted_at IS NULL AND pd.min_stock_level > 0
			GROUP BY pd.id, pd.min_stock_level
			HAVING COUNT(pi.id) < pd.min_stock_level
		) sub
	`, orgID).Scan(&m.PartsLowStock)
	if err != nil {
		return m, err
	}

	// Parts fill rate (reservations fulfilled vs total in last 30 days)
	err = r.DB.QueryRow(ctx, `
		SELECT
			COALESCE(
				COUNT(*) FILTER (WHERE state = 'used')::float /
				NULLIF(COUNT(*) FILTER (WHERE state IN ('used', 'released')), 0),
			0)
		FROM part_reservations
		WHERE org_id = $1 AND updated_at >= $2
	`, orgID, thirtyDaysAgo).Scan(&m.PartsFillRate)
	if err != nil {
		return m, err
	}

	// Compliance metrics
	err = r.DB.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE result = 'pending'),
			COUNT(*) FILTER (WHERE result = 'pass'),
			COUNT(*) FILTER (WHERE result = 'fail')
		FROM compliance_items
		WHERE org_id = $1 AND deleted_at IS NULL
	`, orgID).Scan(&m.CompliancePending, &m.CompliancePassed, &m.ComplianceFailed)
	if err != nil {
		return m, err
	}

	// Overdue directives
	err = r.DB.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM aircraft_directive_compliance
		WHERE org_id = $1 AND status = 'overdue'
	`, orgID).Scan(&m.DirectivesOverdue)
	if err != nil {
		return m, err
	}

	// Expiring certifications
	err = r.DB.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE expiry_date <= $2),
			COUNT(*) FILTER (WHERE expiry_date <= $3),
			COUNT(*) FILTER (WHERE expiry_date <= $4),
			COUNT(*) FILTER (WHERE expiry_date < $5)
		FROM employee_certifications
		WHERE org_id = $1 AND status = 'active' AND expiry_date IS NOT NULL
	`, orgID,
		now.AddDate(0, 0, 30),
		now.AddDate(0, 0, 60),
		now.AddDate(0, 0, 90),
		now,
	).Scan(&m.CertsExpiring30d, &m.CertsExpiring60d, &m.CertsExpiring90d, &m.CertsExpired)
	if err != nil {
		return m, err
	}

	// Alert metrics
	err = r.DB.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE level = 'critical')
		FROM alerts
		WHERE org_id = $1 AND resolved = false
	`, orgID).Scan(&m.AlertsUnresolved, &m.AlertsCritical)
	if err != nil {
		return m, err
	}

	// Mechanic metrics
	err = r.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM users WHERE org_id = $1 AND role = 'mechanic' AND deleted_at IS NULL
	`, orgID).Scan(&m.MechanicsTotal)
	if err != nil {
		return m, err
	}

	err = r.DB.QueryRow(ctx, `
		SELECT COUNT(DISTINCT assigned_mechanic_id)
		FROM maintenance_tasks
		WHERE org_id = $1 AND state = 'in_progress' AND deleted_at IS NULL AND assigned_mechanic_id IS NOT NULL
	`, orgID).Scan(&m.MechanicsOnTask)
	if err != nil {
		return m, err
	}

	return m, nil
}
