package postgres

import (
	"errors"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

func TranslateError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505", "23P01":
			return domain.ErrConflict
		case "23503":
			return domain.NewValidationError("invalid reference")
		}
	}
	return err
}
