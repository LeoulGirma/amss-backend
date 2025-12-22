package services

import (
	"context"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type ImportService struct {
	Imports ports.ImportRepository
	Rows    ports.ImportRowRepository
	Jobs    ports.ImportJobQueue
	Clock   app.Clock
}

type ImportCreateInput struct {
	ID        uuid.UUID
	OrgID     *uuid.UUID
	Type      domain.ImportType
	FileName  string
	FilePath  string
	CreatedBy uuid.UUID
}

func (s *ImportService) Create(ctx context.Context, actor app.Actor, input ImportCreateInput) (domain.Import, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleTenantAdmin && actor.Role != domain.RoleAdmin {
		return domain.Import{}, domain.ErrForbidden
	}
	if input.Type == "" {
		return domain.Import{}, domain.NewValidationError("type is required")
	}
	if input.FileName == "" || input.FilePath == "" {
		return domain.Import{}, domain.NewValidationError("file_name and file_path are required")
	}
	orgID := actor.OrgID
	if actor.IsAdmin() && input.OrgID != nil && *input.OrgID != uuid.Nil {
		orgID = *input.OrgID
	}
	impID := input.ID
	if impID == uuid.Nil {
		impID = uuid.New()
	}
	imp := domain.Import{
		ID:        impID,
		OrgID:     orgID,
		Type:      input.Type,
		Status:    domain.ImportStatusPending,
		FileName:  input.FileName,
		FilePath:  input.FilePath,
		CreatedBy: input.CreatedBy,
		CreatedAt: s.Clock.Now(),
		UpdatedAt: s.Clock.Now(),
	}
	created, err := s.Imports.Create(ctx, imp)
	if err != nil {
		return domain.Import{}, err
	}
	if s.Jobs != nil {
		_ = s.Jobs.Enqueue(ctx, created.ID)
	}
	return created, nil
}

func (s *ImportService) Get(ctx context.Context, actor app.Actor, orgID, id uuid.UUID) (domain.Import, error) {
	if !actor.IsAdmin() && orgID != actor.OrgID {
		return domain.Import{}, domain.ErrForbidden
	}
	return s.Imports.GetByID(ctx, orgID, id)
}

func (s *ImportService) ListRows(ctx context.Context, actor app.Actor, filter ports.ImportRowFilter) ([]domain.ImportRow, error) {
	if !actor.IsAdmin() && filter.OrgID != actor.OrgID {
		return nil, domain.ErrForbidden
	}
	if s.Rows == nil {
		return nil, domain.NewValidationError("import rows unavailable")
	}
	return s.Rows.ListByImport(ctx, filter)
}

func (s *ImportService) UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status domain.ImportStatus, summary map[string]any) error {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	return s.Imports.UpdateStatus(ctx, orgID, id, status, summary, s.Clock.Now())
}
