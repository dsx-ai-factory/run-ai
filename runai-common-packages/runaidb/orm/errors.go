package orm

import (
	pgerrors "github.com/run-ai/runai/runai-common-packages/runaidb/internal/postgres/errors"
	"gorm.io/gorm"
)

var (
	ErrRecordNotFound                = gorm.ErrRecordNotFound
	ErrInvalidTransaction            = gorm.ErrInvalidTransaction
	ErrMissingWhereClause            = gorm.ErrMissingWhereClause
	ErrUnsupportedRelation           = gorm.ErrUnsupportedRelation
	ErrPrimaryKeyRequired            = gorm.ErrPrimaryKeyRequired
	ErrModelValueRequired            = gorm.ErrModelValueRequired
	ErrModelAccessibleFieldsRequired = gorm.ErrModelAccessibleFieldsRequired
	ErrDuplicatedKey                 = gorm.ErrDuplicatedKey
	ErrForeignKeyViolated            = gorm.ErrForeignKeyViolated
	ErrCheckConstraintViolated       = gorm.ErrCheckConstraintViolated

	ErrReadOnlyConnection = pgerrors.ErrReadOnlyConnection
	ErrUndefinedTable     = pgerrors.ErrUndefinedTable
)
