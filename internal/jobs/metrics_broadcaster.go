package jobs

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/infra/ws"
	"github.com/aeromaintain/amss/pkg/observability"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// MetricsBroadcaster periodically computes dashboard metrics for each
// connected org and broadcasts them over WebSocket.
type MetricsBroadcaster struct {
	Metrics  ports.MetricsRepository
	Hub      *ws.Hub
	Logger   zerolog.Logger
	Interval time.Duration
}

func (b *MetricsBroadcaster) Run(ctx context.Context) {
	interval := b.Interval
	if interval == 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.broadcast(ctx)
		}
	}
}

func (b *MetricsBroadcaster) broadcast(ctx context.Context) {
	if b.Hub == nil || b.Metrics == nil {
		return
	}

	orgIDs := b.Hub.ConnectedOrgIDs()
	if len(orgIDs) == 0 {
		return
	}

	observability.IncJobRun("metrics_broadcaster")

	for _, orgIDStr := range orgIDs {
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			continue
		}

		metrics, err := b.Metrics.GetDashboardMetrics(ctx, orgID)
		if err != nil {
			observability.IncJobFailure("metrics_broadcaster")
			b.Logger.Error().Err(err).Str("org_id", orgIDStr).Msg("failed to compute dashboard metrics")
			continue
		}

		b.Hub.Broadcast(orgIDStr, ws.EventDashboardMetrics, metrics)
	}
}
