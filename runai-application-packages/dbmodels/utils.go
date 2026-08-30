package dbmodels

import (
	"time"

	"gorm.io/gorm"
)

func GormDeletedAtToNullableTime(time gorm.DeletedAt) *time.Time {
	if !time.Valid {
		return nil
	}

	return &time.Time
}
