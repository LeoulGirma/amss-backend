package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	OrgID        uuid.UUID
	Email        string
	Role         Role
	PasswordHash string
	LastLogin    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}
