package services

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type CertificationService struct {
	Certs         ports.CertificationRepository
	AircraftTypes ports.AircraftTypeRepository
	Audit         ports.AuditRepository
	Clock         app.Clock
}

// --- Certification Types (reference data) ---

func (s *CertificationService) ListCertTypes(ctx context.Context) ([]domain.CertificationType, error) {
	return s.Certs.ListCertTypes(ctx)
}

// --- Employee Certifications ---

type CertCreateInput struct {
	UserID            uuid.UUID
	CertTypeID        uuid.UUID
	CertificateNumber string
	IssueDate         time.Time
	ExpiryDate        *time.Time
	DocumentURL       string
}

func (s *CertificationService) CreateCert(ctx context.Context, actor app.Actor, input CertCreateInput) (domain.EmployeeCertification, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return domain.EmployeeCertification{}, domain.ErrForbidden
	}

	// Verify cert type exists
	if _, err := s.Certs.GetCertTypeByID(ctx, input.CertTypeID); err != nil {
		return domain.EmployeeCertification{}, err
	}

	now := s.Clock.Now()
	cert := domain.EmployeeCertification{
		ID:                uuid.New(),
		OrgID:             actor.OrgID,
		UserID:            input.UserID,
		CertTypeID:        input.CertTypeID,
		CertificateNumber: input.CertificateNumber,
		IssueDate:         input.IssueDate,
		ExpiryDate:        input.ExpiryDate,
		Status:            domain.CertStatusActive,
		DocumentURL:       input.DocumentURL,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	created, err := s.Certs.CreateCert(ctx, cert)
	if err != nil {
		return domain.EmployeeCertification{}, err
	}

	if s.Audit != nil {
		entry := domain.AuditLog{
			ID:         uuid.New(),
			OrgID:      actor.OrgID,
			EntityType: "employee_certification",
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

func (s *CertificationService) ListCertsByUser(ctx context.Context, actor app.Actor, userID uuid.UUID) ([]domain.EmployeeCertification, error) {
	if !actor.IsAdmin() && actor.Role != domain.RoleTenantAdmin && actor.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return s.Certs.ListCertsByUser(ctx, actor.OrgID, userID)
}

func (s *CertificationService) UpdateCert(ctx context.Context, actor app.Actor, certID uuid.UUID, status *domain.CertificationStatus, expiryDate *time.Time, documentURL *string) (domain.EmployeeCertification, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return domain.EmployeeCertification{}, domain.ErrForbidden
	}

	cert, err := s.Certs.GetCertByID(ctx, actor.OrgID, certID)
	if err != nil {
		return domain.EmployeeCertification{}, err
	}

	if status != nil {
		cert.Status = *status
	}
	if expiryDate != nil {
		cert.ExpiryDate = expiryDate
	}
	if documentURL != nil {
		cert.DocumentURL = *documentURL
	}
	cert.UpdatedAt = s.Clock.Now()

	return s.Certs.UpdateCert(ctx, cert)
}

func (s *CertificationService) ListExpiringCerts(ctx context.Context, actor app.Actor, daysAhead int) ([]domain.EmployeeCertification, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	before := s.Clock.Now().AddDate(0, 0, daysAhead)
	return s.Certs.ListExpiringCerts(ctx, actor.OrgID, before)
}

// --- Type Ratings ---

type TypeRatingCreateInput struct {
	UserID         uuid.UUID
	AircraftTypeID uuid.UUID
	CertID         uuid.UUID
	EndorsementDate time.Time
}

func (s *CertificationService) CreateTypeRating(ctx context.Context, actor app.Actor, input TypeRatingCreateInput) (domain.EmployeeTypeRating, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return domain.EmployeeTypeRating{}, domain.ErrForbidden
	}

	// Verify the certification exists and belongs to the user
	cert, err := s.Certs.GetCertByID(ctx, actor.OrgID, input.CertID)
	if err != nil {
		return domain.EmployeeTypeRating{}, err
	}
	if cert.UserID != input.UserID {
		return domain.EmployeeTypeRating{}, domain.NewValidationError("certification does not belong to user")
	}

	rating := domain.EmployeeTypeRating{
		ID:              uuid.New(),
		OrgID:           actor.OrgID,
		UserID:          input.UserID,
		AircraftTypeID:  input.AircraftTypeID,
		CertID:          input.CertID,
		EndorsementDate: input.EndorsementDate,
		Status:          domain.TypeRatingActive,
		CreatedAt:       s.Clock.Now(),
	}

	return s.Certs.CreateTypeRating(ctx, rating)
}

func (s *CertificationService) ListTypeRatingsByUser(ctx context.Context, actor app.Actor, userID uuid.UUID) ([]domain.EmployeeTypeRating, error) {
	if !actor.IsAdmin() && actor.Role != domain.RoleTenantAdmin && actor.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return s.Certs.ListTypeRatingsByUser(ctx, actor.OrgID, userID)
}

// --- Skills ---

type SkillCreateInput struct {
	UserID            uuid.UUID
	SkillTypeID       uuid.UUID
	ProficiencyLevel  int
	QualificationDate time.Time
	ExpiryDate        *time.Time
}

func (s *CertificationService) CreateSkill(ctx context.Context, actor app.Actor, input SkillCreateInput) (domain.EmployeeSkill, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return domain.EmployeeSkill{}, domain.ErrForbidden
	}

	if input.ProficiencyLevel < 1 || input.ProficiencyLevel > 5 {
		return domain.EmployeeSkill{}, domain.NewValidationError("proficiency_level must be between 1 and 5")
	}

	skill := domain.EmployeeSkill{
		ID:                uuid.New(),
		OrgID:             actor.OrgID,
		UserID:            input.UserID,
		SkillTypeID:       input.SkillTypeID,
		ProficiencyLevel:  input.ProficiencyLevel,
		QualificationDate: input.QualificationDate,
		ExpiryDate:        input.ExpiryDate,
		CreatedAt:         s.Clock.Now(),
	}

	return s.Certs.CreateSkill(ctx, skill)
}

