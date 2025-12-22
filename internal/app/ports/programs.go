package ports

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type MaintenanceProgramRepository interface {
	GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.MaintenanceProgram, error)
	Create(ctx context.Context, program domain.MaintenanceProgram) (domain.MaintenanceProgram, error)
	Update(ctx context.Context, program domain.MaintenanceProgram) (domain.MaintenanceProgram, error)
	SoftDelete(ctx context.Context, orgID, id uuid.UUID, at time.Time) error
	List(ctx context.Context, filter MaintenanceProgramFilter) ([]domain.MaintenanceProgram, error)
	ListDueCalendar(ctx context.Context, now time.Time, limit int) ([]domain.MaintenanceProgram, error)
	GetByName(ctx context.Context, orgID uuid.UUID, name string, aircraftID *uuid.UUID) (domain.MaintenanceProgram, error)
}

type MaintenanceProgramFilter struct {
	OrgID      *uuid.UUID
	AircraftID *uuid.UUID
	Limit      int
	Offset     int
}
