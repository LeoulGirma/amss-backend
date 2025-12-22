package handlers

import (
	"net/http"

	"github.com/aeromaintain/amss/internal/api/rest/middleware"
	"github.com/aeromaintain/amss/internal/app"
	"github.com/google/uuid"
)

func actorFromRequest(r *http.Request) (app.Actor, bool) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		return app.Actor{}, false
	}
	return app.Actor{
		UserID: principal.UserID,
		OrgID:  principal.OrgID,
		Role:   principal.Role,
	}, true
}

func servicesFromRequest(r *http.Request) (middleware.ServiceRegistry, bool) {
	return middleware.ServicesFromContext(r.Context())
}

func resolveOrgID(actor app.Actor, input string) (uuid.UUID, error) {
	if actor.IsAdmin() && input != "" {
		return uuid.Parse(input)
	}
	return actor.OrgID, nil
}