func (s *CertificationService) ListSkillsByUser(ctx context.Context, actor app.Actor, userID uuid.UUID) ([]domain.EmployeeSkill, error) {
	if !actor.IsAdmin() && actor.Role != domain.RoleTenantAdmin && actor.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return s.Certs.ListSkillsByUser(ctx, actor.OrgID, userID)
}

// --- Qualification Check ---

func (s *CertificationService) CheckQualification(ctx context.Context, orgID, userID uuid.UUID, taskType domain.TaskType, aircraftTypeID *uuid.UUID) (domain.QualificationCheckResult, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	now := s.Clock.Now()
	result := domain.QualificationCheckResult{Qualified: true}

	// Get requirements for this task type
	requirements, err := s.Certs.ListRequirements(ctx, orgID, taskType, aircraftTypeID)
	if err != nil {
		return result, err
	}

	// If no requirements defined, any mechanic qualifies
	if len(requirements) == 0 {
		return result, nil
	}

	// Get user's certifications
	certs, err := s.Certs.ListCertsByUser(ctx, orgID, userID)
	if err != nil {
		return result, err
	}
	certMap := make(map[uuid.UUID]domain.EmployeeCertification)
	for _, c := range certs {
		certMap[c.CertTypeID] = c
	}

	// Get user's skills
	skills, err := s.Certs.ListSkillsByUser(ctx, orgID, userID)
	if err != nil {
		return result, err
	}
	skillMap := make(map[uuid.UUID]domain.EmployeeSkill)
	for _, sk := range skills {
		skillMap[sk.SkillTypeID] = sk
	}

	for _, req := range requirements {
		// Check certification requirement
		if req.CertTypeID != nil {
			cert, ok := certMap[*req.CertTypeID]
			if !ok {
				result.Qualified = false
				result.MissingCerts = append(result.MissingCerts, req.CertTypeID.String())
				result.Reasons = append(result.Reasons, "missing required certification")
				continue
			}
			if !cert.IsActive(now) {
				result.Qualified = false
				result.ExpiredCerts = append(result.ExpiredCerts, req.CertTypeID.String())
				result.Reasons = append(result.Reasons, "certification expired or inactive")
				continue
			}

			// Check recency if required
			certType, err := s.Certs.GetCertTypeByID(ctx, *req.CertTypeID)
			if err == nil && certType.RecencyRequiredMonths != nil && certType.RecencyPeriodMonths != nil && aircraftTypeID != nil {
				since := now.AddDate(0, -*certType.RecencyPeriodMonths, 0)
				hours, err := s.Certs.GetRecencyHours(ctx, orgID, userID, *aircraftTypeID, since)
				if err == nil {
					// Convert required months to approximate hours (160 hours/month)
					requiredHours := float64(*certType.RecencyRequiredMonths) * 160
					if hours < requiredHours {
						result.Qualified = false
						result.InsufficientRecency = true
						result.Reasons = append(result.Reasons, "insufficient recency hours")
					}
				}
			}
		}

		// Check skill requirement
		if req.SkillTypeID != nil {
			skill, ok := skillMap[*req.SkillTypeID]
			if !ok {
				result.Qualified = false
				result.MissingSkills = append(result.MissingSkills, req.SkillTypeID.String())
				result.Reasons = append(result.Reasons, "missing required skill")
				continue
			}
			if skill.IsExpired(now) {
				result.Qualified = false
				result.MissingSkills = append(result.MissingSkills, req.SkillTypeID.String())
				result.Reasons = append(result.Reasons, "skill qualification expired")
				continue
			}
			if skill.ProficiencyLevel < req.MinProficiencyLevel {
				result.Qualified = false
				result.MissingSkills = append(result.MissingSkills, req.SkillTypeID.String())
				result.Reasons = append(result.Reasons, "insufficient proficiency level")
			}
		}
	}

	// Check type rating if aircraft type is specified
	if aircraftTypeID != nil {
		hasRating, err := s.Certs.HasTypeRating(ctx, orgID, userID, *aircraftTypeID)
		if err != nil {
			return result, err
		}
		if !hasRating {
			result.Qualified = false
			result.MissingTypeRatings = append(result.MissingTypeRatings, aircraftTypeID.String())
			result.Reasons = append(result.Reasons, "missing type rating for aircraft")
		}
	}

	return result, nil
}

