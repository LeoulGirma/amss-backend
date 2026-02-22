package handlers

import (
	"net/http"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
)

type dashboardMetricsResponse struct {
	FleetTotal       int     `json:"fleet_total"`
	FleetOperational int     `json:"fleet_operational"`
	FleetMaintenance int     `json:"fleet_maintenance"`
	FleetGrounded    int     `json:"fleet_grounded"`
	FleetAvailRate   float64 `json:"fleet_availability_rate"`

	TasksScheduled  int     `json:"tasks_scheduled"`
	TasksInProgress int     `json:"tasks_in_progress"`
	TasksCompleted  int     `json:"tasks_completed"`
	TasksOverdue    int     `json:"tasks_overdue"`
	OnTimeRate      float64 `json:"on_time_rate"`
	AvgTATHours     float64 `json:"avg_tat_hours"`

	PartsInStock      int     `json:"parts_in_stock"`
	PartsLowStock     int     `json:"parts_low_stock"`
	PartsExpiringSoon int     `json:"parts_expiring_soon"`
	PartsFillRate     float64 `json:"parts_fill_rate"`

	CompliancePending int `json:"compliance_pending"`
	CompliancePassed  int `json:"compliance_passed"`
	ComplianceFailed  int `json:"compliance_failed"`
	DirectivesOverdue int `json:"directives_overdue"`

	CertsExpiring30d int `json:"certs_expiring_30d"`
	CertsExpiring60d int `json:"certs_expiring_60d"`
	CertsExpiring90d int `json:"certs_expiring_90d"`
	CertsExpired     int `json:"certs_expired"`

	AlertsUnresolved int `json:"alerts_unresolved"`
	AlertsCritical   int `json:"alerts_critical"`

	MechanicsTotal int `json:"mechanics_total"`
	MechanicsOnTask int `json:"mechanics_on_task"`

	ComputedAt time.Time `json:"computed_at"`
}

func GetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Metrics == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}

	metrics, err := servicesReg.Metrics.GetDashboardMetrics(r.Context(), actor)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapMetrics(metrics))
}

func mapMetrics(m domain.DashboardMetrics) dashboardMetricsResponse {
	return dashboardMetricsResponse{
		FleetTotal:        m.FleetTotal,
		FleetOperational:  m.FleetOperational,
		FleetMaintenance:  m.FleetMaintenance,
		FleetGrounded:     m.FleetGrounded,
		FleetAvailRate:    m.FleetAvailRate,
		TasksScheduled:    m.TasksScheduled,
		TasksInProgress:   m.TasksInProgress,
		TasksCompleted:    m.TasksCompleted,
		TasksOverdue:      m.TasksOverdue,
		OnTimeRate:        m.OnTimeRate,
		AvgTATHours:       m.AvgTATHours,
		PartsInStock:      m.PartsInStock,
		PartsLowStock:     m.PartsLowStock,
		PartsExpiringSoon: m.PartsExpiringSoon,
		PartsFillRate:     m.PartsFillRate,
		CompliancePending: m.CompliancePending,
		CompliancePassed:  m.CompliancePassed,
		ComplianceFailed:  m.ComplianceFailed,
		DirectivesOverdue: m.DirectivesOverdue,
		CertsExpiring30d:  m.CertsExpiring30d,
		CertsExpiring60d:  m.CertsExpiring60d,
		CertsExpiring90d:  m.CertsExpiring90d,
		CertsExpired:      m.CertsExpired,
		AlertsUnresolved:  m.AlertsUnresolved,
		AlertsCritical:    m.AlertsCritical,
		MechanicsTotal:    m.MechanicsTotal,
		MechanicsOnTask:   m.MechanicsOnTask,
		ComputedAt:        m.ComputedAt,
	}
}
