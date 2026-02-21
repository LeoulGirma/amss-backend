package domain

import (
	"time"

	"github.com/google/uuid"
)

type AircraftStatus string

const (
	AircraftOperational AircraftStatus = "operational"
	AircraftMaintenance AircraftStatus = "maintenance"
	AircraftGrounded    AircraftStatus = "grounded"
)

type Aircraft struct {
	ID               uuid.UUID
	OrgID            uuid.UUID
	TailNumber       string
	Model            string
	AircraftTypeID   *uuid.UUID
	LastMaintenance  *time.Time
	NextDue          *time.Time
	Status           AircraftStatus
	CapacitySlots    int
	FlightHoursTotal int
	CyclesTotal      int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}
