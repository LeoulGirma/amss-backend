package services

import (
	"context"
	"errors"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
)

type AuditQueryService struct {
	Repo ports.AuditQueryRepository
}

func (s *AuditQueryService) List(ctx context.Context, actor app.Actor, filter ports.AuditLogFilter) ([]domain.AuditLog, error) {
	if s == nil || s.Repo == nil {
		return nil, errors.New("audit repository unavailable")
	}
	if !actor.IsAdmin() {
		filter.OrgID = &actor.OrgID
	}
	return s.Repo.List(ctx, filter)
}
