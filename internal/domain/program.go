package domain

import (
	"time"

	"github.com/google/uuid"
)

type MaintenanceProgramIntervalType string

const (
	ProgramIntervalFlightHours MaintenanceProgramIntervalType = "flight_hours"
	ProgramIntervalCycles      MaintenanceProgramIntervalType = "cycles"
	ProgramIntervalCalendar    MaintenanceProgramIntervalType = "calendar"
)

type MaintenanceProgram struct {
	ID            uuid.UUID
	OrgID         uuid.UUID
	AircraftID    *uuid.UUID
	Name          string
	IntervalType  MaintenanceProgramIntervalType
	IntervalValue int
	LastPerformed *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}
