package grpcapi

import (
	"strings"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/google/uuid"
)

func resolveOrgID(actor app.Actor, orgID string) (uuid.UUID, error) {
	if actor.IsAdmin() && strings.TrimSpace(orgID) != "" {
		return uuid.Parse(strings.TrimSpace(orgID))
	}
	return actor.OrgID, nil
}

func actorForOrg(actor app.Actor, orgID uuid.UUID) app.Actor {
	actor.OrgID = orgID
	return actor
}