// --- Qualified Mechanics Lookup ---

func (s *CertificationService) GetQualifiedMechanics(ctx context.Context, actor app.Actor, taskType domain.TaskType, aircraftTypeID *uuid.UUID) ([]uuid.UUID, error) {
	return s.Certs.GetQualifiedMechanics(ctx, actor.OrgID, taskType, aircraftTypeID)
}

// --- Task Skill Requirements ---

type RequirementCreateInput struct {
	TaskType            domain.TaskType
	AircraftTypeID      *uuid.UUID
	CertTypeID          *uuid.UUID
	SkillTypeID         *uuid.UUID
	MinProficiencyLevel int
	IsCertifyingRole    bool
	IsInspectionRole    bool
}

func (s *CertificationService) CreateRequirement(ctx context.Context, actor app.Actor, input RequirementCreateInput) (domain.TaskSkillRequirement, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return domain.TaskSkillRequirement{}, domain.ErrForbidden
	}

	if input.CertTypeID == nil && input.SkillTypeID == nil {
		return domain.TaskSkillRequirement{}, domain.NewValidationError("cert_type_id or skill_type_id is required")
	}

	req := domain.TaskSkillRequirement{
		ID:                  uuid.New(),
		OrgID:               actor.OrgID,
		TaskType:            input.TaskType,
		AircraftTypeID:      input.AircraftTypeID,
		CertTypeID:          input.CertTypeID,
		SkillTypeID:         input.SkillTypeID,
		MinProficiencyLevel: input.MinProficiencyLevel,
		IsCertifyingRole:    input.IsCertifyingRole,
		IsInspectionRole:    input.IsInspectionRole,
		CreatedAt:           s.Clock.Now(),
	}

	return s.Certs.CreateRequirement(ctx, req)
}
