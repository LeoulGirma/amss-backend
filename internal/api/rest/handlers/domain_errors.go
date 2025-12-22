package handlers

import (
	"errors"
	"net/http"

	"github.com/aeromaintain/amss/internal/domain"
)

func writeDomainError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
	case errors.Is(err, domain.ErrForbidden):
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "forbidden")
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "not found")
	case errors.Is(err, domain.ErrConflict):
		writeError(w, r, http.StatusConflict, "CONFLICT", "conflict")
	case errors.Is(err, domain.ErrValidation):
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "validation error")
	default:
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "internal error")
	}
}
