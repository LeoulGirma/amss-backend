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

type AircraftRepository struct {
	DB *pgxpool.Pool
}

func (r *AircraftRepository) GetStatus(ctx context.Context, orgID, id uuid.UUID) (domain.AircraftStatus, error) {
	if r == nil || r.DB == nil {
		return "", domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT status
		FROM aircraft
		WHERE org_id=$1 AND id=$2 AND deleted_at IS NULL
	`, orgID, id)
	var status domain.AircraftStatus
	if err := row.Scan(&status); err != nil {
		if err == pgx.ErrNoRows {
			return "", domain.ErrNotFound
		}
		return "", err
	}
	return status, nil
}

func (r *AircraftRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.Aircraft, error) {
	if r == nil || r.DB == nil {
		return domain.Aircraft{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, tail_number, model, last_maintenance, next_due, status, capacity_slots, flight_hours_total, cycles_total, deleted_at, created_at, updated_at
		FROM aircraft
		WHERE org_id=$1 AND id=$2 AND deleted_at IS NULL
	`, orgID, id)
	return scanAircraft(row)
}

func (r *AircraftRepository) GetByTailNumber(ctx context.Context, orgID uuid.UUID, tailNumber string) (domain.Aircraft, error) {
	if r == nil || r.DB == nil {
		return domain.Aircraft{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, tail_number, model, last_maintenance, next_due, status, capacity_slots, flight_hours_total, cycles_total, deleted_at, created_at, updated_at
		FROM aircraft
		WHERE org_id=$1 AND tail_number=$2 AND deleted_at IS NULL
	`, orgID, tailNumber)
	return scanAircraft(row)
}

func (r *AircraftRepository) Create(ctx context.Context, aircraft domain.Aircraft) (domain.Aircraft, error) {
	if r == nil || r.DB == nil {
		return domain.Aircraft{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		INSERT INTO aircraft
			(id, org_id, tail_number, model, last_maintenance, next_due, status, capacity_slots, flight_hours_total, cycles_total, created_at, updated_at, deleted_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, org_id, tail_number, model, last_maintenance, next_due, status, capacity_slots, flight_hours_total, cycles_total, deleted_at, created_at, updated_at
	`, aircraft.ID, aircraft.OrgID, aircraft.TailNumber, aircraft.Model, aircraft.LastMaintenance, aircraft.NextDue, aircraft.Status, aircraft.CapacitySlots, aircraft.FlightHoursTotal, aircraft.CyclesTotal, aircraft.CreatedAt, aircraft.UpdatedAt, aircraft.DeletedAt)
	created, err := scanAircraft(row)
	if err != nil {
		return domain.Aircraft{}, TranslateError(err)
	}
	return created, nil
}

func (r *AircraftRepository) Update(ctx context.Context, aircraft domain.Aircraft) (domain.Aircraft, error) {
	if r == nil || r.DB == nil {
		return domain.Aircraft{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		UPDATE aircraft
		SET tail_number=$1, model=$2, last_maintenance=$3, next_due=$4, status=$5, capacity_slots=$6, flight_hours_total=$7, cycles_total=$8, updated_at=$9
		WHERE org_id=$10 AND id=$11 AND deleted_at IS NULL
		RETURNING id, org_id, tail_number, model, last_maintenance, next_due, status, capacity_slots, flight_hours_total, cycles_total, deleted_at, created_at, updated_at
	`, aircraft.TailNumber, aircraft.Model, aircraft.LastMaintenance, aircraft.NextDue, aircraft.Status, aircraft.CapacitySlots, aircraft.FlightHoursTotal, aircraft.CyclesTotal, aircraft.UpdatedAt, aircraft.OrgID, aircraft.ID)
	updated, err := scanAircraft(row)
	if err != nil {
		return domain.Aircraft{}, TranslateError(err)
	}
	return updated, nil
}

func (r *AircraftRepository) SoftDelete(ctx context.Context, orgID, id uuid.UUID, at time.Time) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	cmd, err := r.DB.Exec(ctx, `
		UPDATE aircraft
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

func (r *AircraftRepository) List(ctx context.Context, filter ports.AircraftFilter) ([]domain.Aircraft, error) {
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
	if filter.Status != nil {
		add("status=", *filter.Status)
	}
	if filter.Model != "" {
		add("model ILIKE ", "%"+filter.Model+"%")
	}
	if filter.TailNumber != "" {
		add("tail_number ILIKE ", "%"+filter.TailNumber+"%")
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
		SELECT id, org_id, tail_number, model, last_maintenance, next_due, status, capacity_slots, flight_hours_total, cycles_total, deleted_at, created_at, updated_at
		FROM aircraft
		WHERE deleted_at IS NULL`
	if len(clauses) > 0 {
		query += " AND " + strings.Join(clauses, " AND ")
	}
	args = append(args, limit, offset)
	query += " ORDER BY tail_number ASC LIMIT $" + itoa(len(args)-1) + " OFFSET $" + itoa(len(args))

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Aircraft
	for rows.Next() {
		item, err := scanAircraft(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanAircraft(row pgx.Row) (domain.Aircraft, error) {
	var aircraft domain.Aircraft
	var lastMaintenance *time.Time
	var nextDue *time.Time
	if err := row.Scan(&aircraft.ID, &aircraft.OrgID, &aircraft.TailNumber, &aircraft.Model, &lastMaintenance, &nextDue, &aircraft.Status, &aircraft.CapacitySlots, &aircraft.FlightHoursTotal, &aircraft.CyclesTotal, &aircraft.DeletedAt, &aircraft.CreatedAt, &aircraft.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.Aircraft{}, domain.ErrNotFound
		}
		return domain.Aircraft{}, err
	}
	aircraft.LastMaintenance = lastMaintenance
	aircraft.NextDue = nextDue
	return aircraft, nil
}
