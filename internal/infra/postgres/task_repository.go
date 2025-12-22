package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	DB *pgxpool.Pool
}

func (r *TaskRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.MaintenanceTask, error) {
	if r == nil || r.DB == nil {
		return domain.MaintenanceTask{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, aircraft_id, program_id, type, state, start_time, end_time, assigned_mechanic_id, notes, deleted_at, created_at, updated_at
		FROM maintenance_tasks
		WHERE org_id=$1 AND id=$2 AND deleted_at IS NULL
	`, orgID, id)
	return scanTask(row)
}

func (r *TaskRepository) Create(ctx context.Context, task domain.MaintenanceTask) (domain.MaintenanceTask, error) {
	if r == nil || r.DB == nil {
		return domain.MaintenanceTask{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		INSERT INTO maintenance_tasks
			(id, org_id, aircraft_id, program_id, type, state, start_time, end_time, assigned_mechanic_id, notes, created_at, updated_at, deleted_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, org_id, aircraft_id, program_id, type, state, start_time, end_time, assigned_mechanic_id, notes, deleted_at, created_at, updated_at
	`, task.ID, task.OrgID, task.AircraftID, task.ProgramID, task.Type, task.State, task.StartTime, task.EndTime, task.AssignedMechanicID, task.Notes, task.CreatedAt, task.UpdatedAt, task.DeletedAt)
	created, err := scanTask(row)
	if err != nil {
		return domain.MaintenanceTask{}, TranslateError(err)
	}
	return created, nil
}

func (r *TaskRepository) Update(ctx context.Context, task domain.MaintenanceTask) (domain.MaintenanceTask, error) {
	if r == nil || r.DB == nil {
		return domain.MaintenanceTask{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		UPDATE maintenance_tasks
		SET program_id=$1, type=$2, start_time=$3, end_time=$4, assigned_mechanic_id=$5, notes=$6, updated_at=$7
		WHERE org_id=$8 AND id=$9 AND deleted_at IS NULL
		RETURNING id, org_id, aircraft_id, program_id, type, state, start_time, end_time, assigned_mechanic_id, notes, deleted_at, created_at, updated_at
	`, task.ProgramID, task.Type, task.StartTime, task.EndTime, task.AssignedMechanicID, task.Notes, task.UpdatedAt, task.OrgID, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		return domain.MaintenanceTask{}, TranslateError(err)
	}
	return updated, nil
}

func (r *TaskRepository) SoftDelete(ctx context.Context, orgID, id uuid.UUID, at time.Time) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	cmd, err := r.DB.Exec(ctx, `
		UPDATE maintenance_tasks
		SET deleted_at=$1, updated_at=$1
		WHERE org_id=$2 AND id=$3 AND deleted_at IS NULL
	`, at, orgID, id)
	if err != nil {
		return TranslateError(err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TaskRepository) List(ctx context.Context, filter ports.TaskFilter) ([]domain.MaintenanceTask, error) {
	if r == nil || r.DB == nil {
		return nil, nil
	}
	clauses := make([]string, 0, 6)
	args := make([]any, 0, 8)
	add := func(condition string, value any) {
		args = append(args, value)
		clauses = append(clauses, condition+"$"+itoa(len(args)))
	}
	if filter.OrgID != nil {
		add("org_id=", *filter.OrgID)
	}
	if filter.AircraftID != nil {
		add("aircraft_id=", *filter.AircraftID)
	}
	if filter.State != nil {
		add("state=", *filter.State)
	}
	if filter.Type != nil {
		add("type=", *filter.Type)
	}
	if filter.StartFrom != nil {
		add("start_time >= ", *filter.StartFrom)
	}
	if filter.StartTo != nil {
		add("start_time <= ", *filter.StartTo)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, org_id, aircraft_id, program_id, type, state, start_time, end_time, assigned_mechanic_id, notes, deleted_at, created_at, updated_at
		FROM maintenance_tasks
		WHERE deleted_at IS NULL`
	if len(clauses) > 0 {
		query += " AND " + strings.Join(clauses, " AND ")
	}
	args = append(args, limit, offset)
	query += " ORDER BY start_time DESC LIMIT $" + itoa(len(args)-1) + " OFFSET $" + itoa(len(args))

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []domain.MaintenanceTask
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *TaskRepository) UpdateState(ctx context.Context, orgID, id uuid.UUID, newState domain.TaskState, notes string, now time.Time) (domain.MaintenanceTask, error) {
	if r == nil || r.DB == nil {
		return domain.MaintenanceTask{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		UPDATE maintenance_tasks
		SET state=$1, notes=$2, updated_at=$3
		WHERE org_id=$4 AND id=$5 AND deleted_at IS NULL
		RETURNING id, org_id, aircraft_id, program_id, type, state, start_time, end_time, assigned_mechanic_id, notes, deleted_at, created_at, updated_at
	`, newState, notes, now, orgID, id)

	task, err := scanTask(row)
	if err != nil {
		return domain.MaintenanceTask{}, TranslateError(err)
	}
	return task, nil
}

func (r *TaskRepository) HasActiveForProgram(ctx context.Context, orgID, programID uuid.UUID) (bool, error) {
	if r == nil || r.DB == nil {
		return false, nil
	}
	row := r.DB.QueryRow(ctx, `
		SELECT 1
		FROM maintenance_tasks
		WHERE org_id=$1 AND program_id=$2 AND deleted_at IS NULL AND state IN ('scheduled','in_progress')
		LIMIT 1
	`, orgID, programID)
	var marker int
	if err := row.Scan(&marker); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func scanTask(row pgx.Row) (domain.MaintenanceTask, error) {
	var task domain.MaintenanceTask
	var programID *uuid.UUID
	var assignedID *uuid.UUID
	if err := row.Scan(&task.ID, &task.OrgID, &task.AircraftID, &programID, &task.Type, &task.State, &task.StartTime, &task.EndTime, &assignedID, &task.Notes, &task.DeletedAt, &task.CreatedAt, &task.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.MaintenanceTask{}, domain.ErrNotFound
		}
		return domain.MaintenanceTask{}, err
	}
	task.ProgramID = programID
	task.AssignedMechanicID = assignedID
	return task, nil
}
