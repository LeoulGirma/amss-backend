package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/aeromaintain/amss/pkg/observability"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// AlertTrigger periodically scans for conditions that warrant alerts:
// - Expiring certifications (30/60/90 day)
// - Overdue maintenance tasks
// - Overdue compliance directives
type AlertTrigger struct {
	DB       *pgxpool.Pool
	Alerts   ports.AlertRepository
	Logger   zerolog.Logger
	Interval time.Duration
}

func (t *AlertTrigger) Run(ctx context.Context) {
	interval := t.Interval
	if interval == 0 {
		interval = 15 * time.Minute
	}

	// Initial run on startup
	t.process(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.process(ctx)
		}
	}
}

func (t *AlertTrigger) process(ctx context.Context) {
	if t.DB == nil || t.Alerts == nil {
		return
	}
	observability.IncJobRun("alert_trigger")

	t.checkExpiringCerts(ctx)
	t.checkOverdueTasks(ctx)
	t.checkOverdueDirectives(ctx)
}

func (t *AlertTrigger) checkExpiringCerts(ctx context.Context) {
	now := time.Now().UTC()
	thresholds := []struct {
		days  int
		level domain.AlertLevel
	}{
		{30, domain.AlertCritical},
		{60, domain.AlertWarning},
		{90, domain.AlertInfo},
	}

	for _, th := range thresholds {
		deadline := now.AddDate(0, 0, th.days)
		rows, err := t.DB.Query(ctx, `
			SELECT ec.id, ec.org_id, ec.user_id, ct.name, ec.expiry_date
			FROM employee_certifications ec
			JOIN certification_types ct ON ct.id = ec.cert_type_id
			WHERE ec.status = 'active'
			  AND ec.expiry_date IS NOT NULL
			  AND ec.expiry_date <= $1
			  AND ec.expiry_date > $2
			  AND NOT EXISTS (
			    SELECT 1 FROM alerts a
			    WHERE a.entity_type = 'employee_certification'
			      AND a.entity_id = ec.id
			      AND a.category = 'cert_expiring'
			      AND a.resolved = false
			  )
		`, deadline, now)
		if err != nil {
			t.Logger.Error().Err(err).Msg("failed to query expiring certs")
			continue
		}

		for rows.Next() {
			var certID, orgID, userID uuid.UUID
			var certName string
			var expiryDate time.Time
			if err := rows.Scan(&certID, &orgID, &userID, &certName, &expiryDate); err != nil {
				continue
			}

			daysUntil := int(time.Until(expiryDate).Hours() / 24)
			alert := domain.Alert{
				ID:         uuid.New(),
				OrgID:      orgID,
				Level:      th.level,
				Category:   "cert_expiring",
				Title:      fmt.Sprintf("Certification expiring: %s", certName),
				Description: fmt.Sprintf("Expires in %d days (user %s)", daysUntil, userID),
				EntityType: "employee_certification",
				EntityID:   certID,
				CreatedAt:  now,
			}
			if _, err := t.Alerts.Create(ctx, alert); err != nil {
				t.Logger.Error().Err(err).Msg("failed to create cert expiry alert")
			}
		}
		rows.Close()
	}
}

func (t *AlertTrigger) checkOverdueTasks(ctx context.Context) {
	now := time.Now().UTC()
	rows, err := t.DB.Query(ctx, `
		SELECT mt.id, mt.org_id, mt.end_time
		FROM maintenance_tasks mt
		WHERE mt.state IN ('scheduled', 'in_progress')
		  AND mt.deleted_at IS NULL
		  AND mt.end_time < $1
		  AND NOT EXISTS (
		    SELECT 1 FROM alerts a
		    WHERE a.entity_type = 'maintenance_task'
		      AND a.entity_id = mt.id
		      AND a.category = 'task_overdue'
		      AND a.resolved = false
		  )
	`, now)
	if err != nil {
		t.Logger.Error().Err(err).Msg("failed to query overdue tasks")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var taskID, orgID uuid.UUID
		var endTime time.Time
		if err := rows.Scan(&taskID, &orgID, &endTime); err != nil {
			continue
		}

		hoursOverdue := int(now.Sub(endTime).Hours())
		level := domain.AlertWarning
		if hoursOverdue > 24 {
			level = domain.AlertCritical
		}

		alert := domain.Alert{
			ID:          uuid.New(),
			OrgID:       orgID,
			Level:       level,
			Category:    "task_overdue",
			Title:       "Maintenance task overdue",
			Description: fmt.Sprintf("Task has been overdue for %d hours", hoursOverdue),
			EntityType:  "maintenance_task",
			EntityID:    taskID,
			CreatedAt:   now,
		}
		if level == domain.AlertCritical {
			escalateAt := now.Add(time.Hour)
			alert.AutoEscalateAt = &escalateAt
		}
		if _, err := t.Alerts.Create(ctx, alert); err != nil {
			t.Logger.Error().Err(err).Msg("failed to create task overdue alert")
		}
	}
}

func (t *AlertTrigger) checkOverdueDirectives(ctx context.Context) {
	now := time.Now().UTC()
	rows, err := t.DB.Query(ctx, `
		SELECT adc.id, adc.org_id, cd.title
		FROM aircraft_directive_compliance adc
		JOIN compliance_directives cd ON cd.id = adc.directive_id
		WHERE adc.status = 'pending'
		  AND adc.next_due_date IS NOT NULL
		  AND adc.next_due_date < $1
		  AND NOT EXISTS (
		    SELECT 1 FROM alerts a
		    WHERE a.entity_type = 'aircraft_directive_compliance'
		      AND a.entity_id = adc.id
		      AND a.category = 'directive_overdue'
		      AND a.resolved = false
		  )
	`, now)
	if err != nil {
		t.Logger.Error().Err(err).Msg("failed to query overdue directives")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var complianceID, orgID uuid.UUID
		var directiveTitle string
		if err := rows.Scan(&complianceID, &orgID, &directiveTitle); err != nil {
			continue
		}

		alert := domain.Alert{
			ID:          uuid.New(),
			OrgID:       orgID,
			Level:       domain.AlertCritical,
			Category:    "directive_overdue",
			Title:       fmt.Sprintf("Directive overdue: %s", directiveTitle),
			Description: "Compliance deadline has passed",
			EntityType:  "aircraft_directive_compliance",
			EntityID:    complianceID,
			CreatedAt:   now,
		}
		escalateAt := now.Add(time.Hour)
		alert.AutoEscalateAt = &escalateAt

		if _, err := t.Alerts.Create(ctx, alert); err != nil {
			t.Logger.Error().Err(err).Msg("failed to create directive overdue alert")
		}

		// Update compliance status to overdue
		_, _ = t.DB.Exec(ctx, `
			UPDATE aircraft_directive_compliance SET status = 'overdue', updated_at = $1
			WHERE id = $2 AND org_id = $3
		`, now, complianceID, orgID)
	}
}
