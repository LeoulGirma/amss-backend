package app

import (
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type Actor struct {
	UserID uuid.UUID
	OrgID  uuid.UUID
	Role   domain.Role
}

func (a Actor) IsAdmin() bool {
	return a.Role == domain.RoleAdmin
}
