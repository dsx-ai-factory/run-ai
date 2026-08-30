package configprovider

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type ErrCode string

const (
	ErrCodeUnknown   ErrCode = "UNKNOWN"
	ErrCodeNotFound  ErrCode = "NOT_FOUND"
	ErrCodeDuplicate ErrCode = "DUPLICATE"
)

type Error struct {
	Code    ErrCode
	Message string
}

func NewError(message string, code ErrCode) *Error {
	return &Error{
		Message: message,
		Code:    code,
	}
}

func (e Error) Error() string {
	return e.Message
}

const pgErrCodeUniqueViolation = "23505"

func isDuplicateKeyError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == pgErrCodeUniqueViolation {
		return true
	}

	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) && pgxErr.Code == pgErrCodeUniqueViolation {
		return true
	}

	return false
}
