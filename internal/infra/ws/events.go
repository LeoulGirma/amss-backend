package ws

// WebSocket event types for real-time updates.
const (
	// Dashboard
	EventDashboardMetrics = "dashboard:metrics_snapshot"

	// Aircraft
	EventAircraftAOG       = "aircraft:aog"
	EventAircraftStatusChanged = "aircraft:status_changed"

	// Tasks
	EventTaskOverdue       = "task:overdue"
	EventTaskStateChanged  = "task:state_changed"
	EventTaskAssigned      = "task:assigned"
	EventTaskRescheduled   = "task:rescheduled"

	// Compliance
	EventCertExpiring      = "compliance:cert_expiring"
	EventDirectiveOverdue  = "compliance:directive_overdue"
	EventComplianceSignOff = "compliance:sign_off"

	// Scheduling
	EventScheduleConflict  = "schedule:conflict"

	// Alerts
	EventAlertCreated      = "alert:created"
	EventAlertEscalated    = "alert:escalated"

	// Parts
	EventPartsLowStock     = "parts:low_stock"
)
