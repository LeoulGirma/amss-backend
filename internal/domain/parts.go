package domain

import (
	"time"

	"github.com/google/uuid"
)

type PartItemStatus string

type PartReservationState string

const (
	PartItemInStock  PartItemStatus = "in_stock"
	PartItemUsed     PartItemStatus = "used"
	PartItemDisposed PartItemStatus = "disposed"
)

const (
	ReservationReserved PartReservationState = "reserved"
	ReservationUsed     PartReservationState = "used"
	ReservationReleased PartReservationState = "released"
)

type PartDefinition struct {
	ID        uuid.UUID
	OrgID     uuid.UUID
	Name      string
	Category  string
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PartItem struct {
	ID           uuid.UUID
	OrgID        uuid.UUID
	DefinitionID uuid.UUID
	SerialNumber string
	Status       PartItemStatus
	ExpiryDate   *time.Time
	DeletedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PartReservation struct {
	ID         uuid.UUID
	OrgID      uuid.UUID
	TaskID     uuid.UUID
	PartItemID uuid.UUID
	State      PartReservationState
	Quantity   int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (r PartReservation) CanTransition(newState PartReservationState) error {
	if r.State == newState {
		return nil
	}
	if r.State != ReservationReserved {
		return NewConflictError("reservation must be in reserved state")
	}
	if newState != ReservationUsed && newState != ReservationReleased {
		return NewValidationError("invalid reservation transition")
	}
	return nil
}
