package jobs

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/pkg/observability"
	"github.com/rs/zerolog"
)

type RetentionCleaner struct {
	Orgs      ports.OrganizationRepository
	Retention ports.RetentionRepository
	Policies  *services.OrgPolicyService
	Logger    zerolog.Logger
	Interval  time.Duration
}

func (c *RetentionCleaner) Run(ctx context.Context) {
	interval := c.Interval
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.processOnce(ctx)
		}
	}
}

func (c *RetentionCleaner) processOnce(ctx context.Context) {
	if c.Orgs == nil || c.Retention == nil {
		return
	}
	observability.IncJobRun("retention_cleaner")
	now := time.Now().UTC()
	limit := 100
	offset := 0
	var hadError bool

	for {
		orgs, err := c.Orgs.List(ctx, ports.OrganizationFilter{Limit: limit, Offset: offset})
		if err != nil {
			hadError = true
			c.Logger.Error().Err(err).Msg("retention list orgs failed")
			break
		}
		if len(orgs) == 0 {
			break
		}
		for _, org := range orgs {
			retentionInterval := 365 * 24 * time.Hour
			if c.Policies != nil {
				policy, err := c.Policies.Get(ctx, org.ID)
				if err != nil {
					hadError = true
					c.Logger.Error().Err(err).Str("org_id", org.ID.String()).Msg("retention policy lookup failed")
					continue
				}
				if policy.RetentionInterval > 0 {
					retentionInterval = policy.RetentionInterval
				}
			}
			cutoff := now.Add(-retentionInterval)
			stats, err := c.Retention.Cleanup(ctx, org.ID, cutoff)
			if err != nil {
				hadError = true
				c.Logger.Error().Err(err).Str("org_id", org.ID.String()).Msg("retention cleanup failed")
				continue
			}
			if stats.Total() > 0 {
				c.Logger.Info().
					Str("org_id", org.ID.String()).
					Int64("deleted_total", stats.Total()).
					Int64("part_reservations", stats.PartReservations).
					Int64("compliance_items", stats.ComplianceItems).
					Int64("maintenance_tasks", stats.MaintenanceTasks).
					Int64("programs", stats.Programs).
					Int64("part_items", stats.PartItems).
					Int64("part_definitions", stats.PartDefinitions).
					Int64("aircraft", stats.Aircraft).
					Int64("users", stats.Users).
					Msg("retention cleanup complete")
			}
		}
		offset += len(orgs)
		if len(orgs) < limit {
			break
		}
	}

	if hadError {
		observability.IncJobFailure("retention_cleaner")
	}
}
