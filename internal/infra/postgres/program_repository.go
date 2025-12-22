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

type MaintenanceProgramRepository struct {
	DB *pgxpool.Pool
}

func (r *MaintenanceProgramRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.MaintenanceProgram, error) {
	if r == nil || r.DB == nil {
		return domain.MaintenanceProgram{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, aircraft_id, name, interval_type, interval_value, last_performed, created_at, updated_at, deleted_at
		FROM maintenance_programs
		WHERE org_id=$1 AND id=$2 AND deleted_at IS NULL
	`, orgID, id)
	return scanProgram(row)
}

func (r *MaintenanceProgramRepository) GetByName(ctx context.Context, orgID uuid.UUID, name string, aircraftID *uuid.UUID) (domain.MaintenanceProgram, error) {
	if r == nil || r.DB == nil {
		return domain.MaintenanceProgram{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, aircraft_id, name, interval_type, interval_value, last_performed, created_at, updated_at, deleted_at
		FROM maintenance_programs
		WHERE org_id=$1 AND name=$2 AND aircraft_id IS NOT DISTINCT FROM $3 AND deleted_at IS NULL
	`, orgID, name, aircraftID)
	return scanProgram(row)
}

func (r *MaintenanceProgramRepository) Create(ctx context.Context, program domain.MaintenanceProgram) (domain.MaintenanceProgram, error) {
	if r == nil || r.DB == nil {
		return domain.MaintenanceProgram{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		INSERT INTO maintenance_programs
			(id, org_id, aircraft_id, name, interval_type, interval_value, last_performed, created_at, updated_at, deleted_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, org_id, aircraft_id, name, interval_type, interval_value, last_performed, created_at, updated_at, deleted_at
	`, program.ID, program.OrgID, program.AircraftID, program.Name, program.IntervalType, program.IntervalValue, program.LastPerformed, program.CreatedAt, program.UpdatedAt, program.DeletedAt)
	created, err := scanProgram(row)
	if err != nil {
		return domain.MaintenanceProgram{}, TranslateError(err)
	}
	return created, nil
}

func (r *MaintenanceProgramRepository) Update(ctx context.Context, program domain.MaintenanceProgram) (domain.MaintenanceProgram, error) {
	if r == nil || r.DB == nil {
		return domain.MaintenanceProgram{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		UPDATE maintenance_programs
		SET aircraft_id=$1, name=$2, interval_type=$3, interval_value=$4, last_performed=$5, updated_at=$6
		WHERE org_id=$7 AND id=$8 AND deleted_at IS NULL
		RETURNING id, org_id, aircraft_id, name, interval_type, interval_value, last_performed, created_at, updated_at, deleted_at
	`, program.AircraftID, program.Name, program.IntervalType, program.IntervalValue, program.LastPerformed, program.UpdatedAt, program.OrgID, program.ID)
	updated, err := scanProgram(row)
	if err != nil {
		return domain.MaintenanceProgram{}, TranslateError(err)
	}
	return updated, nil
}

func (r *MaintenanceProgramRepository) SoftDelete(ctx context.Context, orgID, id uuid.UUID, at time.Time) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	cmd, err := r.DB.Exec(ctx, `
		UPDATE maintenance_programs
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

func (r *MaintenanceProgramRepository) List(ctx context.Context, filter ports.MaintenanceProgramFilter) ([]domain.MaintenanceProgram, error) {
	if r == nil || r.DB == nil {
		return nil, nil
	}
	clauses := make([]string, 0, 3)
	args := make([]any, 0, 5)
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
		SELECT id, org_id, aircraft_id, name, interval_type, interval_value, last_performed, created_at, updated_at, deleted_at
		FROM maintenance_programs
		WHERE deleted_at IS NULL`
	if len(clauses) > 0 {
		query += " AND " + strings.Join(clauses, " AND ")
	}
	args = append(args, limit, offset)
	query += " ORDER BY created_at DESC LIMIT $" + itoa(len(args)-1) + " OFFSET $" + itoa(len(args))

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var programs []domain.MaintenanceProgram
	for rows.Next() {
		prog, err := scanProgram(rows)
		if err != nil {
			return nil, err
		}
		programs = append(programs, prog)
	}
	return programs, rows.Err()
}

func (r *MaintenanceProgramRepository) ListDueCalendar(ctx context.Context, now time.Time, limit int) ([]domain.MaintenanceProgram, error) {
	if r == nil || r.DB == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.Query(ctx, `
		SELECT id, org_id, aircraft_id, name, interval_type, interval_value, last_performed, created_at, updated_at, deleted_at
		FROM maintenance_programs
		WHERE interval_type='calendar'
		  AND deleted_at IS NULL
		  AND aircraft_id IS NOT NULL
		  AND (
		    last_performed IS NULL
		    OR last_performed + (interval_value || ' days')::interval <= $1
		  )
		ORDER BY created_at ASC
		LIMIT $2
	`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var programs []domain.MaintenanceProgram
	for rows.Next() {
		prog, err := scanProgram(rows)
		if err != nil {
			return nil, err
		}
		programs = append(programs, prog)
	}
	return programs, rows.Err()
}

func scanProgram(row pgx.Row) (domain.MaintenanceProgram, error) {
	var program domain.MaintenanceProgram
	var aircraftID *uuid.UUID
	var lastPerformed *time.Time
	if err := row.Scan(&program.ID, &program.OrgID, &aircraftID, &program.Name, &program.IntervalType, &program.IntervalValue, &lastPerformed, &program.CreatedAt, &program.UpdatedAt, &program.DeletedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.MaintenanceProgram{}, domain.ErrNotFound
		}
		return domain.MaintenanceProgram{}, err
	}
	program.AircraftID = aircraftID
	program.LastPerformed = lastPerformed
	return program, nil
}
