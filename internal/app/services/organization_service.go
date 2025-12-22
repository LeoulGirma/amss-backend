package services

import (
	"context"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type OrganizationService struct {
	Organizations ports.OrganizationRepository
	Clock         app.Clock
}

func (s *OrganizationService) Create(ctx context.Context, actor app.Actor, name string) (domain.Organization, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleAdmin {
		return domain.Organization{}, domain.ErrForbidden
	}
	if name == "" {
		return domain.Organization{}, domain.NewValidationError("name is required")
	}
	org := domain.Organization{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: s.Clock.Now(),
		UpdatedAt: s.Clock.Now(),
	}
	return s.Organizations.Create(ctx, org)
}

func (s *OrganizationService) Get(ctx context.Context, actor app.Actor, id uuid.UUID) (domain.Organization, error) {
	if actor.Role == domain.RoleAdmin {
		return s.Organizations.GetByID(ctx, id)
	}
	if actor.Role != domain.RoleTenantAdmin {
		return domain.Organization{}, domain.ErrForbidden
	}
	if actor.OrgID != id {
		return domain.Organization{}, domain.ErrForbidden
	}
	return s.Organizations.GetByID(ctx, id)
}

func (s *OrganizationService) List(ctx context.Context, actor app.Actor, filter ports.OrganizationFilter) ([]domain.Organization, error) {
	if actor.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	return s.Organizations.List(ctx, filter)
}

func (s *OrganizationService) Update(ctx context.Context, actor app.Actor, id uuid.UUID, name string) (domain.Organization, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if name == "" {
		return domain.Organization{}, domain.NewValidationError("name is required")
	}
	if actor.Role == domain.RoleAdmin {
		org, err := s.Organizations.GetByID(ctx, id)
		if err != nil {
			return domain.Organization{}, err
		}
		org.Name = name
		org.UpdatedAt = s.Clock.Now()
		return s.Organizations.Update(ctx, org)
	}
	if actor.Role != domain.RoleTenantAdmin {
		return domain.Organization{}, domain.ErrForbidden
	}
	if actor.OrgID != id {
		return domain.Organization{}, domain.ErrForbidden
	}
	org, err := s.Organizations.GetByID(ctx, id)
	if err != nil {
		return domain.Organization{}, err
	}
	org.Name = name
	org.UpdatedAt = s.Clock.Now()
	return s.Organizations.Update(ctx, org)
}
