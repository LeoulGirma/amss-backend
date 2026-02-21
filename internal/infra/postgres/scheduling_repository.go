package postgres

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Task Dependency Repository ---

type TaskDependencyRepository struct {
	DB *pgxpool.Pool
}

func (r *TaskDependencyRepository) Create(ctx context.Context, dep domain.TaskDependency) (domain.TaskDependency, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO task_dependencies (id, org_id, task_id, depends_on_task_id, dependency_type, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, org_id, task_id, depends_on_task_id, dependency_type, created_at
	`, dep.ID, dep.OrgID, dep.TaskID, dep.DependsOnTaskID, dep.DependencyType, dep.CreatedAt)
	var created domain.TaskDependency
	if err := row.Scan(&created.ID, &created.OrgID, &created.TaskID, &created.DependsOnTaskID, &created.DependencyType, &created.CreatedAt); err != nil {
		return domain.TaskDependency{}, TranslateError(err)
	}
	return created, nil
}

func (r *TaskDependencyRepository) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	cmd, err := r.DB.Exec(ctx, `
		DELETE FROM task_dependencies WHERE org_id=$1 AND id=$2
	`, orgID, id)
	if err != nil {
		return TranslateError(err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TaskDependencyRepository) ListByTask(ctx context.Context, orgID, taskID uuid.UUID) ([]domain.TaskDependency, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, org_id, task_id, depends_on_task_id, dependency_type, created_at
		FROM task_dependencies
		WHERE org_id=$1 AND task_id=$2
		ORDER BY created_at
	`, orgID, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDependencies(rows)
}

func (r *TaskDependencyRepository) ListDependents(ctx context.Context, orgID, taskID uuid.UUID) ([]domain.TaskDependency, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, org_id, task_id, depends_on_task_id, dependency_type, created_at
		FROM task_dependencies
		WHERE org_id=$1 AND depends_on_task_id=$2
		ORDER BY created_at
	`, orgID, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDependencies(rows)
}

func scanDependencies(rows pgx.Rows) ([]domain.TaskDependency, error) {
	var items []domain.TaskDependency
	for rows.Next() {
		var d domain.TaskDependency
		if err := rows.Scan(&d.ID, &d.OrgID, &d.TaskID, &d.DependsOnTaskID, &d.DependencyType, &d.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

// --- Schedule Change Repository ---

type ScheduleChangeRepository struct {
	DB *pgxpool.Pool
}

func (r *ScheduleChangeRepository) Create(ctx context.Context, event domain.ScheduleChangeEvent) (domain.ScheduleChangeEvent, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO schedule_change_events
			(id, org_id, task_id, change_type, reason, old_start_time, new_start_time,
			 old_end_time, new_end_time, triggered_by, affected_task_ids, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, org_id, task_id, change_type, reason, old_start_time, new_start_time,
		          old_end_time, new_end_time, triggered_by, affected_task_ids, created_at
	`, event.ID, event.OrgID, event.TaskID, event.ChangeType, event.Reason,
		event.OldStartTime, event.NewStartTime, event.OldEndTime, event.NewEndTime,
		event.TriggeredBy, event.AffectedTaskIDs, event.CreatedAt)

	var created domain.ScheduleChangeEvent
	if err := row.Scan(&created.ID, &created.OrgID, &created.TaskID, &created.ChangeType,
		&created.Reason, &created.OldStartTime, &created.NewStartTime,
		&created.OldEndTime, &created.NewEndTime, &created.TriggeredBy,
		&created.AffectedTaskIDs, &created.CreatedAt); err != nil {
		return domain.ScheduleChangeEvent{}, TranslateError(err)
	}
	return created, nil
}

func (r *ScheduleChangeRepository) ListByTask(ctx context.Context, orgID, taskID uuid.UUID) ([]domain.ScheduleChangeEvent, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, org_id, task_id, change_type, reason, old_start_time, new_start_time,
		       old_end_time, new_end_time, triggered_by, affected_task_ids, created_at
		FROM schedule_change_events
		WHERE org_id=$1 AND task_id=$2
		ORDER BY created_at DESC
	`, orgID, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ScheduleChangeEvent
	for rows.Next() {
		var e domain.ScheduleChangeEvent
		if err := rows.Scan(&e.ID, &e.OrgID, &e.TaskID, &e.ChangeType, &e.Reason,
			&e.OldStartTime, &e.NewStartTime, &e.OldEndTime, &e.NewEndTime,
			&e.TriggeredBy, &e.AffectedTaskIDs, &e.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}
