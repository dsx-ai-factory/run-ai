package pgerrors

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	ErrCodeInsufficientPrivilege = "42501"
	ErrCodeDuplicateObject       = "42710"
	ErrCodeReadOnlyTransaction   = "25006"
	ErrCodeUndefinedTable        = "42P01"
)

var (
	ErrReadOnlyConnection = errors.New("write operation not allowed on read-only database connection")
	ErrUndefinedTable     = errors.New("table does not exist")
)

var codeToErr = map[string]error{
	ErrCodeReadOnlyTransaction: ErrReadOnlyConnection,
	ErrCodeUndefinedTable:      ErrUndefinedTable,
}

func Translate(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	sentinel, ok := codeToErr[pgErr.Code]
	if !ok || errors.Is(err, sentinel) {
		return err
	}

	return fmt.Errorf("%w: %w", sentinel, err)
}

func HasCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}
