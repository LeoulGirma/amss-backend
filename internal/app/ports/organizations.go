package ports

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type OrganizationRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.Organization, error)
	Create(ctx context.Context, org domain.Organization) (domain.Organization, error)
	Update(ctx context.Context, org domain.Organization) (domain.Organization, error)
	SoftDelete(ctx context.Context, id uuid.UUID, at time.Time) error
	List(ctx context.Context, filter OrganizationFilter) ([]domain.Organization, error)
}

type OrganizationFilter struct {
	Name   string
	Limit  int
	Offset int
}
