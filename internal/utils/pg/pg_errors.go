package pg

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

func HandlePgError(err error) error {

	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("unique constraint violation")
		case "23503":
			return fmt.Errorf("foreign key violation")
		case "23502":
			return fmt.Errorf("not null violation")
		default:
			return fmt.Errorf("database error: %v", pgErr.Message)
		}
	}
	return err
}
