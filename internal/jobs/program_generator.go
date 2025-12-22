package jobs

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/aeromaintain/amss/pkg/observability"
	"github.com/rs/zerolog"
)

type ProgramGenerator struct {
	Programs *services.MaintenanceProgramService
	Logger   zerolog.Logger
}

func (g *ProgramGenerator) Run(ctx context.Context) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.process(ctx)
		}
	}
}

func (g *ProgramGenerator) process(ctx context.Context) {
	if g.Programs == nil {
		return
	}
	observability.IncJobRun("program_generator")
	actor := app.Actor{
		UserID: uuidNew(),
		OrgID:  uuidNew(),
		Role:   domain.RoleAdmin,
	}
	created, err := g.Programs.GenerateDueTasks(ctx, actor, 100)
	if err != nil {
		observability.IncJobFailure("program_generator")
		g.Logger.Error().Err(err).Msg("program generation failed")
		return
	}
	if created > 0 {
		g.Logger.Info().Int("tasks", created).Msg("program tasks generated")
	}
}
