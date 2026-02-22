package domain

import "time"

// DashboardMetrics is a point-in-time snapshot of an organization's KPIs.
type DashboardMetrics struct {
	// Fleet
	FleetTotal       int     `json:"fleet_total"`
	FleetOperational int     `json:"fleet_operational"`
	FleetMaintenance int     `json:"fleet_maintenance"`
	FleetGrounded    int     `json:"fleet_grounded"`
	FleetAvailRate   float64 `json:"fleet_availability_rate"` // operational / total

	// Tasks
	TasksScheduled  int     `json:"tasks_scheduled"`
	TasksInProgress int     `json:"tasks_in_progress"`
	TasksCompleted  int     `json:"tasks_completed"`  // last 30 days
	TasksOverdue    int     `json:"tasks_overdue"`     // end_time < now, still scheduled/in_progress
	OnTimeRate      float64 `json:"on_time_rate"`      // completed on time / total completed (30d)
	AvgTATHours     float64 `json:"avg_tat_hours"`     // average turnaround time (30d)

	// Parts
	PartsInStock       int     `json:"parts_in_stock"`
	PartsLowStock      int     `json:"parts_low_stock"`      // below min_stock_level
	PartsExpiringSoon  int     `json:"parts_expiring_soon"`  // within 30 days
	PartsFillRate      float64 `json:"parts_fill_rate"`      // reservations fulfilled / total

	// Compliance
	CompliancePending  int `json:"compliance_pending"`
	CompliancePassed   int `json:"compliance_passed"`
	ComplianceFailed   int `json:"compliance_failed"`
	DirectivesOverdue  int `json:"directives_overdue"`

	// Certifications
	CertsExpiring30d int `json:"certs_expiring_30d"`
	CertsExpiring60d int `json:"certs_expiring_60d"`
	CertsExpiring90d int `json:"certs_expiring_90d"`
	CertsExpired     int `json:"certs_expired"`

	// Alerts
	AlertsUnresolved int `json:"alerts_unresolved"`
	AlertsCritical   int `json:"alerts_critical"`

	// Mechanics
	MechanicsTotal    int `json:"mechanics_total"`
	MechanicsOnTask   int `json:"mechanics_on_task"` // assigned to in_progress tasks

	ComputedAt time.Time `json:"computed_at"`
}
