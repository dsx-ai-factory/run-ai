package models

import (
	"fmt"

	"gorm.io/gorm"
)

const TenantsSchema = "tenants_manager"

type Tenant struct {
	Id        int32 `gorm:"primaryKey"`
	DeletedAt gorm.DeletedAt
	Name      string
}

func (Tenant) TableName() string {
	return fmt.Sprintf("%s.tenants", TenantsSchema)
}
