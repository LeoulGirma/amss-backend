package domain

import (
	"time"

	"github.com/google/uuid"
)

// DirectiveType represents the type of compliance directive
type DirectiveType string

const (
	DirectiveTypeAD   DirectiveType = "ad"
	DirectiveTypeSB   DirectiveType = "sb"
	DirectiveTypeEO   DirectiveType = "eo"
	DirectiveTypeTCDS DirectiveType = "tcds"
	DirectiveTypeSTC  DirectiveType = "stc"
)

// DirectiveApplicability indicates whether a directive is mandatory
type DirectiveApplicability string

const (
	DirectiveMandatory   DirectiveApplicability = "mandatory"
	DirectiveRecommended DirectiveApplicability = "recommended"
	DirectiveOptional    DirectiveApplicability = "optional"
)

// DirectiveComplianceStatus represents the compliance status per aircraft
type DirectiveComplianceStatus string

const (
	ComplianceStatusPending       DirectiveComplianceStatus = "pending"
	ComplianceStatusInProgress    DirectiveComplianceStatus = "in_progress"
	ComplianceStatusCompliant     DirectiveComplianceStatus = "compliant"
	ComplianceStatusNotApplicable DirectiveComplianceStatus = "not_applicable"
	ComplianceStatusOverdue       DirectiveComplianceStatus = "overdue"
)

// RegulatoryAuthority represents a regulatory body
type RegulatoryAuthority struct {
	ID                   uuid.UUID
	Code                 string
	Name                 string
	Country              string
	RecordRetentionYears int
	CreatedAt            time.Time
}

// OrgRegulatoryRegistration tracks an organization's registration with an authority
type OrgRegulatoryRegistration struct {
	ID                 uuid.UUID
	OrgID              uuid.UUID
	AuthorityID        uuid.UUID
	RegistrationNumber string
	Scope              string
	EffectiveDate      time.Time
	ExpiryDate         *time.Time
	Status             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// ComplianceDirective represents an AD, SB, or other directive
type ComplianceDirective struct {
	ID                    uuid.UUID
	OrgID                 uuid.UUID
	AuthorityID           uuid.UUID
	DirectiveType         DirectiveType
	ReferenceNumber       string
	Title                 string
	Description           string
	Applicability         DirectiveApplicability
	AffectedAircraftTypes []uuid.UUID
	EffectiveDate         time.Time
	ComplianceDeadline    *time.Time
	RecurrenceInterval    string
	SupersededBy          *uuid.UUID
	SourceURL             string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// IsOverdue checks if the directive compliance deadline has passed
func (d ComplianceDirective) IsOverdue(now time.Time) bool {
	if d.ComplianceDeadline == nil {
		return false
	}
	return now.After(*d.ComplianceDeadline)
}

// AircraftDirectiveCompliance tracks compliance status for a specific aircraft
type AircraftDirectiveCompliance struct {
	ID             uuid.UUID
	OrgID          uuid.UUID
	AircraftID     uuid.UUID
	DirectiveID    uuid.UUID
	Status         DirectiveComplianceStatus
	ComplianceDate *time.Time
	NextDueDate    *time.Time
	TaskID         *uuid.UUID
	SignedOffBy    *uuid.UUID
	SignedOffAt    *time.Time
	Notes          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ComplianceTemplate defines a document template for a regulatory authority
type ComplianceTemplate struct {
	ID              uuid.UUID
	AuthorityID     uuid.UUID
	TemplateCode    string
	Name            string
	Description     string
	RequiredFields  map[string]any
	TemplateContent string
	CreatedAt       time.Time
}
