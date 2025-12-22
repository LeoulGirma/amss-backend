package handlers

import (
	"net/http"
	"time"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type reportSummaryResponse struct {
	Tasks      reportTaskSummary       `json:"tasks"`
	Aircraft   reportAircraftSummary   `json:"aircraft"`
	Parts      reportPartsSummary      `json:"parts"`
	Compliance reportComplianceSummary `json:"compliance"`
}

type reportTaskSummary struct {
	Scheduled  int `json:"scheduled"`
	InProgress int `json:"in_progress"`
	Completed  int `json:"completed"`
	Cancelled  int `json:"cancelled"`
}

type reportAircraftSummary struct {
	Total int `json:"total"`
}

type reportPartsSummary struct {
	InStock  int `json:"in_stock"`
	Used     int `json:"used"`
	Disposed int `json:"disposed"`
}

type reportComplianceSummary struct {
	Pending int `json:"pending"`
	Signed  int `json:"signed"`
}

type reportComplianceResponse struct {
	Total    int `json:"total"`
	Pass     int `json:"pass"`
	Fail     int `json:"fail"`
	Pending  int `json:"pending"`
	Signed   int `json:"signed"`
	Unsigned int `json:"unsigned"`
}

func GetReportSummary(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Reports == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	orgID := actor.OrgID
	if actor.IsAdmin() {
		if org := r.URL.Query().Get("org_id"); org != "" {
			parsed, err := uuid.Parse(org)
			if err != nil {
				writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
				return
			}
			orgID = parsed
		}
	}
	summary, err := servicesReg.Reports.Summary(r.Context(), actor, orgID)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, reportSummaryResponse{
		Tasks: reportTaskSummary{
			Scheduled:  summary.TasksScheduled,
			InProgress: summary.TasksInProgress,
			Completed:  summary.TasksCompleted,
			Cancelled:  summary.TasksCancelled,
		},
		Aircraft: reportAircraftSummary{
			Total: summary.AircraftTotal,
		},
		Parts: reportPartsSummary{
			InStock:  summary.PartsInStock,
			Used:     summary.PartsUsed,
			Disposed: summary.PartsDisposed,
		},
		Compliance: reportComplianceSummary{
			Pending: summary.CompliancePending,
			Signed:  summary.ComplianceSigned,
		},
	})
}

func GetComplianceReport(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Reports == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	filter, err := parseComplianceReportFilter(r, actor)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	report, err := servicesReg.Reports.Compliance(r.Context(), actor, filter)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, reportComplianceResponse{
		Total:    report.Total,
		Pass:     report.Pass,
		Fail:     report.Fail,
		Pending:  report.Pending,
		Signed:   report.Signed,
		Unsigned: report.Unsigned,
	})
}

func parseComplianceReportFilter(r *http.Request, actor app.Actor) (ports.ComplianceReportFilter, error) {
	query := r.URL.Query()
	orgID := actor.OrgID
	if actor.IsAdmin() {
		if org := query.Get("org_id"); org != "" {
			parsed, err := uuid.Parse(org)
			if err != nil {
				return ports.ComplianceReportFilter{}, errInvalid("invalid org_id")
			}
			orgID = parsed
		}
	}
	filter := ports.ComplianceReportFilter{OrgID: orgID}
	if taskID := query.Get("task_id"); taskID != "" {
		id, err := uuid.Parse(taskID)
		if err != nil {
			return ports.ComplianceReportFilter{}, errInvalid("invalid task_id")
		}
		filter.TaskID = &id
	}
	if result := query.Get("result"); result != "" {
		value := domain.ComplianceResult(result)
		switch value {
		case domain.CompliancePass, domain.ComplianceFail, domain.CompliancePending:
		default:
			return ports.ComplianceReportFilter{}, errInvalid("invalid result")
		}
		filter.Result = &value
	}
	if signed := query.Get("signed"); signed != "" {
		value, err := parseBool(signed)
		if err != nil {
			return ports.ComplianceReportFilter{}, errInvalid("invalid signed")
		}
		filter.Signed = &value
	}
	if from := query.Get("from"); from != "" {
		value, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return ports.ComplianceReportFilter{}, errInvalid("invalid from")
		}
		filter.From = &value
	}
	if to := query.Get("to"); to != "" {
		value, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return ports.ComplianceReportFilter{}, errInvalid("invalid to")
		}
		filter.To = &value
	}
	return filter, nil
}
