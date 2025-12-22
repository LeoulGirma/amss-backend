package postgres

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RetentionRepository struct {
	DB *pgxpool.Pool
}

func (r *RetentionRepository) Cleanup(ctx context.Context, orgID uuid.UUID, cutoff time.Time) (ports.RetentionStats, error) {
	stats := ports.RetentionStats{}
	if r == nil || r.DB == nil {
		return stats, nil
	}

	var err error
	stats.PartReservations, err = execDelete(ctx, r.DB, `
		DELETE FROM part_reservations
		WHERE org_id=$1 AND (
			task_id IN (
				SELECT id FROM maintenance_tasks
				WHERE org_id=$1 AND deleted_at IS NOT NULL AND deleted_at < $2
			)
			OR part_item_id IN (
				SELECT id FROM part_items
				WHERE org_id=$1 AND deleted_at IS NOT NULL AND deleted_at < $2
			)
		)
	`, orgID, cutoff)
	if err != nil {
		return stats, err
	}

	stats.ComplianceItems, err = execDelete(ctx, r.DB, `
		DELETE FROM compliance_items
		WHERE org_id=$1 AND deleted_at IS NOT NULL AND deleted_at < $2
	`, orgID, cutoff)
	if err != nil {
		return stats, err
	}

	stats.MaintenanceTasks, err = execDelete(ctx, r.DB, `
		DELETE FROM maintenance_tasks
		WHERE org_id=$1 AND deleted_at IS NOT NULL AND deleted_at < $2
			AND NOT EXISTS (
				SELECT 1 FROM part_reservations
				WHERE org_id=$1 AND task_id=maintenance_tasks.id
			)
			AND NOT EXISTS (
				SELECT 1 FROM compliance_items
				WHERE org_id=$1 AND task_id=maintenance_tasks.id
			)
	`, orgID, cutoff)
	if err != nil {
		return stats, err
	}

	stats.Programs, err = execDelete(ctx, r.DB, `
		DELETE FROM maintenance_programs
		WHERE org_id=$1 AND deleted_at IS NOT NULL AND deleted_at < $2
			AND NOT EXISTS (
				SELECT 1 FROM maintenance_tasks
				WHERE org_id=$1 AND program_id=maintenance_programs.id
			)
	`, orgID, cutoff)
	if err != nil {
		return stats, err
	}

	stats.PartItems, err = execDelete(ctx, r.DB, `
		DELETE FROM part_items
		WHERE org_id=$1 AND deleted_at IS NOT NULL AND deleted_at < $2
			AND NOT EXISTS (
				SELECT 1 FROM part_reservations
				WHERE org_id=$1 AND part_item_id=part_items.id
			)
	`, orgID, cutoff)
	if err != nil {
		return stats, err
	}

	stats.PartDefinitions, err = execDelete(ctx, r.DB, `
		DELETE FROM part_definitions
		WHERE org_id=$1 AND deleted_at IS NOT NULL AND deleted_at < $2
			AND NOT EXISTS (
				SELECT 1 FROM part_items
				WHERE org_id=$1 AND part_definition_id=part_definitions.id
			)
	`, orgID, cutoff)
	if err != nil {
		return stats, err
	}

	stats.Aircraft, err = execDelete(ctx, r.DB, `
		DELETE FROM aircraft
		WHERE org_id=$1 AND deleted_at IS NOT NULL AND deleted_at < $2
			AND NOT EXISTS (
				SELECT 1 FROM maintenance_tasks
				WHERE org_id=$1 AND aircraft_id=aircraft.id
			)
	`, orgID, cutoff)
	if err != nil {
		return stats, err
	}

	stats.Users, err = execDelete(ctx, r.DB, `
		DELETE FROM users
		WHERE org_id=$1 AND deleted_at IS NOT NULL AND deleted_at < $2
			AND NOT EXISTS (
				SELECT 1 FROM maintenance_tasks
				WHERE org_id=$1 AND assigned_mechanic_id=users.id
			)
			AND NOT EXISTS (
				SELECT 1 FROM compliance_items
				WHERE org_id=$1 AND sign_off_user_id=users.id
			)
			AND NOT EXISTS (
				SELECT 1 FROM audit_logs
				WHERE org_id=$1 AND user_id=users.id
			)
			AND NOT EXISTS (
				SELECT 1 FROM imports
				WHERE org_id=$1 AND created_by=users.id
			)
			AND NOT EXISTS (
				SELECT 1 FROM refresh_tokens
				WHERE org_id=$1 AND user_id=users.id
			)
	`, orgID, cutoff)
	if err != nil {
		return stats, err
	}

	return stats, nil
}

func execDelete(ctx context.Context, db *pgxpool.Pool, query string, args ...any) (int64, error) {
	cmd, err := db.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}
