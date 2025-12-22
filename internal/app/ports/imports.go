package ports

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type ImportRepository interface {
	GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.Import, error)
	Create(ctx context.Context, imp domain.Import) (domain.Import, error)
	Update(ctx context.Context, imp domain.Import) (domain.Import, error)
	UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status domain.ImportStatus, summary map[string]any, updatedAt time.Time) error
}

type ImportRowRepository interface {
	Create(ctx context.Context, row domain.ImportRow) error
	Update(ctx context.Context, row domain.ImportRow) error
	ListByImport(ctx context.Context, filter ImportRowFilter) ([]domain.ImportRow, error)
}

type ImportRowFilter struct {
	OrgID    uuid.UUID
	ImportID uuid.UUID
	Status   *domain.ImportRowStatus
	Limit    int
	Offset   int
}

type ImportJobQueue interface {
	Enqueue(ctx context.Context, importID uuid.UUID) error
}
