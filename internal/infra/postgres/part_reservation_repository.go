package postgres

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PartReservationRepository struct {
	DB *pgxpool.Pool
}

func (r *PartReservationRepository) Create(ctx context.Context, reservation domain.PartReservation) error {
	if r == nil || r.DB == nil {
		return nil
	}
	_, err := r.DB.Exec(ctx, `
		INSERT INTO part_reservations (id, org_id, task_id, part_item_id, state, quantity, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, reservation.ID, reservation.OrgID, reservation.TaskID, reservation.PartItemID, reservation.State, reservation.Quantity, reservation.CreatedAt, reservation.UpdatedAt)
	return TranslateError(err)
}

func (r *PartReservationRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.PartReservation, error) {
	if r == nil || r.DB == nil {
		return domain.PartReservation{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, task_id, part_item_id, state, quantity, created_at, updated_at
		FROM part_reservations
		WHERE org_id=$1 AND id=$2
	`, orgID, id)
	var reservation domain.PartReservation
	if err := row.Scan(&reservation.ID, &reservation.OrgID, &reservation.TaskID, &reservation.PartItemID, &reservation.State, &reservation.Quantity, &reservation.CreatedAt, &reservation.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.PartReservation{}, domain.ErrNotFound
		}
		return domain.PartReservation{}, err
	}
	return reservation, nil
}

func (r *PartReservationRepository) ListByTask(ctx context.Context, orgID, taskID uuid.UUID) ([]domain.PartReservation, error) {
	if r == nil || r.DB == nil {
		return nil, nil
	}
	rows, err := r.DB.Query(ctx, `
		SELECT id, org_id, task_id, part_item_id, state, quantity, created_at, updated_at
		FROM part_reservations
		WHERE org_id=$1 AND task_id=$2
	`, orgID, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []domain.PartReservation
	for rows.Next() {
		var reservation domain.PartReservation
		if err := rows.Scan(&reservation.ID, &reservation.OrgID, &reservation.TaskID, &reservation.PartItemID, &reservation.State, &reservation.Quantity, &reservation.CreatedAt, &reservation.UpdatedAt); err != nil {
			return nil, err
		}
		reservations = append(reservations, reservation)
	}
	return reservations, rows.Err()
}

func (r *PartReservationRepository) UpdateState(ctx context.Context, orgID, id uuid.UUID, state domain.PartReservationState, now time.Time) error {
	if r == nil || r.DB == nil {
		return nil
	}
	_, err := r.DB.Exec(ctx, `
		UPDATE part_reservations
		SET state=$1, updated_at=$2
		WHERE org_id=$3 AND id=$4
	`, state, now, orgID, id)
	return TranslateError(err)
}

func (r *PartReservationRepository) ReleaseByTask(ctx context.Context, orgID, taskID uuid.UUID, now time.Time) error {
	if r == nil || r.DB == nil {
		return nil
	}
	_, err := r.DB.Exec(ctx, `
		UPDATE part_reservations
		SET state='released', updated_at=$1
		WHERE org_id=$2 AND task_id=$3 AND state='reserved'
	`, now, orgID, taskID)
	return TranslateError(err)
}
