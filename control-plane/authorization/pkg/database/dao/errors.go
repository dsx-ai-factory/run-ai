package dao

import (
	"errors"

	"gorm.io/gorm"
)

func isDuplicateKeyError(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
