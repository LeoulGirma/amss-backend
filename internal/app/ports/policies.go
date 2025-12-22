package ports

import (
	"context"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type OrgPolicyRepository interface {
	GetByOrgID(ctx context.Context, orgID uuid.UUID) (domain.OrgPolicy, error)
	Upsert(ctx context.Context, policy domain.OrgPolicy) (domain.OrgPolicy, error)
}
