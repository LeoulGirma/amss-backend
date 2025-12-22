package domain

import (
	"time"

	"github.com/google/uuid"
)

type ImportType string

const (
	ImportTypeAircraft ImportType = "aircraft"
	ImportTypeParts    ImportType = "parts"
	ImportTypePrograms ImportType = "programs"
)

type ImportStatus string

const (
	ImportStatusPending    ImportStatus = "pending"
	ImportStatusValidating ImportStatus = "validating"
	ImportStatusApplying   ImportStatus = "applying"
	ImportStatusCompleted  ImportStatus = "completed"
	ImportStatusFailed     ImportStatus = "failed"
)

type ImportRowStatus string

const (
	ImportRowPending ImportRowStatus = "pending"
	ImportRowValid   ImportRowStatus = "valid"
	ImportRowInvalid ImportRowStatus = "invalid"
	ImportRowApplied ImportRowStatus = "applied"
)

type Import struct {
	ID        uuid.UUID
	OrgID     uuid.UUID
	Type      ImportType
	Status    ImportStatus
	FileName  string
	FilePath  string
	CreatedBy uuid.UUID
	Summary   map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ImportRow struct {
	ID        uuid.UUID
	OrgID     uuid.UUID
	ImportID  uuid.UUID
	RowNumber int
	Raw       map[string]any
	Status    ImportRowStatus
	Errors    []string
	CreatedAt time.Time
	UpdatedAt time.Time
}
