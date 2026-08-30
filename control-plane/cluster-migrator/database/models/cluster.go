package models

import (
	"gorm.io/gorm"
)

type Cluster struct {
	Id        string `gorm:"primaryKey"`
	TenantId  int32
	DeletedAt gorm.DeletedAt
	Name      string
}

func (Cluster) TableName() string {
	return "clusters"
}
