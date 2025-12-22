package domain

import (
	"errors"
	"fmt"
)

var (
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrValidation   = errors.New("validation")
	ErrUnauthorized = errors.New("unauthorized")
)

func NewValidationError(message string) error {
	return fmt.Errorf("%w: %s", ErrValidation, message)
}

func NewConflictError(message string) error {
	return fmt.Errorf("%w: %s", ErrConflict, message)
}
