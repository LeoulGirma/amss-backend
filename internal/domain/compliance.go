package domain

import (
	"time"

	"github.com/google/uuid"
)

type ComplianceResult string

const (
	CompliancePass    ComplianceResult = "pass"
	ComplianceFail    ComplianceResult = "fail"
	CompliancePending ComplianceResult = "pending"
)

type ComplianceItem struct {
	ID            uuid.UUID
	OrgID         uuid.UUID
	TaskID        uuid.UUID
	Description   string
	Result        ComplianceResult
	SignOffUserID *uuid.UUID
	SignOffTime   *time.Time
	DeletedAt     *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (c ComplianceItem) CanSignOff(role Role) error {
	if role != RoleMechanic {
		return ErrForbidden
	}
	if c.SignOffTime != nil {
		return NewConflictError("compliance item already signed off")
	}
	if c.Result == CompliancePending {
		return NewValidationError("result must be pass or fail before sign off")
	}
	return nil
}
