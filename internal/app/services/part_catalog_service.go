package services

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type PartCatalogService struct {
	Definitions ports.PartDefinitionRepository
	Items       ports.PartItemRepository
	Clock       app.Clock
}

func (s *PartCatalogService) CreateDefinition(ctx context.Context, actor app.Actor, orgID uuid.UUID, name, category string) (domain.PartDefinition, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleMechanic && actor.Role != domain.RoleAdmin {
		return domain.PartDefinition{}, domain.ErrForbidden
	}
	resolvedOrg := actor.OrgID
	if actor.IsAdmin() && orgID != uuid.Nil {
		resolvedOrg = orgID
	}

	def := domain.PartDefinition{
		ID:        uuid.New(),
		OrgID:     resolvedOrg,
		Name:      name,
		Category:  category,
		CreatedAt: s.Clock.Now(),
		UpdatedAt: s.Clock.Now(),
	}

	created, err := s.Definitions.Create(ctx, def)
	if err != nil {
		return domain.PartDefinition{}, err
	}
	return created, nil
}

func (s *PartCatalogService) ListDefinitions(ctx context.Context, actor app.Actor, filter ports.PartDefinitionFilter) ([]domain.PartDefinition, error) {
	if !actor.IsAdmin() {
		filter.OrgID = &actor.OrgID
	}
	return s.Definitions.List(ctx, filter)
}

func (s *PartCatalogService) UpdateDefinition(ctx context.Context, actor app.Actor, orgID, id uuid.UUID, name, category string) (domain.PartDefinition, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleMechanic && actor.Role != domain.RoleAdmin {
		return domain.PartDefinition{}, domain.ErrForbidden
	}
	if !actor.IsAdmin() {
		orgID = actor.OrgID
	}
	def, err := s.Definitions.GetByID(ctx, orgID, id)
	if err != nil {
		return domain.PartDefinition{}, err
	}
	def.Name = name
	def.Category = category
	def.UpdatedAt = s.Clock.Now()

	updated, err := s.Definitions.Update(ctx, def)
	if err != nil {
		return domain.PartDefinition{}, err
	}
	return updated, nil
}

func (s *PartCatalogService) DeleteDefinition(ctx context.Context, actor app.Actor, orgID, id uuid.UUID) error {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleMechanic && actor.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	if !actor.IsAdmin() {
		orgID = actor.OrgID
	}
	return s.Definitions.SoftDelete(ctx, orgID, id, s.Clock.Now())
}

func (s *PartCatalogService) CreateItem(ctx context.Context, actor app.Actor, orgID uuid.UUID, defID uuid.UUID, serial string, status domain.PartItemStatus, expiry *time.Time) (domain.PartItem, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleMechanic && actor.Role != domain.RoleAdmin {
		return domain.PartItem{}, domain.ErrForbidden
	}
	resolvedOrg := actor.OrgID
	if actor.IsAdmin() && orgID != uuid.Nil {
		resolvedOrg = orgID
	}
	if status == "" {
		status = domain.PartItemInStock
	}

	item := domain.PartItem{
		ID:           uuid.New(),
		OrgID:        resolvedOrg,
		DefinitionID: defID,
		SerialNumber: serial,
		Status:       status,
		ExpiryDate:   expiry,
		CreatedAt:    s.Clock.Now(),
		UpdatedAt:    s.Clock.Now(),
	}

	created, err := s.Items.Create(ctx, item)
	if err != nil {
		return domain.PartItem{}, err
	}
	return created, nil
}

func (s *PartCatalogService) ListItems(ctx context.Context, actor app.Actor, filter ports.PartItemFilter) ([]domain.PartItem, error) {
	if !actor.IsAdmin() {
		filter.OrgID = &actor.OrgID
	}
	return s.Items.List(ctx, filter)
}

func (s *PartCatalogService) UpdateItem(ctx context.Context, actor app.Actor, orgID, id uuid.UUID, status *domain.PartItemStatus, expiry *time.Time) (domain.PartItem, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleMechanic && actor.Role != domain.RoleAdmin {
		return domain.PartItem{}, domain.ErrForbidden
	}
	if !actor.IsAdmin() {
		orgID = actor.OrgID
	}
	item, err := s.Items.GetByID(ctx, orgID, id)
	if err != nil {
		return domain.PartItem{}, err
	}
	if status != nil {
		item.Status = *status
	}
	if expiry != nil {
		item.ExpiryDate = expiry
	}
	item.UpdatedAt = s.Clock.Now()

	updated, err := s.Items.Update(ctx, item)
	if err != nil {
		return domain.PartItem{}, err
	}
	return updated, nil
}

func (s *PartCatalogService) DeleteItem(ctx context.Context, actor app.Actor, orgID, id uuid.UUID) error {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleMechanic && actor.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	if !actor.IsAdmin() {
		orgID = actor.OrgID
	}
	return s.Items.SoftDelete(ctx, orgID, id, s.Clock.Now())
}
