package domain

import (
	"time"

	"github.com/google/uuid"
)

// CertificationAuthority represents a regulatory body
type CertificationAuthority string

const (
	AuthorityFAA   CertificationAuthority = "faa"
	AuthorityEASA  CertificationAuthority = "easa"
	AuthorityECAA  CertificationAuthority = "ecaa"
	AuthorityICAO  CertificationAuthority = "icao"
	AuthorityOther CertificationAuthority = "other"
)

// CertificationStatus represents the status of a certification
type CertificationStatus string

const (
	CertStatusActive    CertificationStatus = "active"
	CertStatusExpired   CertificationStatus = "expired"
	CertStatusSuspended CertificationStatus = "suspended"
	CertStatusRevoked   CertificationStatus = "revoked"
)

// TypeRatingStatus represents the status of a type rating
type TypeRatingStatus string

const (
	TypeRatingActive  TypeRatingStatus = "active"
	TypeRatingLapsed  TypeRatingStatus = "lapsed"
)

// SkillCategory represents the category of a specialized skill
type SkillCategory string

const (
	SkillCategoryStructural SkillCategory = "structural"
	SkillCategoryAvionics   SkillCategory = "avionics"
	SkillCategoryEngine     SkillCategory = "engine"
	SkillCategoryNDT        SkillCategory = "ndt"
	SkillCategoryGeneral    SkillCategory = "general"
)

// CertificationType defines a type of certification recognized by the system
type CertificationType struct {
	ID                    uuid.UUID
	Code                  string
	Name                  string
	Authority             CertificationAuthority
	HasExpiry             bool
	RecencyRequiredMonths *int
	RecencyPeriodMonths   *int
	CreatedAt             time.Time
}

// AircraftType defines an aircraft type for type rating purposes
type AircraftType struct {
	ID           uuid.UUID
	ICAOCode     string
	Manufacturer string
	Model        string
	Series       string
	CreatedAt    time.Time
}

// EmployeeCertification represents a mechanic's certification
type EmployeeCertification struct {
	ID                uuid.UUID
	OrgID             uuid.UUID
	UserID            uuid.UUID
	CertTypeID        uuid.UUID
	CertificateNumber string
	IssueDate         time.Time
	ExpiryDate        *time.Time
	Status            CertificationStatus
	VerifiedBy        *uuid.UUID
	VerifiedAt        *time.Time
	DocumentURL       string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// IsExpired checks if the certification has expired
func (c EmployeeCertification) IsExpired(now time.Time) bool {
	if c.ExpiryDate == nil {
		return false
	}
	return now.After(*c.ExpiryDate)
}

// IsActive checks if the certification is currently active and not expired
func (c EmployeeCertification) IsActive(now time.Time) bool {
	return c.Status == CertStatusActive && !c.IsExpired(now)
}

// EmployeeTypeRating represents a mechanic's type rating for a specific aircraft type
type EmployeeTypeRating struct {
	ID             uuid.UUID
	OrgID          uuid.UUID
	UserID         uuid.UUID
	AircraftTypeID uuid.UUID
	CertID         uuid.UUID
	EndorsementDate time.Time
	Status         TypeRatingStatus
	CreatedAt      time.Time
}

// SkillType defines a specialized skill
type SkillType struct {
	ID        uuid.UUID
	Code      string
	Name      string
	Category  SkillCategory
	CreatedAt time.Time
}

// EmployeeSkill represents a mechanic's specialized skill
type EmployeeSkill struct {
	ID               uuid.UUID
	OrgID            uuid.UUID
	UserID           uuid.UUID
	SkillTypeID      uuid.UUID
	ProficiencyLevel int
	QualificationDate time.Time
	ExpiryDate       *time.Time
	CreatedAt        time.Time
}

// IsExpired checks if the skill qualification has expired
func (s EmployeeSkill) IsExpired(now time.Time) bool {
	if s.ExpiryDate == nil {
		return false
	}
	return now.After(*s.ExpiryDate)
}

// EmployeeRecencyLog tracks maintenance hours for recency compliance
type EmployeeRecencyLog struct {
	ID             uuid.UUID
	OrgID          uuid.UUID
	UserID         uuid.UUID
	AircraftTypeID uuid.UUID
	TaskID         uuid.UUID
	WorkDate       time.Time
	HoursLogged    float64
	CreatedAt      time.Time
}

// TaskSkillRequirement defines what certifications/skills a task requires
type TaskSkillRequirement struct {
	ID                  uuid.UUID
	OrgID               uuid.UUID
	TaskType            TaskType
	AircraftTypeID      *uuid.UUID
	CertTypeID          *uuid.UUID
	SkillTypeID         *uuid.UUID
	MinProficiencyLevel int
	IsCertifyingRole    bool
	IsInspectionRole    bool
	CreatedAt           time.Time
}

// QualificationCheckResult holds the result of a mechanic qualification check
type QualificationCheckResult struct {
	Qualified          bool
	MissingCerts       []string
	MissingTypeRatings []string
	MissingSkills      []string
	ExpiredCerts       []string
	InsufficientRecency bool
	Reasons            []string
}
