package middleware

import (
	"context"
	"net/http"

	"github.com/aeromaintain/amss/internal/app/services"
)

type servicesKey struct{}

type ServiceRegistry struct {
	Tasks         *services.TaskService
	Parts         *services.PartReservationService
	Compliance    *services.ComplianceService
	AuditQuery    *services.AuditQueryService
	Catalog       *services.PartCatalogService
	Organizations *services.OrganizationService
	Users         *services.UserService
	Aircraft      *services.AircraftService
	Programs      *services.MaintenanceProgramService
	Imports       *services.ImportService
	Webhooks      *services.WebhookService
	Policies      *services.OrgPolicyService
	Reports        *services.ReportService
	Certifications *services.CertificationService
	Directives     *services.DirectiveService
	Alerts         *services.AlertService
	Scheduling     *services.SchedulingService
}

func InjectServices(registry ServiceRegistry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), servicesKey{}, registry)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ServicesFromContext(ctx context.Context) (ServiceRegistry, bool) {
	value := ctx.Value(servicesKey{})
	if value == nil {
		return ServiceRegistry{}, false
	}
	registry, ok := value.(ServiceRegistry)
	return registry, ok
}
