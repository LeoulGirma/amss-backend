package services

import (
	"context"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
)

type MetricsService struct {
	Metrics ports.MetricsRepository
	Clock   app.Clock
}

// GetDashboardMetrics returns a point-in-time snapshot of all dashboard KPIs
// for the given organization.
func (s *MetricsService) GetDashboardMetrics(ctx context.Context, actor app.Actor) (domain.DashboardMetrics, error) {
	return s.Metrics.GetDashboardMetrics(ctx, actor.OrgID)
}
