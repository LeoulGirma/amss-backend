package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RetentionStats struct {
	PartReservations int64
	ComplianceItems  int64
	MaintenanceTasks int64
	Programs         int64
	PartItems        int64
	PartDefinitions  int64
	Aircraft         int64
	Users            int64
}

func (s RetentionStats) Total() int64 {
	return s.PartReservations + s.ComplianceItems + s.MaintenanceTasks + s.Programs +
		s.PartItems + s.PartDefinitions + s.Aircraft + s.Users
}

type RetentionRepository interface {
	Cleanup(ctx context.Context, orgID uuid.UUID, cutoff time.Time) (RetentionStats, error)
}
