package models

import "gorm.io/gorm"

type Logo struct {
	gorm.Model
	TenantId uint    `gorm:"index;unique"`
	LogoPng  *[]byte `gorm:"not null"`
}
