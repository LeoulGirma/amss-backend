package services

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type DirectiveService struct {
	Directives ports.DirectiveRepository
	Aircraft   ports.AircraftRepository
	Audit      ports.AuditRepository
	Clock      app.Clock
}

// --- Authorities ---

func (s *DirectiveService) ListAuthorities(ctx context.Context) ([]domain.RegulatoryAuthority, error) {
	return s.Directives.ListAuthorities(ctx)
}

// --- Directives ---

type DirectiveCreateInput struct {
	AuthorityID           uuid.UUID
	DirectiveType         domain.DirectiveType
	ReferenceNumber       string
	Title                 string
	Description           string
	Applicability         domain.DirectiveApplicability
	AffectedAircraftTypes []uuid.UUID
	EffectiveDate         time.Time
	ComplianceDeadline    *time.Time
	RecurrenceInterval    string
	SourceURL             string
}

func (s *DirectiveService) CreateDirective(ctx context.Context, actor app.Actor, input DirectiveCreateInput) (domain.ComplianceDirective, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin && actor.Role != domain.RoleAuditor {
		return domain.ComplianceDirective{}, domain.ErrForbidden
	}

	if input.ReferenceNumber == "" || input.Title == "" {
		return domain.ComplianceDirective{}, domain.NewValidationError("reference_number and title are required")
	}

	now := s.Clock.Now()
	directive := domain.ComplianceDirective{
		ID:                    uuid.New(),
		OrgID:                 actor.OrgID,
		AuthorityID:           input.AuthorityID,
		DirectiveType:         input.DirectiveType,
		ReferenceNumber:       input.ReferenceNumber,
		Title:                 input.Title,
		Description:           input.Description,
		Applicability:         input.Applicability,
		AffectedAircraftTypes: input.AffectedAircraftTypes,
		EffectiveDate:         input.EffectiveDate,
		ComplianceDeadline:    input.ComplianceDeadline,
		RecurrenceInterval:    input.RecurrenceInterval,
		SourceURL:             input.SourceURL,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	created, err := s.Directives.CreateDirective(ctx, directive)
	if err != nil {
		return domain.ComplianceDirective{}, err
	}

	if s.Audit != nil {
		entry := domain.AuditLog{
			ID:         uuid.New(),
			OrgID:      actor.OrgID,
			EntityType: "compliance_directive",
			EntityID:   created.ID,
			Action:     domain.AuditActionCreate,
			UserID:     actor.UserID,
			RequestID:  uuid.Nil,
			Timestamp:  now,
		}
		_ = s.Audit.Insert(ctx, entry)
	}

	return created, nil
}

func (s *DirectiveService) ListDirectives(ctx context.Context, actor app.Actor, filter ports.DirectiveFilter) ([]domain.ComplianceDirective, error) {
	if !actor.IsAdmin() {
		filter.OrgID = &actor.OrgID
	}
	return s.Directives.ListDirectives(ctx, filter)
}

func (s *DirectiveService) GetDirective(ctx context.Context, actor app.Actor, id uuid.UUID) (domain.ComplianceDirective, error) {
	d, err := s.Directives.GetDirectiveByID(ctx, id)
	if err != nil {
		return domain.ComplianceDirective{}, err
	}
	if !actor.IsAdmin() && actor.OrgID != d.OrgID {
		return domain.ComplianceDirective{}, domain.ErrForbidden
	}
	return d, nil
}

// --- Aircraft Compliance ---

func (s *DirectiveService) ListAircraftCompliance(ctx context.Context, actor app.Actor, filter ports.AircraftComplianceFilter) ([]domain.AircraftDirectiveCompliance, error) {
	if !actor.IsAdmin() {
		filter.OrgID = &actor.OrgID
	}
	return s.Directives.ListAircraftCompliance(ctx, filter)
}

type ComplianceUpdateInput struct {
	AircraftID  uuid.UUID
	DirectiveID uuid.UUID
	Status      domain.DirectiveComplianceStatus
	TaskID      *uuid.UUID
	Notes       string
}

func (s *DirectiveService) UpdateAircraftCompliance(ctx context.Context, actor app.Actor, input ComplianceUpdateInput) (domain.AircraftDirectiveCompliance, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin &&
		actor.Role != domain.RoleScheduler && actor.Role != domain.RoleMechanic {
		return domain.AircraftDirectiveCompliance{}, domain.ErrForbidden
	}

	now := s.Clock.Now()
	compliance := domain.AircraftDirectiveCompliance{
		ID:          uuid.New(),
		OrgID:       actor.OrgID,
		AircraftID:  input.AircraftID,
		DirectiveID: input.DirectiveID,
		Status:      input.Status,
		TaskID:      input.TaskID,
		Notes:       input.Notes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if input.Status == domain.ComplianceStatusCompliant {
		compliance.ComplianceDate = &now
		compliance.SignedOffBy = &actor.UserID
		compliance.SignedOffAt = &now

		// For recurring directives, compute the next due date
		directive, err := s.Directives.GetDirectiveByID(ctx, input.DirectiveID)
		if err == nil && directive.RecurrenceInterval != "" {
			nextDue := computeNextDue(now, directive.RecurrenceInterval)
			if nextDue != nil {
				compliance.NextDueDate = nextDue
			}
		}
	}

	return s.Directives.UpsertAircraftCompliance(ctx, compliance)
}

// ScanFleetForDirective creates compliance records for all aircraft affected by a directive
func (s *DirectiveService) ScanFleetForDirective(ctx context.Context, actor app.Actor, directiveID uuid.UUID) (int, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return 0, domain.ErrForbidden
	}

	directive, err := s.Directives.GetDirectiveByID(ctx, directiveID)
	if err != nil {
		return 0, err
	}

	// Get all aircraft in the org
	aircraftList, err := s.Aircraft.List(ctx, ports.AircraftFilter{OrgID: &actor.OrgID, Limit: 200})
	if err != nil {
		return 0, err
	}

	// Build set of affected aircraft type IDs
	affectedTypes := make(map[uuid.UUID]bool)
	for _, atID := range directive.AffectedAircraftTypes {
		affectedTypes[atID] = true
	}

	now := s.Clock.Now()
	count := 0
	for _, ac := range aircraftList {
		// If directive has specific types and aircraft has a type set, check match
		if len(affectedTypes) > 0 && ac.AircraftTypeID != nil {
			if !affectedTypes[*ac.AircraftTypeID] {
				continue
			}
		}

		compliance := domain.AircraftDirectiveCompliance{
			ID:          uuid.New(),
			OrgID:       actor.OrgID,
			AircraftID:  ac.ID,
			DirectiveID: directiveID,
			Status:      domain.ComplianceStatusPending,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		// Set initial next_due_date from directive's compliance deadline
		if directive.ComplianceDeadline != nil {
			t := time.Date(directive.ComplianceDeadline.Year(), directive.ComplianceDeadline.Month(), directive.ComplianceDeadline.Day(), 0, 0, 0, 0, time.UTC)
			compliance.NextDueDate = &t
		}
		if _, err := s.Directives.UpsertAircraftCompliance(ctx, compliance); err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

// computeNextDue parses a recurrence interval string and returns the next due date.
// Supported formats: "30d", "90d", "180d", "365d", "1y", "2y", "6m", "12m"
func computeNextDue(from time.Time, interval string) *time.Time {
	if len(interval) < 2 {
		return nil
	}

	unit := interval[len(interval)-1]
	numStr := interval[:len(interval)-1]

	var num int
	for _, ch := range numStr {
		if ch < '0' || ch > '9' {
			return nil
		}
		num = num*10 + int(ch-'0')
	}
	if num <= 0 {
		return nil
	}

	var next time.Time
	switch unit {
	case 'd':
		next = from.AddDate(0, 0, num)
	case 'm':
		next = from.AddDate(0, num, 0)
	case 'y':
		next = from.AddDate(num, 0, 0)
	default:
		return nil
	}
	return &next
}

// --- Templates ---

func (s *DirectiveService) ListTemplates(ctx context.Context, authorityID uuid.UUID) ([]domain.ComplianceTemplate, error) {
	return s.Directives.ListTemplates(ctx, authorityID)
}
